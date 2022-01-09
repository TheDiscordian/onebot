// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/TheDiscordian/onebot/onelib"
	"github.com/lunixbochs/struc"
)

const (
	// NAME is same as filename, minus extension
	NAME = "d2lb"
	// LONGNAME is what's presented to the user
	LONGNAME = "Diablo II Leaderboards"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

var (
	// Path to ladder.D2DV
	d2lbLadderPath string
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	d2lbLadderPath = onelib.GetTextConfig(NAME, "ladder_path")
	return new(D2LBPlugin)
}

func formatXp(xp int) string {
	outMsg := "%d"
	if xp >= 1000000 {
		outMsg = "%dK"
		xp /= 1000
	}
	p := message.NewPrinter(language.English)
	return p.Sprintf(outMsg, xp)
}

func createTable(title string, chars []*CharInfo) (text, formattedText string) {
	text = fmt.Sprintf("Diablo II %s Ladder:\n", title)
	formattedText = fmt.Sprintf("<strong>Diablo II %s Ladder:</strong><br />\n<table><tr><th> # </th><th> Name </th><th> Class </th><th> Level </th><th> XP </th></tr><br />\n", title)
	for i, char := range chars { // getTitle(c Class, expansion bool, difficulty int, hardcore bool)
		text += fmt.Sprintf("    %d. %s [%s] Lvl.%d (%sxp)\n", i+1, getTitle(char.Class, char.expansion, char.difficulty, char.hardcore)+char.CharName, classToString(char.Class), char.Level, formatXp(char.Experience))
		var highlight, highlightCloser string
		switch i {
		case 0:
			highlight = `<font color="#C9B037">`
			highlightCloser = "</font>"
		case 1:
			highlight = `<font color="#D7D7D7">`
			highlightCloser = "</font>"
		case 2:
			highlight = `<font color="#6A3805">`
			highlightCloser = "</font>"
		}
		formattedText += fmt.Sprintf("<tr><td> %s%d%s  </td><th>  <strong>%s%s%s</strong>  </th><td> %s  </td><td> %d  </td><td> %s</td></tr>\n", highlight, i+1, highlightCloser, highlight, getTitle(char.Class, char.expansion, char.difficulty, char.hardcore)+char.CharName, highlightCloser, classToString(char.Class), char.Level, formatXp(char.Experience))
		if i <= 2 {
			formattedText += "</font>"
		}
	}
	formattedText += "</table>"
	return
}

func d2xpLb(msg onelib.Message, sender onelib.Sender) {
	_, _, exp, _ := getLeaderboards(10)
	text, formattedText := createTable("Expansion", exp)
	sender.Location().SendFormattedText(text, formattedText)
}

func d2xphcLb(msg onelib.Message, sender onelib.Sender) {
	_, _, _, exphc := getLeaderboards(10)
	text, formattedText := createTable("Expansion Hardcore", exphc)
	sender.Location().SendFormattedText(text, formattedText)
}

func d2Lb(msg onelib.Message, sender onelib.Sender) {
	d2, _, _, _ := getLeaderboards(10)
	text, formattedText := createTable("Standard", d2)
	sender.Location().SendFormattedText(text, formattedText)
}

func d2hcLb(msg onelib.Message, sender onelib.Sender) {
	_, hc, _, _ := getLeaderboards(10)
	text, formattedText := createTable("Standard Hardcore", hc)
	sender.Location().SendFormattedText(text, formattedText)
}

func d2AllLb(msg onelib.Message, sender onelib.Sender) {
	d2, hc, exp, exphc := getLeaderboards(3)
	text, formattedText := createTable("Standard", d2)
	text2, formattedText2 := createTable("Standard Hardcore", hc)
	text3, formattedText3 := createTable("Expansion", exp)
	text4, formattedText4 := createTable("Expansion Hardcore", exphc)
	sender.Location().SendFormattedText(text+"\n"+text2+"\n"+text3+"\n"+text4, formattedText+"<br>"+formattedText2+"<br>"+formattedText3+"<br>"+formattedText4)
}

// D2LBPlugin is an object for satisfying the Plugin interface.
type D2LBPlugin int

// Name returns the name of the plugin, usually the filename.
func (dp *D2LBPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (dp *D2LBPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (dp *D2LBPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (dp *D2LBPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"d2xp": d2xpLb, "d2": d2Lb, "d2hc": d2hcLb, "d2xphc": d2xphcLb, "d2all": d2AllLb}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (dp *D2LBPlugin) Remove() {
}

type LadderHeader struct {
	MaxType  int `struc:"int32,little"`
	Checksum int `struc:"int32,little"`
}

type CharInfo struct {
	Experience    int    `struc:"int32,little"`
	Status        int    `struc:"int8,little"`
	ActsCompleted int    `struc:"int8,little"` // this doesn't seem to update after completing act IV : /
	Level         int    `struc:"int8,little"`
	Class         Class  `struc:"int8,little"`
	CharName      string `struc:"[16]int8,little"`

	// statuses
	expansion  bool
	difficulty int
	hardcore   bool
	dead       bool
}

type LadderIndex struct {
	Type   int `struc:"int32,little"`
	Offset int `struc:"int32,little"`
	Number int `struc:"int32,little"`
}

const (
	ExpansionBit = 1 << 5
	HardcoreBit  = 1 << 2
	DeadBit      = 1 << 3
)

type Class byte

const (
	Amazon Class = iota
	Sorceress
	Necromancer
	Paladin
	Barbarian
	Druid
	Assassin
)

func classToString(c Class) string {
	switch c {
	case Amazon:
		return "ama"
	case Sorceress:
		return "sor"
	case Necromancer:
		return "nec"
	case Paladin:
		return "pal"
	case Barbarian:
		return "bar"
	case Druid:
		return "dru"
	case Assassin:
		return "asn"
	}
	return ""
}

func getGender(c Class) string {
	switch c {
	case Amazon, Sorceress, Assassin:
		return "f"
	default:
		return "m"
	}
}

func getTitle(c Class, expansion bool, difficulty int, hardcore bool) string {
	gender := getGender(c)
	if expansion {
		switch difficulty {
		case 1:
			if !hardcore {
				return "Slayer "
			} else {
				return "Destroyer "
			}
		case 2:
			if !hardcore {
				return "Champion "
			} else {
				return "Conqueror "
			}
		case 3:
			if hardcore {
				return "Guardian "
			} else {
				if gender == "f" {
					return "Matriarch "
				} else {
					return "Patriarch "
				}
			}
		}
	} else {
		if gender == "f" {
			switch difficulty {
			case 1:
				if !hardcore {
					return "Dame "
				} else {
					return "Countess "
				}
			case 2:
				if !hardcore {
					return "Lady "
				} else {
					return "Duchess "
				}
			case 3:
				if !hardcore {
					return "Baroness "
				} else {
					return "Queen "
				}
			}
		} else {
			switch difficulty {
			case 1:
				if !hardcore {
					return "Sir "
				} else {
					return "Count "
				}
			case 2:
				if !hardcore {
					return "Lord "
				} else {
					return "Duke "
				}
			case 3:
				if !hardcore {
					return "Baron "
				} else {
					return "King "
				}
			}
		}
	}
	return ""
}

// count must be positive and non-zero, or you won't get any results
func getLeaderboards(count int) (norm, hc, exp, expHc []*CharInfo) {
	f, err := os.Open(d2lbLadderPath)
	if err != nil {
		onelib.Error.Println(err)
		return
	}
	ladderHeader := new(LadderHeader)
	err = struc.Unpack(f, ladderHeader)
	if err != nil {
		onelib.Error.Println(err)
		return
	}

	// Iterate past ... whatever this data is
	for i := 0; i < ladderHeader.MaxType; i++ {
		ladderIndex := new(LadderIndex)
		err = struc.Unpack(f, ladderIndex)
		if err != nil {
			onelib.Error.Println(err)
			return
		}
	}

	norm = make([]*CharInfo, 0, count)
	hc = make([]*CharInfo, 0, count)
	exp = make([]*CharInfo, 0, count)
	expHc = make([]*CharInfo, 0, count)

	// Throw all the chars into a map, because there will be duplicates
	chars := make(map[string]*CharInfo)
	charInfo := new(CharInfo)
	var i int
	for ; err == nil; err = struc.Unpack(f, charInfo) {
		charInfo.CharName = strings.Replace(charInfo.CharName, "\x00", "", -1)
		if charInfo.CharName == "" || charInfo.CharName[0] == 0 || chars[charInfo.CharName] != nil {
			continue
		}
		if err == nil {
			i++
			if charInfo.Status&ExpansionBit == ExpansionBit {
				charInfo.expansion = true
			}
			if charInfo.Status&HardcoreBit == HardcoreBit {
				charInfo.hardcore = true
				if charInfo.Status&DeadBit == DeadBit {
					charInfo.dead = true
				}
			}

			//fmt.Printf("%+v\n", charInfo)
			if !charInfo.expansion {
				charInfo.difficulty = charInfo.ActsCompleted / 4
				if !charInfo.hardcore && len(norm) < count {
					norm = append(norm, charInfo)
				} else if charInfo.hardcore && len(hc) < count {
					hc = append(hc, charInfo)
				}
			} else {
				charInfo.difficulty = charInfo.ActsCompleted / 5
				if !charInfo.hardcore && len(exp) < count {
					exp = append(exp, charInfo)
				} else if charInfo.hardcore {
					if len(expHc) >= count {
						return
					}
					expHc = append(expHc, charInfo)
				}
			}
			chars[charInfo.CharName] = charInfo
			charInfo = new(CharInfo)
		}
	}
	return
}
