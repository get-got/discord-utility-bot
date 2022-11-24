//go:build windows

package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func reboot() {
	dubLog("Main", color.HiGreenString, "Rebooting...")
	if err := exec.Command("cmd", "/C", "shutdown", "/r").Run(); err != nil {
		dubLog("Main", color.HiRedString, "Failed to initiate reboot:", err)
	}
}
