package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	correctRoles(s, p.GuildID, &p.Presence)
}
func roleFix(s *discordgo.Session, g *discordgo.Guild) {
	log.Println("Correcting Roles in guild: ", g.Name)
	botID = s.State.User.ID
	for _, p := range g.Presences {
		correctRoles(s, g.ID, p)
	}
}
func correctRoles(s *discordgo.Session, g string, p *discordgo.Presence) {
	if authenticate(s, g, p.User) {
		log.Println("Admin, cannot modify role")
		return
	}
	var updatedRoles []string
	var role string

	guildRoles, e := s.GuildRoles(g)
	if err(e, "") {
		return
	}
	for _, gr := range guildRoles {
		if p.User.Bot {
			if gr.Name == "Bots" {
				role = "Bot"
				updatedRoles = append(updatedRoles, gr.ID)
			}
		} else {
			if p.Game != nil {
				if gr.Name == p.Game.Name {
					role = gr.Name
					updatedRoles = append(updatedRoles, gr.ID)
				}
			} else {
				guild, e := s.Guild(g)
				if err(e, "") {
					return
				}
				role = guild.Name
				updatedRoles = append(updatedRoles, guild.ID)
			}
			if role == "" {
				for _, gr := range guildRoles {
					if gr.Name == "Other Games" {
						updatedRoles = append(updatedRoles, gr.ID)
						role = "Other Games"
					}
				}
			}
			if role == "" {
				guild, e := s.Guild(g)
				if err(e, "") {
					return
				}
				role = guild.Name
				updatedRoles = append(updatedRoles, guild.ID)
				log.Println("No role by name \"Other Games\", putting user in default role")
			}
		}
	}
	log.Println("Changing user role to:", role)
	e = s.GuildMemberEdit(g, p.User.ID, updatedRoles)
	if err(e, "") {
		return
	}
}
