package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/xcutr"
)

func main() {
	cfg := cfx.MustGetUserHome() + "/.cfx/rpi4.cluster.json"
	var config xcutr.Config
	if err := cfx.LoadJsonFile(cfg, &config); err != nil {
		logrus.Fatal(err)
	}

	cmdMan, err := xcutr.NewCmdMan(&config, xcutr.StdIO{
		Out: os.Stdout,
		Err: os.Stderr,
		In:  os.Stdin,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	if len(os.Args) == 1 {
		logrus.Fatal("Not enough arguments")
	}

	if os.Args[1] == "sudo" {
		if len(os.Args) == 2 {
			logrus.Fatal("Not enough arguments")
		}
		if config.SudoPass == "" {
			config.SudoPass = cfx.AskPassword("sudo password")
		}

		if err := cmdMan.ExecSudo(strings.Join(os.Args[2:], " ")); err != nil {
			logrus.Fatal(err)
		}
		return
	}

	if err := cmdMan.Exec(strings.Join(os.Args[1:], " ")); err != nil {
		logrus.Fatal(err)
	}
}
