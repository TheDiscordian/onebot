// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"

	"github.com/TheDiscordian/onebot/onelib"
)

// Not sure how much flexibility I plan to add to this.

const (
	// NAME is same as filename, minus extension
	NAME = "bnetdbridge"
	// LONGNAME is what's presented to the user
	LONGNAME = "BNetd Bridge"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

var (
	bnetdChannel string // channel to bridge
	bnetdDest    string // json `[{"protocol":"matrix","channel":"!example:matrix.thedisco.zone"}]`

	monitor *onelib.Monitor
)

func loadConfig() {
	bnetdChannel = onelib.GetTextConfig(NAME, "channel")
	bnetdDest = onelib.GetTextConfig(NAME, "dest")
}

// Load returns the Plugin object.
func Load() onelib.Plugin {
	loadConfig()
	channels := make([]*BridgedChannel, 0, 1)
	err := json.Unmarshal([]byte(bnetdDest), &channels)
	if err != nil {
		onelib.Error.Println(err)
	}
	return &BNetdBridge{Channels: channels}
}

type BridgedChannel struct {
	Protocol string
	Channel  string
}

// BNetdBridge is an object for satisfying the Plugin interface.
type BNetdBridge struct {
	Channels []*BridgedChannel
}

// Name returns the name of the plugin, usually the filename.
func (bnb *BNetdBridge) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (bnb *BNetdBridge) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (bnb *BNetdBridge) Version() string {
	return VERSION
}

func (bnb *BNetdBridge) OnMessageWithText(from onelib.Sender, msg onelib.Message) {
	if from.Protocol() == "irc_bnetd" {
		if from.Location().UUID() != onelib.UUID(bnetdChannel) {
			return
		}
		for _, chn := range bnb.Channels {
			proto := onelib.Protocols.Get(chn.Protocol)
			proto.SendText(onelib.UUID(chn.Channel), fmt.Sprintf("[%s] %s", from.DisplayName(), msg.Text()))
		}
	} else {
		ircBnetd := onelib.Protocols.Get("irc_bnetd")
		for _, chn := range bnb.Channels {
			if from.Protocol() == chn.Protocol && from.Location().UUID() == onelib.UUID(chn.Channel) {
				ircBnetd.SendText(onelib.UUID(bnetdChannel), fmt.Sprintf("[%s] %s", from.DisplayName(), msg.Text()))
				return
			}
		}
	}
}

// Implements returns a map of commands and monitor the plugin implements.
func (bnb *BNetdBridge) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	monitor = &onelib.Monitor{
		OnMessageWithText: bnb.OnMessageWithText,
		OnMessageUpdate:   bnb.OnMessageWithText,
	}
	return nil, monitor
}

// Remove is called when the plugin is about to be terminated.
func (bnb *BNetdBridge) Remove() {
	onelib.Monitors.Delete(monitor)
}
