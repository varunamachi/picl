package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/varunamachi/picl/cmn"
)

var (
	ErrNotFound            = errors.New("client.http.notFound")
	ErrUnauthorized        = errors.New("client.http.unauthorized")
	ErrForbidden           = errors.New("client.http.forbidden")
	ErrInternalServerError = errors.New("client.http.internalServerError")
	ErrOtherStatus         = errors.New("client.http.otherStatus")

	ErrInvalidResponse = errors.New("client.http.invalidResponse")
	ErrClientError     = errors.New("client.http.clientError")
)

type ApiResult struct {
	resp   *http.Response
	err    error
	target string
	code   int
}

func newApiResult(req *http.Request, resp *http.Response) *ApiResult {

	target := "[" + req.Method + " " + req.URL.Path + "]"
	res := &ApiResult{
		resp:   resp,
		target: target,
		code:   resp.StatusCode,
	}

	var err *cmn.Error

	switch resp.StatusCode {
	case http.StatusNotFound:
		err = cmn.Errf(ErrNotFound, "Not found: %s", target)
	case http.StatusUnauthorized:
		err = cmn.Errf(ErrUnauthorized, "Unauthorized: %s", target)
	case http.StatusForbidden:
		err = cmn.Errf(ErrUnauthorized, "Forbidden: %s", target)
	case http.StatusInternalServerError:
		err = cmn.Errf(ErrUnauthorized, "Internal Server Error: %s", target)
	default:
		if resp.StatusCode > 400 {
			err = cmn.Errf(
				ErrOtherStatus, "HTTP Error: %d - %s", resp.StatusCode, target)
		}
	}
	if err != nil {
		logrus.WithError(err).Error(err.String())
		res.err = err
	}
	return res
}

func newErrorResult(req *http.Request, err error, msg string) *ApiResult {
	target := ""
	if req != nil {
		target = req.Method + " " + req.URL.Path
		msg = msg + " - [" + target + "]"
	}

	return &ApiResult{
		err:    cmn.Errf(err, msg),
		target: target,
	}
}

func (ar *ApiResult) LoadClose(out interface{}) error {
	defer func() {
		if ar.resp != nil && ar.resp.Body != nil {
			ar.resp.Body.Close()
		}
	}()

	if ar.Error() != nil {
		return ar.Error()
	}

	ar.err = json.NewDecoder(ar.resp.Body).Decode(out)
	return ar.err
}

func (ar *ApiResult) Error() error {
	if ar.err != nil {
		return ar.err
	}

	if ar.resp == nil || ar.resp.Body == nil {
		ar.err = cmn.Errf(ErrInvalidResponse, "No valid http response received")
	}
	return ar.err
}

type Client struct {
	*http.Client
	address     string
	contextRoot string
	token       string
}

func DefaultTransport() *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 1 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 1 * time.Second,
	}
}

func New(address, contextRoot string) *Client {
	return &Client{
		address:     address,
		contextRoot: contextRoot,
		Client: &http.Client{
			Timeout:   time.Second * 1,
			Transport: DefaultTransport(),
		},
	}
}

func NewCustom(
	address, contextRoot string,
	transport *http.Transport,
	timeout time.Duration) *Client {
	return &Client{
		address:     address,
		contextRoot: contextRoot,
		Client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

func (client *Client) createUrl(args ...string) string {
	var buffer bytes.Buffer
	if _, err := buffer.WriteString(client.address); err != nil {
		logrus.Fatal(err)
	}
	// if _, err := buffer.WriteString("/"); err != nil {
	// 	logrus.Fatal(err)
	// }
	if _, err := buffer.WriteString(client.contextRoot); err != nil {
		logrus.Fatal(err)
	}
	// if _, err := buffer.WriteString("/"); err != nil {
	// 	logrus.Fatal(err)
	// }
	for i := 0; i < len(args); i++ {
		if _, err := buffer.WriteString(args[i]); err != nil {
			logrus.Fatal(err)
		}
		if i < len(args)-1 {
			if _, err := buffer.WriteString("/"); err != nil {
				logrus.Fatal(err)
			}
		}
	}
	return buffer.String()
}

func (client *Client) putOrPost(
	gtx context.Context,
	method string,
	content interface{},
	urlArgs ...string) *ApiResult {

	url := client.createUrl(urlArgs...)
	data, err := json.Marshal(content)
	if err != nil {
		return newErrorResult(nil, err, "Failed to marshal data")
	}

	req, err := http.NewRequestWithContext(
		gtx, method, url, bytes.NewBuffer(data))

	// We assume JSON
	req.Header.Set("Content-Type", "application/json")
	if client.token != "" {
		authHeader := fmt.Sprintf("Bearer %s", client.token)
		req.Header.Add("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return newErrorResult(req, err, "Failed to perform http request")
	}
	return newApiResult(req, resp)
}

func (client *Client) Get(gtx context.Context, urlArgs ...string) *ApiResult {

	apiURL := client.createUrl(urlArgs...)
	req, err := http.NewRequestWithContext(gtx, "GET", apiURL, nil)
	if err != nil {
		newErrorResult(req, err, "Failed to create http request")
	}

	if client.token != "" {
		authHeader := fmt.Sprintf("Bearer %s", client.token)
		req.Header.Add("Authorization", authHeader)
	}

	resp, err := client.Do(req)

	if err != nil {
		return newErrorResult(req, err, "Failed to perform http request")
	}

	return newApiResult(req, resp)
}

func (client *Client) Post(
	gtx context.Context,
	content interface{},
	urlArgs ...string) *ApiResult {
	return client.putOrPost(gtx, echo.POST, content, urlArgs...)
}

//Put - performs a put request
func (client *Client) Put(
	gtx context.Context,
	content interface{},
	urlArgs ...string) *ApiResult {
	return client.putOrPost(gtx, echo.PUT, content, urlArgs...)
}

//Delete - performs a delete request
func (client *Client) Delete(
	gtx context.Context,
	urlArgs ...string) *ApiResult {
	apiURL := client.createUrl(urlArgs...)
	req, err := http.NewRequestWithContext(gtx, echo.DELETE, apiURL, nil)
	if err != nil {
		newErrorResult(req, err, "Failed to create http request")
	}

	if client.token != "" {
		authHeader := fmt.Sprintf("Bearer %s", client.token)
		req.Header.Add("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return newErrorResult(req, err, "Failed to perform http request")
	}

	return newApiResult(req, resp)
}

func (client *Client) UserLogin(
	gtx context.Context, userID, password string) error {
	authData := &AuthData{
		AuthType: "user",
		Data: map[string]string{
			userID:   userID,
			password: password,
		},
	}
	return client.Login(gtx, authData)
}

type AuthData struct {
	AuthType string      `json:"authType"`
	Data     interface{} `json:"data"`
}

func (client *Client) Login(gtx context.Context, authData *AuthData) error {

	if authData == nil {
		return nil
	}

	loginResult := struct {
		Token string `json:"token"`
	}{}

	rr := client.Post(gtx, authData, "login")
	if err := rr.LoadClose(&loginResult); err != nil {
		return err
	}
	client.token = loginResult.Token
	return nil
}
