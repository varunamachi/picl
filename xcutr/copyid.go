package xcutr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/varunamachi/picl/cmn"
)

var (
	ErrAuthKeyFileRead    = "xcutr.ssh.ak.file"
	ErrAuthKeyFileInvalud = "xcutr.ssh.ak.format"
)

type AuthzKeysRow struct {
	Options string
	KeyType string
	Key     string
	Comment string
}

func CopyId(sshCfg []*SshConnOpts) error {
	pubKey, err := GetPublicKeyFileContent()
	if err != nil {
		return err
	}

	pubRow, err := processLine(strings.TrimSpace(pubKey))
	if err != nil {
		return err
	}

	failures := 0
	for _, opts := range sshCfg {
		copier, err := newCopier(opts)
		if err != nil {
			failures++
			continue
		}
		if err = copier.copyId(pubRow); err != nil {
			failures++
			fmt.Printf("Skipping ID copy for %s", opts.Host)
			continue
		}
	}
	if failures != 0 {
		msg := fmt.Sprintf(
			"could not copy id to all nodes (%d out of %d failed)",
			failures, len(sshCfg))
		fmt.Println(msg)
		return errors.New(msg)
	}

	return nil
}

type idCopier struct {
	conn         *SshConn
	fcon         *sftp.Client
	authzKeyPath string
	stdIO        StdIO
}

func newCopier(opts *SshConnOpts) (*idCopier, error) {
	conn, err := NewConn(opts)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(conn.client)
	if err != nil {
		return nil, cmn.Errf(err, "failed to create sftp client")
	}

	return &idCopier{
		conn: conn,
		fcon: sftpClient,
		authzKeyPath: filepath.Join(
			"/home", conn.opts.UserName, ".ssh", "authorized_keys"),
		stdIO: StdIO{
			Out: NewNodeWriter(conn.Name(), os.Stdout, color(conn.opts.Color)),
			Err: NewNodeWriter(conn.Name(), os.Stderr, color(conn.opts.Color)),
			In:  os.Stdin,
		},
	}, nil
}

func (cpr *idCopier) info(msg string, args ...interface{}) {
	fmt.Fprintf(cpr.stdIO.Out, msg, args...)
}

func (cpr *idCopier) err(msg string, args ...interface{}) {
	fmt.Fprintf(cpr.stdIO.Err, msg, args...)
}

func (cpr *idCopier) copyId(pubKey *AuthzKeysRow) error {

	cpr.info("reading authorized_keys file from %s", cpr.authzKeyPath)
	rows, err := cpr.readAuthorizedKeys()
	if err != nil {
		return err
	}

	backupFilePath := ""
	if len(rows) == 0 {
		backupFilePath = cpr.authzKeyPath + "_" +
			time.Now().Format("20060102_150405")

		cmd := fmt.Sprintf("cp %s %s", cpr.authzKeyPath, backupFilePath)
		if err = cpr.conn.Exec(cmd, nil); err != nil {
			return cmn.Errf(err, "failed to back up authorized_keys file")
		}
	}

	success := false
	defer func() {
		if !success && backupFilePath != "" {
			cpr.info("restoring backed up authorized keys file")
			cmd := fmt.Sprintf("rm -rf %s", cpr.authzKeyPath)
			if err = cpr.conn.Exec(cmd, nil); err != nil {
				cpr.err(
					"failed to remove incomplete authorized_keys file: %v", err)
				return
			}
			cmd = fmt.Sprintf("mv %s %s", backupFilePath, cpr.authzKeyPath)
			if err = cpr.conn.Exec(cmd, nil); err != nil {
				cpr.err("failed to back up authorized_keys file: %s", err)
				return
			}
		}
	}()

	for _, row := range rows {
		if row.KeyType == pubKey.KeyType && row.Key == pubKey.Key {
			cpr.info("public key already exists in authorized_keys file")
			return nil
		}
	}
	rows = append(rows, pubKey)

	file, err := cpr.fcon.Create(cpr.authzKeyPath)
	if err != nil {
		err = cmn.Errf(err, "failed to create/open authorized_keys to write")
		cpr.err(err.Error())
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			cpr.err("failed to close authroized_keys file: %v", err)
		}
	}()

	if err = cpr.writeAuthorizedKeys(file, rows); err != nil {
		err = cmn.Errf(err, "failed to update authorized_keys file")
		cpr.err(err.Error())
		return err
	}

	success = true
	return nil

}

func (cpr *idCopier) readAuthorizedKeys() ([]*AuthzKeysRow, error) {

	file, err := cpr.fcon.Open(cpr.authzKeyPath)
	if err == sftp.ErrSSHFxNoSuchFile {
		return make([]*AuthzKeysRow, 0, 1), nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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

func (cpr *idCopier) writeAuthorizedKeys(
	writer io.Writer, keys []*AuthzKeysRow) error {

	for _, key := range keys {
		_, err := fmt.Fprintf(
			writer,
			"%s %s %s %s",
			key.Options, key.KeyType, key.Key, key.Comment)
		if err != nil {
			return cmn.Errf(
				err, "failed to write a row into authorized_keys file")
		}
	}
	return nil
}

func (cpr *idCopier) verifyConnection() error {
	// try to connect with public key and check if the copy id worked
	return nil
}

func processLine(line string) (*AuthzKeysRow, error) {
	azk := AuthzKeysRow{}

	parts := strings.Fields(line)
	index := 0
	if !startsWithKey(line) {
		azk.Options = strings.TrimSpace(parts[index])
		index++
	}
	azk.KeyType = strings.TrimSpace(parts[index])
	index++

	azk.Key = strings.TrimSpace(parts[index])
	index++

	azk.Comment = strings.TrimSpace(parts[index])

	return nil, nil
}

func startsWithKey(part string) bool {
	return part == "ssh-rsa" ||
		part == "ssh-dss" ||
		part == "ssh-ed25519" ||
		part == "ecdsa-sha" ||
		part == "sk-ecdsa-sha" ||
		part == "sk-ssh-ed25519"
}
