//go:build !windows

package main

import (
	"fmt"
	"syscall"

	"github.com/fatih/color"
	lsysinfo "github.com/zcalusic/sysinfo"
)

func reboot() {
	dubLog("Main", logLevelInfo, color.HiGreenString, "Rebooting...")
	syscall.Sync()
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
		dubLog("Main", logLevelError, color.HiRedString, "Failed to initiate reboot:", err)
	}
}

func getPlatformKeys() [][]string {
	var sys lsysinfo.SysInfo
	sys.GetSysInfo()
	return [][]string{
		{"{{lsysCpuModel}}", sys.CPU.Model},
		{"{{lsysCpuSpeed}}", fmt.Sprintf("%0.1f GHz", sys.CPU.Speed/1000)},
		{"{{lsysCpuCores}}", fmt.Sprint(sys.CPU.Cores)},
		{"{{lsysCpuThreads}}", fmt.Sprint(sys.CPU.Threads)},
	}
}
