package main

const (
	projectName    = "discord-utility-bot"
	projectLabel   = "Discord Utility Bot"
	projectVersion = "0.0.3-dev"
	projectIcon    = "https://cdn.discordapp.com/attachments/716861000745222164/1045416792530624724/trree.png"

	projectRepo          = "get-got/discord-utility-bot"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	configFileBase = "settings"
	databasePath   = "database"
	cachePath      = "cache"

	defaultReact = "✅"
)

var (
	configFile  string
	configFileC bool
)
