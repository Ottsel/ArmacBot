package main

import (
	"flag"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	botID       string
	adminRoleID string
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
	}
	roleFix(s, event.Guild)
}
func authenticate(s *discordgo.Session, g string, u *discordgo.User) bool {
	user, e := s.GuildMember(g, u.ID)
	if err(e, "") {
		return false
	}
	roles, e := s.GuildRoles(g)
	if err(e, "") {
		return false
	}
	for _, ar := range roles {
		if ar.Permissions == 8 {
			adminRoleID = ar.ID
		}
	}
	if adminRoleID != "" {
		for _, r := range user.Roles {
			if r == adminRoleID {
				return true
			}
		}
	} else {
		log.Println("Make sure admins only have the permission \"Administrator,\" they override other permissions anyway. ;)")
		return false
	}
	return false
}

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
	var role string

	guild, e := s.Guild(g)
	if err(e, "") {
		return
	}
	guildRoles, e := s.GuildRoles(g)
	if err(e, "") {
		return
	}

	var updatedRoles []string

	for _, gr := range guildRoles {
		if p.Game != nil {
			if gr.Name == p.Game.Name {
				role = gr.Name
				updatedRoles = append(updatedRoles, gr.ID)
			} else {
				if gr.Name == "Other Games" {
					role = "Other Games"
					updatedRoles = append(updatedRoles, gr.ID)
				} else {
					updatedRoles = append(updatedRoles, "1")
					role = guild.Name

					log.Println("No role by name \"Other Games\", putting user in default role")
				}
			}
		} else {
			updatedRoles = append(updatedRoles, "1")
		}
		if p.User.Bot {
			if gr.Name == "Bots" {
				role = "Bot"
				updatedRoles = append(updatedRoles, gr.ID)
			}
		}
	}
	log.Println("Changing user role to:", role)
	e = s.GuildMemberEdit(g, p.User.ID, updatedRoles)
	if err(e, "") {
		return
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
