package config

import (
	"fmt"

	"github.com/varunamachi/picl/cmn"
	"github.com/varunamachi/picl/cmn/client"
	"github.com/varunamachi/picl/mon"
	"github.com/varunamachi/picl/xcutr"
)

type Provider interface {
	ExecuterConfig() *xcutr.Config
	MonitorConfig() *mon.Config
}

type piclConfig struct {
	Name     string `json:"name"`
	SudoPass string `json:"sudoPass"`
	Monitor  struct {
		Height int    `json:"height"`
		Width  int    `json:"width"`
		GoArch string `json:"goArch"`
	} `json:"monitor"`
	Hosts []struct {
		Name     string `json:"name"`
		Host     string `json:"host"`
		Executer struct {
			SshPort   int                 `json:"sshPort"`
			UserName  string              `json:"userName"`
			AuthMehod xcutr.SshAuthMethod `json:"authMethod"`
			AuthData  map[string]string   `json:"authData"`
			Color     string              `json:"color"`
		} `json:"executer"`
		Agent struct {
			Port     int             `json:"port"`
			Protocol string          `json:"protocol"`
			AuthData client.AuthData `json:"authData"`
		} `json:"agent"`
	} `json:"hosts"`
}

type configProvider struct {
	path string
	eCfg *xcutr.Config
	mCfg *mon.Config
}

func (cp *configProvider) ExecuterConfig() *xcutr.Config {
	return cp.eCfg
}

func (cp *configProvider) MonitorConfig() *mon.Config {
	return cp.mCfg
}

func New() (Provider, error) {
	cp := configProvider{}

	cfg := piclConfig{}
	if err := cmn.LoadJsonFile(cp.path, &cfg); err != nil {
		// Log and return appropriate error
		return &cp, err
	}

	cp.eCfg = &xcutr.Config{
		Name:     cfg.Name,
		SudoPass: cfg.SudoPass,
		Opts:     make([]*xcutr.SshConnOpts, 0, len(cfg.Hosts)),
	}
	cp.mCfg = &mon.Config{
		Name:        cfg.Name,
		Height:      cfg.Monitor.Height,
		Width:       cfg.Monitor.Width,
		GoArch:      cfg.Monitor.GoArch,
		AgentConfig: make([]*mon.AgentConfig, len(cfg.Hosts)),
	}

	for _, h := range cfg.Hosts {
		cp.eCfg.Opts = append(cp.eCfg.Opts, &xcutr.SshConnOpts{
			Name:      h.Name,
			Host:      h.Host,
			Port:      h.Executer.SshPort,
			UserName:  h.Executer.UserName,
			AuthMehod: h.Executer.AuthMehod,
			AuthData:  h.Executer.AuthData,
			Color:     h.Executer.Color,
		})

		protocol := h.Agent.Protocol
		if protocol == "" {
			protocol = "http"
		}
		port := h.Agent.Port
		if port == 0 {
			port = 8000
		}
		address := fmt.Sprintf("%s://%s:%d", protocol, h.Host, port)
		cp.mCfg.AgentConfig = append(cp.mCfg.AgentConfig, &mon.AgentConfig{
			Name:     h.Name,
			Address:  address,
			AuthData: h.Agent.AuthData,
		})
	}

	return &cp, nil
}
