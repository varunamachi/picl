package mon

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

func RunAgent(address string) error {
	server := echo.New()
	server.GET("/api/v0/cur", handleSysInfo)
	server.GET("/api/v0/host", hostInfo)
	server.GET("/", func(etx echo.Context) error {
		return etx.String(http.StatusOK, "42")
	})
	return server.Start(address)
}

func handleSysInfo(etx echo.Context) error {
	info, err := systemInfo(etx.Request().Context())
	if err != nil {
		return err
	}

	return etx.JSON(http.StatusOK, info)
}

func hostInfo(etx echo.Context) error {
	h, err := host.Info()
	if err != nil {
		log.Error().Err(err).Msg("failed to get host inforamtion")
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "failed to get host information",
			Internal: err,
		}
	}

	days := h.Uptime / 60 / 60 / 24
	hours := (h.Uptime / 60 / 60) % 24
	minute := (h.Uptime / 60) % 60
	seconds := h.Uptime % 60

	humanUptime := fmt.Sprintf("%d Days, %d Hours, %d Minutes, %d Seconds",
		days, hours, minute, seconds)

	return etx.JSON(http.StatusOK, map[string]interface{}{
		"hostname":    h.Hostname,
		"hostId":      h.HostID,
		"kernalArch":  h.KernelArch,
		"uptime":      h.Uptime,
		"humanUptime": humanUptime,
	})
}
