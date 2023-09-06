package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/hashicorp/go-version"
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

//#region Requests

func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJSONwithHeaders(url string, target interface{}, headers map[string]string) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

//#endregion

//#region Github Release Checking

type githubReleaseApiObject struct {
	TagName string `json:"tag_name"`
}

func isLatestGithubRelease() bool {
	githubReleaseApiObject := new(githubReleaseApiObject)
	err := getJSON(projectReleaseApiURL, githubReleaseApiObject)
	if err != nil {
		dubLog("Version", logLevelInfo, color.RedString, "Error fetching current Release JSON: %s", err)
		return true
	}

	thisVersion, err := version.NewVersion(projectVersion)
	if err != nil {
		dubLog("Version", logLevelInfo, color.RedString, "Error parsing current version: %s", err)
		return true
	}

	latestVersion, err := version.NewVersion(githubReleaseApiObject.TagName)
	if err != nil {
		dubLog("Version", logLevelInfo, color.RedString, "Error parsing latest version: %s", err)
		return true
	}

	if latestVersion.GreaterThan(thisVersion) {
		return false
	}

	return true
}

func shortenTime(input string) string {
	input = strings.ReplaceAll(input, " nanoseconds", "ns")
	input = strings.ReplaceAll(input, " nanosecond", "ns")
	input = strings.ReplaceAll(input, " microseconds", "μs")
	input = strings.ReplaceAll(input, " microsecond", "μs")
	input = strings.ReplaceAll(input, " milliseconds", "ms")
	input = strings.ReplaceAll(input, " millisecond", "ms")
	input = strings.ReplaceAll(input, " seconds", "s")
	input = strings.ReplaceAll(input, " second", "s")
	input = strings.ReplaceAll(input, " minutes", "m")
	input = strings.ReplaceAll(input, " minute", "m")
	input = strings.ReplaceAll(input, " hours", "h")
	input = strings.ReplaceAll(input, " hour", "h")
	input = strings.ReplaceAll(input, " days", "d")
	input = strings.ReplaceAll(input, " day", "d")
	input = strings.ReplaceAll(input, " weeks", "w")
	input = strings.ReplaceAll(input, " week", "w")
	input = strings.ReplaceAll(input, " months", "mo")
	input = strings.ReplaceAll(input, " month", "mo")
	return input
}

//#endregion

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
