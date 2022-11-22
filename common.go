package main

import (
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

func uptime() time.Duration {
	return time.Since(timeLaunched)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	dubLog("Main", color.HiRedString, "EXIT IN 15 SECONDS - Uptime was %s...", durafmt.Parse(time.Since(timeLaunched)).String())
	dubLog("Main", color.HiCyanString, "---------------------------------------------------------------------")
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

/* logging system:
[module]



*/

func dubLog(group string, colorFunc func(string, ...interface{}) string, line string, p ...interface{}) {
	colorPrefix := group
	switch group {

	case "Main":
		colorPrefix = color.CyanString("[~]")
		break
	case "Debug":
		colorPrefix = color.HiYellowString("[Debug]")
		break
	case "Info":
		colorPrefix = color.CyanString("[Info]")
		break
	case "Version":
		colorPrefix = color.HiMagentaString("[Version]")
		break

	case "Settings":
		colorPrefix = color.GreenString("[Settings]")
		break

	case "Setup":
		colorPrefix = color.HiGreenString("[Setup]")
		break

	case "Discord":
		colorPrefix = color.HiBlueString("[Discord]")
		break

	case "Spotify":
		colorPrefix = color.HiGreenString("[Spotify]")
		break
	}
	log.Println(colorPrefix, colorFunc(line, p...))

	if false {
		// send to discord log channel(s) (group, line)
	}
}
