package main

import "github.com/fatih/color"

const (
	projectName    = "discord-utilities"
	projectLabel   = "Discord Utilities"
	projectVersion = "0.0.0-dev"
	projectIcon    = "https://cdn.discordapp.com/icons/780985109608005703/9dc25f1b91e6d92664590254e0797fad.webp?size=256"

	projectRepo          = "get-got/discord-utilities"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	configFileBase = "settings"
	databasePath   = "database"
	cachePath      = "cache"

	defaultReact = "✅"
)

var (
	logPrefixSetup = color.HiGreenString("[Setup]")

	logPrefixDebug = color.HiYellowString("[Debug]")

	logPrefixInfo = color.CyanString("[Info]")

	logPrefixSettings = color.GreenString("[Settings]")
	logPrefixVersion  = color.HiMagentaString("[Version]")
	logPrefixDiscord  = color.HiBlueString("[Discord]")
)

var (
	configFile  string
	configFileC bool
)
