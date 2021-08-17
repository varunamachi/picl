package xcutr

import (
	"errors"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	ErrCmdExec = errors.New("xcutr.cmd.failed")
)

type StdIO struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

type Config struct {
	Name     string         `json:"name"`
	SudoPass string         `json:"sudoPass"`
	Opts     []*SshConnOpts `json:"opts"`
}

type CmdMan struct {
	conns  []*SshConn
	config *Config
	io     StdIO
}

// func NewCmdManFromConfigFile(
// 	configPath string, stdIO StdIO) (*CmdMan, error) {
// 	var config Config
// 	if err := cfx.LoadJsonFile(configPath, config); err != nil {
// 		return nil, err
// 	}
// 	return NewCmdMan(&config, stdIO)
// }

func NewCmdMan(config *Config, stdIO StdIO) (*CmdMan, error) {
	conns := make([]*SshConn, 0, len(config.Opts))
	for _, opts := range config.Opts {
		conn, err := NewConn(opts)
		if err != nil {
			logrus.Warn("Diconnecting established connections")
			for name, conn := range conns {
				if err = conn.Close(); err != nil {
					logrus.WithError(err).WithField("conn", name).
						Warn("Failed to disconnect")
				}
			}
			return nil, err
		}
		conns = append(conns, conn)
	}
	return &CmdMan{
		conns:  conns,
		config: config,
		io:     stdIO,
	}, nil
}

func (cm *CmdMan) Exec(cmd string) error {
	failed := 0
	var wg sync.WaitGroup
	wg.Add(len(cm.conns))
	for _, conn := range cm.conns {
		conn := conn
		go func() {
			if err := conn.Exec(cmd, &cm.io); err != nil {
				failed++
			}
			// fmt.Fprintf(cm.io.Out, "\n\n")
			wg.Done()
		}()
	}

	wg.Wait()
	if failed != 0 {
		return NewErrf(ErrCmdExec,
			"Failed to execute command on %d targets", failed)
	}
	return nil
}

func (cm *CmdMan) ExecSudo(cmd string) error {
	failed := 0
	var wg sync.WaitGroup
	wg.Add(len(cm.conns))
	for _, conn := range cm.conns {
		conn := conn
		go func() {
			if err := conn.ExecSudo(
				cmd, cm.config.SudoPass, &cm.io); err != nil {
				failed++
			}
			// fmt.Fprintf(cm.io.Out, "\n\n")
			wg.Done()
		}()
	}
	wg.Wait()
	if failed != 0 {
		return NewErrf(ErrCmdExec,
			"Failed to execute sudo command on %d targets", failed)
	}
	return nil
}
