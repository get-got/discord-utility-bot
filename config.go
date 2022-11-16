package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
}

//#endregion

//#region Configuration

var (
	cdDebugOutput bool = false
)

func defaultConfiguration() configuration {
	return configuration{
		// Required
		Credentials: configurationCredentials{
			Token: placeholderToken,
		},
		// Setup
		Admins:              []string{},
		DebugOutput:         cdDebugOutput,
		ExitOnBadConnection: false,
		DiscordLogLevel:     discordgo.LogError,
	}
}

type configuration struct {
	Constants map[string]string `json:"_constants,omitempty"`
	// Required
	Credentials configurationCredentials `json:"credentials"` // required
	// Setup
	Admins              []string `json:"admins"`                        // optional
	DebugOutput         bool     `json:"debugOutput"`                   // optional, defaults
	DiscordLogLevel     int      `json:"discordLogLevel,omitempty"`     // optional, defaults
	ExitOnBadConnection bool     `json:"exitOnBadConnection,omitempty"` // optional, defaults
	//TODO: mapping for channels or servers, whitelist or blacklist, however it will listen
}

type constStruct struct {
	Constants map[string]string `json:"_constants,omitempty"`
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
	log.Println(logPrefixSettings, color.YellowString("Loading from \"%s\"...", configFile))
	// Load settings
	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println(logPrefixSettings, color.HiRedString("Failed to open file...\t%s", err))
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
			log.Println(logPrefixSettings, color.HiRedString("Failed to parse settings file...\t%s", err))
			log.Println(logPrefixSettings, color.MagentaString("Please ensure you're following proper JSON format syntax."))
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
				log.Println(logPrefixSettings, color.HiRedString("Failed to re-parse settings file after replacing constants...\t%s", err))
				log.Println(logPrefixSettings, color.MagentaString("Please ensure you're following proper JSON format syntax."))
				properExit()
			}
			newConfig.Constants = nil
		}
		config = newConfig

		// Debug Output
		if config.DebugOutput {
			s, err := json.MarshalIndent(config, "", "\t")
			if err != nil {
				log.Println(logPrefixSettings, logPrefixDebug, color.HiRedString("Failed to output...\t%s", err))
			} else {
				log.Println(logPrefixSettings, logPrefixDebug, color.HiYellowString("Parsed into JSON:\n\n"),
					color.YellowString(string(s)),
				)
			}
		}

		// Credentials Check
		if config.Credentials.Token == "" || config.Credentials.Token == placeholderToken {
			log.Println(logPrefixSettings, color.HiRedString("No valid discord login found. Token is invalid..."))
			log.Println(logPrefixSettings, color.HiYellowString("Please save your credentials & info into \"%s\" then restart...", configFile))
			log.Println(logPrefixSettings, color.MagentaString("If your credentials are already properly saved, please ensure you're following proper JSON format syntax."))
			properExit()
		}
	}
}
