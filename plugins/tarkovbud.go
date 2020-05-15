// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "tarkovbud"
	// LONGNAME is what's presented to the user
	LONGNAME = "Tarkov Buddy Plugin"
	// VERSION of the script
	VERSION = "v0.0.1"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return new(TarkovBuddy)
}

// TarkovBuddy is a placeholder type, currently just used to satisfy Plugin interface
type TarkovBuddy int

// Name returns the name of the plugin, usually the filename.
func (tb *TarkovBuddy) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (tb *TarkovBuddy) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (tb *TarkovBuddy) Version() string {
	return VERSION
}

// tarkovMap searches gamemaps
func tarkovMap(msg onelib.Message, sender onelib.Sender) {
	interchange := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/e/e5/InterchangeMap_Updated_4.24.2020.png?version=c1114bd10889074ca8c8d85e3d1fb04b"
	reserve := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c0/ReserveMap3d.jpg?version=2b5fcc2b5f557535a42002e31c17c113"
	woods := "https://cdn.gamerjournalist.com/primary/2020/01/Escape-From-Tarkov-Woods-Map-Guide-2020-scaled.jpg"
	shoreline := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/d/d4/Shoreline_marvelin_2_updated.png?version=f28651df0d566bdc1996aeeacb496d15"
	labs := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/0/0b/TheLabMapFull.png?version=8743e690fbd315e114f51540347eca29"
	factory := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/8/83/Factory_0.8.png?version=91f04c0c3f62388c86e3fbebdd0abcdf"
	customs := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c8/Customs_Nuxx_20190106_1.2.png?version=a3b44edf49616eaad2736c6523c977b0"
	maphelp := "https://escapefromtarkov.gamepedia.com/Map_of_Tarkov"

	if txt := msg.Text(); txt == "interchange" {
		sender.Location().SendText(interchange)
	} else if txt == "reserve" {
		sender.Location().SendText(reserve)
	} else if txt == "woods" {
		sender.Location().SendText(woods)
	} else if txt == "shoreline" {
		sender.Location().SendText(shoreline)
	} else if txt == "labs" {
		sender.Location().SendText(labs)
	} else if txt == "factory" {
		sender.Location().SendText(factory)
	} else if txt == "customs" {
		sender.Location().SendText(customs)
	} else {
		sender.Location().SendText("Sorry, try this: " + maphelp)
	}
}

func tarkovBoss(msg onelib.Message, sender onelib.Sender) {
	// find boss searching by map
	customs := "https://escapefromtarkov.gamepedia.com/Reshala"
	reserve := "https://escapefromtarkov.gamepedia.com/Glukar"
	interchange := "https://escapefromtarkov.gamepedia.com/Killa"
	woods := "https://escapefromtarkov.gamepedia.com/Shturman"
	// find boss searching by name
	reshala := "https://escapefromtarkov.gamepedia.com/Reshala"
	glukar := "https://escapefromtarkov.gamepedia.com/Glukar"
	killa := "https://escapefromtarkov.gamepedia.com/Killa"
	shturman := "https://escapefromtarkov.gamepedia.com/Shturman"

	bosshelp := "https://escapefromtarkov.gamepedia.com/Characters#Bosses"

	if txt := msg.Text(); txt == "interchange" {
		sender.Location().SendText(interchange)
	} else if txt == "reserve" {
		sender.Location().SendText(reserve)
	} else if txt == "woods" {
		sender.Location().SendText(woods)
	} else if txt == "customs" {
		sender.Location().SendText(customs)
	} else if txt == "reshala" {
		sender.Location().SendText(reshala)
	} else if txt == "glukar" {
		sender.Location().SendText(glukar)
	} else if txt == "shturman" {
		sender.Location().SendText(shturman)
	} else if txt == "killa" {
		sender.Location().SendText(killa)
	} else {
		sender.Location().SendText("Sorry, try this: " + bosshelp)
	}
}

// Implements returns the function to call tarkovMap.
func (tb *TarkovBuddy) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"tarkov map": tarkovMap}, nil
	return map[string]onelib.Command{"tarkov boss": tarkovBoss}, nil
}

// Remove is required.
func (tb *TarkovBuddy) Remove() {
}
