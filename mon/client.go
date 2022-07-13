package mon

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/varunamachi/libx/httpx"
	"github.com/varunamachi/libx/iox"
)

type loginResult struct {
	Token string `json:"token"`
}

func Login(
	gtx context.Context,
	client *httpx.Client,
	authData httpx.AuthData) error {

	if authData == nil {
		return nil
	}

	rr := client.Post(gtx, authData, "/api/v1/auth/user")
	lres := &loginResult{}
	if err := rr.LoadClose(&lres); err != nil {
		return err
	}
	client.SetToken(lres.Token)
	return nil
}

func CreateClient(ctx *cli.Context) (
	*httpx.Client, error) {
	host := ctx.String("host")
	ignCertErrs := ctx.Bool("ignore-cert-errors")
	timeOut := ctx.Int("timeout-secs")

	tp := httpx.DefaultTransport()
	if ignCertErrs {
		tp.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: ignCertErrs,
		}
	}
	client := httpx.NewCustom(host, "", tp, time.Duration(timeOut)*time.Second)

	userId := ctx.String("user-id")
	if userId == "" {
		return client, nil
	}
	password := ctx.String("password")
	if password == "" {
		password = iox.AskPassword(fmt.Sprintf("Password for '%s'", userId))
	}
	err := Login(ctx.Context, client, httpx.AuthData{
		"userId":   userId,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func WithClientFlags(
	withAuth bool,
	flags ...cli.Flag) []cli.Flag {
	flags = append(flags,
		&cli.StringFlag{
			Name: "host",
			Usage: "Full address of the host with URL scheme, host name/IP " +
				"and port",
			EnvVars: []string{
				"MK__CLIENT_REMOTE_HOST",
				"LIBX_CLIENT_REMOTE_HOST",
			},
			Required: true,
		},
		&cli.BoolFlag{
			Name: "ignore-cert-errors",
			Usage: "Ignore certificate errors while connecting to a HTTPS " +
				"service",
			Value: false,
			EnvVars: []string{
				"MK__CLIENT_IGNORE_CERT_ERR",
				"LIBX_CLIENT_IGNORE_CERT_ERR",
			},
		},
		&cli.IntFlag{
			Name:  "timeout-secs",
			Usage: "Time out in seconds",
			Value: 20,
			EnvVars: []string{
				"MK__CLIENT_TIMEOUT_SECS",
				"LIBX_CLIENT_TIMEOUT_SECS",
			},
		},
	)
	if withAuth {
		flags = append(flags,
			&cli.StringFlag{
				Name:     "user-id",
				Usage:    "User present in the remote service",
				Required: false,
				EnvVars: []string{
					"MK__CLIENT_USER_ID",
					"LIBX_CLIENT_USER_ID",
				},
			},
			&cli.StringFlag{
				Name: "password",
				Usage: "Password for the user, only use for development " +
					"purposes",
				Required: false,
				Hidden:   true,
				EnvVars: []string{
					"MK__CLIENT_PASSWORD",
					"LIBX_CLIENT_PASSWORD",
				},
			},
		)
	}

	return flags
}
