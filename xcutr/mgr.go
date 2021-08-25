package xcutr

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
)

var (
	ErrCmdExec     = errors.New("xcutr.cmd.failed")
	ErrInvalidNode = errors.New("xcutr.node.invalid")
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
	conns   []*SshConn
	connMap map[string]*SshConn
	config  *Config
	io      StdIO
}

type ExecOpts struct {
	Included []string
	Excluded []string
}

func NewCmdMan(config *Config, stdIO StdIO) (*CmdMan, error) {
	conns := make([]*SshConn, 0, len(config.Opts))
	connMap := make(map[string]*SshConn)
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
		connMap[opts.Name] = conn
	}
	return &CmdMan{
		conns:   conns,
		connMap: connMap,
		config:  config,
		io:      stdIO,
	}, nil
}

// func (cm *CmdMan) Include(nodes []string) *CmdMan {
// 	if len(nodes) == 0 {
// 		return cm
// 	}
// 	// for _, conn := range cm.conns {
// 	// 	_, found := nodes[conn.Name()]
// 	// 	conn.disabled = !found
// 	// }
// 	for _, node := range nodes {

// 	}
// 	return cm
// }

// func (cm *CmdMan) Exclude(nodes map[string]struct{}) *CmdMan {
// 	if len(nodes) == 0 {
// 		return cm
// 	}
// 	for _, conn := range cm.conns {
// 		_, found := nodes[conn.Name()]
// 		conn.disabled = found
// 	}
// 	return cm
// }

func (cm *CmdMan) connList(opts *ExecOpts) []*SshConn {
	if len(opts.Included) != 0 && len(opts.Excluded) != 0 {
		return cm.conns
	}

	conns := make([]*SshConn, 0, len(cm.conns))
	if len(opts.Included) != 0 {
		for _, inc := range opts.Included {
			if conn, found := cm.connMap[inc]; found {
				conns = append(conns, conn)
			}
		}

	} else if len(opts.Excluded) != 0 {
		for _, con := range cm.conns {
			exclude := false
			for _, ex := range opts.Excluded {
				if con.Name() == ex {
					exclude = true
					break
				}
			}
			if !exclude {
				conns = append(conns, con)
			}
		}
	}
	return conns
}

func (cm *CmdMan) Exec(cmd string, opts *ExecOpts) error {
	failed := 0
	conns := cm.connList(opts)
	if len(conns) == 0 {
		logrus.Warn("Could find any node that satisfies current config")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(conns))
	for _, conn := range conns {
		conn := conn
		go func() {
			if err := conn.Exec(cmd, &cm.io); err != nil {
				failed++
			}
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

func (cm *CmdMan) ExecSudo(cmd string, opts *ExecOpts) error {
	failed := 0
	conns := cm.connList(opts)
	if len(conns) == 0 {
		logrus.Warn("Could find any node that satisfies current config")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(conns))
	for _, conn := range conns {
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

func (cm *CmdMan) Pull(node, remotePath, localPath string) error {
	conn := cm.connMap[node]
	if conn == nil {
		logrus.WithField("nodeName", node).Error("Invalid node name given")
		return NewErrf(ErrInvalidNode, "Invalid node name given: %s", node)
	}

	sftpClient, err := sftp.NewClient(conn.client)
	if err != nil {
		const msg = "Failed to create SFTP client"
		logrus.WithError(err).WithField("nodeName", node).Error(msg)
		return NewErrf(err, msg)
	}

	remote, err := sftpClient.Open(remotePath)
	if err != nil {
		const msg = "Failed to read remote file"
		logrus.WithError(err).WithFields(logrus.Fields{
			"nodeName":   node,
			"remotePath": remotePath,
		}).Error(msg)
		return NewErrf(err, msg)
	}

	local, err := os.Create(localPath)
	if err != nil {
		const msg = "Failed to create local file"
		logrus.WithError(err).WithFields(logrus.Fields{
			"localPath": remotePath,
		}).Error(msg)
		return NewErrf(err, msg)
	}

	_, err = io.Copy(local, remote)
	if err != nil {
		const msg = "Failed to copy remote file to local"
		logrus.WithError(err).WithFields(logrus.Fields{
			"nodeName":   node,
			"remotePath": remotePath,
			"localPath":  remotePath,
		}).Error(msg)
		return NewErrf(err, msg)
	}

	return nil
}

func (cm *CmdMan) Push(localPath, remotePath string, opts *ExecOpts) error {
	// local, err := os.Open(localPath)
	// if err != nil {
	// 	// TODO - log and stuff
	// 	return err
	// }

	// conns := cm.connList(opts)
	// if len(conns) == 0 {
	// 	logrus.Warn("Could find any node that satisfies current config")
	// 	return nil
	// }

	// var wg sync.WaitGroup
	// wg.Add(len(conns))
	// for _, conn := range conns {
	// 	conn := conn
	// 	go func() {
	// 		client, err := sftp.NewClient(conn.client)
	// 	}()
	// }

	// wg.Wait()
	// if failed != 0 {
	// 	return NewErrf(ErrCmdExec,
	// 		"Failed to execute command on %d targets", failed)
	// }

	return nil
}

func (cm *CmdMan) Replicate(node, remotePath string) error {
	return nil
}
