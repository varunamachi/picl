package mon

import (
	_ "embed"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/xcutr"
)

var (
	ErrExecutablePath = errors.New("mon.build.execPath")
)

//go:embed run.sh
var script []byte

func Build(fxRootPath, goArch string) (string, error) {

	// go build -ldflags "-s -w" -race -o "$root/_local/bin/fx"

	cmdDir := filepath.Join(fxRootPath, "cmd", "fx")
	output := filepath.Join(fxRootPath, "_local", "bin", goArch, "fx")

	cmd := exec.Command(
		"go", "build",
		"-ldflags", "-s -w",
		"-o", output,
		cmdDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOARCH="+goArch)
	if err := cmd.Run(); err != nil {
		const msg = "failed to run go build"
		logrus.WithError(err).Error(msg)
		return "", cfx.Errf(err, msg)
	}

	return output, nil
}

func InstallAgent(cmdMan *xcutr.CmdMan, exePath string) error {

	err := cmdMan.Exec("mkdir -p /opt/bin", &xcutr.ExecOpts{
		WithSudo: true,
	})
	if err != nil {
		return cfx.Errf(err, "failed to create destination directory")
	}

	// err = cmdMan.Exec("killall fx", &xcutr.ExecOpts{})

	err = cmdMan.Push(exePath, "/opt/bin/fx", &xcutr.CopyOpts{
		ExecOpts: xcutr.ExecOpts{
			WithSudo: true,
		},
		DupFilePolicy: xcutr.Replace,
	})
	if err != nil {
		return cfx.Errf(err, "failed to copy agent executable")
	}

	err = cmdMan.PushData(script, "/opt/bin/run.sh", &xcutr.CopyOpts{
		ExecOpts: xcutr.ExecOpts{
			WithSudo: true,
		},
		DupFilePolicy: xcutr.Replace,
	})
	if err != nil {
		return cfx.Errf(err, "failed to copy run script")
	}

	err = cmdMan.Exec("chmod -R 755 /opt/bin/*", &xcutr.ExecOpts{
		WithSudo: true,
	})
	if err != nil {
		return cfx.Errf(err, "failed to agent executable permission")
	}

	err = cmdMan.Exec("/opt/bin/run.sh", &xcutr.ExecOpts{
		WithSudo: true,
	})
	if err != nil {
		return cfx.Errf(err, "failed to start agent")
	}

	return nil
}

func BuildAndInstall(cmdMan *xcutr.CmdMan, fxRootPath, goArch string) error {
	logrus.Info("Building the executable for arch ", goArch)
	output, err := Build(fxRootPath, goArch)
	if err != nil {
		return err
	}

	logrus.Info("Deploying executable")
	if err := InstallAgent(cmdMan, output); err != nil {
		return err
	}

	return nil
}
