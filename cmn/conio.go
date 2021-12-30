package cmn

import (
	"fmt"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

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
