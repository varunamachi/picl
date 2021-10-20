package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/clusterfox/agent"
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
			return agent.RunAgent(fmt.Sprintf(":%d", port))
		},
	}
}

func getMonitorCmd() *cli.Command {
	return &cli.Command{
		Name:        "monitor",
		Description: "Start the monitor",
		Usage:       "Start the monitor",
		Flags:       []cli.Flag{},
		Action: func(etx *cli.Context) error {
			port := etx.Int("port")
			return agent.RunAgent(fmt.Sprintf(":%d", port))
		},
	}
}

func getBuildInstallCmd() *cli.Command {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		const msg = "Couldnt get main file path"
		logrus.Fatal(msg)
	}

	fxRootPath, err := filepath.Abs(filename + "/../..")
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
				Name: "fx-root",
				Usage: "Root of the clusterfox repo. Default is assumes its " +
					"the same repo where running version of fx is built",
				Value: fxRootPath,
			},
			&cli.StringFlag{
				Name:  "arch",
				Usage: "ISA of target machine, sets the GOARCH for the build",
				Value: "",
			},
		},
		Action: func(etx *cli.Context) error {
			return nil
		},
	}
}
