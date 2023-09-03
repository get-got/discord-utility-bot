package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

func isBotAdmin(m *discordgo.Message) bool {
	// No Admins
	if len(config.DiscordAdmins) == 0 {
		return true
	}
	// Bypass Check
	if isServerPermitted(m.GuildID) {
		serverConfig := getPermittedServerConfig(m.GuildID)
		if serverConfig.UnlockCommands {
			return true
		}
	}
	//
	return stringInSlice(m.Author.ID, config.DiscordAdmins)
}

func getUserIdentifier(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

// For command case-insensitivity
func messageToLower(message *discordgo.Message) *discordgo.Message {
	newMessage := *message
	newMessage.Content = strings.ToLower(newMessage.Content)
	return &newMessage
}

func hasPerms(channelID string, permission int64) bool {
	if !config.DiscordCheckPerms {
		return true
	}

	sourceChannel, err := bot.State.Channel(channelID)
	if sourceChannel != nil && err == nil {
		switch sourceChannel.Type {
		case discordgo.ChannelTypeDM:
			return true
		case discordgo.ChannelTypeGroupDM:
			return true
		case discordgo.ChannelTypeGuildText:
			perms, err := bot.UserChannelPermissions(user.ID, channelID)
			if err == nil {
				return perms&permission == permission
			}
			dubLog("Discord", logLevelError, color.HiRedString, "Failed to check permissions (%d) for %s:\t%s", permission, channelID, err)
		}
	}
	return false
}

func dataKeyReplacement(input string) string {
	//TODO: Case-insensitive key replacement. -- If no streamlined way to do it, convert to lower to find substring location but replace normally
	if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
		timeNow := time.Now()
		keys := [][]string{
			{"{{goVersion}}", runtime.Version()},
			{"{{dgVersion}}", discordgo.VERSION},
			{"{{dubVersion}}", projectVersion},
			{"{{apiVersion}}", discordgo.APIVersion},
			{"{{numServers}}", fmt.Sprint(len(bot.State.Guilds))},
			{"{{numAdmins}}", fmt.Sprint(len(config.DiscordAdmins))},
			{"{{timeNowShort}}", timeNow.Format("3:04pm")},
			{"{{timeNowShortTZ}}", timeNow.Format("3:04pm MST")},
			{"{{timeNowMid}}", timeNow.Format("3:04pm MST 1/2/2006")},
			{"{{timeNowLong}}", timeNow.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeNowShort24}}", timeNow.Format("15:04")},
			{"{{timeNowShortTZ24}}", timeNow.Format("15:04 MST")},
			{"{{timeNowMid24}}", timeNow.Format("15:04 MST 2/1/2006")},
			{"{{timeNowLong24}}", timeNow.Format("15:04:05 MST - 2 January, 2006")},
			{"{{uptime}}", durafmt.ParseShort(time.Since(timeLaunched)).String()},
		}
		for _, key := range keys {
			if strings.Contains(input, key[0]) {
				input = strings.ReplaceAll(input, key[0], key[1])
			}
		}
	}
	return input
}

//#region Embeds

func getEmbedColor(channelID string) int {
	var err error
	var color *string
	var channelInfo *discordgo.Channel

	// Use Defined Color
	if color != nil {
		// Defined as Role, fetch role color
		if *color == "role" || *color == "user" {
			botColor := bot.State.UserColor(user.ID, channelID)
			if botColor != 0 {
				return botColor
			}
			goto color_random
		}
		// Defined as Random, jump below (not preferred method but seems to work flawlessly)
		if *color == "random" || *color == "rand" {
			goto color_random
		}

		var colorString string = *color

		// Input is Hex
		colorString = strings.ReplaceAll(colorString, "#", "")
		if convertedHex, err := strconv.ParseUint(colorString, 16, 64); err == nil {
			return int(convertedHex)
		}

		// Input is Int
		if convertedInt, err := strconv.Atoi(colorString); err == nil {
			return convertedInt
		}

		// Definition is invalid since hasn't returned, so defaults to below...
	}

	// User color
	channelInfo, err = bot.State.Channel(channelID)
	if err == nil {
		if channelInfo.Type != discordgo.ChannelTypeDM && channelInfo.Type != discordgo.ChannelTypeGroupDM {
			if bot.State.UserColor(user.ID, channelID) != 0 {
				return bot.State.UserColor(user.ID, channelID)
			}
		}
	}

	// Random color
color_random:
	var randomColor string = randomcolor.GetRandomColorInHex()
	if convertedRandom, err := strconv.ParseUint(strings.ReplaceAll(randomColor, "#", ""), 16, 64); err == nil {
		return int(convertedRandom)
	}

	return 16777215 // white
}

// Shortcut function for quickly constructing a styled embed with Title & Description
func buildEmbed(channelID string, title string, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       getEmbedColor(channelID),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: projectIcon,
			Text:    fmt.Sprintf("%s v%s", projectName, projectVersion),
		},
	}
}

// Shortcut function for quickly replying a styled embed with Title & Description
func replyEmbed(m *discordgo.Message, title string, description string) (*discordgo.Message, error) {
	if m != nil {
		if hasPerms(m.ChannelID, discordgo.PermissionSendMessages) {
			return bot.ChannelMessageSendComplex(m.ChannelID,
				&discordgo.MessageSend{
					Content: m.Author.Mention(),
					Embed:   buildEmbed(m.ChannelID, title, description),
				},
			)
		}
		dubLog("Discord", logLevelError, color.HiRedString, fmtBotSendPerm, m.ChannelID)
	}
	return nil, nil
}

//#endregion
