package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(m.Message)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.EditedTimestamp != nil {
		if *m.EditedTimestamp != time.Now() {
			handleMessage(m.Message)
		}
	}
}

func handleMessage(m *discordgo.Message) int64 {
	//log.Println(m.Content)
	return -1
}
