// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"strings"

	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "parrot"
	// LONGNAME is what's presented to the user
	LONGNAME = "Parrot Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	pp := new(ParrotPlugin)
	pp.msgs = make(map[onelib.UUID]string, 1)
	pp.monitor = &onelib.Monitor{
		OnMessageWithText: pp.OnMessageWithText,
	}
	return pp
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func revParrot(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendText(reverse(msg.Text()))
}

func parrot(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendFormattedText(msg.Text(), msg.FormattedText())
}

func formatParrot(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendFormattedText(msg.Text(), msg.Text())
}

// ParrotPlugin is an object for satisfying the Plugin interface.
type ParrotPlugin struct {
	monitor *onelib.Monitor
	msgs    map[onelib.UUID]string
}

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

func (pp *ParrotPlugin) OnMessageWithText(from onelib.Sender, msg onelib.Message) {
	var splitMsg []string
	txt := msg.Text()
	if strings.HasPrefix(txt, "s/") || strings.HasPrefix(txt, "ss/") {
		splitMsg = strings.Split(txt, "/")
		if len(splitMsg) < 2 {
			return
		}
	} else {
		pp.msgs[from.Location().UUID()] = txt
		return
	}
	loc := from.Location()
	outMsg := strings.ReplaceAll(pp.msgs[loc.UUID()], splitMsg[1], splitMsg[2])
	loc.SendText(outMsg)
}

// Implements returns a map of commands and monitor the plugin implements.
func (pp *ParrotPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"say": parrot, "s": parrot, "r": revParrot, "rev": revParrot, "format": formatParrot, "form": formatParrot}, pp.monitor
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (pp *ParrotPlugin) Remove() {
}
