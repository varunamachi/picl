package mon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/httpx"
	"golang.org/x/sync/errgroup"
)

type AgentConfig struct {
	Name     string          `json:"name"`
	Address  string          `json:"address"`
	AuthData *httpx.AuthData `json:"authData"`
}

type Config struct {
	Name        string         `json:"name"`
	Height      int            `json:"height"`
	Width       int            `json:"width"`
	GoArch      string         `json:"goArch"`
	AgentConfig []*AgentConfig `json:"agentConfig"`
}

func (cfg *Config) PrintSampleJSON() {
	cfg.AgentConfig = []*AgentConfig{
		{},
	}

	j, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal MonitorConfig to JSON")
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
	config   *Config
	clients  []*httpx.Client
	handler  Handler
	relayCtl *RelayController
	server   *httpx.Server
}

func NewMonitor(
	gtx context.Context,
	config *Config,
	realyConfig *RelayConfig,
	hdl Handler,
	server *httpx.Server) (*Monitor, error) {
	mon := &Monitor{
		config:  config,
		clients: make([]*httpx.Client, 0, len(config.AgentConfig)),
		handler: hdl,
		server:  server,
	}

	for _, conf := range config.AgentConfig {
		client := httpx.NewCustomClient(
			conf.Address, "/api/v0", httpx.DefaultTransport(),
			100*time.Millisecond)
		if conf.AuthData != nil {

			if err := Login(gtx, client, *conf.AuthData); err != nil {
				msg := "failed to login to agent"
				log.Error().Err(err).Str("conf", conf.Name).Msg(msg)
				return nil, errx.Errf(err, msg, conf.Name)
			}
		}
		mon.clients = append(mon.clients, client)
	}
	var err error
	mon.relayCtl, err = NewRelayController(realyConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize GPIO, " +
			"disabling related features...")
		// return nil, err
	}
	mon.server.WithAPIs(getRelayEndpoints(mon.relayCtl)...)
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
					return errx.Wrap(err)
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
						return errx.Wrap(err)
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
	monConfig *Config
}

func NewSimpleHandler(cfg *Config) (Handler, context.Context, error) {
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

func NewNoOpHandler(cfg *Config) (Handler, context.Context, error) {
	return &noOpHandler{}, context.Background(), nil
}

func (sh *noOpHandler) Handle(
	gtx context.Context, resp *AgentResponse) error {
	return nil
}

func (sh *noOpHandler) Close() error {
	return nil
}
