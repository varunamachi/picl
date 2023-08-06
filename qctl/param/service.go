package param

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/varunamachi/libx/httpx"
)

var (
	ErrParamNotFound     = errors.New("qctl.paramNotFound")
	ErrInvalidParamValue = errors.New("qctl.invalidParamValue")
)

type Operator interface {
	Get() (any, error)
	Set(value any) error
	Reset() error
	Default() (any, error)
}

type Service struct {
	operators map[string]Operator
}

func (s *Service) Endpoints(gtx context.Context) []*httpx.Endpoint {
	return []*httpx.Endpoint{}
}

func (s *Service) GetAllEp(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {
		return nil
	}

	return &httpx.Endpoint{
		Method:     echo.GET,
		Path:       "/qctl",
		Category:   "qtcl",
		Desc:       "Get all parameters",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}

}
