package mon

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SysInfo struct {
	CPUTemp     float64 `json:"cpuTemp"`
	CPUUsagePct float64 `json:"cpuUsage"`
	MemUsagePct float64 `json:"memUsage"`
}

func systemInfo(gtx context.Context) (*SysInfo, error) {
	temp, err := os.ReadFile(
		"/sys/class/thermal/thermal_zone0/temp")

	info := SysInfo{}
	if err != nil {
		return &info, err
	}

	strTemp := strings.TrimSpace(string(temp))
	info.CPUTemp, err = strconv.ParseFloat(strTemp, 64)

	if err != nil {
		return &info, err
	}

	vmem, err := mem.VirtualMemoryWithContext(gtx)
	if err != nil {
		return &info, err
	}
	info.MemUsagePct = vmem.UsedPercent

	usage, err := cpu.PercentWithContext(gtx, 0, false)
	if err != nil {
		return &info, err
	}
	info.CPUUsagePct = usage[0]

	return &info, nil
}
