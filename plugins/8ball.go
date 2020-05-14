// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
)

const (
	// NAME is same as filename, minus extension
	NAME = "8ball"
	// LONGNAME is what's presented to the user
	LONGNAME = "8Ball Plugin"
	// VERSION of the script
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(EightBallPlugin)
}

func eightball(msg onelib.Message, sender onelib.Sender) {
	text := msg.Text()
	if len(text) < 3 {
		text = fmt.Sprintf("Predicts the future. Usage: %s8ball <y/n question>", onelib.DefaultPrefix)
	} else {
		eightball_answers := []string{"As I see it, yes", "It is certain", "It is decidedly so", "Most likely",
			"Outlook good", "Signs point to yes", "Without a doubt", "Yes", "Yes, definitely",
			"You may rely on it", "Reply hazy, try again", "Ask again later",
			"Better not tell you now", "Cannot predict now", "Concentrate and ask again",
			"Don't count on it", "My reply is no", "My sources say no", "Outlook not so good",
			"Very doubtful"}
		text = eightball_answers[rand.Intn(len(eightball_answers))]
	}
	sender.Location().SendText(text)
}

// EightBallPlugin is an object for satisfying the Plugin interface.
type EightBallPlugin int

// Name returns the name of the plugin, usually the filename.
func (eb *EightBallPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (eb *EightBallPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (eb *EightBallPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (eb *EightBallPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"8b": eightball, "8ball": eightball}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (eb *EightBallPlugin) Remove() {
}
