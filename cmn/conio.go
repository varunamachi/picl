package cmn

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

//askSecret - asks password from user, does not echo charectors
func askSecret() (secret string, err error) {
	var pbyte []byte
	pbyte, err = term.ReadPassword(int(syscall.Stdin))
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

func (uir *UserInputReader) Int(name string) int {
	fmt.Fprint(uir.output, "Please enter a number for '", name, "'*: ")
	str, err := uir.readString()
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

func (uir *UserInputReader) Float(name string) float64 {
	fmt.Fprint(uir.output, "Please enter real number for '", name, "'*: ")
	str, err := uir.readString()
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

func (uir *UserInputReader) String(name string) string {
	fmt.Fprint(uir.output, "Please enter a string for ", name, "*: ")
	str, err := uir.readString()
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read string: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		fmt.Fprintln(uir.output, "Empty string given")
	}
	return str
}

func (uir *UserInputReader) BoolOr(question string, def bool) bool {

	msg := " [y|N|q]: "
	if def {
		msg = " [Y|n|q]: "
	}

	for {
		fmt.Fprint(uir.output, question, msg)
		str, err := uir.readString()
		if err != nil {
			fmt.Fprintln(uir.output, "Failed to read boolean: ", err.Error())
			os.Exit(1)
		}

		switch {
		case str == "":
			return def
		case EqFold(str, "y", "yes", "true", "on"):
			return true
		case EqFold(str, "n", "no", "false", "off"):
			return false
		case EqFold(str, "q", "Q"):
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid input, try again")
		}
	}
}

func (uir *UserInputReader) IntOr(name string, def int) int {
	fmt.Fprintf(uir.output,
		"Please enter integer value for '%s' [%d]: ", name, def)
	str, err := uir.readString()
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

func (uir *UserInputReader) FloatOr(name string, def float64) float64 {
	fmt.Fprintf(uir.output,
		"Please enter real number value for '%s' [%.2f]: ", name, def)
	str, err := uir.readString()
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

func (uir *UserInputReader) StringOr(name string, def string) string {
	fmt.Fprintf(uir.output, "Please enter string for '%s' [%s]: ", name, def)
	str, err := uir.readString()
	if err != nil {
		fmt.Fprintln(uir.output, "Failed to read integer: ", err.Error())
		os.Exit(1)
	}
	if str == "" {
		return def
	}
	return str
}

func (uir *UserInputReader) Select(
	name string, options []string, def string) string {

	fmt.Fprintf(
		uir.output,
		"Choose one of the following options for '%s':\n",
		name)
	for idx, opt := range options {
		if opt == def {
			fmt.Fprintf(uir.output, "\t\t%d. [%s]\n", idx+1, opt)
			continue
		}
		fmt.Fprintf(uir.output, "\t\t%d. %s\n", idx+1, opt)
	}

	for {
		fmt.Fprintf(
			uir.output,
			"\tEnter value between 1 and %d (inclusive) for '%s' [%s]: ",
			len(options), name, def)
		str, err := uir.readString()
		if err != nil {
			fmt.Fprintln(uir.output, "Failed to read an option: ", err.Error())
			os.Exit(1)
		}
		if str == "" {
			return def
		}
		idx, err := strconv.Atoi(str)
		if err == nil && idx > 0 && idx <= len(options) {
			return options[idx-1]
		}
	}
}

//Secret - asks password from user, does not echo charectors
func (uir *UserInputReader) Secret(msg string) string {
	for {
		fmt.Fprint(uir.output, msg, ": ")
		pbyte, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintln(uir.output, "Error getting password")
			os.Exit(2)
			return ""
		}
		str := strings.TrimSpace(string(pbyte))
		fmt.Println()
		if str != "" {
			return str
		}
	}
}

func (uir *UserInputReader) readString() (string, error) {
	str, err := uir.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(str), nil
}
