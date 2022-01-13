package cmn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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

type UserInputReader struct {
	output io.Writer
	input  io.Reader
	reader bufio.Reader
}

func NewUserInputReader(input io.Reader, output io.Writer) *UserInputReader {
	return &UserInputReader{
		input:  input,
		output: output,
		reader: *bufio.NewReader(input),
	}
}

func StdUserInputReader() *UserInputReader {
	return &UserInputReader{
		input:  os.Stdin,
		output: os.Stdout,
		reader: *bufio.NewReader(os.Stdin),
	}
}

func (uir *UserInputReader) ReadInt(name string) int {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Fprintln(uir.output, "Invalid integer given: ", err.Error())
		os.Exit(2)
	}
	return val
}

func (uir *UserInputReader) ReadFloat(name string) float64 {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Fprintln(uir.output, "Invalid read number given: ", err.Error())
		os.Exit(2)
	}
	return val
}

func (uir *UserInputReader) ReadString(name string) string {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read string: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		fmt.Fprintln(uir.output, "Empty string given: ", err.Error())
		os.Exit(2)
	}
	return str
}

func (uir *UserInputReader) ReadBoolOr(question string, def bool) bool {

	msg := " [y|N]: "
	if def {
		msg = " [Y|n]: "
	}

	fmt.Fprint(uir.output, question, msg)
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}

	if str == "" {
		return def
	}

	str = strings.ToLower(str)
	if str == "y" || str == "yes" || str == "on" {
		return true
	} else if str == "n" || str == "no" || str == "off" {
		return false
	}
	fmt.Fprintln(uir.output, "Invalid bool value", str, "given")
	os.Exit(2)
	return false
}

func (uir *UserInputReader) ReadIntOr(name string, def int) int {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		return def
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Fprintln(uir.output, "Invalid integer given: ", err.Error())
		os.Exit(2)
	}
	return val
}

func (uir *UserInputReader) ReadFloatOr(name string, def float64) float64 {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}

	if str == "" {
		return def
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Fprintln(uir.output, "Invalid read number given: ", err.Error())
		os.Exit(2)
	}
	return val
}

func (uir *UserInputReader) ReadStringOr(name string, def string) string {
	fmt.Fprint(uir.output, "Please enter ", name, ": ")
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		return def
	}
	return str
}

func (uir *UserInputReader) ReadOption(
	name string, options []string, def string) string {

	// fmt.Fprint(uir.output, "Please enter ", name, ": ")
	buf := bytes.NewBufferString("Please enter ")
	buf.WriteString(name)
	buf.WriteString("(")

	found := false
	for i, o := range options {
		if def == o {
			buf.WriteString("[")
			buf.WriteString(o)
			buf.WriteString("]")
		} else {
			buf.WriteString(o)
		}

		if i != len(options)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(")")

	if !found {
		fmt.Fprintln(uir.output,
			"Default value: '", def, "' is not part of options")
		os.Exit(1)
	}

	fmt.Fprintln(uir.output, buf.String())
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read an option: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		return def
	}

	for _, o := range options {
		if str == o {
			return str
		}
	}

	fmt.Fprintln(uir.output, "Invalid option ", str, " given")
	return ""
}
