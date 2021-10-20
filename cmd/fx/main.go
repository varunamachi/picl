package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:        "fx",
		Description: "Clusterfox!",
		Commands: []*cli.Command{
			getExecCmd(),
			getPullCmd(),
			getPushCmd(),
			getReplicateCmd(),
			getAgentCmd(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		// logrus.Fatal(err)
		os.Exit(-1)
	}
}
