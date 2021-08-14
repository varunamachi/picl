package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/xcutr"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var config xcutr.Config
	if err := cfx.LoadJsonFile("config.json", config); err != nil {
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
			config.SudoPass = AskPassword("sudo password")
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

//askSecret - asks password from user, does not echo charectors
func askSecret() (secret string, err error) {
	var pbyte []byte
	pbyte, err = terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		secret = string(pbyte)
		fmt.Println()
	}
	return secret, err
}

//AskPassword - asks password, prints the given name before asking
func AskPassword(name string) (secret string) {
	fmt.Print(name + ": ")
	secret, err := askSecret()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get secret")
	}
	return secret
}
