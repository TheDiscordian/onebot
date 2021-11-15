// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package discord

import (
	"github.com/TheDiscordian/onebot/onelib"
	"github.com/bwmarrin/discordgo"
)

var DiscordAdminId onelib.UUID

// Literally just to expose the session so other features of discordgo can be used if needed: https://pkg.go.dev/github.com/bwmarrin/discordgo
type DiscordLocation struct {
	Client                       *DiscordClient // pointer to originating client
	displayName, nickname, topic string
	Uuid                         onelib.UUID
	GuildID                      onelib.UUID // useful for roles
}

func (ml *DiscordLocation) DisplayName() string {
	return ml.displayName
}

func (ml *DiscordLocation) Nickname() string {
	return ml.nickname
}

func (ml *DiscordLocation) Topic() string {
	return ml.topic
}

func (ml *DiscordLocation) UUID() onelib.UUID {
	return ml.Uuid
}

func (ml *DiscordLocation) Send(msg onelib.Message) {
	ml.Client.Send(ml.Uuid, msg)
}

func (ml *DiscordLocation) SendText(text string) {
	ml.Client.SendText(ml.Uuid, text)
}

func (ml *DiscordLocation) SendFormattedText(text, formattedText string) {
	ml.Client.SendFormattedText(ml.Uuid, text, formattedText)
}

func (ml *DiscordLocation) Protocol() string {
	return "discord"
}

type DiscordClient struct {
	*discordgo.Session
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (dc *DiscordClient) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (dc *DiscordClient) SendText(to onelib.UUID, text string) {
	_, err := dc.Session.ChannelMessageSend(string(to), text)
	if err != nil {
		onelib.Error.Println(err)
	}
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
// FIXME currently ignores formatted text
func (dc *DiscordClient) SendFormattedText(to onelib.UUID, text, formattedText string) {
	dc.SendText(to, text)
}
