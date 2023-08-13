package param

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/varunamachi/libx/httpx"
)

type Service struct {
	operators map[string]Operator
	pg        []*ControlGroup
}

func (s *Service) Endpoints(gtx context.Context) []*httpx.Endpoint {
	return []*httpx.Endpoint{}
}

func (s *Service) GetParamList(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {
		return etx.JSON(http.StatusOK, s.pg)
	}

	return &httpx.Endpoint{
		Method:     echo.GET,
		Path:       "/qctl",
		Category:   "qtcl",
		Desc:       "Get parameter list",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}

func (s *Service) GetValues(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {
		vals := map[string]any{}

		for id, op := range s.operators {
			val, err := op.Get()
			if err != nil {
				return err
			}
			vals[id] = val
		}
		return httpx.SendJSON(etx, vals)
	}

	return &httpx.Endpoint{
		Method:     echo.GET,
		Path:       "/qctl/value",
		Category:   "qtcl",
		Desc:       "Get all parameter values",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}

func (s *Service) GetValue(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {

		return nil
	}

	return &httpx.Endpoint{
		Method:     echo.GET,
		Path:       "/qctl/value/:id",
		Category:   "qtcl",
		Desc:       "Get value for a parameter",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}

func (s *Service) GetDefaultValue(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {

		return nil
	}

	return &httpx.Endpoint{
		Method:     echo.GET,
		Path:       "/qctl/value/:id/default",
		Category:   "qtcl",
		Desc:       "Get value default for a parameter",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}

func (s *Service) SetValue(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {

		return nil
	}

	return &httpx.Endpoint{
		Method:     echo.PUT,
		Path:       "/qctl/value/:id/default",
		Category:   "qtcl",
		Desc:       "Set value for a parameter",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}

func (s *Service) SetDefault(gtx context.Context) *httpx.Endpoint {

	handler := func(etx echo.Context) error {

		return nil
	}

	return &httpx.Endpoint{
		Method:     echo.PUT,
		Path:       "/qctl/value/:id",
		Category:   "qtcl",
		Desc:       "Set default value for a parameter",
		Version:    "v1",
		Role:       "",
		Permission: "",
		Handler:    handler,
	}
}
