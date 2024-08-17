package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/httpx"
	"github.com/varunamachi/picl/config"
	"github.com/varunamachi/picl/mon"
	"github.com/varunamachi/picl/xcutr"
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
		Action: func(ctx *cli.Context) error {
			port := ctx.Int("port")
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
		Action: func(ctx *cli.Context) error {
			port := ctx.Uint("port")
			handler := ctx.String("handler")

			provider, err := config.NewFromCli(ctx)
			if err != nil {
				return errx.Wrap(err)
			}

			monConfig := provider.MonitorConfig()

			hdl, gtx, err := newHandler(handler, monConfig)
			if err != nil {
				return errx.Wrap(err)
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
				gtx,
				monConfig,
				rcfg,
				hdl,
				httpx.NewServer(printer, nil))

			if err != nil {
				log.Error().Err(err).Msg("failed to initialize monitor")
				return errx.Wrap(err)
			}
			if err = monitor.Run(gtx, uint32(port)); err != nil {
				log.Error().Err(err).Msg("failed to run monitor")
				return errx.Wrap(err)
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
		log.Fatal().Msg(msg)
	}

	fxRootPath, err := filepath.Abs(filename + "/../../..")
	if err != nil {
		const msg = "couldnt get main file path"
		log.Error().Err(err).Msg(msg)
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
		Action: func(ctx *cli.Context) error {
			// config := ctx.String("config")
			root := ctx.String("picl-root")
			arch := ctx.String("arch")

			provider, err := config.NewFromCli(ctx)
			if err != nil {
				return errx.Wrap(err)
			}

			cmdMan, err := xcutr.NewCmdMan(
				provider.ExecuterConfig(), xcutr.StdIO{
					Out: os.Stdout,
					Err: os.Stderr,
					In:  os.Stdin,
				})
			if err != nil {
				return errx.Wrap(err)
			}

			if err := mon.BuildAndInstall(cmdMan, root, arch); err != nil {
				return errx.Wrap(err)
			}

			return nil
		},
	}
}
