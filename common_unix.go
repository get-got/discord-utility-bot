// +build !windows

package main

func reboot() {
	dubLog(logPrefixHere, color.HiGreenString, "Rebooting...")
	syscall.Sync()
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}
