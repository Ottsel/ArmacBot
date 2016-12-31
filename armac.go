package main

import (
	"C"
	"encoding/json"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	sounds      []os.FileInfo
	adminRoleID string
	playing     bool = false
)

type Configuration struct {
	BotToken             string
	GuildID              string
	SoundboardCommandKey string
	AdminCommandKey      string
	CommandChannelName   string
	SoundboardMessageID  string
}

var cfg Configuration

func init() {
	if _, e := os.Stat("sounds"); os.IsNotExist(e) {
		log.Println("No 'sounds' directory found, creating one")
		os.Mkdir("sounds", os.ModeDir)
	}
	sounds, _ = ioutil.ReadDir("sounds")

	if _, e := os.Stat("config.json"); os.IsNotExist(e) {
		os.Create("config.json")
		configText := []byte("{\n\t\"BotToken\": \"\",\n\t\"GuildID\": \"\",\n\n\t\"SoundboardCommandKey\": \"!\",\n\t\"AdminCommandKey\": \"*\",\n\n\t\"CommandChannelName\": \"\",\n \n\t\"SoundboardMessageID\": \"\"\n}")
		ioutil.WriteFile("config.json", configText, os.ModePerm)
		log.Println("No config file found, creating one. Please configure and restart.")
		return
	}
	configFile, _ := os.Open("config.json")
	decoder := json.NewDecoder(configFile)
	cfg = Configuration{}
	e := decoder.Decode(&cfg)
	if e != nil {
		log.Println("Error:", e)
		return
	}
}
func main() {
	dg, e := discordgo.New("Bot " + cfg.BotToken)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	dg.AddHandlerOnce(ready)
	dg.AddHandler(messageCreate)
	dg.AddHandler(presenceUpdate)
	dg.AddHandler(voiceState)
	if e := dg.Open(); e != nil {
		log.Println("Error:", e)
		return
	}
	<-make(chan struct{})
	return
}
func ready(s *discordgo.Session, event *discordgo.Event) {
	go func() {
		time.Sleep(time.Second * 2)
		guild, e := s.Guild(cfg.GuildID)
		if e != nil {
			log.Println("Error:", e)
		}
		for _, p := range guild.Presences {
			correctRoles(s, p)
		}
		listSounds(s)
	}()
}
func messageCreate(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if !mc.Author.Bot {
		if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"forget ") {
			if authenticate(s, mc.Author) {
				log.Println("Authenticated user:", mc.Author.Username)
				forget(s, mc)
			} else {
				log.Println("Failed to authenticate user:", mc.Author.Username)
			}
		}
		if strings.HasPrefix(mc.Content, cfg.SoundboardCommandKey) {
			listSounds(s)
			channel, e := s.Channel(mc.ChannelID)
			if e != nil {
				log.Println("Error:", e)
			}
			if channel.Name != cfg.CommandChannelName {
				e := s.ChannelMessageDelete(channel.ID, mc.ID)
				if e != nil {
					log.Println("Couldn't delete message:", mc.Content)
					log.Println("Error:", e)
				}
			}
			if mc.Content == cfg.SoundboardCommandKey+"stop" {
				if playing {
					dgvoice.KillPlayer()
					playing = false
					return
				} else {
					return
				}
			}
			soundString := strings.Replace(mc.Content, cfg.SoundboardCommandKey, "", 1) + ".mp3"
			sounds, _ = ioutil.ReadDir("sounds")
			for _, f := range sounds {
				if f.Name() == soundString {
					if playing {
						dgvoice.KillPlayer()
						playing = false
					}
					playSound(s, mc.Author, soundString)
				}
			}
		}
		if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"help") {
			if authenticate(s, mc.Author) {
				help(s, mc)
			}
		}
	}
}
func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	if authenticate(s, p.User) || (p.User.Bot) {
		log.Println("Admin/Bot user, cannot modify role")
		return
	}
	var updatedRoles []string
	var role string

	guildRoles, e := s.GuildRoles(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, gr := range guildRoles {
		if p.Game != nil {
			if gr.Name == p.Game.Name {
				role = gr.Name
				updatedRoles = append(updatedRoles, gr.ID)
			}
		} else {
			guild, e := s.Guild(cfg.GuildID)
			if e != nil {
				log.Println("Error:", e)
			}
			role = guild.Name
			updatedRoles = append(updatedRoles, guild.ID)
		}
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
		guild, e := s.Guild(cfg.GuildID)
		if e != nil {
			log.Println("Error:", e)
		}
		role = guild.Name
		updatedRoles = append(updatedRoles, guild.ID)
		log.Println("No role by name \"Other Games\", putting user in default role")
	}
	log.Println("Changing user role to:", role)
	if e := s.GuildMemberEdit(cfg.GuildID, p.User.ID, updatedRoles); e != nil {
		log.Println("Error:", e)
	}
}
func voiceState(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	user, e := s.GuildMember(cfg.GuildID, vsu.UserID)
	if e != nil {
		log.Println("Error:", e)
	}
	if !user.User.Bot && !vsu.SelfMute && !vsu.SelfDeaf {
		file := (strings.ToLower(user.User.Username) + ".mp3")
		playSound(s, user.User, file)
	}
}
func correctRoles(s *discordgo.Session, p *discordgo.Presence) {
	if authenticate(s, p.User) || (p.User.Bot) {
		log.Println("Admin/Bot user, cannot modify role")
		return
	}
	var updatedRoles []string
	var role string

	guildRoles, e := s.GuildRoles(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, gr := range guildRoles {
		if p.Game != nil {
			if gr.Name == p.Game.Name {
				role = gr.Name
				updatedRoles = append(updatedRoles, gr.ID)
			}
		} else {
			guild, e := s.Guild(cfg.GuildID)
			if e != nil {
				log.Println("Error:", e)
			}
			role = guild.Name
			updatedRoles = append(updatedRoles, guild.ID)
		}
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
		guild, e := s.Guild(cfg.GuildID)
		if e != nil {
			log.Println("Error:", e)
		}
		role = guild.Name
		updatedRoles = append(updatedRoles, guild.ID)
		log.Println("No role by name \"Other Games\", putting user in default role")
	}
	log.Println("Changing user role to:", role)
	if e := s.GuildMemberEdit(cfg.GuildID, p.User.ID, updatedRoles); e != nil {
		log.Println("Error:", e)
	}
}
func authenticate(s *discordgo.Session, u *discordgo.User) bool {
	user, e := s.GuildMember(cfg.GuildID, u.ID)
	if e != nil {
		log.Println("Error:", e)
		return false
	}
	roles, e := s.GuildRoles(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
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
func playSound(s *discordgo.Session, user *discordgo.User, file string) {
	guild, e := s.Guild(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, v := range guild.VoiceStates {
		if v.UserID == user.ID {
			if v.ChannelID != "" {
				var e error
				vc, e := s.ChannelVoiceJoin(cfg.GuildID, v.ChannelID, false, false)
				if e != nil {
					log.Println("Error:", e)
					return
				}
				log.Println("Attempting to play audio file \"" + file + "\" for user: " + user.Username)
				if playing {
					dgvoice.KillPlayer()
					playing = false
				}
				dgvoice.PlayAudioFile(vc, ("sounds/" + file))
				dgvoice.KillPlayer()
			} else {
				return
			}
		}
	}
}
func listSounds(s *discordgo.Session) {
	var sounds string
	files, _ := ioutil.ReadDir("sounds")
	for _, f := range files {
		if !strings.HasSuffix(strings.ToLower(f.Name()), ".ogg") {
			sounds += cfg.SoundboardCommandKey + strings.Replace(f.Name(), ".mp3", "", 1) + "\n"
		}
	}
	guildChannels, e := s.GuildChannels(cfg.GuildID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, c := range guildChannels {
		if cfg.CommandChannelName == c.Name {
			m, e := s.ChannelMessageEdit(c.ID, cfg.SoundboardMessageID, sounds)
			if e != nil {
				log.Println("Couldn't edit message:", m.Content)
				log.Println("Error:", e)
				return
			} else {
				log.Println("Sounds Updated")
			}
		}
	}
}
func forget(session *discordgo.Session, mc *discordgo.MessageCreate) {
	channel, e := session.Channel(mc.ChannelID)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	var untilMsg = strings.Replace(mc.Content, "*forget ", "", -1)
	messages, e := session.ChannelMessages(mc.ChannelID, 100, channel.LastMessageID, untilMsg)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	for _, m := range messages {
		if m.ID != cfg.SoundboardMessageID {
			if e := session.ChannelMessageDelete(mc.ChannelID, m.ID); e != nil {
				log.Println("Couldn't delete message: " + m.Content)
				log.Println("Error:", e)
				return
			} else {
				log.Println("Deleting message: " + m.Content)
			}
		}
	}
}
func help(session *discordgo.Session, mc *discordgo.MessageCreate) {
	if m, e := session.ChannelMessageSend(mc.ChannelID, (cfg.AdminCommandKey + "`help - Shows this dialog.`")); e != nil {
		log.Println("Could not post message: ", m)
		log.Println("Error:", e)
	}
	if m, e := session.ChannelMessageSend(mc.ChannelID, ("`" + cfg.AdminCommandKey + "forget [messageID] - Deletes messages down to a specified message.`")); e != nil {
		log.Println("Could not post message: ", m)
		log.Println("Error:", e)
	}
}
