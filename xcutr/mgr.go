package xcutr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"github.com/varunamachi/picl/cmn"
	"golang.org/x/sync/errgroup"
)

var (
	ErrCmdExec      = errors.New("xcutr.cmd.failed")
	ErrInvalidNode  = errors.New("xcutr.node.invalid")
	ErrFileNotFound = errors.New("xcutr.file.notFound")
)

type StdIO struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

type Config struct {
	Name string `json:"name"`
	// SudoPass string         `json:"sudoPass"`
	Opts []*SshConnOpts `json:"opts"`
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
	WithSudo bool
}

type ExistingFilePolicy int

const (
	Ignore ExistingFilePolicy = iota
	Replace
)

type CopyOpts struct {
	ExecOpts
	DupFilePolicy ExistingFilePolicy
}

func NewCmdMan(config *Config, stdIO StdIO) (*CmdMan, error) {
	conns := make([]*SshConn, 0, len(config.Opts))
	connMap := make(map[string]*SshConn)
	for _, opts := range config.Opts {
		conn, err := NewConn(opts)
		if err != nil {
			logrus.Warn("Failed to connect to ", opts.Host)
			continue
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

func (cm *CmdMan) connList(opts *ExecOpts) []*SshConn {
	if len(opts.Included) == 0 && len(opts.Excluded) == 0 {
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
			var err error
			if opts.WithSudo {
				err = conn.ExecSudo(cmd, &cm.io)
			} else {
				err = conn.Exec(cmd, &cm.io)
			}
			if err != nil {
				failed++
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if failed != 0 {
		return NewErrf(ErrCmdExec,
			"failed to execute command on %d targets", failed)
	}
	return nil
}

func (cm *CmdMan) Pull(node, remotePath, localPath string) error {
	conn := cm.connMap[node]
	if conn == nil {
		logrus.WithField("nodeName", node).Error("invalid node name given")
		return NewErrf(ErrInvalidNode, "invalid node name given: %s", node)
	}

	sftpClient, err := sftp.NewClient(conn.client)
	if err != nil {
		const msg = "failed to create SFTP client"
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

func (cm *CmdMan) Push(localPath, remoteDest string, opts *CopyOpts) error {

	if !cmn.ExistsAsFile(localPath) {
		const msg = "Push: source file does not exist"
		logrus.WithFields(logrus.Fields{
			"localPath": localPath,
		}).Error(msg)
		return NewErrf(ErrFileNotFound, msg)
	}

	// local, err := os.Open(localPath)
	// if err != nil {
	// 	logrus.WithError(err).
	// 		WithFields(logrus.Fields{
	// 			"localPath": localPath,
	// 		}).
	// 		Error("Failed to open source file")
	// }
	// defer local.Close()

	data, err := os.ReadFile(localPath)
	if err != nil {
		logrus.WithError(err).
			WithFields(logrus.Fields{
				"localPath": localPath,
			}).
			Error("Failed to open source file")
		return cmn.Errf(err, "failed to open file to push")
	}

	return cm.PushData(data, remoteDest, opts)
}

func (cm *CmdMan) PushData(
	data []byte, remoteDest string, opts *CopyOpts) error {
	remotePath := remoteDest
	if opts.WithSudo {
		tempName := uuid.NewString()
		remotePath = "/tmp/" + tempName
	}

	conns := cm.connList(&opts.ExecOpts)
	if len(conns) == 0 {
		logrus.Warn("Could find any node that satisfies current config")
		return nil
	}

	// var wg sync.WaitGroup
	// wg.Add(len(conns))

	eg := errgroup.Group{}

	for _, conn := range conns {
		conn := conn
		eg.Go(func() error {

			// defer wg.Done()
			// local, err := os.Open(localPath)
			// if err != nil {
			// 	logrus.WithError(err).
			// 		WithFields(logrus.Fields{
			// 			"localPath": localPath,
			// 		}).
			// 		Error("Failed to open source file")
			// }
			// defer local.Close()

			reader := bytes.NewBuffer(data)
			err := copy(conn, remotePath, opts.DupFilePolicy, reader)
			if err != nil {
				return err
			}

			if opts.WithSudo {
				cmd := fmt.Sprintf("mv %s %s", remotePath, remoteDest)
				if err := conn.ExecSudo(cmd, &cm.io); err != nil {
					return cmn.Errf(err,
						"with sudo: failed to move temp file to destination")
				}

				cmd = fmt.Sprintf("rm -f %s", remotePath)
				if err := conn.Exec(cmd, &cm.io); err != nil {
					return cmn.Errf(err,
						"failed to remove temp file")
				}
			}
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil

}

func (cm *CmdMan) Replicate(node, remoteDest string, opts *CopyOpts) error {

	conn := cm.connMap[node]
	if conn == nil {
		logrus.WithField("nodeName", node).
			Error("Invalid source node name given")
		return NewErrf(ErrInvalidNode,
			"Invalid source node name given: %s", node)
	}

	client, err := sftp.NewClient(conn.client)
	if err != nil {
		const msg = "Failed to create SFTP client"
		logrus.WithField("nodeName", node).Error(msg)
		return NewErrf(ErrInvalidNode, msg)
	}

	if !remoteExists(client, remoteDest) {
		const msg = "Remote source file does not exist"
		logrus.WithError(err).
			WithFields(logrus.Fields{
				"node":       conn.Name(),
				"remotePath": remoteDest,
			}).
			Error(msg)
		return NewErrf(ErrInvalidNode, msg)
	}

	conns := cm.connList(&opts.ExecOpts)
	if len(conns) == 0 || (len(conns) == 1 && conns[0].Name() == node) {
		logrus.Warn("Could find any node that satisfies current config")
		return nil
	}

	remotePath := remoteDest
	if opts.WithSudo {
		tempName := uuid.NewString()
		remotePath = "/tmp/" + tempName
	}

	var wg sync.WaitGroup
	wg.Add(len(conns))
	failed := 0
	for _, conn := range conns {
		conn := conn
		if conn.Name() == node {
			// Dont do anything for source itself
			wg.Done()
			continue
		}
		go func() {
			defer wg.Done()

			source, err := client.Open(remoteDest)
			if err != nil {
				logrus.WithError(err).
					WithFields(logrus.Fields{
						"node":       conn.Name(),
						"remotePath": remoteDest,
					}).
					Error("Failed to open remote source file")
			}
			defer source.Close()

			err = copy(conn, remotePath, opts.DupFilePolicy, source)
			if err != nil {
				failed++
				return
			}

			if opts.WithSudo {

				parent := filepath.Dir(remoteDest)
				cmd := fmt.Sprintf("mkdir -p %s", parent)
				if err := conn.ExecSudo(cmd, &cm.io); err != nil {
					failed++
					return
				}

				cmd = fmt.Sprintf("mv %s %s", remotePath, remoteDest)
				if err := conn.ExecSudo(cmd, &cm.io); err != nil {
					failed++
					return
				}

				cmd = fmt.Sprintf("rm -f %s", remotePath)
				if err := conn.Exec(cmd, &cm.io); err != nil {
					failed++
					return
				}
			}

		}()
	}

	wg.Wait()
	if failed != 0 {
		logrus.WithField("failedCount", failed).Error("Finished with errors")
		return NewErrf(ErrCmdExec,
			"Failed to perform replication on %d targets", failed)
	}

	return nil
}

func (cm *CmdMan) Remove(remotePath string, opts *ExecOpts) error {

	conns := cm.connList(opts)
	if len(conns) == 0 {
		logrus.Warn("Could find any node that satisfies current config")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(conns))
	failed := 0
	for _, conn := range conns {
		conn := conn
		go func() {
			defer wg.Done()
			client, err := sftp.NewClient(conn.client)
			if err != nil {
				logrus.WithError(err).
					WithField("node", conn.Name()).
					Error("failed to create SFTP client")
				failed++
			}

			if remoteExists(client, remotePath) {
				if err = client.Remove(remotePath); err != nil {
					const msg = "failed to remove remote file"
					logrus.WithError(err).
						WithFields(logrus.Fields{
							"node":       conn.Name(),
							"remotePath": remotePath,
						}).
						Error(msg)
					failed++
					return
				}
			}
		}()
	}

	wg.Wait()
	if failed != 0 {
		return NewErrf(ErrCmdExec,
			"failed to execute remove command on %d targets", failed)
	}
	return nil
}

func remoteExists(client *sftp.Client, remote string) bool {
	_, err := client.Stat(remote)
	return !os.IsNotExist(err)
}

func copy(
	conn *SshConn,
	remotePath string,
	dupPolicy ExistingFilePolicy,
	source io.Reader) error {
	client, err := sftp.NewClient(conn.client)
	if err != nil {
		const msg = "Failed to create SFTP client"
		logrus.WithError(err).
			WithField("node", conn.Name()).
			Error(msg)
		return NewErrf(err, msg)
	}

	if remoteExists(client, remotePath) {
		if dupPolicy == Ignore {
			return nil
		}
		if dupPolicy == Replace {
			if err = client.Remove(remotePath); err != nil {
				const msg = "Failed to remove remote file"
				logrus.WithError(err).
					WithFields(logrus.Fields{
						"node":       conn.Name(),
						"remotePath": remotePath,
					}).
					Error(msg)
				return NewErrf(err, msg)
			}
		}
	}

	parent := filepath.Dir(remotePath)
	if err = client.MkdirAll(parent); err != nil {
		const msg = "Failed to create remote directory structure"
		logrus.WithError(err).
			WithFields(logrus.Fields{
				"node":          conn.Name(),
				"remoteDirPath": parent,
			}).
			Error(msg)
		return NewErrf(err, msg)
	}

	remote, err := client.Create(remotePath)
	if err != nil {
		const msg = "Failed to create remote file"
		logrus.WithError(err).
			WithFields(logrus.Fields{
				"node":       conn.Name(),
				"remotePath": remotePath,
			}).
			Error(msg)
		return NewErrf(err, msg)
	}
	defer remote.Close()

	if _, err = io.Copy(remote, source); err != nil {
		const msg = "Failed to copy content to remote file"
		logrus.WithError(err).
			WithFields(logrus.Fields{
				"node":       conn.Name(),
				"remotePath": remotePath,
			}).
			Error(msg)
		return NewErrf(err, msg)
	}

	return nil
}
