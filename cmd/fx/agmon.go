package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/mon"
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
				Usage: "Name of the config, as: ~/.fx/'config'.monitor.json",
				Value: "default",
			},
		},
		Action: func(etx *cli.Context) error {
			cfg := etx.String("config")
			cfgPath := filepath.Join(
				cfx.MustGetUserHome(), ".fx", cfg+".monitor.json")
			var config mon.MonitorConfig
			if err := cfx.LoadJsonFile(cfgPath, &config); err != nil {
				logrus.
					WithError(err).
					WithField("config", cfg).
					Error("Failed to load config")

				config.PrintSampleJSON()
				return err
			}

			gtx := context.Background()
			monitor, err := mon.NewMonitor(
				gtx,
				&config,
				&mon.TuiHandler{})

			if err != nil {
				return err
			}

			return monitor.Run(gtx)
		},
	}
}

func getBuildInstallCmd() *cli.Command {

	//<root>/cmd/fx/agmon.go
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
		Description: "Build fx and install it to nodes",
		Usage:       "Build fx and install it to nodes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Server group configuration to use",
				EnvVars: []string{"CFX_GROUP_CONFIG"},
				Value:   "default",
			},
			&cli.StringFlag{
				Name: "fx-root",
				Usage: "Root of the clusterfox repo. Default is assumes its " +
					"the same repo where running version of fx is built",
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
			root := etx.String("fx-root")
			arch := etx.String("arm64")

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
