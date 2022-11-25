package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

/*

ROADMAP BEFORE RETIRING PYTHON BOT:
- Admin & Channel/Server Restriction
- Fluidly construct embeds, message error checking
- SG
- Reboot
- Exit
- Help
- Info

FUTURE ROADMAP:
- Channel Wiper functionality

*/

var (
	// Bot
	bot      *discordgo.Session
	botReady bool = false
	user     *discordgo.User
	dgr      *exrouter.Route
	// General
	loop                 chan os.Signal
	timeLaunched         time.Time
	timePresenceUpdated  time.Time
	timeConfigReloaded   time.Time
	configReloadLastTime time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	timeLaunched = time.Now()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	dubLog("Main", color.HiCyanString, "Welcome to %s v%s", projectName, projectVersion)
	dubLog("Version", color.CyanString, "%s / discord-go v%s / Discord API v%s", runtime.Version(), discordgo.VERSION, discordgo.APIVersion)
}

func main() {
	//var err error

	// Config
	loadConfig()

	botLogin()
	botReady = true

	// Startup Done
	dubLog("Main", color.YellowString, "Startup finished, took %s...", uptime())
	dubLog("Discord", color.HiCyanString, "%s v%s is online and connected to %d guilds", projectLabel, projectVersion, len(bot.State.Guilds))
	dubLog("Main", color.RedString, "CTRL+C to exit...")

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
					dubLog("Discord", color.YellowString, "Closing Discord connections...")
					bot.Client.CloseIdleConnections()
					bot.CloseWithCode(1001)
					bot = nil
					dubLog("Discord", color.RedString, "Discord connections closed!")
					if config.ExitOnBadConnection {
						properExit()
					} else {
						dubLog("Discord", color.GreenString, "Logging in...")
						botLogin()
						dubLog("Discord", color.HiGreenString, "Reconnected! The bot *should* resume working...")
						// Log Status
						//logStatusMessage(logStatusReconnect)
					}
				}
				gate, err := bot.Gateway()
				if err != nil || gate == "" {
					dubLog("Discord", color.HiYellowString, "Bot encountered a gateway error: GATEWAY: %s,\tERR: %s", gate, err)
					doReconnect()
				} else if time.Since(bot.LastHeartbeatAck).Seconds() > 4*60 {
					dubLog("Discord", color.HiYellowString, "Bot has not received a heartbeat from Discord in 4 minutes...")
					doReconnect()
				}
			}
		}
	}()

	// Settings Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		dubLog("Settings", color.HiRedString, "Error creating NewWatcher:\t%s", err)
	}
	defer watcher.Close()
	err = watcher.Add(configFile)
	if err != nil {
		dubLog("Settings", color.HiRedString, "Error adding watcher for settings:\t%s", err)
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// It double-fires the event without time check, might depend on OS but this works anyways
					if time.Now().Sub(configReloadLastTime).Milliseconds() > 1 {
						time.Sleep(1 * time.Second)
						dubLog("Settings", color.YellowString, "Detected changes in \"%s\", reloading...", configFile)
						loadConfig()
						dubLog("Settings", color.HiYellowString, "Reloaded...")

						updateDiscordPresence()
					}
					configReloadLastTime = time.Now()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				dubLog("Settings", color.HiRedString, "Watcher Error:\t%s", err)
			}
		}
	}()

	//#endregion

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	dubLog("Discord", color.GreenString, "Logging out of discord...")
	bot.Close()

	dubLog("Main", color.HiRedString, "Exiting...")
}

var (
	spotifyClient  *spotify.Client
	spotifyContext context.Context
)

func loadAPIs() {
	// Spotify
	if config.Credentials.SpotifyClientID != "" && config.Credentials.SpotifyClientSecret != "" {

		dubLog("Spotify", color.MagentaString, "Connecting to Spotify API...")
		spotifyConfig := &clientcredentials.Config{
			ClientID:     config.Credentials.SpotifyClientID,
			ClientSecret: config.Credentials.SpotifyClientSecret,
			TokenURL:     spotifyauth.TokenURL,
		}
		spotifyContext = context.Background()
		spotifyToken, err := spotifyConfig.Token(spotifyContext)
		if err != nil {
			dubLog("Spotify", color.HiRedString, "Error getting Spotify token: %s", err)
		} else {
			spotifyClient = spotify.New(spotifyauth.New().Client(spotifyContext, spotifyToken))
			_, err = spotifyClient.GetCategories(spotifyContext)
			if err != nil {
				dubLog("Spotify", color.HiRedString, "Error connecting to Spotify: %s", err)
			} else {
				dubLog("Spotify", color.HiGreenString, "Connected to Spotify API!")
			}
		}
	}
}

func botLogin() {

	loadAPIs()

	var err error

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		dubLog("Discord", color.GreenString, "Connecting to Discord via Token...")
		input := config.Credentials.Token
		if input[:3] != "Bot" {
			input = "Bot " + input
		}
		bot, err = discordgo.New(input)
	} else {
		dubLog("Discord", color.HiRedString, "No valid credentials for Discord...")
		properExit()
	}
	if err != nil {
		dubLog("Discord", color.HiRedString, "Error logging in: %s", err)
		properExit()
	}

	// Connect Bot
	err = bot.Open()
	if err != nil {
		dubLog("Discord", color.HiRedString, "Discord login failed:\t%s", err)
		properExit()
	}
	bot.ShouldReconnectOnError = true
	dur, err := time.ParseDuration(string(config.DiscordTimeout) + "s")
	if err != nil {
		dur, _ = time.ParseDuration("180s")
	}
	bot.Client.Timeout = dur

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		user = bot.State.User
		if user == nil {
			dubLog("Discord", color.HiRedString, "Error obtaining user details: %s", err)
			loop <- syscall.SIGINT
		}
	} else if user == nil {
		dubLog("Discord", color.HiRedString, "No error encountered obtaining user details, but it's empty...")
		loop <- syscall.SIGINT
	}

	// Event Handlers
	dgr = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timePresenceUpdated = time.Now()
	updateDiscordPresence()
}
