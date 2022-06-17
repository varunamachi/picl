package qctl

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "qctl",
		Description: "Quick control commands",
		Usage:       "Quick control commands",
		Subcommands: []*cli.Command{
			listControllersCmd(),
			getStatesCmd(),
			setStateCmd(),
			setDefaultStateCmd(),
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}

}

func getStatesCmd() *cli.Command {
	return &cli.Command{
		Name:        "get-state",
		Description: "Get state of one/all switches",
		Usage:       "Get state of one/all switches",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "controller",
				Usage:    "Controller to select",
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func listControllersCmd() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Description: "List all the relay controllers in the network",
		Usage:       "List all the relay controllers in the network",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "service-name",
				Usage: "mDNS service name",
				Value: "_relayctl",
			},
		},
		Action: func(ctx *cli.Context) error {
			service := ctx.String("service-name")
			ctls, err := discover(service)
			if err != nil {
				return err
			}
			for _, ctl := range ctls {
				fmt.Printf("%20s %30s %4d %20v",
					ctl.ShortName,
					ctl.Name,
					ctl.Port,
					ctl.AddrIP4,
				)

			}
			return nil
		},
	}
}

func setStateCmd() *cli.Command {
	return &cli.Command{
		Name:        "set-state",
		Description: "Set state of a switch true/on or false/off",
		Usage:       "Set state of a switch true/on or false/off",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "controller",
				Usage:    "Controller to select",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "slot",
				Usage:    "Switch number",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "state",
				Usage:    "Switch state",
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}

func setDefaultStateCmd() *cli.Command {
	return &cli.Command{
		Name:        "set-default-state",
		Description: "Set the default state of switch",
		Usage:       "Set the default state of switch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "controller",
				Usage:    "Controller to select",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "slot",
				Usage:    "Switch number",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "def-state",
				Usage:    "New default witch state",
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
