// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "parrot"
	// LONGNAME is what's presented to the user
	LONGNAME = "Parrot Plugin"
	// VERSION of the script
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(ParrotPlugin)
}

func parrot(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendText(msg.Text())
}

// ParrotPlugin is an object for satisfying the Plugin interface.
type ParrotPlugin int

// Name returns the name of the plugin, usually the filename.
func (pp *ParrotPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (pp *ParrotPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (pp *ParrotPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (pp *ParrotPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"say": parrot}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (pp *ParrotPlugin) Remove() {
}
