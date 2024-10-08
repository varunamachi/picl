package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/httpx"
	"github.com/varunamachi/libx/iox"
	"github.com/varunamachi/libx/str"
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
	Password  string              `json:"password"`
	AuthMehod xcutr.SshAuthMethod `json:"authMethod"`
	// AuthData  map[string]string   `json:"authData"`
	Color string `json:"color"`
}

type agent struct {
	Port     int             `json:"port"`
	Protocol string          `json:"protocol"`
	AuthData *httpx.AuthData `json:"authData"`
}

type host struct {
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Executer executer `json:"executer"`
	Agent    agent    `json:"agent"`
}

type PiclConfig struct {
	Name    string  `json:"name"`
	Monitor monitor `json:"monitor"`
	Hosts   []*host `json:"hosts"`
}

type configProvider struct {
	// path string
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
		iox.MustGetUserHome(), ".picl", cfg+".config.json")
	if iox.ExistsAsFile(cfgPath) {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			err := fmt.Errorf("failed to read configuration %s: %w", cfg, err)
			log.Error().Err(err).Msg("")
			return nil, err
		}
		return New(data)
	}

	cfgPath = filepath.Join(
		iox.MustGetUserHome(), ".picl", cfg+".config.json.enc")
	if iox.ExistsAsFile(cfgPath) {
		pw := iox.AskPassword("Please Enter Config Password")

		data, err := os.ReadFile(cfgPath)
		if err != nil {
			e := fmt.Errorf("failed to read configuration %s", cfg)
			log.Error().Err(err).Msg(e.Error())
			return nil, e
		}

		data, err = iox.NewCryptor(pw).Decrypt(data)
		if err != nil {
			e := fmt.Errorf(
				"failed to decrypt configuration %s", cfg)
			log.Error().Err(err).Msg(e.Error())
			return nil, e
		}

		return New(data)
	}

	err := fmt.Errorf("could not find configuration %s", cfg)
	log.Error().Err(err).Msg("")
	return nil, err
}

