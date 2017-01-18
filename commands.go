package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func forget(session *discordgo.Session, mc *discordgo.MessageCreate) {
	channel, e := session.Channel(mc.ChannelID)
	if err(e, "") {
		return
	}
	var untilMsg = strings.Replace(mc.Content, "*forget ", "", -1)
	messages, e := session.ChannelMessages(mc.ChannelID, 100, channel.LastMessageID, untilMsg, "")
	if err(e, "") {
		return
	}
	for _, m := range messages {
		if m.ID != cfg.SoundboardMessageID {
			e := session.ChannelMessageDelete(mc.ChannelID, m.ID)
			if err(e, "Couldn't delete message: "+m.Content) {
				return
			} else {
				log.Println("Deleting message: ", m.Content)
			}
		}
	}
}
func help(session *discordgo.Session, mc *discordgo.MessageCreate) {
	m, e := session.ChannelMessageSend(mc.ChannelID, (cfg.AdminCommandKey + "`help - Shows this dialog.`"))
	if err(e, "Couldn't post message: "+m.Content) {
		return
	}
	m, e = session.ChannelMessageSend(mc.ChannelID, ("`" + cfg.AdminCommandKey + "forget [messageID] - Deletes messages down to a specified message.`"))
	if err(e, "Couldn't post message: "+m.Content) {
		return
	}
}
