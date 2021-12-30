package main

import (
	"os"

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
		},
	}
	if err := app.Run(os.Args); err != nil {
		// logrus.Fatal(err)
		os.Exit(-1)
	}
}
