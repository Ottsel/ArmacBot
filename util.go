package main

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	soundDir   string
	configPath string
	configText []byte = []byte("{\n\t\"SoundboardCommandKey\": \"!\",\n\t\"AdminCommandKey\": \"*\",\n\n\t\"CommandChannelID\": \"\",\n \n\t\"SoundboardMessageID\": \"\"\n}")
)

type Configuration struct {
	SoundboardCommandKey string
	AdminCommandKey      string
	CommandChannelID     string
	SoundboardMessageID  string
}

var cfg Configuration

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
		if ar.Name == "Admin" {
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
		log.Println("No role by name of \"Admin\", Things might not go so well :/")
		return false
	}
	return false
}
func config(g *discordgo.Guild) {
	soundDir = strings.ToLower(strings.Replace(g.Name, " ", "", -1)) + "/sounds"
	configPath = strings.ToLower(strings.Replace(g.Name, " ", "", -1)) + "/config.json"
	if _, e := os.Stat(soundDir); os.IsNotExist(e) {
		log.Println("No '" + soundDir + "' directory found, creating one")
		os.MkdirAll(soundDir, os.ModeDir)
	}
	sounds, _ = ioutil.ReadDir(soundDir)

	if _, e := os.Stat(configPath); os.IsNotExist(e) {
		os.Create(configPath)
		ioutil.WriteFile(configPath, configText, os.ModePerm)
		log.Println("No '" + configPath + "' file found, creating one.")
	}
	configFile, _ := os.Open(configPath)
	decoder := json.NewDecoder(configFile)
	cfg = Configuration{}
	e := decoder.Decode(&cfg)
	if err(e, "") {
		return
	}
}
func isConfigured(g *discordgo.Guild) bool {
	config(g)
	if cfg.CommandChannelID == "" || cfg.SoundboardMessageID == "" {
		return false
	}
	return true
}
func writeConfig(g *discordgo.Guild, c string, m string) {
	newConfigText := []byte("{\n\t\"SoundboardCommandKey\": \"!\",\n\t\"AdminCommandKey\": \"*\",\n\n\t\"CommandChannelID\": \"" + c + "\",\n \n\t\"SoundboardMessageID\": \"" + m + "\"\n}")
	ioutil.WriteFile(configPath, newConfigText, os.ModePerm)
	log.Println("Configuring Guild: " + g.Name)
}
func getGuild(s *discordgo.Session, c string) *discordgo.Guild {
	channel, e := s.State.Channel(c)
	if err(e, "") {
		return nil
	}
	g, e := s.State.Guild(channel.GuildID)
	if err(e, "") {
		return nil
	}
	return g
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
