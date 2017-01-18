package main

import (
	"C"
	"flag"
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
	botID       string
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
	dg.AddHandler(messageCreate)
	dg.AddHandler(presenceUpdate)
	dg.AddHandler(voiceState)
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
	if !isConfigured(event.Guild) {
		s.ChannelMessageSend(event.Guild.ID, "Not configured. Type `"+cfg.AdminCommandKey+"config` in the channel you would like to keep commands in.")
	}
	roleFix(s, event.Guild)
}
func messageCreate(s *discordgo.Session, mc *discordgo.MessageCreate) {
	g := getGuild(s, mc.ChannelID)
	config(g)

	/*
	 *	Admin commands
	 */

	//Configure via Discord
	if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"config") {
		if !isConfigured(g) {
			if authenticate(s, g.ID, mc.Author) {
				m, e := s.ChannelMessageSend(mc.ChannelID, "Listing available sounds and pinning message...")
				if err(e, "Couldn't post message: "+m.Content) {
					return
				}
				writeConfig(g, mc.ChannelID, m.ID)
				listSounds(s, g)
				e = s.ChannelMessagePin(m.ChannelID, m.ID)
				if err(e, "") {
					return
				}
			}
		}
	}

	//Corrects the roles of the current guild
	if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"fixroles ") {
		if authenticate(s, g.ID, mc.Author) {
			roleFix(s, g)
		} else {
			log.Println("Failed to authenticate user:", mc.Author.Username)
		}
	}
	//Deletes messages up to a specified message
	if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"forget ") {
		if authenticate(s, g.ID, mc.Author) {
			log.Println("Authenticated user:", mc.Author.Username)
			forget(s, mc)
		} else {
			log.Println("Failed to authenticate user:", mc.Author.Username)
		}
	}

	/*
	 *	Soundboard commands
	 */

	if strings.HasPrefix(mc.Content, cfg.SoundboardCommandKey) {
		channel, e := s.Channel(mc.ChannelID)
		if err(e, "") {
			return
		}
		if channel.ID != cfg.CommandChannelID {
			e := s.ChannelMessageDelete(channel.ID, mc.ID)
			if err(e, "Couldn't delete message:"+mc.Content) {
				return
			}
		}
		//Stop the player
		if mc.Content == cfg.SoundboardCommandKey+"stop" {
			KillPlayer()
			return
		}
		//Play a requested sound
		soundString := strings.Replace(mc.Content, cfg.SoundboardCommandKey, "", 1) + ".mp3"
		sounds, _ = ioutil.ReadDir(soundDir)
		for _, f := range sounds {
			if f.Name() == soundString {
				playSound(s, g, mc.Author, soundString)
			}
		}
	}
	//Post a list of admin commands
	if strings.HasPrefix(mc.Content, cfg.AdminCommandKey+"help") {
		if authenticate(s, g.ID, mc.Author) {
			help(s, mc)
		}
	}
}
func voiceState(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	guild, e := s.Guild(vsu.GuildID)
	if err(e, "") {
		return
	}
	config(guild)
	user, e := s.GuildMember(vsu.GuildID, vsu.UserID)
	if err(e, "") {
		return
	}

	file := (strings.ToLower(user.User.Username) + ".mp3")
	if _, e := os.Stat(soundDir + "/" + file); os.IsNotExist(e) {
		log.Println("No entrance for user: ", user.User.Username)
	} else {
		guild, e := s.Guild(vsu.GuildID)
		if err(e, "") {
			return
		}
		var botVC string
		for _, v := range guild.VoiceStates {
			if v.UserID == botID {
				botVC = v.ChannelID
			}
		}
		for _, v := range guild.VoiceStates {
			if v.UserID == user.User.ID {
				if v.ChannelID != botVC && v.ChannelID != "" {
					playSound(s, guild, user.User, file)
				} else {
					return
				}
			}
		}
	}
}
func playSound(s *discordgo.Session, g *discordgo.Guild, user *discordgo.User, file string) {
	config(g)
	var userVC string
	for _, v := range g.VoiceStates {
		if v.UserID == user.ID {
			userVC = v.ChannelID
		}
	}
	vc, e := s.ChannelVoiceJoin(g.ID, userVC, false, false)
	if err(e, "") {
		return
	}
	go func() {
		time.Sleep(time.Millisecond * 200)
		g, e = s.Guild(g.ID)
		for _, v := range g.VoiceStates {
			if v.UserID == s.State.User.ID {
				if v.ChannelID == userVC {
					listSounds(s, g)
					log.Println("Attempting to play audio file \"" + soundDir + "/" + file + "\" for user: " + user.Username)
					KillPlayer()
					time.Sleep(time.Millisecond * 200)
					PlayAudioFile(vc, (soundDir + "/" + file))
				}
			}
		}
	}()
}
func listSounds(s *discordgo.Session, g *discordgo.Guild) {
	config(g)
	var sounds string
	files, _ := ioutil.ReadDir(soundDir)
	for _, f := range files {
		sounds += cfg.SoundboardCommandKey + strings.Replace(f.Name(), ".mp3", "", 1) + "\n"
	}
	if sounds == "" {
		sounds = "(No mp3 files in directory: " + soundDir + ")"
	}
	m, e := s.ChannelMessageEdit(cfg.CommandChannelID, cfg.SoundboardMessageID, sounds)
	if err(e, "Couldn't edit soundboard message with ID: "+m.ID) {
		return
	} else {
		log.Println("Sounds Updated")
	}
}
