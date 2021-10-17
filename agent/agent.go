package agent

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RunAgent(address string) error {
	server := echo.New()

	server.GET("/api/v0/sys/cur", handleSysInfo)
	server.GET("/", handleSysInfo)
	return server.Start(address)
}

func handleSysInfo(etx echo.Context) error {
	info, err := systemInfo(etx.Request().Context())
	if err != nil {
		return err
	}

	return etx.JSON(http.StatusOK, info)
}

func root(etx echo.Context) error {
	etx.String(http.StatusOK, "Hello!")
	return nil
}
