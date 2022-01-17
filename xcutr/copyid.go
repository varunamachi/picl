package xcutr

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

var (
	ErrAuthKeyFileRead    = "xcutr.ssh.ak.file"
	ErrAuthKeyFileInvalud = "xcutr.ssh.ak.format"
)

func SshCopyId() error {

	return nil
}

type AuthzKeysRow struct {
	Options string
	Key     ssh.PublicKey
	Comment string
}

func CopyDefaultId() error {
	return nil
}

func CopyId(pubKeyFile string) error {
	return nil

	// pk.Marshal()
}

func parseAuthorizedKeys(reader io.Reader) ([]*AuthzKeysRow, error) {

	scanner := bufio.NewScanner(reader)

	rows := make([]*AuthzKeysRow, 0, 20)
	for scanner.Scan() {

		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		row, err := processLine(line)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

func processLine(line string) (*AuthzKeysRow, error) {
	return nil, nil
}
