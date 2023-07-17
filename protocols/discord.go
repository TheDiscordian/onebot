// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"strings"

	"github.com/TheDiscordian/onebot/libs/discord"
	"github.com/TheDiscordian/onebot/onelib"
	"github.com/bwmarrin/discordgo"
)

const (
	// NAME is same as filename, minus extension
	NAME = "discord"
	// LONGNAME is what's presented to the user
	LONGNAME = "Discord"
	// VERSION of the script
	VERSION = "v0.0.0"
)

var (
	// discordAuthUser FIXME set this somewhere so we can filter out messages from ourselves (see: recv)
	discordAuthUser string
	// discordAuthToken if blank, falls back onto pass
	discordAuthToken string
	// discordAdminId is the UUID of the admin of the bot. This should probably be an array
	discordAdminId string

	// our userID
	discordId onelib.UUID
)

func loadConfig() {
	discordAuthToken = onelib.GetTextConfig(NAME, "auth_token")
	discord.DiscordAdminId = onelib.UUID(onelib.GetTextConfig(NAME, "admin_id"))
}

// Load connects to Discord, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()

	if discordAuthToken == "" {
		onelib.Error.Panicln("discordAuthToken can't be blank.")
	}
	client, err := discordgo.New("Bot " + discordAuthToken) // FIXME maybe make this a global of a "discord lib" (wrap discordgo client in our own struct, extending as needed)
	if err != nil {
		onelib.Error.Panicln(err)
	}

	discordSession := &Discord{client: &discord.DiscordClient{Session: client}, prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname}

	client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { // OnMessageCreate...
		if m.Type == discordgo.MessageTypeDefault {
			var (
				displayName string
				sender      *discordSender
				user        *discordgo.User
			)

			msg := &discordMessage{text: m.Content, id: onelib.UUID(m.ID)}
			dc := &discord.DiscordClient{Session: client}
			dl := &discord.DiscordLocation{Client: dc, Uuid: onelib.UUID(m.ChannelID), GuildID: onelib.UUID(m.GuildID)}
			if m.Member != nil && m.Member.Nick != "" {
				displayName = m.Member.Nick
			} else if m.Author != nil {
				displayName = m.Author.Username
			}
			if m.Author != nil {
				user = m.Author
			} else if m.Member != nil && m.Member.User != nil {
				user = m.Member.User
			}

			if user != nil {
				sender = &discordSender{uuid: onelib.UUID(user.ID), username: user.Username + "#" + user.Discriminator, displayName: displayName, location: dl}
			} else {
				onelib.Error.Println("Error processing message, contains no UUID:", m)
				return
			}
			discordSession.recv(onelib.Message(msg), onelib.Sender(sender))
			onelib.Debug.Printf("%s: %s\n", displayName, msg.Text())
		} else {
			onelib.Debug.Printf("Message (type: %v): %v\n", m.Type, m)
		}
	})

	client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		if m.Emoji.ID == "" {
			m.Emoji.ID = m.Emoji.Name
		}

		msg := &discordMessage{id: onelib.UUID(m.MessageID), emoji: &onelib.Emoji{Added: true, ID: onelib.UUID(m.Emoji.ID), Name: m.Emoji.Name}}
		dc := &discord.DiscordClient{Session: client}
		dl := &discord.DiscordLocation{Client: dc, Uuid: onelib.UUID(m.ChannelID), GuildID: onelib.UUID(m.GuildID)}
		sender := &discordSender{uuid: onelib.UUID(m.UserID), location: dl}

		discordSession.update(msg, sender)
	})

	client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
		if m.Emoji.ID == "" {
			m.Emoji.ID = m.Emoji.Name
		}

		msg := &discordMessage{id: onelib.UUID(m.MessageID), emoji: &onelib.Emoji{Added: false, ID: onelib.UUID(m.Emoji.ID), Name: m.Emoji.Name}}
		dc := &discord.DiscordClient{Session: client}
		dl := &discord.DiscordLocation{Client: dc, Uuid: onelib.UUID(m.ChannelID), GuildID: onelib.UUID(m.GuildID)}
		sender := &discordSender{uuid: onelib.UUID(m.UserID), location: dl}

		discordSession.update(msg, sender)
	})

	// Add a handler for the Ready event
	client.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		// Retrieve the user ID
		discordId = onelib.UUID(r.User.ID)
	})

	err = discordSession.client.Open()
	if err != nil {
		onelib.Error.Panicln(err)
		return nil
	}

	return onelib.Protocol(discordSession)
}

