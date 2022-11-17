package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// Multiple use messages to save space and make cleaner.
//TODO: Implement this for more?
const (
	cmderrLackingLocalAdminPerms = "You do not have permission to use this command.\n" +
		"\nTo use this command you must:" +
		"\n• Be set as a bot administrator (in the settings)" +
		"\n• Own this Discord Server" +
		"\n• Have Server Administrator Permissions"
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. Your User ID must be set as a bot administrator in the settings file."
	cmderrChannelNotRegistered = "Specified channel is not registered in the bot settings."
	cmderrHistoryCancelled     = "History cataloging was cancelled."
	fmtBotSendPerm             = "Bot does not have permission to send messages in %s"
)

func handleCommands() *exrouter.Route {
	router := exrouter.New()

	//#region Utility Commands

	router.On("ping", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:ping]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			//if isCommandableChannel(ctx.Msg) {
			beforePong := time.Now()
			pong, err := ctx.Reply("Pong!")
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Error sending pong message:\t%s", err))
			} else {
				afterPong := time.Now()
				latency := bot.HeartbeatLatency().Milliseconds()
				roundtrip := afterPong.Sub(beforePong).Milliseconds()
				mention := ctx.Msg.Author.Mention()
				content := fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
					latency,
					roundtrip,
				)
				if pong != nil {
					bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
						ID:      pong.ID,
						Channel: pong.ChannelID,
						Content: &mention,
						Embed:   buildEmbed(ctx.Msg.ChannelID, "Command — Ping", content),
					})
				}
				// Log
				log.Println(logPrefixHere, color.HiCyanString("%s pinged bot - Latency: %dms, Roundtrip: %dms",
					getUserIdentifier(*ctx.Msg.Author),
					latency,
					roundtrip),
				)
			}
			//}
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot.")

	router.On("help", func(ctx *exrouter.Context) {
	}).Cat("Utility").Alias("h").Desc("Help.")

	//#endregion

	//#region Admin Commands

	router.On("exit", func(ctx *exrouter.Context) {
		/*logPrefixHere := color.CyanString("[dgrouter:exit]")
		if isCommandableChannel(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting...")
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s (bot admin) requested exit, goodbye...", getUserIdentifier(*ctx.Msg.Author)))
				properExit()
			} else {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Exit", cmderrLackingBotAdminPerms)
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s tried to exit but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}*/
	}).Cat("Admin").Alias("reload", "kill").Desc("Exits this program.")

	router.On("reboot", func(ctx *exrouter.Context) {
	}).Cat("Admin").Alias("restart", "shutdown").Desc("Restarts the server.")

	//#endregion

	//#region Discord

	router.On("emoji", func(ctx *exrouter.Context) {
	}).Cat("Discord").Alias("e").Desc("Emoji lookup.")

	router.On("emojis", func(ctx *exrouter.Context) {
	}).Cat("Discord").Desc("Dump server emojis.")

	//#endregion

	//#region Spotify API

	router.On("sg", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:spotifygenres]")
		if spotifyClient != nil {
			msg, page, err := spotifyClient.FeaturedPlaylists(spotifyContext)
			if err != nil {
				log.Println(logPrefixHere, "Couldn't get featured playlists: %v", err)
			} else {
				log.Println(logPrefixHere, msg)
				for _, playlist := range page.Playlists {
					log.Println(logPrefixHere, playlist.Name)
				}
			}
		} else {
			log.Println("Bot is not connected to Spotify...")
		}
	}).Cat("Spotify").Alias("spotifygenres", "spotgen").Desc("Spotify genre lookup by url.")

	//#endregion

	//#region Games

	router.On("minecraft", func(ctx *exrouter.Context) {
	}).Cat("Games").Desc("Minecraft Server Status.")

	router.On("valheim", func(ctx *exrouter.Context) {
	}).Cat("Games").Desc("Valheim Server Status.")

	//#endregion

	//#region Misc...

	router.On("plex", func(ctx *exrouter.Context) {
	}).Cat("Misc").Desc("Plex Status.")

	router.On("webm", func(ctx *exrouter.Context) {
	}).Cat("Misc").Alias("mp4").Desc("WEBM to MP4 Conversion.")

	//#endregion

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(bot, ".", bot.State.User.ID, m.Message)
	})

	return router
}
