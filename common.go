package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

func uptime() time.Duration {
	return time.Since(timeLaunched)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	dubLog("Main", logLevelInfo, color.HiRedString, "EXIT IN 15 SECONDS - Uptime was %s...", durafmt.Parse(time.Since(timeLaunched)).String())
	dubLog("Main", logLevelInfo, color.HiCyanString, "---------------------------------------------------------------------")
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

/* logging system:


implement log leveling


*/

const (
	logLevelOff = iota
	logLevelFatal
	logLevelError
	logLevelWarning
	logLevelInfo
	logLevelDebug
	logLevelVerbose
	logLevelAll
)

func dubLog(group string, logLevel int, colorFunc func(string, ...interface{}) string, line string, p ...interface{}) {
	colorPrefix := group
	switch strings.ToLower(group) {

	case "main":
		colorPrefix = color.CyanString("[~]")
		break
	case "debug":
		colorPrefix = color.HiYellowString("<DEBUG>")
		break
	case "test":
		colorPrefix = color.HiYellowString("<TEST>")
		break
	case "info":
		colorPrefix = color.CyanString("[Info]")
		break
	case "version":
		colorPrefix = color.HiMagentaString("[Version]")
		break

	case "settings":
		colorPrefix = color.GreenString("[Settings]")
		break

	case "setup":
		colorPrefix = color.HiGreenString("[Setup]")
		break

	case "discord":
		colorPrefix = color.HiBlueString("[Discord]")
		break

	case "spotify":
		colorPrefix = color.HiGreenString("[Spotify]")
		break
	}

	if logLevel <= config.LogLevel {
		//TODO: trace file+line
		log.Println(colorPrefix, colorFunc(line, p...))
	}

	if bot != nil && botReady {
		for _, channelConfig := range config.OutputChannels {
			if channelConfig.OutputProgram {
				outputToChannel := func(channel string) {
					if channel != "" {
						if !hasPerms(channel, discordgo.PermissionSendMessages) {
							dubLog("Log", logLevelError, color.HiRedString, fmtBotSendPerm, channel)
						} else {
							if _, err := bot.ChannelMessageSend(channel, fmt.Sprintf("```%s | [%s] %s```", time.Now().Format(time.RFC3339), group, fmt.Sprintf(line, p...))); err != nil {
								dubLog("Log", logLevelError, color.HiRedString, "Failed to send message...\t%s", err)
							}
						}
					}
				}
				outputToChannel(channelConfig.Channel)
				if channelConfig.Channels != nil {
					for _, ch := range *channelConfig.Channels {
						outputToChannel(ch)
					}
				}
			}
		}
	}
}
