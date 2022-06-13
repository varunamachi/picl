package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/libx/iox"
	"github.com/varunamachi/picl/config"
)

func getInteractiveSetupCmd() *cli.Command {
	return &cli.Command{
		Name:        "setup",
		Description: "Set the picl configuration up interactively",
		Usage:       "Set the picl configuration up interactively",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Name of the picl config",
				Value:   "default",
				EnvVars: []string{"PICL_CONFIG"},
			},
			&cli.BoolFlag{
				Name:  "use-defaults",
				Usage: "Use default options where possible",
				Value: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			cfgName := ctx.String("config")
			useDefaults := ctx.Bool("use-defaults")

			if useDefaults {
				return config.CreateConfigWithDefaults(cfgName)
			}
			return config.CreateConfig(cfgName)
		},
	}
}

func getCopyIdCmd() *cli.Command {
	return &cli.Command{
		Name: "copy-id",
		Description: "Copy current machines public key to target " +
			"nodes (like ssh-copy-id)",
		Usage: "Copy current machines public key to target " +
			"nodes (like ssh-copy-id)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Name of the picl config",
				Value:   "default",
				EnvVars: []string{"PICL_CONFIG"},
			},
		},
		Action: func(ctx *cli.Context) error {
			provider, err := config.NewFromCli(ctx)
			if err != nil {
				return err
			}
			return config.CopySshId(provider)
		},
	}
}

func getEncryptCmd() *cli.Command {
	return &cli.Command{
		Name:        "encrypt-config",
		Description: "Encrypts configuration file identified by config name",
		Usage:       "Encrypts configuration file identified by config name",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "Name of the picl config",
				EnvVars:  []string{"PICL_CONFIG"},
				Required: true,
			},
			&cli.StringFlag{
				Name: "out",
				Usage: "Where to output the file, if not given " +
					"default location and file naming scheme is used",
				Required: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			cfg := ctx.String("config")
			out := ctx.String("out")
			if out == "" {
				out = filepath.Join(
					iox.MustGetUserHome(), ".picl", cfg+".config.json.enc")
			} else {
				out = filepath.Join(
					iox.MustGetUserHome(), ".picl", out+".config.json.enc")
			}
			cfgPath := filepath.Join(
				iox.MustGetUserHome(), ".picl", cfg+".config.json")

			if !iox.ExistsAsFile(cfgPath) {
				err := fmt.Errorf("could not find config for '%s'", cfg)
				log.Error().Err(err).Msg("")
				return err
			}

			var pw string
			for {
				pw := iox.AskPassword("Config Encryption Password")
				if pw != "" {
					break
				}
			}

			configFile, err := os.Open(cfgPath)
			if err != nil {
				const msg = "failed to open config file"
				log.Error().Err(err).Msg(msg)
				return errx.Errf(err, msg)
			}
			defer configFile.Close()

			return iox.EncryptToFile(configFile, out, pw)

		},
	}
}

func getDecryptCmd() *cli.Command {
	return &cli.Command{
		Name:        "decrypt-config",
		Description: "Dencrypts configuration file identified by config name",
		Usage:       "Dencrypts configuration file identified by config name",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Name of the picl config",
				Value:   "default",
				EnvVars: []string{"PICL_CONFIG"},
			},
			&cli.StringFlag{
				Name: "out",
				Usage: "Where to output the file, if not given " +
					"default location and file naming scheme is used",
				Required: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			cfg := ctx.String("config")
			out := ctx.String("out")
			if out == "" {
				out = filepath.Join(
					iox.MustGetUserHome(), ".picl", cfg+".config.json")
			} else {
				out = filepath.Join(
					iox.MustGetUserHome(), ".picl", out+".config.json")
			}
			cfgPath := filepath.Join(
				iox.MustGetUserHome(), ".picl", cfg+".config.json.enc")

			if !iox.ExistsAsFile(cfgPath) {
				err := fmt.Errorf("could not find config for '%s'", cfg)
				log.Error().Err(err).Msg("")
				return err
			}

			var pw string
			for {
				pw := iox.AskPassword("Config Encryption Password")
				if pw != "" {
					break
				}
			}

			outFile, err := os.Create(out)
			if err != nil {
				const msg = "failed create to output file"
				log.Error().Err(err).Msg(msg)
				return errx.Errf(err, msg)
			}
			defer outFile.Close()

			err = iox.DecryptFromFile(cfgPath, pw, outFile)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt file")
				return err
			}

			return nil
		},
	}
}
