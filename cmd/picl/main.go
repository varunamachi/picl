package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx"
	"github.com/varunamachi/libx/errx"
)

func main() {
	cApp := cli.App{
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
			getInteractiveSetupCmd(),
			getCopyIdCmd(),
			getEncryptCmd(),
			getDecryptCmd(),
		},
		Usage: "If no valid subcommand is given - it acts as 'exec' " +
			"subcommand. I.e It treats the argument as a " +
			"command that needs to be executed on all the nodes. ",
		Flags: withCmdManFlags(&cli.StringFlag{
			Name:  "log-level",
			Value: "info",
			Usage: "Give log level, one of: 'trace', 'debug', " +
				"'info', 'warn', 'error'",
		}),
		Before: func(ctx *cli.Context) error {
			log.Logger = log.Output(
				zerolog.ConsoleWriter{Out: os.Stderr}).
				With().Caller().Logger()
			logLevel := ctx.String("log-level")
			if logLevel != "" {
				level := zerolog.InfoLevel
				switch logLevel {
				case "trace":
					level = zerolog.TraceLevel
				case "debug":
					level = zerolog.DebugLevel
				case "info":
					level = zerolog.InfoLevel
				case "warn":
					level = zerolog.WarnLevel
				case "error":
					level = zerolog.ErrorLevel
				}
				zerolog.SetGlobalLevel(level)
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() == 0 {
				cli.ShowAppHelp(ctx)
				return nil
			}
			cmdMan, opts, err := getCmdMgrAndOpts(ctx)
			if err != nil {
				return errx.Wrap(err)
			}

			cmd := strings.Join(ctx.Args().Slice(), " ")
			if err := cmdMan.Exec(cmd, opts); err != nil {
				return errx.Wrap(err)
			}
			return nil
		},
	}

	app := libx.NewCustomApp(&cApp)
	if err := app.Run(os.Args); err != nil {

		fmt.Println()
		fmt.Println()
		errx.PrintSomeStack(err)
		log.Fatal().Err(err).Msg("")
	}
}
