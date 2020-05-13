// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	NAME     = "firstprotocol"  // Same as filename, minus extension
	LONGNAME = "First Protocol" // Name presented to user
	VERSION  = 0                // Version of the script (higher regarded as newer)
)

func Load() onelib.Protocol {
	/*
	   Code to be executed on-load goes here (connects)
	*/
	return onelib.Protocol(&FirstProtocol{prefix: onelib.DefaultPrefix, nickname: onelib.DefaultNickname})
}

type FirstProtocol struct {
	/*
	   Store useful data here such as connected rooms, admins, nickname, accepted prefixes, etc
	*/
	prefix   string
	nickname string
}

func (fp *FirstProtocol) Name() string {
	return NAME
}

func (fp *FirstProtocol) LongName() string {
	return LONGNAME
}

func (fp *FirstProtocol) Version() int {
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

// recv should be called after you've recieved data and built a Message object
func (fp *FirstProtocol) recv(msg onelib.Message, sender onelib.Sender) {
	onelib.ProcessMessage(fp.prefix, msg, sender)
}

func (fp *FirstProtocol) Remove() {
	/*
	   Unload code goes here (disconnects)
	*/
}
