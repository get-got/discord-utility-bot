package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	// Bot
	bot  *discordgo.Session
	user *discordgo.User
	dgr  *exrouter.Route
	// Gen
	loop                 chan os.Signal
	startTime            time.Time
	timeLastUpdated      time.Time
	configReloadLastTime time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString("Welcome to %s v%s", projectName, projectVersion))
	log.Println(color.CyanString("discord-go v%s using Discord API v%s", discordgo.VERSION, discordgo.APIVersion))
}

func main() {
	//var err error

	// Config
	loadConfig()

	botLogin()

	// Startup Done
	log.Println(color.YellowString("Startup finished, took %s...", uptime()))
	log.Println(color.HiCyanString("%s v%s is online and connected to %d guilds", projectLabel, projectVersion, len(bot.State.Guilds)))
	log.Println(color.RedString("CTRL+C to exit..."))

	//#region Background Tasks

	// Tickers
	ticker5m := time.NewTicker(5 * time.Minute)
	ticker1m := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker5m.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				//updateDiscordPresence()
			case <-ticker1m.C:
				doReconnect := func() {
					log.Println(logPrefixDiscord, color.YellowString("Closing Discord connections..."))
					bot.Client.CloseIdleConnections()
					bot.CloseWithCode(1001)
					bot = nil
					log.Println(logPrefixDiscord, color.RedString("Discord connections closed!"))
					if config.ExitOnBadConnection {
						properExit()
					} else {
						log.Println(logPrefixDiscord, color.GreenString("Logging in..."))
						botLogin()
						log.Println(logPrefixDiscord, color.HiGreenString("Reconnected! The bot *should* resume working..."))
						// Log Status
						//logStatusMessage(logStatusReconnect)
					}
				}
				gate, err := bot.Gateway()
				if err != nil || gate == "" {
					log.Println(logPrefixDiscord, color.HiYellowString("Bot encountered a gateway error: GATEWAY: %s,\tERR: %s", gate, err))
					doReconnect()
				} else if time.Since(bot.LastHeartbeatAck).Seconds() > 4*60 {
					log.Println(logPrefixDiscord, color.HiYellowString("Bot has not received a heartbeat from Discord in 4 minutes..."))
					doReconnect()
				}
			}
		}
	}()

	//#endregion

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	log.Println(color.GreenString("Logging out of discord..."))
	bot.Close()

	log.Println(color.HiRedString("Exiting... "))
}

var (
	spotifyClient  *spotify.Client
	spotifyContext context.Context
)

func loadAPIs() {
	logPrefixHere := color.HiMagentaString("[APIs]")
	// Spotify
	if config.Credentials.SpotifyClientID != "" && config.Credentials.SpotifyClientSecret != "" {
		logPrefixHere = color.HiMagentaString("[API:Spotify]")

		log.Println(logPrefixHere, color.MagentaString("Connecting to Spotify API..."))
		spotifyConfig := &clientcredentials.Config{
			ClientID:     config.Credentials.SpotifyClientID,
			ClientSecret: config.Credentials.SpotifyClientSecret,
			TokenURL:     spotifyauth.TokenURL,
		}
		spotifyContext = context.Background()
		spotifyToken, err := spotifyConfig.Token(spotifyContext)
		if err != nil {
			log.Println(logPrefixHere, color.HiRedString("Error getting Spotify token: %s", err))
		} else {
			spotifyClient = spotify.New(spotifyauth.New().Client(spotifyContext, spotifyToken))
			_, err = spotifyClient.GetCategories(spotifyContext)
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Error connecting to Spotify: %s", err))
			} else {
				log.Println(logPrefixHere, color.HiGreenString("Connected to Spotify API!"))
			}
		}
	}
}

func botLogin() {

	loadAPIs()

	var err error

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		log.Println(logPrefixDiscord, color.GreenString("Connecting to Discord via Token..."))
		input := config.Credentials.Token
		if input[:3] != "Bot" {
			input = "Bot " + input
		}
		bot, err = discordgo.New(input)
	} else {
		log.Println(logPrefixDiscord, color.HiRedString("No valid credentials for Discord..."))
		properExit()
	}
	if err != nil {
		log.Println(logPrefixDiscord, color.HiRedString("Error logging in: %s", err))
		properExit()
	}

	// Connect Bot
	err = bot.Open()
	if err != nil {
		log.Println(logPrefixDiscord, color.HiRedString("Discord login failed:\t%s", err))
		properExit()
	}
	bot.ShouldReconnectOnError = true

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		user = bot.State.User
		if user == nil {
			log.Println(logPrefixDiscord, color.HiRedString("Error obtaining user details: %s", err))
			loop <- syscall.SIGINT
		}
	} else if user == nil {
		log.Println(logPrefixDiscord, color.HiRedString("No error encountered obtaining user details, but it's empty..."))
		loop <- syscall.SIGINT
	}

	// Event Handlers
	dgr = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timeLastUpdated = time.Now()
	//updateDiscordPresence()
}
