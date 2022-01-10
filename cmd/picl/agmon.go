package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/picl/cmn"
	"github.com/varunamachi/picl/mon"
)

func getAgentCmd() *cli.Command {
	return &cli.Command{
		Name:        "agent",
		Description: "Run as an agent service with REST APIs exposed",
		Usage:       "Run as an agent",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "port",
				Usage: "Port on which the service runs",
				Value: 20202,
			},
		},
		Action: func(etx *cli.Context) error {
			port := etx.Int("port")
			return mon.RunAgent(fmt.Sprintf(":%d", port))
		},
	}
}

func getMonitorCmd() *cli.Command {
	return &cli.Command{
		Name:        "monitor",
		Description: "Start the monitor",
		Usage:       "Start the monitor",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Name of the config, as: ~/.picl/'config'.monitor.json",
				Value: "default",
			},
			&cli.UintFlag{
				Name:  "port",
				Usage: "Port for exposing monitor related REST endpoints",
				Value: 8000,
			},
			&cli.StringFlag{
				Name:  "handler",
				Usage: "Handler type, one of: tui | simple | noop",
				Value: "tui",
			},
		},
		Action: func(etx *cli.Context) error {
			cfg := etx.String("config")
			port := etx.Uint("port")
			handler := etx.String("handler")

			cfgPath := filepath.Join(
				cmn.MustGetUserHome(), ".picl", cfg+".monitor.json")
			var config mon.Config
			if err := cmn.LoadJsonFile(cfgPath, &config); err != nil {
				logrus.
					WithError(err).
					WithField("config", cfg).
					Error("Failed to load config")

				config.PrintSampleJSON()
				return err
			}

			hdl, gtx, err := newHandler(handler, &config)
			if err != nil {
				return err
			}
			var printer io.Writer
			if handler != "tui" {
				printer = os.Stdout
			}
			defer hdl.Close()

			rcfg := &mon.RelayConfig{
				GpioPins:       []uint8{22, 23, 24, 25},
				IsNormallyOpen: true,
			}
			monitor, err := mon.NewMonitor(
				gtx, &config, rcfg, hdl, cmn.NewServer(printer))

			if err != nil {
				logrus.WithError(err).Error("failed to initialize monitor")
				return err
			}
			if err = monitor.Run(gtx, uint32(port)); err != nil {
				logrus.WithError(err).Error("failed to run monitor")
				return err
			}

			return nil
		},
	}
}

func newHandler(hdl string, cfg *mon.Config) (
	mon.Handler, context.Context, error) {
	switch hdl {
	case "simple":
		return mon.NewSimpleHandler(cfg)
	case "noop":
		return mon.NewNoOpHandler(cfg)
	case "tui":
		return mon.NewTuiHandler(cfg)
	}
	return nil, nil, fmt.Errorf("invalid handler '%s' selected", hdl)
}

func getBuildInstallCmd() *cli.Command {

	//<root>/cmd/picl/agmon.go
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		const msg = "Couldnt get main file path"
		logrus.Fatal(msg)
	}

	fxRootPath, err := filepath.Abs(filename + "/../../..")
	if err != nil {
		const msg = "couldnt get main file path"
		logrus.WithError(err).Error(msg)
	}

	return &cli.Command{
		Name:        "build-install",
		Description: "Build picl and install it to nodes",
		Usage:       "Build picl and install it to nodes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Server group configuration to use",
				EnvVars: []string{"cmn_GROUP_CONFIG"},
				Value:   "default",
			},
			&cli.StringFlag{
				Name: "picl-root",
				Usage: "Root of the picl repo. Default is assumes its " +
					"the same repo where running version of picl is built",
				Value: fxRootPath,
			},
			&cli.StringFlag{
				Name:  "arch",
				Usage: "ISA of target machine, sets the GOARCH for the build",
				Value: "arm64",
			},
		},
		Action: func(etx *cli.Context) error {
			config := etx.String("config")
			root := etx.String("picl-root")
			arch := etx.String("arch")

			cmdMan, err := createCmdManager(config)
			if err != nil {
				return err
			}

			if err := mon.BuildAndInstall(cmdMan, root, arch); err != nil {
				return err
			}

			return nil
		},
	}
}
