package main

import "github.com/bwmarrin/discordgo"

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

//#endregion
