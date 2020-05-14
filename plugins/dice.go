// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"strconv"
)

const (
	NAME     = "dice"        // Same as filename, minus extension
	LONGNAME = "Dice Plugin" // Name presented to user
	// Version of the script
	VERSION = "v0.0.0"
)

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
		text = fmt.Sprintf("Rolls one die. Usage: %sroll <number of sides> (min 2)\n", onelib.DefaultPrefix)
	}
	sender.Location().SendText(text)
}

type DicePlugin int

func (dp *DicePlugin) Name() string {
	return NAME
}

func (dp *DicePlugin) LongName() string {
	return LONGNAME
}

func (dp *DicePlugin) Version() string {
	return VERSION
}

func (dp *DicePlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"roll": roll, "r": roll}, nil
}

func (dp *DicePlugin) Remove() {
}
