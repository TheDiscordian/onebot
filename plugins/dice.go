// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"strconv"
)

const (
	// NAME is same as filename, minus extension
	NAME = "dice"
	// LONGNAME is what's presented to the user
	LONGNAME = "Dice Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(DicePlugin)
}

func roll(msg onelib.Message, sender onelib.Sender) {
	var text string
	droll := 0
	if mtext := msg.Text(); mtext == "" {
		droll = 20
	} else {
		droll, _ = strconv.Atoi(mtext)
	}
	if droll > 1 {
		text = fmt.Sprintf("You rolled a %d.", rand.Intn(droll)+1)
	} else {
		text = fmt.Sprintf("Rolls one die. Usage: %sroll <number of sides> (min 2)", onelib.DefaultPrefix)
	}
	sender.Location().SendText(text)
}

// DicePlugin is an object for satisfying the Plugin interface.
type DicePlugin int

// Name returns the name of the plugin, usually the filename.
func (dp *DicePlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (dp *DicePlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (dp *DicePlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (dp *DicePlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"roll": roll, "r": roll}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (dp *DicePlugin) Remove() {
}
