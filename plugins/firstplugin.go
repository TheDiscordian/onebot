// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	NAME     = "firstplugin"  // Same as filename, minus extension
	LONGNAME = "First Plugin" // Name presented to user
	VERSION  = 0              // Version of the script (higher regarded as newer)
)

func Load() onelib.Plugin {
	/*
	   Code to be executed on-load goes here
	*/
	return new(FirstPlugin)
}

type FirstPlugin int

func (fp *FirstPlugin) Name() string {
	return NAME
}

func (fp *FirstPlugin) LongName() string {
	return LONGNAME
}

func (fp *FirstPlugin) Version() int {
	return VERSION
}

func (fp *FirstPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return nil, nil
}

func (fp *FirstPlugin) Remove() {
	/*
	   Unload code goes here (if any)
	*/
}