type discordMessage struct {
	id                  onelib.UUID
	formattedText, text string
	emoji               *onelib.Emoji
}

func (mm *discordMessage) UUID() onelib.UUID {
	return mm.id
}

func (mm *discordMessage) Reaction() *onelib.Emoji {
	return mm.emoji
}

func (mm *discordMessage) Text() string {
	return mm.text
}

func (mm *discordMessage) FormattedText() string {
	return mm.formattedText
}

func (mm *discordMessage) StripPrefix(prefix string) onelib.Message {
	if len(mm.text) > len(prefix) {
		prefix = prefix + " "
	}
	return onelib.Message(&discordMessage{text: strings.Replace(mm.text, prefix, "", 1), formattedText: strings.Replace(mm.formattedText, prefix, "", 1)})
}

func (mm *discordMessage) Raw() []byte {
	return []byte(mm.text)
}

type discordSender struct {
	displayName, username string
	location              *discord.DiscordLocation
	uuid                  onelib.UUID
}

func (ms *discordSender) Self() bool {
	return ms.uuid == discordId
}

func (ms *discordSender) DisplayName() string {
	return ms.displayName
}

func (ms *discordSender) Username() string {
	return ms.username
}

func (ms *discordSender) UUID() onelib.UUID {
	return ms.uuid
}

func (ms *discordSender) Location() onelib.Location {
	return ms.location
}

func (ms *discordSender) Protocol() string {
	return NAME
}

func (ms *discordSender) Send(msg onelib.Message) {
	ms.location.Client.Send(ms.uuid, msg)
}

func (ms *discordSender) SendText(text string) {
	ms.location.Client.SendText(ms.uuid, text)
}

func (ms *discordSender) SendFormattedText(text, formattedText string) {
	ms.location.Client.SendFormattedText(ms.uuid, text, formattedText)
}

// Discord is the Protocol object used for handling anything Discord related.
type Discord struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
	client   *discord.DiscordClient
}

// Name returns the name of the plugin, usually the filename.
func (dis *Discord) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (dis *Discord) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (dis *Discord) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (dis *Discord) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (dis *Discord) Send(to onelib.UUID, msg onelib.Message) {
	dis.client.Send(to, msg)
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (dis *Discord) SendText(to onelib.UUID, text string) {
	dis.client.SendText(to, text)
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (dis *Discord) SendFormattedText(to onelib.UUID, text, formattedText string) {
	dis.client.SendFormattedText(to, text, formattedText)
}

// GetUserDisplayName returns a user's display name from a UUID
//func (dis *Discord) GetUserDisplayName(uuid onelib.UUID) string

// recv should be called after you've recieved data and built a Message object
func (dis *Discord) recv(msg onelib.Message, sender onelib.Sender) {
	if string(sender.UUID()) != discordAuthUser {
		onelib.ProcessMessage([]string{dis.prefix}, msg, sender)
	}
}

// update should be called after you've recieved an edit or reaction
func (dis *Discord) update(msg onelib.Message, sender onelib.Sender) {
	if string(sender.UUID()) != discordAuthUser {
		onelib.ProcessUpdate(msg, sender)
	}
}

// Remove
func (dis *Discord) Remove() {
	dis.client.Session.Close()
}
