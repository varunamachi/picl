package mon

import (
	"context"
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/varunamachi/clusterfox/cfx"
)

type TuiHandler struct {
	cfg    *MonitorConfig
	table  *widgets.Table
	values []*SysInfo
}

func NewTuiHandler(cfg *MonitorConfig) (Handler, context.Context, error) {
	if err := ui.Init(); err != nil {
		return nil, nil, cfx.Errf(err, "failed to initialize termui")

	}

	table := widgets.NewTable()
	table.RowStyles[0] = ui.NewStyle(
		ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, cfg.Width, cfg.Height)
	table.TextAlignment = ui.AlignCenter
	table.Rows = make([][]string, len(cfg.AgentConfig)+1)
	uiEvents := ui.PollEvents()

	gtx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			e := <-uiEvents
			switch e.ID {
			case "q", "<C-c>":
				cancel()
				return
			}

		}
	}()

	hdl := &TuiHandler{
		cfg:    cfg,
		table:  table,
		values: make([]*SysInfo, len(cfg.AgentConfig)),
	}
	return hdl, gtx, nil
}

func (t *TuiHandler) Close() error {
	ui.Close()
	return nil
}

func (t *TuiHandler) Handle(gtx context.Context, resp *AgentResponse) error {

	t.values[resp.Index] = resp.Data
	t.table.Rows[0] = []string{"Name", "Temp", "CPU Usage", "RAM Usage"}

	for index, ag := range t.cfg.AgentConfig {

		select {
		case <-gtx.Done():
			return gtx.Err()
		default:
		}

		val := t.values[index]
		if val == nil {
			t.table.Rows[index+1] = []string{
				ag.Name,
				fmt.Sprintf("N/A"),
				fmt.Sprintf("N/A"),
				fmt.Sprintf("N/A"),
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		t.table.Rows[index+1] = []string{
			ag.Name,
			fmt.Sprintf("%.2f", val.CPUTemp/1000),
			fmt.Sprintf("%.2f%%", val.CPUUsagePct),
			fmt.Sprintf("%.2f%%", val.MemUsagePct),
		}
	}

	ui.Render(t.table)
	return nil
}
