package mon

import (
	_ "embed"

	"github.com/varunamachi/clusterfox/cfx"
	"github.com/varunamachi/clusterfox/xcutr"
)

//go:embed run.sh
var script []byte

func BuildAndInstall(goArch string) error {
	return nil
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