func New(data []byte) (Provider, error) {
	cfg := PiclConfig{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return new(&cfg)
}

func new(cfg *PiclConfig) (Provider, error) {
	cp := configProvider{}
	cp.eCfg = &xcutr.Config{
		Name: cfg.Name,
		Opts: make([]*xcutr.SshConnOpts, len(cfg.Hosts)),
	}
	cp.mCfg = &mon.Config{
		Name:        cfg.Name,
		Height:      cfg.Monitor.Height,
		Width:       cfg.Monitor.Width,
		GoArch:      cfg.Monitor.GoArch,
		AgentConfig: make([]*mon.AgentConfig, len(cfg.Hosts)),
	}

	for i, h := range cfg.Hosts {
		cp.eCfg.Opts[i] = &xcutr.SshConnOpts{
			Name:      h.Name,
			Host:      h.Host,
			Port:      h.Executer.SshPort,
			UserName:  h.Executer.UserName,
			AuthMehod: h.Executer.AuthMehod,
			Password:  h.Executer.Password,
			Color:     h.Executer.Color,
		}

		protocol := h.Agent.Protocol
		if protocol == "" {
			protocol = "http"
		}
		port := h.Agent.Port
		if port == 0 {
			port = 20202
		}
		address := fmt.Sprintf("%s://%s:%d", protocol, h.Host, port)
		cp.mCfg.AgentConfig[i] = &mon.AgentConfig{
			Name:     h.Name,
			Address:  address,
			AuthData: h.Agent.AuthData,
		}
	}

	return &cp, nil
}

func CreateConfigTemplate(configName string, numHosts int) error {

	config := PiclConfig{
		Name: "",
		Monitor: monitor{
			Height: 20,
			Width:  60,
			GoArch: "AARCH64",
		},
		Hosts: make([]*host, numHosts),
	}

	if numHosts == 0 {
		numHosts = 1
	}

	for i := 0; i < numHosts; i++ {
		config.Hosts = append(config.Hosts, &host{
			Name: fmt.Sprintf("host_%d", i),
			Host: fmt.Sprintf("host%d", i),
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
	}
	return generateConfig(&config, configName, false, "")
}

func CreateConfig(name string) error {

	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current user")
	}

	gtr := iox.StdUserInputReader()

	numHosts := gtr.Int("Number of Hosts")

	conf := PiclConfig{
		Name:  name,
		Hosts: make([]*host, numHosts),
	}

	colors := []string{
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"white",
	}

	useCmnUser := gtr.BoolOr("Use Common User Name (SSH)?", true)
	var cmnUser, cmnPwd string
	if useCmnUser {
		cmnUser = gtr.StringOr("SSH User Name", user.Username)
		msg := fmt.Sprintf("Common SSH Password for '%s'", cmnUser)
		cmnPwd = gtr.Secret(msg)
	}

	conf.Monitor.Height = gtr.IntOr("Monitor Height", 20)
	conf.Monitor.Width = gtr.IntOr("Monitor Width", 60)
	conf.Monitor.GoArch = gtr.Select("Architecture", []string{
		"386",
		"amd64",
		"arm",
		"arm64",
	}, "arm64")
	conf.Hosts = make([]*host, numHosts)

	agentPort := gtr.IntOr("Agent Port", 20202)
	agentProto := gtr.Select(
		"Agent Protocol", []string{"http", "https"}, "http")

	fmt.Println()
	for i := 0; i < numHosts; i++ {

		for {
			msg := fmt.Sprintf(
				"Host-%d Name & Address (space separated) (q to quit)", i+1)
			hostStr := gtr.String(msg)
			parts := strings.Fields(hostStr)
			if len(parts) == 2 {
				host := &host{
					Name: strings.TrimSpace(parts[0]),
					Host: strings.TrimSpace(parts[1]),
					Executer: executer{
						SshPort:  22,
						Color:    colors[i%(len(colors)-1)],
						UserName: cmnUser,
						Password: cmnPwd,
					},
					Agent: agent{
						Port:     agentPort,
						Protocol: agentProto,
					},
				}
				if !useCmnUser {
					msg := fmt.Sprintf("SSH Username for %s", host.Host)
					host.Executer.UserName = gtr.String(msg)
					msg = fmt.Sprintf("SSH Password for %s@%s",
						host.Executer.UserName, host.Host)
					host.Executer.Password = gtr.Secret(msg)
				}
				conf.Hosts[i] = host
				break
			} else if str.EqFold(hostStr, "q") {
				os.Exit(0)
			}
		}

	}

	var pw string
	encrypt := gtr.BoolOr("Do you want to encrypt the config?", false)
	if encrypt {
		pw = gtr.Secret("Password for encryption")
	}

	if err := generateConfig(&conf, name, encrypt, pw); err != nil {
		return errx.Wrap(err)
	}

	provider, err := new(&conf)
	if err != nil {
		return errx.Wrap(err)
	}

	copyId := gtr.BoolOr("Copy SSH Public Key to Nodes (ssh-copy-id)? ", true)
	if copyId {

		opts := provider.ExecuterConfig().Opts
		for _, opt := range opts {
			opt.AuthMehod = xcutr.SshAuthPassword
		}

		if err := xcutr.CopyId(opts); err != nil {
			return errx.Wrap(err)
		}
	}

	return nil
}

func CreateConfigWithDefaults(name string) error {

	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current user")
	}

	gtr := iox.StdUserInputReader()

	numHosts := gtr.Int("Number of Hosts")

	conf := PiclConfig{
		Name:  name,
		Hosts: make([]*host, numHosts),
	}

	colors := []string{
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
		"white",
	}

	cmnUser := user.Username
	msg := fmt.Sprintf("Common SSH Password for '%s'", cmnUser)
	cmnPwd := gtr.Secret(msg)

	conf.Monitor = monitor{
		Height: 20,
		Width:  60,
		GoArch: "arm64",
	}
	conf.Hosts = make([]*host, numHosts)

	fmt.Println()
	for i := 0; i < numHosts; i++ {

		for {
			msg := fmt.Sprintf(
				"Host-%d Name & Address (space separated) (q to quit)", i+1)
			hostStr := gtr.String(msg)
			parts := strings.Fields(hostStr)
			if len(parts) == 2 {
				host := &host{
					Name: strings.TrimSpace(parts[0]),
					Host: strings.TrimSpace(parts[1]),
					Executer: executer{
						SshPort:   22,
						Color:     colors[i%(len(colors)-1)],
						UserName:  cmnUser,
						Password:  cmnPwd,
						AuthMehod: xcutr.SshAuthPublicKey,
					},
					Agent: agent{
						Port:     20202,
						Protocol: "http",
					},
				}
				conf.Hosts[i] = host
				break
			} else if str.EqFold(hostStr, "q") {
				os.Exit(0)
			}
		}

	}

	var pw string
	encrypt := gtr.BoolOr("Do you want to encrypt the config?", false)
	if encrypt {
		pw = gtr.Secret("Password for encryption")
	}
	if err := generateConfig(&conf, name, encrypt, pw); err != nil {
		return errx.Wrap(err)
	}

	provider, err := new(&conf)
	if err != nil {
		return errx.Wrap(err)
	}
	copyId := gtr.BoolOr("Copy SSH Public Key to Nodes (ssh-copy-id)? ", true)
	if copyId {
		if err = CopySshId(provider); err != nil {
			return errx.Wrap(err)
		}
	}

	return nil
}

func CopySshId(provider Provider) error {
	opts := provider.ExecuterConfig().Opts
	for _, opt := range opts {
		opt.AuthMehod = xcutr.SshAuthPassword
	}

	if err := xcutr.CopyId(opts); err != nil {
		return errx.Wrap(err)
	}
	return nil
}

func generateConfig(
	config *PiclConfig, configName string, encrypt bool, pw string) error {
	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return errx.Wrap(err)
	}

	ext := ".config.json"
	if encrypt {
		ext = ".config.json.enc"
	}
	dir := filepath.Join(iox.MustGetUserHome(), ".picl")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errx.Errf(err, "failed to create config dir at '%s'", dir)
	}

	path := filepath.Join(dir, configName+ext)
	configFile, err := os.Create(path)
	if err != nil {
		return errx.Errf(err, "failed to create config file at '%s'", path)
	}
	defer configFile.Close()

	if encrypt {
		jsonData, err = iox.NewCryptor(pw).Encrypt(jsonData)
		if err != nil {
			return errx.Wrap(err)
		}
		_, err = configFile.Write(jsonData)
		return errx.Wrap(err)
	}

	_, err = configFile.WriteString(string(jsonData))
	return errx.Wrap(err)
}
