package relayctl

import (
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/str"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "qctl",
		Description: "Quick control commands",
		Usage:       "Quick control commands",
		Subcommands: []*cli.Command{
			listControllersCmd(),
			getStatesCmd(),
			getDefaultStatesCmd(),
			setStateCmd(),
			setDefaultStateCmd(),
			setAllStatesCmd(),
		},
		Action: func(ctx *cli.Context) error {
			fmt.Println("hello!")
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
				Name:  "service",
				Usage: "mDNS service name",
				Value: "_relayctl",
			},
		},
		Action: func(ctx *cli.Context) error {
			service := ctx.String("service")
			ctls, err := discover(service)
			if err != nil {
				return err
			}
			for ctl := range ctls {
				fmt.Printf("%20s %40s %4d %20v\n",
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

func getStatesCmd() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Description: "Get state of one/all switches",
		Usage:       "Get state of one/all switches",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ctlr",
				Usage: "Controller to select",
				Value: "boxer",
			},
			// &cli.StringFlag{
			// 	Name:  "service",
			// 	Usage: "mDNS service name",
			// 	Value: "_relayctl",
			// },
		},
		Action: func(ctx *cli.Context) error {
			ch, err := getClientTo("_relayctl", ctx.String("ctlr"))
			if err != nil {
				return err
			}

			cwrap := <-ch
			if cwrap.err != nil {
				return cwrap.err
			}
			res := cwrap.client.Get(ctx.Context, "state")
			states := make([]bool, 0, 4)
			if err := res.LoadClose(&states); err != nil {
				return errx.Errf(err, "failed to get switch state")
			}
			for idx, state := range states {
				fmt.Println(idx, state)
			}
			return nil
		},
	}
}

func getDefaultStatesCmd() *cli.Command {
	return &cli.Command{
		Name:        "get-def",
		Description: "Get default state of one/all switches",
		Usage:       "Get default state of one/all switches",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ctlr",
				Usage: "Controller to select",
				Value: "boxer",
			},
			// &cli.StringFlag{
			// 	Name:  "service",
			// 	Usage: "mDNS service name",
			// 	Value: "_relayctl",
			// },
		},
		Action: func(ctx *cli.Context) error {
			ch, err := getClientTo("_relayctl", ctx.String("ctlr"))
			if err != nil {
				return err
			}

			cwrap := <-ch
			if cwrap.err != nil {
				return cwrap.err
			}
			res := cwrap.client.Get(ctx.Context, "stateDefault")
			states := make([]bool, 0, 4)
			if err := res.LoadClose(&states); err != nil {
				return errx.Errf(err, "failed to get switch default state")
			}
			for idx, state := range states {
				fmt.Println(idx, state)
			}
			return nil
		},
	}
}

func setStateCmd() *cli.Command {
	return &cli.Command{
		Name:        "set",
		Description: "Set state of a switch to true/on or false/off",
		Usage:       "Set state of a switch to true/on or false/off",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ctlr",
				Usage:    "Controller to select",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "slot",
				Usage:    "Switch number",
				Required: true,
			},
			&cli.StringFlag{
				Name: "state",
				Usage: "New default witch state: For ON: true | 1 | on and " +
					"for OFF: false | 0 | off",
			},
		},
		Action: func(ctx *cli.Context) error {
			stateStr := ctx.String("state")
			state := false
			if str.EqFold(stateStr, "true", "1", "on") {
				state = true
			}

			ch, err := getClientTo("_relayctl", ctx.String("ctlr"))
			if err != nil {
				return err
			}

			cw := <-ch
			if cw.err != nil {
				return cw.err
			}

			url := fmt.Sprintf("set?slot=%d&state=%t", ctx.Int("slot"), state)
			res := cw.client.Post(ctx.Context, nil, url)
			if err := res.Error(); err != nil {
				return errx.Errf(err, "failed to set default switch state")
			}
			return nil
		},
	}
}

func setAllStatesCmd() *cli.Command {
	return &cli.Command{
		Name:        "set-all",
		Description: "Set states of all the switches",
		Usage:       "Set states of all the switches",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ctlr",
				Usage:    "Controller to select",
				Required: true,
			},
			&cli.StringFlag{
				Name: "state",
				Usage: "New default witch state: For ON: true | 1 | on and " +
					"for OFF: false | 0 | off",
			},
		},
		Action: func(ctx *cli.Context) error {
			stateStr := ctx.String("state")
			state := false
			if str.EqFold(stateStr, "true", "1", "on") {
				state = true
			}

			ch, err := getClientTo("_relayctl", ctx.String("ctlr"))
			if err != nil {
				return err
			}

			cw := <-ch
			if cw.err != nil {
				return cw.err
			}

			url := fmt.Sprintf("setAll?state=%t", state)
			res := cw.client.Post(ctx.Context, nil, url)
			if err := res.Error(); err != nil {
				if !errors.Is(err, io.ErrUnexpectedEOF) {
					return errx.Errf(err, "failed to set all switch states")
				}
				// Ignore - this needs to be fixed in firmware
			}
			return nil
		},
	}
}

func setDefaultStateCmd() *cli.Command {
	return &cli.Command{
		Name:        "set-def",
		Description: "Set the default state of switch",
		Usage:       "Set the default state of switch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ctlr",
				Usage:    "Controller to select",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "slot",
				Usage:    "Switch number",
				Required: true,
			},
			&cli.StringFlag{
				Name: "def-state",
				Usage: "New default witch state: For ON: true | 1 | on and " +
					"for OFF: false | 0 | off",
			},
		},
		Action: func(ctx *cli.Context) error {
			stateStr := ctx.String("def-state")
			state := false
			if str.EqFold(stateStr, "true", "1", "on") {
				state = true
			}

			ch, err := getClientTo("_relayctl", ctx.String("ctlr"))
			if err != nil {
				return err
			}

			cw := <-ch
			if cw.err != nil {
				return cw.err
			}

			url := fmt.Sprintf(
				"stateDefault?slot=%d&state=%t", ctx.Int("slot"), state)
			res := cw.client.Post(ctx.Context, nil, url)
			if err := res.Error(); err != nil {
				return errx.Errf(err, "failed to get controller switch states")
			}
			return nil
		},
	}
}
