//go:build windows

package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func reboot() {
	dubLog("Main", logLevelInfo, color.HiGreenString, "Rebooting...")
	if err := exec.Command("cmd", "/C", "shutdown", "/r").Run(); err != nil {
		dubLog("Main", logLevelError, color.HiRedString, "Failed to initiate reboot:", err)
	}
}
