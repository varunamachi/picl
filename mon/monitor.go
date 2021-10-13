package mon

import (
	"fmt"
	"net/http"
)

type AgentConfig struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type MonitorConfig struct {
	ScreenWidth  int
	ScreenHeight int
	AgentConfig  []*AgentConfig
}

type Monitor struct {
	config  *MonitorConfig
	clients map[string]*http.Client
}

func NewMonitor(config *MonitorConfig) (*Monitor, error) {
	mon := &Monitor{
		config:  config,
		clients: make(map[string]*http.Client),
	}

	for _, conf := range config.AgentConfig {
		// TODO - creat clients
		fmt.Println(conf.Name)
	}

	return mon, nil

}

func (mon *Monitor) Run() error {
	return nil
}
