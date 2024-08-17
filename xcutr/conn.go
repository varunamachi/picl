package xcutr

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	fc "github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/iox"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SshAuthMethod string

const (
	SshAuthPublicKey SshAuthMethod = "PublicKey"
	SshAuthPassword  SshAuthMethod = "Password"
)

type SshConnOpts struct {
	Name      string        `json:"name"`
	Host      string        `json:"host"`
	Port      int           `json:"port"`
	UserName  string        `json:"userName"`
	Password  string        `json:"password"`
	AuthMehod SshAuthMethod `json:"authMethod"`
	KeyFile   string        `json:"keyFile"`
	Color     string        `json:"color"`
}

func (opts *SshConnOpts) String() string {
	return fmt.Sprintf("[%s] %s@%s:%d",
		opts.AuthMehod, opts.UserName, opts.Host, opts.Port)
}

func (opts *SshConnOpts) FillDefaults() {
	if opts.AuthMehod == "" {
		opts.AuthMehod = SshAuthPublicKey
	}
	if opts.Port == 0 {
		opts.Port = 22
	}
	if opts.UserName == "" {
		if user, err := user.Current(); err != nil {
			log.Error().Err(err).Msg("Failed to get current user")
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
		const msg = "failed to connect to remote host"
		log.Fatal().Err(err).Str("opts", opts.String()).Msg(msg)
		return nil, errx.Errf(err, msg)
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
		return errx.Wrap(err)
	}
	return nil
}

func (conn *SshConn) Exec(cmd string, stdIO *StdIO) error {
	sess, err := conn.createSession()
	if err != nil {
		return errx.Wrap(err)
	}

	if stdIO == nil {
		stdIO = &StdIO{
			Out: os.Stdout,
			Err: os.Stderr,
			In:  os.Stdin,
		}
	}

	defer closeSession(sess)
	sess.Stdout = NewNodeWriter(conn.Name(), stdIO.Out, color(conn.opts.Color))
	sess.Stderr = NewNodeWriter(conn.Name(), stdIO.Err, color(conn.opts.Color))
	sess.Stdin = stdIO.In
	if err := sess.Run(cmd); err != nil {
		// log.Fatal().Err(err).WithField("cmd", cmd).
		// 	Error("Command execution failed")
		return errx.Errf(err, "Command %s failed to execute", cmd)
	}
	return nil
}

func (conn *SshConn) ExecSudo(cmd string, stdIO *StdIO) error {
	sess, err := conn.createSession()
	if err != nil {
		return errx.Wrap(err)
	}

	if stdIO == nil {
		stdIO = &StdIO{
			Out: os.Stdout,
			Err: os.Stderr,
			In:  os.Stdin,
		}
	}

	cmd = "sudo -S " + cmd
	sess.Stdout = NewNodeWriter(conn.Name(), stdIO.Out, color(conn.opts.Color))
	sess.Stderr = NewNodeWriter(conn.Name(), stdIO.Err, color(conn.opts.Color))
	fmt.Fprintln(sess.Stderr)
	sess.Stdin = strings.NewReader(conn.opts.Password)

	if err := sess.Run(cmd); err != nil {
		// log.Fatal().Err(err).WithField("cmd", cmd).
		// 	Error("Command execution failed")
		return errx.Errf(err, "Command %s failed to execute", cmd)
	}
	return nil
}

func (conn *SshConn) createSession() (*ssh.Session, error) {
	session, err := conn.client.NewSession()
	if err != nil {
		const msg = "failed to create SSH session"
		log.Fatal().Err(err).Str("opts", conn.opts.String()).Msg(msg)
		return nil, errx.Errf(err, msg)
	}
	return session, nil
}

func closeSession(sess *ssh.Session) error {
	err := sess.Close()
	if err != nil && err != io.EOF {
		const msg = "Failed to close ssh session"
		log.Error().Err(err).Msg(msg)
		return errx.Errf(err, msg)
	}
	return nil
}

func getPrivateKeyConfig(opts *SshConnOpts) (*ssh.ClientConfig, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err)
	}

	key, err := GetPrivateKeyFileContent(opts)
	if err != nil {
		return nil, err
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		const msg = "unable read the private key"
		log.Error().Err(err).Msg(msg)
		return nil, errx.Errf(err, msg)
	}

	khFile := filepath.Join(home, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(khFile)
	if err != nil {
		const msg = "could not create hostkeycallback function"
		log.Fatal().Err(err).Str("path", khFile).Msg(msg)
		return nil, errx.Errf(err, msg)
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
	home := iox.MustGetUserHome()
	khFile := filepath.Join(home, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(khFile)
	if err != nil {
		const msg = "could not create hostkeycallback function"
		log.Fatal().Err(err).Str("path", khFile).Msg(msg)
		return nil, errx.Errf(err, msg)
	}

	return &ssh.ClientConfig{
		User: opts.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.Password),
		},
		HostKeyCallback: hostKeyCallback,
	}, nil

}

func color(color string) fc.Attribute {
	switch color {
	case "red":
		return fc.FgRed
	case "green":
		return fc.FgGreen
	case "yellow":
		return fc.FgYellow
	case "blue":
		return fc.FgBlue
	case "magenta":
		return fc.FgMagenta
	case "cyan":
		return fc.FgCyan
	case "white":
		return fc.FgWhite
	}

	switch rand.Intn(10) {
	case 1:
		return fc.FgRed
	case 2:
		return fc.FgGreen
	case 3:
		return fc.FgYellow
	case 4:
		return fc.FgBlue
	case 5:
		return fc.FgMagenta
	case 6:
		return fc.FgCyan
	case 7:
		return fc.FgWhite
	case 8:
		return fc.BgRed
	case 9:
		return fc.BgGreen
	case 10:
		return fc.BgYellow
	case 11:
		return fc.BgBlue
	case 12:
		return fc.BgMagenta
	case 13:
		return fc.BgCyan
	case 14:
		return fc.BgWhite
	}
	return fc.FgWhite
}

func GetPrivateKeyFileContent(opts *SshConnOpts) ([]byte, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	pkFile := filepath.Join(home, ".ssh", "id_rsa")
	if !iox.ExistsAsFile(pkFile) {
		pkFile = filepath.Join(home, ".ssh", "id_ed25519")
	}

	if opts.KeyFile != "" {
		pkFile = filepath.Join(home, ".ssh", opts.KeyFile)
	}
	key, err := os.ReadFile(pkFile)
	if err != nil {
		const msg = "Unable read the private key"
		log.Error().Err(err).Msg("")
		return nil, errx.Errf(err, msg)
	}

	return key, nil
}

func GetPublicKeyFileContent() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	pkFile := filepath.Join(home, ".ssh", "id_rsa.pub")
	if !iox.ExistsAsFile(pkFile) {
		pkFile = filepath.Join(home, ".ssh", "id_ed25519.pub")
	}

	key, err := os.ReadFile(pkFile)
	if err != nil {
		const msg = "Unable read the public key"
		log.Error().Err(err).Msg("")
		return "", errx.Errf(err, msg)
	}

	return string(key), nil
}
