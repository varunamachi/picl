package mon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/varunamachi/picl/cmn"
	"github.com/varunamachi/picl/cmn/client"
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

func (cfg *MonitorConfig) PrintSampleJSON() {
	cfg.AgentConfig = []*AgentConfig{
		{},
	}

	j, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal MonitorConfig to JSON")
		return
	}
	fmt.Println(string(j))
}

type Handler interface {
	Handle(gtx context.Context, resp *AgentResponse) error
	Close() error
}

type AgentResponse struct {
	Index int
	Data  *SysInfo
	Err   error
}

type Monitor struct {
	config   *MonitorConfig
	clients  []*client.Client
	handler  Handler
	relayCtl *RelayController
	server   *cmn.Server
}

func NewMonitor(
	gtx context.Context,
	config *MonitorConfig,
	realyConfig *RelayConfig,
	hdl Handler,
	server *cmn.Server) (*Monitor, error) {
	mon := &Monitor{
		config:  config,
		clients: make([]*client.Client, 0, len(config.AgentConfig)),
		handler: hdl,
		server:  server,
	}

	for _, conf := range config.AgentConfig {
		client := client.NewCustom(
			conf.Address, "/api/v0", client.DefaultTransport(),
			100*time.Millisecond)
		if conf.AuthData.Data != nil {
			if err := client.Login(gtx, &conf.AuthData); err != nil {
				msg := "failed to login to agent"
				logrus.WithError(err).Error(msg, conf.Name)
				return nil, cmn.Errf(err, msg, conf.Name)
			}
		}
		mon.clients = append(mon.clients, client)
	}
	var err error
	mon.relayCtl, err = NewRelayController(realyConfig)
	if err != nil {
		logrus.WithError(err).Warn("failed to initialize GPIO, " +
			"disabling related features...")
		// return nil, err
	}
	mon.server.AddEndpoints(getRelayEndpoints(mon.relayCtl)...)
	return mon, nil
}

func (mon *Monitor) Run(
	gtx context.Context, port uint32) error {

	out := make(chan *AgentResponse)
	defer func() {
		close(out)
		if mon.relayCtl != nil {
			mon.relayCtl.Close()
		}
	}()
	eg := errgroup.Group{}

	eg.Go(func() error {
		return mon.poll(gtx, &eg, out)
	})

	eg.Go(func() error {
		for {
			select {
			case <-gtx.Done():
				mon.server.Close()
				return gtx.Err()
			case resp := <-out:
				if err := mon.handler.Handle(gtx, resp); err != nil {
					return err
				}
			}
		}
	})
	eg.Go(func() error {
		return mon.server.Start(port)
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

			for index, client := range mon.clients {
				index := index
				client := client
				eg.Go(func() error {
					info := &SysInfo{}
					res := client.Get(gtx, "/cur")
					if err := res.LoadClose(&info); err != nil {
						dataOut <- &AgentResponse{Index: index, Err: err}
						return err
					}
					dataOut <- &AgentResponse{Index: index, Data: info}
					return nil
				})
			}
		}
		time.Sleep(1 * time.Second)
	}
}

type simpleHandler struct {
	monConfig *MonitorConfig
}

func NewSimpleHandler(cfg *MonitorConfig) (Handler, context.Context, error) {
	return &simpleHandler{
		monConfig: cfg,
	}, context.Background(), nil
}

func (sh *simpleHandler) Handle(
	gtx context.Context, resp *AgentResponse) error {

	node := sh.monConfig.AgentConfig[resp.Index]
	if resp.Err != nil {
		fmt.Println(resp.Err)
		return nil
	}

	fmt.Printf("%2d.  %10s   Tmp: %4.2f   CPU: %4.2f%%   Mem: %4.2f%%\n",
		resp.Index,
		node.Name,
		resp.Data.CPUTemp/1000,
		resp.Data.CPUUsagePct,
		resp.Data.MemUsagePct)
	return nil
}

func (sh *simpleHandler) Close() error {
	return nil
}

type noOpHandler struct {
}

func NewNoOpHandler(cfg *MonitorConfig) (Handler, context.Context, error) {
	return &noOpHandler{}, context.Background(), nil
}

func (sh *noOpHandler) Handle(
	gtx context.Context, resp *AgentResponse) error {
	return nil
}

func (sh *noOpHandler) Close() error {
	return nil
}
