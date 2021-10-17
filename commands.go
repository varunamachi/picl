package clusterfox

import (
	"errors"
	"fmt"
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
			Usage:        "Execute commands on multiple machines",
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
			Name:         "pull",
			Usage:        "Copy a file from remote to local",
			Description:  "Copy a file from remote to local",
			BashComplete: cli.DefaultAppComplete,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Usage:   "Server group configuration to use",
					EnvVars: []string{"CFX_GROUP_CONFIG"},
					Value:   "default",
				},
				&cli.StringFlag{
					Name:     "local-path",
					Usage:    "Local destination file path",
					Required: true,
				},
				&cli.StringFlag{
					Name: "remote",
					Usage: "Remote source file path, should be of the " +
						"form <nodeName>:<remotePath>",
					Required: true,
				},
			},
			Action: func(ctx *cli.Context) error {
				local := ctx.String("local-path")
				remote := ctx.String("remote")

				parts := strings.SplitN(remote, ":", 2)
				fmt.Println(parts)
				if len(parts) != 2 {
					return cfx.Errf(errors.New("invalid remote format"),
						"Invalud remote file provided, should be of the form: "+
							" <nodeName>:<remotePath>")
				}

				cmdMan, _, err := getCmdMgrAndOpts(ctx)
				if err != nil {
					return err
				}
				err = cmdMan.Pull(parts[0], parts[1], local)
				return err
			},
		},
		{
			Name:         "push",
			Usage:        "Push a file from local to remote",
			Description:  "Copy a file from local to remote",
			BashComplete: cli.DefaultAppComplete,
			Flags: withCommonFlags(
				&cli.StringFlag{
					Name:     "local-path",
					Usage:    "Local destination file path",
					Required: true,
				},
				&cli.StringFlag{
					Name: "remote-path",
					Usage: "Remote destination file path " +
						"(without node name)",
					Required: true,
				},
				&cli.StringFlag{
					Name: "fileConflictPolicy",
					Usage: "What should happen if file already exists in " +
						"remote, supports these options: ignore | replace",
					Value: "ignore",
				},
			),
			Action: func(ctx *cli.Context) error {
				local := ctx.String("local-path")
				remote := ctx.String("remote-path")
				policy := toFileConfictPolicy(ctx.String("fileConflictPolicy"))
				cmdMan, opts, err := getCmdMgrAndOpts(ctx)
				if err != nil {
					return err
				}
				copyOpts := xcutr.CopyOpts{
					ExecOpts:      *opts,
					DupFilePolicy: policy,
				}

				return cmdMan.Push(local, remote, &copyOpts)
			},
		},
		{
			Name:  "replicate",
			Usage: "Replicate a file from one remote node to other nodes",
			Description: "Replicate a file from one remote node to others, " +
				"with same path",
			BashComplete: cli.DefaultAppComplete,
			Flags: withCommonFlags(
				&cli.StringFlag{
					Name: "remote",
					Usage: "Remote source file path, should be of the " +
						"form <nodeName>:<remotePath>",
					Required: true,
				},
				&cli.StringFlag{
					Name: "fileConflictPolicy",
					Usage: "What should happen if file already exists in " +
						"remote, supports these options: ignore | replace",
					Value: "ignore",
				},
			),
			Action: func(ctx *cli.Context) error {
				policy := toFileConfictPolicy(ctx.String("fileConflictPolicy"))

				remote := ctx.String("remote")
				parts := strings.SplitN(remote, ":", 2)
				if len(parts) != 2 {
					return cfx.Errf(errors.New("invalid remote format"),
						"Invalud remote file provided, should be of the form: "+
							" <nodeName>:<remotePath>")
				}

				cmdMan, opts, err := getCmdMgrAndOpts(ctx)
				if err != nil {
					return err
				}
				copyOpts := xcutr.CopyOpts{
					ExecOpts:      *opts,
					DupFilePolicy: policy,
				}
				return cmdMan.Replicate(parts[0], parts[1], &copyOpts)
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
	execOpts.WithSudo = ctx.Bool("sudo")
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
		&cli.BoolFlag{
			Name:    "sudo",
			Usage:   "Runs command with sudo privilege",
			EnvVars: []string{"CFX_EXEC_SUDO"},
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

func toFileConfictPolicy(str string) xcutr.ExistingFilePolicy {
	switch str {
	case "ignore":
		return xcutr.Ignore
	case "replace":
		return xcutr.Replace
	}
	return xcutr.Ignore
}
