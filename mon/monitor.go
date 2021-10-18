package mon

import (
	"context"
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/varunamachi/clusterfox/agent"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/cfx/client"
	"golang.org/x/sync/errgroup"
)

type AgentConfig struct {
	Name     string          `json:"name"`
	Address  string          `json:"address"`
	AuthData client.AuthData `json:"authData"`
}

type MonitorConfig struct {
	Name        string         `json:"name"`
	Height      int            `json:"height"`
	Width       int            `json:"width"`
	GoArch      string         `json:"goArch"`
	AgentConfig []*AgentConfig `json:"agentConfig"`
}

type Handler interface {
	Handle(resp *AgentResponse) error
}

type AgentResponse struct {
	Index int
	Data  *agent.SysInfo
	Err   error
}

type Monitor struct {
	config  *MonitorConfig
	clients []*client.Client
	handler Handler
}

func NewMonitor(
	gtx context.Context,
	config *MonitorConfig,
	hdl Handler) (*Monitor, error) {
	mon := &Monitor{
		config:  config,
		clients: make([]*client.Client, 0, len(config.AgentConfig)),
		handler: hdl,
	}

	for _, conf := range config.AgentConfig {
		client := client.New(conf.Address, "/api/v0")
		if err := client.Login(gtx, &conf.AuthData); err != nil {
			return nil, err
		}
		mon.clients = append(mon.clients, client)
	}
	return mon, nil
}

func (mon *Monitor) Run(
	gtx context.Context) error {

	out := make(chan *AgentResponse)
	defer close(out)
	eg := errgroup.Group{}

	eg.Go(func() error {
		return mon.poll(gtx, &eg, out)
	})

	eg.Go(func() error {
		for {
			select {
			case <-gtx.Done():
				return gtx.Err()
			case resp := <-out:
				if err := mon.handler.Handle(resp); err != nil {
					return err
				}
			default:
			}
		}
	})

	return eg.Wait()

}

func (mon *Monitor) poll(
	gtx context.Context,
	eg *errgroup.Group,
	dataOut chan<- *AgentResponse) error {

	for {
		select {
		case <-gtx.Done():
			return gtx.Err()
		default:
			// No-op
		}

		for index, client := range mon.clients {
			index := index
			client := client
			eg.Go(func() error {
				info := &agent.SysInfo{}
				res := client.Get(gtx, "/sysinfo/cur")
				if err := res.LoadClose(&info); err != nil {
					dataOut <- &AgentResponse{Index: index, Err: err}
					return err
				}
				dataOut <- &AgentResponse{Index: index, Data: info}
				return nil
			})
		}

		time.Sleep(1 * time.Second)
	}
}

type TuiHandler struct {
	cfg    *MonitorConfig
	table  *widgets.Table
	values []*agent.SysInfo
}

func NewTuiHandler(cfg *MonitorConfig) (Handler, error) {
	if err := ui.Init(); err != nil {
		return nil, cfx.Errf(err, "failed to initialize termui")

	}
	defer ui.Close()

	table := widgets.NewTable()
	// table.Rows = [][]string{
	// 	[]string{"Name", "Temp", "CPU Usage", "RAM Usage"},
	// }
	table.RowStyles[0] = ui.NewStyle(
		ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)

	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, cfg.Width, cfg.Height)
	table.TextAlignment = ui.AlignRight
	table.Rows = make([][]string, len(cfg.AgentConfig)+1)

	hdl := &TuiHandler{
		cfg:   cfg,
		table: table,
	}

	return hdl, nil
}

func (t TuiHandler) Handle(resp *AgentResponse) error {

	t.values[resp.Index] = resp.Data

	t.table.Rows[0] = []string{"Name", "Temp", "CPU Usage", "RAM Usage"}

	for index, ag := range t.cfg.AgentConfig {
		val := t.values[index]
		t.table.Rows[index+1] = []string{
			ag.Name,
			fmt.Sprintf("%.2f", val.CPUTemp),
			fmt.Sprintf("%.2f", val.CPUUsagePct),
			fmt.Sprintf("%.2f", val.MemUsagePct),
		}
	}

	ui.Render(t.table)
	// uiEvents := ui.PollEvents()
	// for {
	// 	e := <-uiEvents
	// 	switch e.ID {
	// 	case "q", "<C-c>":
	// 		return nil
	// 	}
	// }
	return nil
}
