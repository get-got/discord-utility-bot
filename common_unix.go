//go:build !windows

package main

import (
	"syscall"

	"github.com/fatih/color"
)

func reboot() {
	dubLog("Main", logLevelInfo, color.HiGreenString, "Rebooting...")
	syscall.Sync()
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART); err != nil {
		dubLog("Main", logLevelError, color.HiRedString, "Failed to initiate reboot:", err)
	}
}
