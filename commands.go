package clusterfox

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/xcutr"
)

func GetCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:         "exec",
			Description:  "Execute commands on multiple machines",
			BashComplete: cli.DefaultAppComplete,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Usage:   "Server group configuration to use",
					EnvVars: []string{"CFX_GROUP_CONFIG"},
					Value:   "default",
				},
			},
			Action: func(ctx *cli.Context) error {
				cfg := ctx.String("config")

				cfgPath := filepath.Join(
					cfx.MustGetUserHome(), ".fx", cfg+".cluster.json")
				var config xcutr.Config
				if err := cfx.LoadJsonFile(cfgPath, &config); err != nil {
					logrus.
						WithError(err).
						WithField("config", cfg).
						Error("Failed to load config")
					return err
				}

				cmdMan, err := xcutr.NewCmdMan(&config, xcutr.StdIO{
					Out: os.Stdout,
					Err: os.Stderr,
					In:  os.Stdin,
				})
				if err != nil {
					return err
				}

				cmd := strings.Join(ctx.Args().Slice(), " ")
				if err := cmdMan.Exec(cmd); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name: "exec-sudo",
			Description: "Execute commands on multiple machines " +
				"with sudo permission",
			BashComplete: cli.DefaultAppComplete,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Usage:   "Server group configuration to use",
					EnvVars: []string{"CFX_GROUP_CONFIG"},
					Value:   "default",
				},
			},
			Action: func(ctx *cli.Context) error {
				cfg := ctx.String("config")

				cfgPath := filepath.Join(
					cfx.MustGetUserHome(), ".fx", cfg+".cluster.json")
				var config xcutr.Config
				if err := cfx.LoadJsonFile(cfgPath, &config); err != nil {
					logrus.
						WithError(err).
						WithField("config", cfg).
						Error("Failed to load config")
					return err
				}

				if config.SudoPass == "" {
					config.SudoPass = cfx.AskPassword("sudo password")
				}

				cmdMan, err := xcutr.NewCmdMan(&config, xcutr.StdIO{
					Out: os.Stdout,
					Err: os.Stderr,
					In:  os.Stdin,
				})
				if err != nil {
					return err
				}

				cmd := strings.Join(ctx.Args().Slice(), " ")
				if err := cmdMan.ExecSudo(cmd); err != nil {
					return err
				}
				return nil
			},
		},
	}
}
