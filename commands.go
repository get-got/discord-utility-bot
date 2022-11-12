package main

import (
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
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
)

func handleCommands() *exrouter.Route {
	router := exrouter.New()

	//#region Utility Commands

	router.On("ping", func(ctx *exrouter.Context) {
		/*logPrefixHere := color.CyanString("[dgrouter:ping]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			if isCommandableChannel(ctx.Msg) {
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
						if selfbot {
							bot.ChannelMessageEdit(pong.ChannelID, pong.ID, fmt.Sprintf("%s **Command — Ping**\n\n%s", mention, content))
						} else {
							bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
								ID:      pong.ID,
								Channel: pong.ChannelID,
								Content: &mention,
								Embed:   buildEmbed(ctx.Msg.ChannelID, "Command — Ping", content),
							})
						}
					}
					// Log
					log.Println(logPrefixHere, color.HiCyanString("%s pinged bot - Latency: %dms, Roundtrip: %dms",
						getUserIdentifier(*ctx.Msg.Author),
						latency,
						roundtrip),
					)
				}
			}
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}*/
	}).Cat("Utility").Alias("test").Desc("Pings the bot")

	router.On("help", func(ctx *exrouter.Context) {
		/*logPrefixHere := color.CyanString("[dgrouter:help]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			if isGlobalCommandAllowed(ctx.Msg) {
				text := ""
				for _, cmd := range router.Routes {
					if cmd.Category != "Admin" || isBotAdmin(ctx.Msg) {
						text += fmt.Sprintf("• \"%s\" : %s",
							cmd.Name,
							cmd.Description,
						)
						if len(cmd.Aliases) > 0 {
							text += fmt.Sprintf("\n— Aliases: \"%s\"", strings.Join(cmd.Aliases, "\", \""))
						}
						text += "\n\n"
					}
				}
				_, err := replyEmbed(ctx.Msg, "Command — Help", fmt.Sprintf("Use commands as ``\"%s<command> <arguments?>\"``\n```%s```\n%s", config.CommandPrefix, text, projectRepoURL))
				// Failed to send
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s asked for help", getUserIdentifier(*ctx.Msg.Author)))
			}
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}*/
	}).Cat("Utility").Alias("commands").Desc("Outputs this help menu")

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
	}).Alias("reload", "kill").Cat("Admin").Desc("Kills the bot")

	//#endregion

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, ".", bot.State.User.ID, messageToLower(m.Message))
	})

	return router
}
