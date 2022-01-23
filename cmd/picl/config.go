package main

import (
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/picl/config"
)

func getInteractiveSetupCommand() *cli.Command {
	return &cli.Command{
		Name: "setup",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config-name",
				Usage: "Name of the configuration that determines",
				Value: "default",
			},
		},
		Action: func(ctx *cli.Context) error {
			cfgName := ctx.String("config-name")
			return config.CreateConfig(cfgName)
		},
	}
}
