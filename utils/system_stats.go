package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"os/exec"
)

func GetSystemStats() (string, string, string) {
	v, _ := mem.VirtualMemory()
	memoryUsage := fmt.Sprintf("%.2f%%", v.UsedPercent)

	c, _ := cpu.Percent(0, false)
	cpuUsage := fmt.Sprintf("%.2f%%", c[0])

	cmd := exec.Command("ping", "-c", "1", "8.8.8.8")
	if err := cmd.Run(); err != nil {
		return memoryUsage, cpuUsage, "Disconnected"
	}
	return memoryUsage, cpuUsage, "Connected"
}
