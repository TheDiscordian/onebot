// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	NAME     = "parrot"        // Same as filename, minus extension
	LONGNAME = "Parrot Plugin" // Name presented to user
	VERSION  = 0               // Version of the script (higher regarded as newer)
)

func Load() onelib.Plugin {
	/*
	   Code to be executed on-load goes here
	*/
	return new(ParrotPlugin)
}

func parrot(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendText(msg.Text())
}

type ParrotPlugin int

func (pp *ParrotPlugin) Name() string {
	return NAME
}

func (pp *ParrotPlugin) LongName() string {
	return LONGNAME
}

func (pp *ParrotPlugin) Version() int {
	return VERSION
}

func (pp *ParrotPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"say": parrot}, nil
}

func (pp *ParrotPlugin) Remove() {
	/*
	   Unload code goes here (if any)
	*/
}
