package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/muhammadmuzzammil1998/jsonc"
)

var (
	config = defaultConfiguration()
)

//#region Credentials

var (
	placeholderToken string = "REPLACE_WITH_YOUR_TOKEN_OR_DELETE_LINE"
)

type configurationCredentials struct {
	// Login
	Token string `json:"token,omitempty"` // required for bot token (this or login)
	// APIs
	SpotifyClientID     string `json:"spotifyClientID,omitempty"`     // optional
	SpotifyClientSecret string `json:"spotifyClientSecret,omitempty"` // optional
	//twitter?
	//youtube?
	//flickr?
}

//#endregion

//#region Configuration

var (
	cdCommandPrefix  string = "dub "
	cdPresenceStatus string = "{{numServers}} servers"
	cdPresenceType   string = string(discordgo.StatusOnline)
)

func defaultConfiguration() configuration {
	return configuration{
		// Required
		Credentials: configurationCredentials{
			Token: placeholderToken,
		},
		// Setup
		Admins:              []string{},
		DebugOutput:         false,
		MessageOutput:       true,
		CommandPrefix:       cdCommandPrefix,
		DiscordCheckPerms:   true,
		DiscordTimeout:      180,
		ExitOnBadConnection: false,
		DiscordLogLevel:     discordgo.LogError,
		// Appearance
		PresenceEnabled: true,
		PresenceStatus:  &cdPresenceStatus,
		PresenceType:    cdPresenceType,
		PresenceLabel:   discordgo.ActivityType(discordgo.ActivityTypeListening),
	}
}

type configuration struct {
	Constants map[string]string `json:"_constants,omitempty"`

	// Required
	Credentials configurationCredentials `json:"credentials"` // required

	// Setup
	Admins              []string `json:"admins"`                        // optional
	CommandPrefix       string   `json:"commandPrefix"`                 // optional, defaults
	DebugOutput         bool     `json:"debugOutput"`                   // optional, defaults
	MessageOutput       bool     `json:"messageOutput"`                 // optional, defaults
	DiscordLogLevel     int      `json:"discordLogLevel,omitempty"`     // optional, defaults
	DiscordTimeout      int      `json:"discordTimeout,omitempty"`      // optional, defaults
	DiscordCheckPerms   bool     `json:"discordCheckPerms,omitempty"`   // optional, defaults
	ExitOnBadConnection bool     `json:"exitOnBadConnection,omitempty"` // optional, defaults
	//GithubUpdateChecking           bool                        `json:"githubUpdateChecking"`                     // optional, defaults

	// Appearance
	PresenceEnabled bool                   `json:"presenceEnabled,omitempty"` // optional, defaults
	PresenceStatus  *string                `json:"presenceStatus,omitempty"`  // optional, defaults
	PresenceType    string                 `json:"presenceType,omitempty"`    // optional, defaults
	PresenceLabel   discordgo.ActivityType `json:"presenceLabel,omitempty"`   // optional, defaults
	//EmbedColor      *string            `json:"embedColor,omitempty"`   // optional, defaults to role if undefined, then defaults random if no role color

	// Discord
	//All                  *configurationTarget  `json:"all,omitempty"`                  // optional, defaults
	//AllBlacklistChannels *[]string             `json:"allBlacklistChannels,omitempty"` // optional
	//AllBlacklistServers  *[]string             `json:"allBlacklistServers,omitempty"`  // optional
	PermittedServers []configurationTarget `json:"permittedServers"` // required
	OutputChannels   []configurationOutput `json:"outputChannels"`   // required
}

type configurationTarget struct {
	Server  string    `json:"server,omitempty"`  // used for config.PermittedServers
	Servers *[]string `json:"servers,omitempty"` // ---> alternative to Server

	UnlockCommands bool `json:"unlockCommands,omitempty"` // optional, defaults
}

type configurationOutput struct {
	Channel  string    `json:"channel,omitempty"`  // used for config.OutputChannels
	Channels *[]string `json:"channels,omitempty"` // ---> alternative to Channel

	OutputProgram bool `json:"outputProgram,omitempty"` // optional, defaults
}

