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
	dubLog("Main", color.HiRedString, "EXIT IN 15 SECONDS - Uptime was %s...", durafmt.Parse(time.Since(timeLaunched)).String())
	dubLog("Main", color.HiCyanString, "---------------------------------------------------------------------")
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

func dubLog(group string, colorFunc func(string, ...interface{}) string, line string, p ...interface{}) {
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
	log.Println(colorPrefix, colorFunc(line, p...))

	if bot != nil && botReady {
		for _, channelConfig := range config.OutputChannels {
			if channelConfig.OutputProgram {
				if channelConfig.Channel != "" {
					if !hasPerms(channelConfig.Channel, discordgo.PermissionSendMessages) {
						dubLog("Self", color.HiRedString, fmtBotSendPerm, channelConfig.Channel)
					} else {
						if _, err := bot.ChannelMessageSend(channelConfig.Channel, fmt.Sprintf("```%s | [%s] %s```", time.Now().Format(time.RFC3339), group, fmt.Sprintf(line, p...))); err != nil {
							dubLog("Self", color.HiRedString, "Failed to send message...\t%s", err)
						}
					}
				}
				if channelConfig.Channels != nil {
					/*for _, ch := range *channelConfig.Channels {

					}*/
				}
			}
		}
	}

	if false {
		// send to discord log channel(s) (group, line)
	}
}
