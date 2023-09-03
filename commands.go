package main

import (
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/zmb3/spotify/v2"
)

// Multiple use messages to save space and make cleaner.
const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)

func cmderrUnpermittedServer(prefix string, ctx *exrouter.Context) {
	dubLog(prefix, logLevelWarning, color.HiCyanString, "%s tried to use %s command in an unpermitted server (%s)", getUserIdentifier(*ctx.Msg.Author), ctx.Args[0], ctx.Msg.GuildID)
}

func handleCommands() *exrouter.Route {
	router := exrouter.New()

	//#region Utility Commands

	router.On("ping", func(ctx *exrouter.Context) {
		logPrefixHere := "commands:ping"
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				dubLog(logPrefixHere, logLevelError, color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID)
			} else {
				beforePong := time.Now()
				pong, err := ctx.Reply("Pong!")
				if err != nil {
					dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error sending pong message:\t%s", err)
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
					dubLog(logPrefixHere, logLevelInfo, color.HiCyanString, "%s pinged bot - Latency: %dms, Roundtrip: %dms",
						getUserIdentifier(*ctx.Msg.Author),
						latency,
						roundtrip)
				}
			}
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot.")

	router.On("help", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("commands:help")
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				dubLog(logPrefixHere, logLevelError, color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID)
			} else {
				text := ""
				for _, cmd := range router.Routes {
					if cmd.Category != "Admin" || isBotAdmin(ctx.Msg) {
						text += fmt.Sprintf("• \"%s\" : %s", cmd.Name, cmd.Description)
						if len(cmd.Aliases) > 0 {
							text += fmt.Sprintf("\n— Aliases: \"%s\"", strings.Join(cmd.Aliases, "\", \""))
						}
						text += "\n\n"
					}
				}
				content := fmt.Sprintf("Use commands as ``\"%s<command> <arguments?>\"``\n```%s```\n%s",
					config.CommandPrefix, text, projectRepoURL)
				if _, err := replyEmbed(ctx.Msg, "Command — Help", content); err != nil {
					dubLog(logPrefixHere, logLevelError, color.HiRedString, "Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err)
				}
				dubLog(logPrefixHere, logLevelInfo, color.HiCyanString, "%s asked for help", getUserIdentifier(*ctx.Msg.Author))
			}
		}

	}).Cat("Utility").Alias("h").Desc("Help.. please...")

	//#endregion

	//#region Admin Commands

	router.On("execute", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("commands:execute")
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			if !isBotAdmin(ctx.Msg) {
				dubLog(logPrefixHere, logLevelError, color.HiRedString, "%s attempted cli execution but is not admin...", getUserIdentifier(*ctx.Msg.Author))
			} else {
				var cmd *exec.Cmd
				input := ctx.Args.After(1)
				if strings.Contains(input, "sudo") {
					cmd = exec.Command("/bin/sh", "-c", input)
				} else {
					cmd = exec.Command(input)
				}
				stdout, err := cmd.CombinedOutput()
				if err != nil {
					log.Println(err)
					ctx.Reply(fmt.Sprintf("```[ERROR]: %s```", err))
				} else {
					log.Println(string(stdout))
					ctx.Reply(fmt.Sprintf("```%s```", string(stdout)))
				}
			}
		}

	}).Cat("Admin").Alias("exec", "ex", "x").Desc("Command line execution.")

	router.On("exit", func(ctx *exrouter.Context) {
		logPrefixHere := "commands:exit"
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			if !isBotAdmin(ctx.Msg) {
				dubLog(logPrefixHere, logLevelError, color.HiRedString, "%s attempted program exit but is not admin...", getUserIdentifier(*ctx.Msg.Author))
			} else {
				dubLog(logPrefixHere, logLevelInfo, color.HiCyanString, "%s commanded program exit; exiting in 15 seconds...", getUserIdentifier(*ctx.Msg.Author))
				if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					dubLog(logPrefixHere, logLevelError, color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID)
				} else {
					if _, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting bot program in 15 seconds..."); err != nil {
						dubLog(logPrefixHere, logLevelError, color.HiRedString, "Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err)
					}
				}
				properExit()
			}
		}
	}).Cat("Admin").Alias("reload", "kill").Desc("[A] Exits this program.")

	router.On("reboot", func(ctx *exrouter.Context) {
		logPrefixHere := "commands:reboot"
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			//TODO: show what syscall is being used if debug enabled
			if !isBotAdmin(ctx.Msg) {
				dubLog(logPrefixHere, logLevelError, color.HiRedString, "%s attempted system reboot but is not admin...", getUserIdentifier(*ctx.Msg.Author))
			} else {
				dubLog(logPrefixHere, logLevelInfo, color.HiGreenString, "%s commanded system reboot; rebooting in 10 seconds...", getUserIdentifier(*ctx.Msg.Author))
				if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					dubLog(logPrefixHere, logLevelError, color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID)
				} else {
					if _, err := replyEmbed(ctx.Msg, "Command — Reboot", "Rebooting host system in 10 seconds..."); err != nil {
						dubLog(logPrefixHere, logLevelError, color.HiRedString, "Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err)
					}
				}
				time.Sleep(10 * time.Second)
				reboot()
			}
		}
	}).Cat("Admin").Alias("restart", "shutdown").Desc("[A] Restarts the server.")

	//#endregion

	//#region Spotify API

	router.On("sg", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("commands:spotifygenres")
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			if spotifyClient == nil {
				dubLog(logPrefixHere, logLevelWarning, color.RedString, "Bot is not connected to Spotify...")
			} else {
				if checkSpotify() == nil {
					if spotErr := loadSpotifyAPI(); spotErr != nil {
						dubLog(logPrefixHere, logLevelWarning, color.RedString,
							"Bot failed to reconnect to Spotify...\t%s", spotErr)
					}
				}

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
					dubLog(logPrefixHere, logLevelWarning, color.HiRedString, "Input is not a valid format: \"\"...", input)
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

					if input_type == "playlist" {
						var genres map[string]int = make(map[string]int)
						var artists []spotify.ID
						playlist, err := spotifyClient.GetPlaylist(spotifyContext, spotify.ID(cleanedInput))
						if err != nil {
							dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error fetching Spotify playlist: %s", err)
						} else {
							isArtistInStack := func(artist spotify.ID) bool {
								for _, a := range artists {
									if a == artist {
										return true
									}
								}
								return false
							}
							for _, track := range playlist.Tracks.Tracks {
								// cache unique artists
								if !isArtistInStack(track.Track.Artists[0].ID) {
									artists = append(artists, track.Track.Artists[0].ID)
								}
							}
							// foreach unique artist
							for _, a := range artists {
								artist, err := spotifyClient.GetArtist(spotifyContext, a)
								if err != nil {
									dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error fetching Spotify artist: %s", err)
								} else {
									for _, genre := range artist.Genres {
										// exists
										if _, ok := genres[genre]; ok {
											genres[genre]++
										} else {
											genres[genre] = 1
										}
									}
								}
							}
							// foreach genre
							keys := make([]string, 0, len(genres))
							for key := range genres {
								keys = append(keys, key)
							}
							sort.SliceStable(keys, func(i, j int) bool {
								return genres[keys[i]] > genres[keys[j]]
							})
							output := fmt.Sprintf("**[%s's](https://open.spotify.com/playlist/%s \"%s\") top genres:**", playlist.Name, playlist.ID.String(), playlist.Name)
							for _, genre := range keys {
								if genres[genre] > 1 {
									output += fmt.Sprintf("\n• %s: %d", strings.Title(genre), genres[genre])
								}
							}
							//TODO: clean this up and better error reporting
							_, err := bot.ChannelMessageSendComplex(ctx.Msg.ChannelID,
								&discordgo.MessageSend{
									Content: ctx.Msg.Author.Mention(),
									Embed: &discordgo.MessageEmbed{
										Title:       "Spotify Genre Search",
										Description: output,
										Color:       getEmbedColor(ctx.Msg.ChannelID),
										Thumbnail: &discordgo.MessageEmbedThumbnail{
											URL: playlist.Images[0].URL,
										},
										Footer: &discordgo.MessageEmbedFooter{
											IconURL: projectIcon,
											Text:    fmt.Sprintf("%s v%s", projectName, projectVersion),
										},
									},
								},
							)
							if err != nil {
								dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error sending command response message: %s", err)
							}
						}
					} else {
						// Output Vars
						var artist_id spotify.ID
						var artist_name string
						var artist_url string
						var artist_image string
						var genres []string

						if input_type == "artist" {
							artist_id = spotify.ID(cleanedInput)
						} else if input_type == "album" {
							album, err := spotifyClient.GetAlbum(spotifyContext, spotify.ID(cleanedInput))
							if err != nil {
								dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error fetching Spotify album: %s", err)
							}
							artist_id = album.Artists[0].ID
						} else if input_type == "track" {
							track, err := spotifyClient.GetTrack(spotifyContext, spotify.ID(cleanedInput))
							if err != nil {
								dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error fetching Spotify track: %s", err)
							}
							artist_id = track.Artists[0].ID
						}

						if artist_id != "" {
							artist, err := spotifyClient.GetArtist(spotifyContext, artist_id)
							if err != nil {
								dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error fetching Spotify artist: %s", err)
							} else {
								artist_name = artist.Name
								artist_url = "https://open.spotify.com/artist/" + artist.ID.String()
								if len(artist.Images) > 0 {
									artist_image = artist.Images[0].URL
								}
								genres = artist.Genres
							}
						}

						output := fmt.Sprintf("**[%s's](%s \"%s\") genres:**", artist_name, artist_url, artist_name)
						if len(genres) == 0 {
							output += "\nWho?..."
						} else {
							for _, genre := range genres {
								genre_link := fmt.Sprintf("https://open.spotify.com/search/genre%%3A%%22%s%%22", strings.ReplaceAll(genre, " ", "%20"))
								output += fmt.Sprintf("\n• [%s](%s \"%s\")", strings.Title(genre), genre_link, genre)
							}
						}
						//TODO: clean this up and better error reporting
						_, err := bot.ChannelMessageSendComplex(ctx.Msg.ChannelID,
							&discordgo.MessageSend{
								Content: ctx.Msg.Author.Mention(),
								Embed: &discordgo.MessageEmbed{
									Title:       "Spotify Genre Search",
									Description: output,
									Color:       getEmbedColor(ctx.Msg.ChannelID),
									Thumbnail: &discordgo.MessageEmbedThumbnail{
										URL: artist_image,
									},
									Footer: &discordgo.MessageEmbedFooter{
										IconURL: projectIcon,
										Text:    fmt.Sprintf("%s v%s", projectName, projectVersion),
									},
								},
							},
						)
						if err != nil {
							dubLog(logPrefixHere, logLevelError, color.HiRedString, "Error sending command response message: %s", err)
						}

					}
				}
			}
		}
	}).Cat("Spotify").Alias("spotifygenres", "spotgen").Desc("Spotify genre lookup by url.")

	//#endregion

	//#region Misc...

	router.On("plex", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("commands:plex")
		if !isServerPermitted(ctx.Msg.GuildID) {
			cmderrUnpermittedServer(logPrefixHere, ctx)
		} else {
			//
		}

	}).Cat("Misc").Desc("<WIP!!> Plex Status.")

	//#endregion

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(bot, ".", bot.State.User.ID, m.Message)
		router.FindAndExecute(bot, config.CommandPrefix, bot.State.User.ID, m.Message)
	})

	return router
}
