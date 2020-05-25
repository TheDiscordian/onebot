// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"strings"

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
	// discordAuthUser
	discordAuthUser string
	// discordAuthToken if blank, falls back onto pass
	discordAuthToken string
	// discordAuthPass
	discordAuthPass string
)

func loadConfig() {
	discordAuthToken = onelib.GetTextConfig(NAME, "auth_token")
}

type discordProtocolMessage struct {
	Format        string `json:"format"`
	Msgtype       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body"`
}

// Load connects to Discord, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()

	if discordAuthToken == "" {
		onelib.Error.Panicln("discordAuthToken can't be blank.")
	}
	client, err := discordgo.New("Bot " + discordAuthToken)
	if err != nil {
		onelib.Error.Panicln(err)
	}

	discordSession := &Discord{client: &discordClient{Session: client}, prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname}

	client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { // OnMessageCreate...
		if m.Type == discordgo.MessageTypeDefault {
			msg := &discordMessage{text: m.Content}
			dc := &discordClient{Session: client}
			dl := &discordLocation{Client: dc, uuid: onelib.UUID(m.ChannelID)}
			var displayName string
			if m.Member != nil && m.Member.Nick != "" {
				displayName = m.Member.Nick
			} else if m.Author != nil {
				displayName = m.Author.Username
			}
			var sender *discordSender
			var user *discordgo.User
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

	discordSession.client.Open()

	return onelib.Protocol(discordSession)
}

type discordMessage struct {
	formattedText, text string
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
	location              *discordLocation
	uuid                  onelib.UUID
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

type discordLocation struct {
	Client                                 *discordClient // pointer to originating client
	displayName, nickname, topic, protocol string
	uuid                                   onelib.UUID
}

func (ml *discordLocation) DisplayName() string {
	return ml.displayName
}

func (ml *discordLocation) Nickname() string {
	return ml.nickname
}

func (ml *discordLocation) Topic() string {
	return ml.topic
}

func (ml *discordLocation) UUID() onelib.UUID {
	return ml.uuid
}

func (ml *discordLocation) Send(msg onelib.Message) {
	ml.Client.Send(ml.uuid, msg)
}

func (ml *discordLocation) SendText(text string) {
	ml.Client.SendText(ml.uuid, text)
}

func (ml *discordLocation) SendFormattedText(text, formattedText string) {
	ml.Client.SendFormattedText(ml.uuid, text, formattedText)
}

func (ml *discordLocation) Protocol() string {
	return NAME
}

type discordClient struct {
	*discordgo.Session
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (dc *discordClient) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (dc *discordClient) SendText(to onelib.UUID, text string) {
	_, err := dc.Session.ChannelMessageSend(string(to), text)
	if err != nil {
		onelib.Error.Println(err)
	}
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
// FIXME currently ignores formatted text
func (dc *discordClient) SendFormattedText(to onelib.UUID, text, formattedText string) {
	dc.SendText(to, text)
}

// Discord is the Protocol object used for handling anything Discord related.
type Discord struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
	client   *discordClient
}

// Name returns the name of the plugin, usually the filename.
func (discord *Discord) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (discord *Discord) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (discord *Discord) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (discord *Discord) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (discord *Discord) Send(to onelib.UUID, msg onelib.Message) {
	discord.client.Send(to, msg)
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (discord *Discord) SendText(to onelib.UUID, text string) {
	discord.client.SendText(to, text)
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (discord *Discord) SendFormattedText(to onelib.UUID, text, formattedText string) {
	discord.client.SendFormattedText(to, text, formattedText)
}

// GetUserDisplayName returns a user's display name from a UUID
//func (discord *Discord) GetUserDisplayName(uuid onelib.UUID) string

// recv should be called after you've recieved data and built a Message object
func (discord *Discord) recv(msg onelib.Message, sender onelib.Sender) {
	if string(sender.UUID()) != discordAuthUser {
		onelib.ProcessMessage(discord.prefix, msg, sender)
	}
}

// Remove currently doesn't do anything.
func (discord *Discord) Remove() {
	discord.client.Session.Close()
}
