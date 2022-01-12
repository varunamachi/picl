package config

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/varunamachi/picl/cmn"
	"github.com/varunamachi/picl/cmn/client"
	"github.com/varunamachi/picl/mon"
	"github.com/varunamachi/picl/xcutr"
)

type Provider interface {
	ExecuterConfig() *xcutr.Config
	MonitorConfig() *mon.Config
}

type monitor struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	GoArch string `json:"goArch"`
}

type executer struct {
	SshPort   int                 `json:"sshPort"`
	UserName  string              `json:"userName"`
	AuthMehod xcutr.SshAuthMethod `json:"authMethod"`
	AuthData  map[string]string   `json:"authData"`
	Color     string              `json:"color"`
}

type agent struct {
	Port     int              `json:"port"`
	Protocol string           `json:"protocol"`
	AuthData *client.AuthData `json:"authData"`
}

type host struct {
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Executer executer `json:"executer"`
	Agent    agent    `json:"agent"`
}

type PiclConfig struct {
	Name     string  `json:"name"`
	SudoPass string  `json:"sudoPass"`
	Monitor  monitor `json:"monitor"`
	Hosts    []*host `json:"hosts"`
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

func NewFromCli(ctx *cli.Context) (Provider, error) {
	cfg := ctx.String("config")
	if cfg == "" {
		cfg = "default"
	}
	cfgPath := filepath.Join(
		cmn.MustGetUserHome(), ".picl", cfg+".cluster.json")

	return New(cfgPath)
}

func New(path string) (Provider, error) {
	cp := configProvider{path: path}

	cfg := PiclConfig{}
	if err := cmn.LoadJsonFile(path, &cfg); err != nil {
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

func CreateConfigTemplate(w io.Writer, numHosts int) error {

	config := PiclConfig{
		Name:     "",
		SudoPass: "",
		Monitor: monitor{
			Height: 0,
			Width:  0,
			GoArch: "AARCH64",
		},
		Hosts: make([]*host, numHosts),
	}
	config.Hosts = append(config.Hosts, &host{
		Executer: executer{
			SshPort:   20,
			UserName:  "",
			AuthMehod: "PublicKey",
			// AuthData:  ,
			Color: "",
		},
		Agent: agent{
			Port:     8000,
			Protocol: "http",
			AuthData: nil,
		},
	})
	return nil

	// 		{
	// 			Name     string `json:"name"`
	// 			Host     string `json:"host"`

	// 		}

	// 		Executer struct {
	// 			SshPort   int                 `json:"sshPort"`
	// 			UserName  string              `json:"userName"`
	// 			AuthMehod xcutr.SshAuthMethod `json:"authMethod"`
	// 			AuthData  map[string]string   `json:"authData"`
	// 			Color     string              `json:"color"`
	// 		} `json:"executer"`
	// 		Agent struct {
	// 			Port     int             `json:"port"`
	// 			Protocol string          `json:"protocol"`
	// 			AuthData client.AuthData `json:"authData"`
	// 		} `json:"agent"`
	// 	} `json:"hosts"`
	// }
}
