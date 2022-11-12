package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
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
	var err error

	// Config
	//loadConfig()

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
				updateDiscordPresence()
			case <-ticker1m.C:
				doReconnect := func() {
					log.Println(color.YellowString("Closing Discord connections..."))
					bot.Client.CloseIdleConnections()
					bot.CloseWithCode(1001)
					bot = nil
					log.Println(color.RedString("Discord connections closed!"))
					if config.ExitOnBadConnection {
						properExit()
					} else {
						log.Println(color.GreenString("Logging in..."))
						botLogin()
						log.Println(color.HiGreenString("Reconnected! The bot *should* resume working..."))
						// Log Status
						logStatusMessage(logStatusReconnect)
					}
				}
				gate, err := bot.Gateway()
				if err != nil || gate == "" {
					log.Println(color.HiYellowString("Bot encountered a gateway error: GATEWAY: %s,\tERR: %s", gate, err))
					doReconnect()
				} else if time.Since(bot.LastHeartbeatAck).Seconds() > 4*60 {
					log.Println(color.HiYellowString("Bot has not received a heartbeat from Discord in 4 minutes..."))
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

func botLogin() {

	var err error

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		log.Println(logPrefixDiscord, color.GreenString("Connecting to Discord via Token..."))
		bot, err = discordgo.New("Bot " + config.Credentials.Token)
	} else {
		log.Println(logPrefixDiscord, color.HiRedString("No valid credentials for Discord..."))
		properExit()
	}
	if err != nil {
		// Newer discordgo throws this error for some reason with Email/Password login
		if err.Error() != "Unable to fetch discord authentication token. <nil>" {
			log.Println(logPrefixDiscord, color.HiRedString("Error logging in: %s", err))
			properExit()
		}
	}

	// Connect Bot
	bot.LogLevel = -1 // to ignore dumb wsapi error
	err = bot.Open()
	if err != nil {
		log.Println(logPrefixDiscord, color.HiRedString("Discord login failed:\t%s", err))
		properExit()
	}
	bot.LogLevel = discordgo.LogError // reset
	bot.ShouldReconnectOnError = true
	//bot.Client.Timeout = 100000

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		user = bot.State.User
		if user == nil {
			log.Println(color.HiRedString("Error obtaining user details: %s", err))
			loop <- syscall.SIGINT
		}
	} else if user == nil {
		log.Println(color.HiRedString("No error encountered obtaining user details, but it's empty..."))
		loop <- syscall.SIGINT
	} else {
		log.Println(color.MagentaString("This is a Bot User"))
		log.Println(color.MagentaString("- Status presence details are limited."))
		log.Println(color.MagentaString("- Access is restricted to servers you have permission to add the bot to."))
	}

	// Event Handlers
	dgr = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()
}
