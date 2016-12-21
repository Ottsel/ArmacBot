# ArmacBot
Golang Moderator and Soundboard Bot for Discord

# Prerequisites

* gcc
* ffmpeg
* opus

# How to setup

Install the bot and run once to generate a config file, then configure at least ```BotToken``` and ```GuildID```
```
go get github.com/ottsel/armacbot/
```
#Config options

## To copy IDs of things like messages, channels and guilds, enable developer mode under User settings -> Appearance. Now you can right click things and copy their IDs!

* `BotToken` - This is your bot user's token.
* `GuildID` - This is your Guild's `ID`.

* `SoundboardCommandKey` - This is the prefix of soundboard commands. (!SOUND_FILE_NAME)
* `AdminCommandKey` - This is the prefix of admin commands. (*help)

* `CommandChannelName` - This is the name of the channel that you want users to type commands into. (Messages in this channel won't be deleted)

* `SoundboardMessageID` - The `ID` of a pinned message in `CommandChannel` that displays all soundboard commands. (Type a message in `CommandChannel` and copy its `ID`)

# Adding soundboard commands and entrance sounds

 * To add a soundboard `command`, put a `mp3` file named what you want the `command` to be in the sounds folder
 * To add a sound file that plays when a specific `user` joins a voice channel, put a `mp3` file with the same name as the `user` in the sounds folder

#Thats it!

You wanted a note down here, didn't you?
