package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(m.Message)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.EditedTimestamp != discordgo.Timestamp("") {
		handleMessage(m.Message)
	}
}

func handleMessage(m *discordgo.Message) int64 {
	// Ignore own messages unless told not to
	log.Println(m.Content)
	return -1
}
