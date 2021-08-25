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
			Flags:        withCommonFlags(),
			Action: func(ctx *cli.Context) error {

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
		},
		{
			Name: "exec-sudo",
			Description: "Execute commands on multiple machines " +
				"with sudo permission",
			BashComplete: cli.DefaultAppComplete,
			Flags:        withCommonFlags(),
			Action: func(ctx *cli.Context) error {
				cmdMan, opts, err := getCmdMgrAndOpts(ctx)
				if err != nil {
					return err
				}

				cmd := strings.Join(ctx.Args().Slice(), " ")
				if err := cmdMan.ExecSudo(cmd, opts); err != nil {
					return err
				}
				return nil
			},
		},
	}

}

func getCmdMgrAndOpts(ctx *cli.Context) (
	*xcutr.CmdMan, *xcutr.ExecOpts, error) {

	cfg := ctx.String("config")
	only := ctx.String("only")
	except := ctx.String("except")

	if only != "" && except != "" {
		logrus.Fatalln(
			"Both 'only' and 'except' options cannot be given simultaneously")
	}

	cfgPath := filepath.Join(
		cfx.MustGetUserHome(), ".fx", cfg+".cluster.json")
	var config xcutr.Config
	if err := cfx.LoadJsonFile(cfgPath, &config); err != nil {
		logrus.
			WithError(err).
			WithField("config", cfg).
			Error("Failed to load config")
		return nil, nil, err
	}

	cmdMgr, err := xcutr.NewCmdMan(&config, xcutr.StdIO{
		Out: os.Stdout,
		Err: os.Stderr,
		In:  os.Stdin,
	})

	if err != nil {
		return nil, nil, err
	}

	execOpts := xcutr.ExecOpts{}
	if only != "" {
		execOpts.Included = strings.Split(only, ",")
	}
	if except != "" {
		execOpts.Excluded = strings.Split(except, ",")
	}
	return cmdMgr, &execOpts, nil
}

func withCommonFlags(flags ...cli.Flag) []cli.Flag {
	common := []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "Server group configuration to use",
			EnvVars: []string{"CFX_GROUP_CONFIG"},
			Value:   "default",
		},
		&cli.StringFlag{
			Name: "only",
			Usage: "Comma seperated list of nodes, only on which " +
				"the commands will be executed",
			EnvVars: []string{"CFX_EXEC_ONLY"},
			Value:   "",
		},
		&cli.StringFlag{
			Name: "except",
			Usage: "Comma seperated list of nodes, except which " +
				"the commands will be executed",
			EnvVars: []string{"CFX_EXEC_EXCEPT"},
			Value:   "",
		},
	}
	return append(common, flags...)
}

// func parseCommaSeperated(commaSeperatedStr string) map[string]struct{} {
// 	vals := strings.Split(commaSeperatedStr, ",")
// 	set := make(map[string]struct{})
// 	for _, val := range vals {
// 		if len(val) != 0 {
// 			set[val] = struct{}{}
// 		}
// 	}
// 	return set
// }
