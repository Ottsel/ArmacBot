package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func main() {
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
	)
	flag.Parse()

	dg, e := discordgo.New("Bot " + *Token)
	if err(e, "") {
		return
	}
	dg.AddHandler(onGuildCreate)
	dg.AddHandler(presenceUpdate)
	e = dg.Open()
	if err(e, "") {
		return
	}
	<-make(chan struct{})
	return
}
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable == true {
		return
	} else {
		roleFix(s, event.Guild)
	}
}

func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	correctRoles(s, p.GuildID, &p.Presence)
}
func roleFix(s *discordgo.Session, g *discordgo.Guild) {
	guildRoles, e := s.GuildRoles(g.ID)
	if err(e, "") {
		return
	}

	log.Println("Correcting Roles in guild:", g.Name+"\n")

	log.Println(strconv.Itoa(len(g.Presences)) + " users online")
	log.Println(strconv.Itoa(len(guildRoles)) + " roles available: \n")
	for _, r := range guildRoles {
		log.Println(r.Name, " - ", r.ID)
	}
	for _, p := range g.Presences {
		correctRoles(s, g.ID, p)
	}
}
func correctRoles(s *discordgo.Session, g string, p *discordgo.Presence) {
	if !p.User.Bot {

		guild, e := s.Guild(g)
		if err(e, "") {
			return
		}
		guildRoles, e := s.GuildRoles(g)
		if err(e, "") {
			return
		}

		var updatedRoles []string
		var role string

		if p.Game == nil {
			updatedRoles = append(updatedRoles, guild.ID)
			role = guild.Name
		} else {
			for _, gr := range guildRoles {
				if gr.Name == p.Game.Name {
					role = gr.Name
					updatedRoles = append(updatedRoles, gr.ID)
				}
			}
			if role == "" {
				for _, gr := range guildRoles {
					if gr.Name == "Other Games" {
						role = "Other Games"
						updatedRoles = append(updatedRoles, gr.ID)
					}
				}
				if role == "" {
					updatedRoles = append(updatedRoles, guild.ID)
					role = guild.Name
				}
			}
		}
		if role == guild.Name && p.Game != nil {
			log.Println("No role by name \"Other Games\", putting user in default role")
		} else {
			log.Println("Changing "+p.User.Username+"'s role to:", role)
		}
		e = s.GuildMemberEdit(g, p.User.ID, updatedRoles)
		if err(e, "") {
			return
		}
	}
}
func err(e error, c string) bool {
	if e != nil {
		if c != "" {
			log.Println(c)
		}
		log.Println("Error:", e)
		return true
	} else {
		return false
	}
}
