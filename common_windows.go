// +build windows

package main

import (
	"os/exec"

	"github.com/fatih/color"
)

func reboot() {
	if err := exec.Command("cmd", "/C", "shutdown", "/r").Run(); err != nil {
		dubLog("Main", color.HiRedString, "Failed to initiate shutdown:", err)
	} else {
		dubLog("Main", color.HiGreenString, "Rebooting...")
	}
}
