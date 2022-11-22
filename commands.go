package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/zmb3/spotify/v2"
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
		logPrefixHere := "[commands:ping]"
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			//if isCommandableChannel(ctx.Msg) {
			beforePong := time.Now()
			pong, err := ctx.Reply("Pong!")
			if err != nil {
				dubLog(logPrefixHere, color.HiRedString, "Error sending pong message:\t%s", err)
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
				dubLog(logPrefixHere, color.HiCyanString, "%s pinged bot - Latency: %dms, Roundtrip: %dms",
					getUserIdentifier(*ctx.Msg.Author),
					latency,
					roundtrip)
			}
			//}
		} else {
			dubLog(logPrefixHere, color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID)
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot.")

	router.On("help", func(ctx *exrouter.Context) {

	}).Cat("Utility").Alias("h").Desc("Help.")

	//#endregion

	//#region Admin Commands

	router.On("exit", func(ctx *exrouter.Context) {

	}).Cat("Admin").Alias("reload", "kill").Desc("Exits this program.")

	router.On("reboot", func(ctx *exrouter.Context) {
		logPrefixHere := "[commands:reboot]"
		dubLog(logPrefixHere, color.HiGreenString, "Attempting to reboot system...")
		reboot()
		properExit()
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
		logPrefixHere := color.CyanString("[commands:spotifygenres]")
		if spotifyClient != nil {

			input := ctx.Args[1]
			input_type := ""
			if strings.Contains(input, "/artist/") {
				input_type = "artist"
			} else if strings.Contains(input, "/album/") {
				input_type = "album"
			} else if strings.Contains(input, "/track/") {
				input_type = "track"
			} else if strings.Contains(input, "/playlist/") {
				input_type = "playlist"
			}
			if input_type == "" {
				dubLog(logPrefixHere, color.HiRedString, "Input is not a valid format...")
			} else {
				cleanedInput := input
				if idx := strings.Index(cleanedInput, "?si="); idx != -1 {
					cleanedInput = cleanedInput[:idx]
				}
				blacklist := []string{
					"spotify:artist:",
					"spotify:album:",
					"spotify:track:",
					"spotify:playlist:",
					"https://open.spotify.com/artist/",
					"https://open.spotify.com/album/",
					"https://open.spotify.com/track/",
					"https://open.spotify.com/playlist/",
					"http://open.spotify.com/artist/",
					"http://open.spotify.com/album/",
					"http://open.spotify.com/track/",
					"http://open.spotify.com/playlist/",
				}
				for _, phrase := range blacklist {
					cleanedInput = strings.ReplaceAll(cleanedInput, phrase, "")
				}

				artist_name := ""
				artist_url := ""
				artist_image := ""
				var genres []string
				if input_type == "artist" {
					artist, err := spotifyClient.GetArtist(spotifyContext, spotify.ID(cleanedInput))
					if err != nil {
						dubLog(logPrefixHere, color.HiRedString, "Error fetching Spotify artist: %s", err)
					} else {
						artist_name = artist.Name
						artist_url = "https://open.spotify.com/artist/" + artist.ID.String()
						if len(artist.Images) > 0 {
							artist_image = artist.Images[0].URL
						}
						genres = artist.Genres
					}
				} else if input_type == "album" {
					album, err := spotifyClient.GetAlbum(spotifyContext, spotify.ID(cleanedInput))
					if err != nil {
						dubLog(logPrefixHere, color.HiRedString, "Error fetching Spotify album: %s", err)
					} else {
						artist, err := spotifyClient.GetArtist(spotifyContext, album.Artists[0].ID)
						if err != nil {
							dubLog(logPrefixHere, color.HiRedString, "Error fetching Spotify artist: %s", err)
						} else {
							artist_name = artist.Name
							artist_url = "https://open.spotify.com/artist/" + artist.ID.String()
							if len(artist.Images) > 0 {
								artist_image = artist.Images[0].URL
							}
							genres = artist.Genres
							if len(album.Genres) > 0 {
								genres = album.Genres
							}
						}
					}
				} else if input_type == "track" {
					track, err := spotifyClient.GetTrack(spotifyContext, spotify.ID(cleanedInput))
					if err != nil {
						dubLog(logPrefixHere, color.HiRedString, "Error fetching Spotify album: %s", err)
					} else {
						artist, err := spotifyClient.GetArtist(spotifyContext, track.Artists[0].ID)
						if err != nil {
							dubLog(logPrefixHere, color.HiRedString, "Error fetching Spotify artist: %s", err)
						} else {
							artist_name = artist.Name
							artist_url = "https://open.spotify.com/artist/" + artist.ID.String()
							if len(artist.Images) > 0 {
								artist_image = artist.Images[0].URL
							}
							genres = artist.Genres
						}
					}
				} else if input_type == "playlist" {
					//TODO:
				}
				dubLog(logPrefixHere, color.HiGreenString, "ARTIST: %s", artist_name)
				dubLog(logPrefixHere, color.HiGreenString, "URL: %s", artist_url)
				dubLog(logPrefixHere, color.HiGreenString, "IMAGE: %s", artist_image)
				dubLog(logPrefixHere, color.HiGreenString, "GENRES: %s", strings.Join(genres, ", "))
			}

			/*

				- clean input
				- ifor artist/album/track/playlist

			*/

			/*msg, page, err := spotifyClient.FeaturedPlaylists(spotifyContext)
			if err != nil {
				dubLog(logPrefixHere, color.HiRedString, "Couldn't get featured playlists: %v", err)
			} else {
				dubLog(logPrefixHere, color.HiCyanString, msg)
				for _, playlist := range page.Playlists {
					dubLog(logPrefixHere, color.HiCyanString, playlist.Name)
				}
			}*/
		} else {
			dubLog(logPrefixHere, color.RedString, "Bot is not connected to Spotify...")
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
