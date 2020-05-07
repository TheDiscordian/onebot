package main

import (
	. "github.com/TheDiscordian/onebot/onetype"
)

const (
	NAME     = "firstplugin"  // Same as filename, minus extension
	LONGNAME = "First Plugin" // Name presented to user
	VERSION  = 0              // Version of the script (higher regarded as newer)
)

func Load() Plugin {
	/*
	   Code to be executed on-load goes here
	*/
	return new(FirstPlugin)
}

type FirstPlugin struct {
}

func (fp *FirstPlugin) Name() string {
	return NAME
}

func (fp *FirstPlugin) LongName() string {
	return LONGNAME
}

func (fp *FirstPlugin) Version() int {
	return VERSION
}

func (fp *FirstPlugin) Implements() ([]Command, Monitor) {
	return nil, nil
}

func (fp *FirstPlugin) Remove() {
	/*
	   Unload code goes here (if any)
	*/
}
