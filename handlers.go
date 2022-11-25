package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(m.Message, false)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.EditedTimestamp != nil {
		if *m.EditedTimestamp != time.Now() {
			handleMessage(m.Message, true)
		}
	}
}

func handleMessage(m *discordgo.Message, edited bool) int64 {
	if config.MessageOutput && m.Author.ID != bot.State.User.ID && !m.Author.Bot {
		prfx := "NEW MESSAGE"
		if edited {
			prfx = "MESSAGE EDIT"
		}
		dubLog(prfx, color.CyanString, "%s/%s/%s - %s: %s (%d attachments)",
			m.GuildID, m.ChannelID, m.ID, getUserIdentifier(*m.Author), m.Content, len(m.Attachments),
		)
	}
	return -1
}
