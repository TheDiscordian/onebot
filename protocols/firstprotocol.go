// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "firstprotocol"
	// LONGNAME is what's presented to the user
	LONGNAME = "First Protocol"
	// VERSION of the script
	VERSION = "v0.0.0"
)

func loadConfig() {
	// firstProtocolServer = onelib.GetTextConfig(NAME, "server")
}

// Load connects to FirstProtocol, and sets up listeners. It's required for OneBot.
func Load() onelib.Protocol {
	loadConfig()
	/*
	   Code to be executed on-load goes here (connects)
	*/
	return onelib.Protocol(&FirstProtocol{prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname})
}

// FirstProtocol is the Protocol object used for handling anything FirstProtocol related.
type FirstProtocol struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
}

// Name returns the name of the plugin, usually the filename.
func (fp *FirstProtocol) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (fp *FirstProtocol) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (fp *FirstProtocol) Version() string {
	return VERSION
}

// NewMessage should generate a message object from something
func (fp *FirstProtocol) NewMessage(raw []byte) onelib.Message {
	return nil
}

// Send sends a Message object to a location specified by to (usually a location or sender UUID).
func (fp *FirstProtocol) Send(to onelib.UUID, msg onelib.Message) {
	// code here
}

// SendText sends text to a location specified by to (usually a location or sender UUID).
func (fp *FirstProtocol) SendText(to onelib.UUID, text string) {
	// code here
}

// SendFormattedText sends formatted text to a location specified by to (usually a location or sender UUID).
func (fp *FirstProtocol) SendFormattedText(to onelib.UUID, text, formattedText string) {
	// code here
}

// recv should be called after you've recieved data and built a Message object
func (fp *FirstProtocol) recv(msg onelib.Message, sender onelib.Sender) {
	onelib.ProcessMessage(fp.prefix, msg, sender)
}

// Remove should disconnect any open connections making it so the bot can forget about the protocol cleanly.
func (fp *FirstProtocol) Remove() {
	/*
	   Unload code goes here (disconnects)
	*/
}