func isServerPermitted(serverID string) bool {
	for _, server := range config.PermittedServers {
		if serverID == server.Server {
			return true
		}
		if server.Servers != nil {
			for _, nestedServer := range *server.Servers {
				if serverID == nestedServer {
					return true
				}
			}
		}
	}
	return false
}

func getPermittedServerConfig(serverID string) configurationTarget {
	for _, server := range config.PermittedServers {
		if serverID == server.Server {
			return server
		}
		if server.Servers != nil {
			for _, nestedServer := range *server.Servers {
				if serverID == nestedServer {
					return server
				}
			}
		}
	}
	return configurationTarget{}
}

//#endregion

func initConfig() {
	if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
		configFile = configFileBase + ".jsonc"
		configFileC = true
	} else {
		configFile = configFileBase + ".json"
		configFileC = false
	}
}

func loadConfig() {
	// Determine json type
	if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
		configFile = configFileBase + ".jsonc"
		configFileC = true
	} else {
		configFile = configFileBase + ".json"
		configFileC = false
	}
	// .
	dubLog("Settings", color.YellowString, "Loading from \"%s\"...", configFile)
	// Load settings
	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		dubLog("Settings", color.HiRedString, "Failed to open file...\t%s", err)
		//createConfig()
		properExit()
	} else {
		fixed := string(configContent)
		// Fix backslashes
		fixed = strings.ReplaceAll(fixed, "\\", "\\\\")
		for strings.Contains(fixed, "\\\\\\") {
			fixed = strings.ReplaceAll(fixed, "\\\\\\", "\\\\")
		}
		//TODO: Not even sure if this is realistic to do but would be nice to have line comma & trailing comma fixing

		// Parse
		newConfig := defaultConfiguration()
		if configFileC {
			err = jsonc.Unmarshal([]byte(fixed), &newConfig)
		} else {
			err = json.Unmarshal([]byte(fixed), &newConfig)
		}
		if err != nil {
			dubLog("Settings", color.HiRedString, "Failed to parse settings file...\t%s", err)
			dubLog("Settings", color.MagentaString, "Please ensure you're following proper JSON format syntax.")
			properExit()
		}
		// Constants
		if newConfig.Constants != nil {
			for key, value := range newConfig.Constants {
				if strings.Contains(fixed, key) {
					fixed = strings.ReplaceAll(fixed, key, value)
				}
			}
			// Re-parse
			newConfig = defaultConfiguration()
			if configFileC {
				err = jsonc.Unmarshal([]byte(fixed), &newConfig)
			} else {
				err = json.Unmarshal([]byte(fixed), &newConfig)
			}
			if err != nil {
				dubLog("Settings", color.HiRedString, "Failed to re-parse settings file after replacing constants...\t%s", err)
				dubLog("Settings", color.MagentaString, "Please ensure you're following proper JSON format syntax.")
				properExit()
			}
			newConfig.Constants = nil
		}
		config = newConfig

		// Debug Output
		if config.DebugOutput {
			dupeConfig := config
			if dupeConfig.Credentials.Token != "" && dupeConfig.Credentials.Token != placeholderToken {
				dupeConfig.Credentials.Token = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.SpotifyClientID != "" {
				dupeConfig.Credentials.SpotifyClientID = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.SpotifyClientSecret != "" {
				dupeConfig.Credentials.SpotifyClientSecret = "STRIPPED_FOR_OUTPUT"
			}
			s, err := json.MarshalIndent(dupeConfig, "", "\t")
			if err != nil {
				dubLog("Debug", color.HiRedString, "Failed to output...\t%s", err)
			} else {
				dubLog("Debug", color.HiYellowString, "Loaded Settings:\n%s", string(s))
			}
		}

		// Credentials Check
		if config.Credentials.Token == "" || config.Credentials.Token == placeholderToken {
			dubLog("Discord", color.HiRedString, "No valid discord login found. Token is invalid...")
			dubLog("Discord", color.HiYellowString, "Please save your credentials & info into \"%s\" then restart...", configFile)
			dubLog("Discord", color.MagentaString, "If your credentials are already properly saved, please ensure you're following proper JSON format syntax.")
			properExit()
		}
	}
}
