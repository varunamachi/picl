package xcutr

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SshAuthMethod string

const (
	SshAuthPublicKey SshAuthMethod = "PublicKey"
	SshAuthPassword  SshAuthMethod = "Password"
)

type SshConnOpts struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	UserName string `json:"userName"`
	// CanSudo   bool              `json:"canSudo"`
	// SudoPass  string            `json:"sudoPass"`
	AuthMehod SshAuthMethod     `json:"authMethod"`
	AuthData  map[string]string `json:"authData"`
}

func (opts *SshConnOpts) String() string {
	return fmt.Sprintf("[%s] %s@%s:%d",
		opts.AuthMehod, opts.UserName, opts.Host, opts.Port)
}

func (opts *SshConnOpts) FillDefaults() {
	if opts.AuthMehod == "" {
		opts.AuthMehod = SshAuthPublicKey
		opts.AuthData = map[string]string{}
	}
	if opts.Port == 0 {
		opts.Port = 22
	}
	if opts.UserName == "" {
		if user, err := user.Current(); err != nil {
			logrus.WithError(err).Error("Failed to get current user")
		} else {
			opts.UserName = user.Username
		}
	}
}

type SshConn struct {
	opts *SshConnOpts
	// session *ssh.Session
	client *ssh.Client
}

func NewConn(opts *SshConnOpts) (*SshConn, error) {
	var config *ssh.ClientConfig
	opts.FillDefaults()

	var err error
	if opts.AuthMehod == SshAuthPublicKey {
		config, err = getPrivateKeyConfig(opts)
	} else {
		config, err = getPasswordConfig(opts)
	}
	if err != nil {

		return nil, err
	}

	address := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		const msg = "Failed to connect to remote host"
		logrus.WithError(err).WithField("opts", opts.String()).Error(msg)
		return nil, NewErrf(err, msg)
	}
	// defer client.Close()

	return &SshConn{opts: opts, client: client}, nil
}

func (conn *SshConn) Name() string {
	return conn.opts.Name
}

func (conn *SshConn) PrintOpts() {
	fmt.Println(conn.opts)
}

func (conn *SshConn) Close() error {
	if err := conn.client.Close(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (conn *SshConn) Exec(cmd string, stdIO *StdIO) error {
	sess, err := conn.createSession()
	if err != nil {
		return err
	}
	defer closeSession(sess)
	sess.Stdout = stdIO.Out
	sess.Stderr = stdIO.Err
	sess.Stdin = stdIO.In
	if err := sess.Run(cmd); err != nil {
		// logrus.WithError(err).WithField("cmd", cmd).
		// 	Error("Command execution failed")
		return NewErrf(err, "Command %s failed to execute", cmd)
	}
	return nil
}

func (conn *SshConn) ExecSudo(cmd, sudoPass string, stdIO *StdIO) error {
	sess, err := conn.createSession()
	if err != nil {
		return err
	}

	cmd = "sudo -S " + cmd
	sess.Stdout = stdIO.Out
	sess.Stderr = stdIO.Err
	fmt.Fprintln(sess.Stderr)
	sess.Stdin = strings.NewReader(sudoPass)
	if err := sess.Run(cmd); err != nil {
		// logrus.WithError(err).WithField("cmd", cmd).
		// 	Error("Command execution failed")
		return NewErrf(err, "Command %s failed to execute", cmd)
	}
	return nil
}

func (conn *SshConn) createSession() (*ssh.Session, error) {
	session, err := conn.client.NewSession()
	if err != nil {
		const msg = "Failed to create SSH session"
		logrus.WithError(err).WithField("opts", conn.opts.String()).Error(msg)
		return nil, NewErrf(err, msg)
	}
	return session, nil
}

func closeSession(sess *ssh.Session) error {
	err := sess.Close()
	if err != nil && err != io.EOF {
		const msg = "Failed to close ssh session"
		logrus.WithError(err).Error(msg)
		return NewErrf(err, msg)
	}
	return nil
}

func getPrivateKeyConfig(opts *SshConnOpts) (*ssh.ClientConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
	}

	pkFile := filepath.Join(home, ".ssh", "id_rsa")
	if keyFile, found := opts.AuthData["keyFile"]; found {
		pkFile = filepath.Join(home, ".ssh", keyFile)
	}
	key, err := ioutil.ReadFile(pkFile)
	if err != nil {
		logrus.WithError(err).Error("Unable read the private key")
		return nil, NewErrf(err, "Unable read the private key %s", err.Error())
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		logrus.WithError(err).Error("Unable read the private key")
		return nil, NewErrf(err, "Unable read the private key")
	}

	khFile := filepath.Join(home, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(khFile)
	if err != nil {
		const msg = "Could not create hostkeycallback function"
		logrus.WithError(err).WithField("path", khFile).Error(msg)
		return nil, NewErrf(err, msg)
	}

	return &ssh.ClientConfig{
		User: opts.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}, nil

}

func getPasswordConfig(opts *SshConnOpts) (*ssh.ClientConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
	}

	khFile := filepath.Join(home, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(khFile)
	if err != nil {
		const msg = "Could not create hostkeycallback function"
		logrus.WithError(err).WithField("path", khFile).Error(msg)
		return nil, NewErrf(err, msg)
	}

	return &ssh.ClientConfig{
		User: opts.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.AuthData["password"]),
		},
		HostKeyCallback: hostKeyCallback,
	}, nil

}
