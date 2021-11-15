// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "firstplugin"
	// LONGNAME is what's presented to the user
	LONGNAME = "First Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	/*
	   Code to be executed on-load goes here
	*/
	return new(FirstPlugin)
}

// FirstPlugin is an object for satisfying the Plugin interface.
type FirstPlugin int

// Name returns the name of the plugin, usually the filename.
func (fp *FirstPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (fp *FirstPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (fp *FirstPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (fp *FirstPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return nil, nil
}

// Remove is called when the plugin is about to be terminated.
func (fp *FirstPlugin) Remove() {
	/*
	   Unload code goes here (if any)
	*/
}
