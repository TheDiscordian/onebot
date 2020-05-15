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
	interchangeFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/e/e5/InterchangeMap_Updated_4.24.2020.png?version=c1114bd10889074ca8c8d85e3d1fb04b>Interchange</a>."
	reserve := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c0/ReserveMap3d.jpg?version=2b5fcc2b5f557535a42002e31c17c113"
	reserveFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c0/ReserveMap3d.jpg?version=2b5fcc2b5f557535a42002e31c17c113>Reserve</a>."
	woods := "https://cdn.gamerjournalist.com/primary/2020/01/Escape-From-Tarkov-Woods-Map-Guide-2020-scaled.jpg"
	woodsFormatted := "Map of <a href=https://cdn.gamerjournalist.com/primary/2020/01/Escape-From-Tarkov-Woods-Map-Guide-2020-scaled.jpg>Woods</a>."
	shoreline := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/d/d4/Shoreline_marvelin_2_updated.png?version=f28651df0d566bdc1996aeeacb496d15"
	shorelineFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/d/d4/Shoreline_marvelin_2_updated.png?version=f28651df0d566bdc1996aeeacb496d15>Shoreline</a>."
	labs := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/0/0b/TheLabMapFull.png?version=8743e690fbd315e114f51540347eca29"
	labsFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/0/0b/TheLabMapFull.png?version=8743e690fbd315e114f51540347eca29>Labs</a>."
	factory := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/8/83/Factory_0.8.png?version=91f04c0c3f62388c86e3fbebdd0abcdf"
	factoryFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/8/83/Factory_0.8.png?version=91f04c0c3f62388c86e3fbebdd0abcdf>Factory</a>."
	customs := "https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c8/Customs_Nuxx_20190106_1.2.png?version=a3b44edf49616eaad2736c6523c977b0"
	customsFormatted := "Map of <a href=https://gamepedia.cursecdn.com/escapefromtarkov_gamepedia/c/c8/Customs_Nuxx_20190106_1.2.png?version=a3b44edf49616eaad2736c6523c977b0>Customs</a>."
	maphelp := "Sorry, try this: https://escapefromtarkov.gamepedia.com/Map_of_Tarkov"
	maphelpFormatted := "Sorry, try this: <a href=https://escapefromtarkov.gamepedia.com/Map_of_Tarkov>Map of Tarkov</a>."

	if txt := msg.Text(); txt == "interchange" {
		sender.Location().SendFormattedText(interchange, interchangeFormatted)
	} else if txt == "reserve" {
		sender.Location().SendFormattedText(reserve, reserveFormatted)
	} else if txt == "woods" {
		sender.Location().SendFormattedText(woods, woodsFormatted)
	} else if txt == "shoreline" {
		sender.Location().SendFormattedText(shoreline, shorelineFormatted)
	} else if txt == "labs" {
		sender.Location().SendFormattedText(labs, labsFormatted)
	} else if txt == "factory" {
		sender.Location().SendFormattedText(factory, factoryFormatted)
	} else if txt == "customs" {
		sender.Location().SendFormattedText(customs, customsFormatted)
	} else {
		sender.Location().SendFormattedText(maphelp, maphelpFormatted)
	}
}

func tarkovBoss(msg onelib.Message, sender onelib.Sender) {
	// find boss searching by map
	customs := "https://escapefromtarkov.gamepedia.com/Reshala is the boss of Customs."
	customsFormatted := "<a href=https://escapefromtarkov.gamepedia.com/Reshala>Reshala</a> is the boss of Customs."
	reserve := "https://escapefromtarkov.gamepedia.com/Glukar is the boss of Reserve."
	reserveFormatted := "<a href=https://escapefromtarkov.gamepedia.com/Glukar>Glukar</a> is the boss of Reserve."
	interchange := "https://escapefromtarkov.gamepedia.com/Killa is the boss of Interchange."
	interchangeFormatted := "<a href=https://escapefromtarkov.gamepedia.com/Killa>Killa</a> is the boss of Interchange."
	woods := "https://escapefromtarkov.gamepedia.com/Shturman is the boss of Woods."
	woodsFormatted := "<a href=https://escapefromtarkov.gamepedia.com/Shturman>Shturman</a> is the boss of Woods."

	bosshelp := "Sorry, try this: https://escapefromtarkov.gamepedia.com/Characters#Bosses"
	bosshelpFormatted := "Sorry, try this: <a href=https://escapefromtarkov.gamepedia.com/Characters#Bosses>Characters#Bosses</a>"

	if txt := msg.Text(); txt == "interchange" || txt == "killa" {
		sender.Location().SendFormattedText(interchange, interchangeFormatted)
	} else if txt == "reserve" || txt == "glukar" {
		sender.Location().SendFormattedText(reserve, reserveFormatted)
	} else if txt == "woods" || txt == "shturman" {
		sender.Location().SendFormattedText(woods, woodsFormatted)
	} else if txt == "customs" || txt == "reshala" {
		sender.Location().SendFormattedText(customs, customsFormatted)
	} else {
		sender.Location().SendFormattedText(bosshelp, bosshelpFormatted)
	}
}

// Implements returns the function to call tarkovMap.
func (tb *TarkovBuddy) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"map": tarkovMap, "boss": tarkovBoss}, nil
}

// Remove is required.
func (tb *TarkovBuddy) Remove() {
}
