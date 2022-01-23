package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:        "picl",
		Description: "Pi cluster controller",
		Authors: []*cli.Author{
			{
				Name: "varunamachi",
			},
		},
		Commands: []*cli.Command{
			getExecCmd(),
			getPullCmd(),
			getPushCmd(),
			getReplicateCmd(),
			getAgentCmd(),
			getMonitorCmd(),
			getBuildInstallCmd(),
			getInteractiveSetupCommand(),
		},
		Usage: "If no valid subcommand is given - it acts as 'exec' " +
			"subcommand. I.e It treats the argument as a " +
			"command that needs to be executed on all the nodes. ",
		Flags: withCmdManFlags(),
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() == 0 {
				cli.ShowAppHelp(ctx)
				return nil
			}
			cmdMan, opts, err := getCmdMgrAndOpts(ctx)
			if err != nil {
				return err
			}

			cmd := strings.Join(ctx.Args().Slice(), " ")
			if err := cmdMan.Exec(cmd, opts); err != nil {
				return err
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		// logrus.Fatal(err)
		os.Exit(-1)
	}
}
