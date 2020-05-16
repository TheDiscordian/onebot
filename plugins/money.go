// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/libs/onecurrency"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
)

const (
	// NAME is same as filename, minus extension
	NAME = "money"
	// LONGNAME is what's presented to the user
	LONGNAME = "Currency Plugin"
	// VERSION of the script
	VERSION = "v0.0.0"

	// DEFAULT_CURRENCY is the default currency symbol
	DEFAULT_CURRENCY = "★"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	onecurrency.InitCurrency(DEFAULT_CURRENCY)
	return new(MoneyPlugin)
}

func work(msg onelib.Message, sender onelib.Sender) {
	const work_max = 25
	const work_min = 15
	roll := rand.Intn(work_max-work_min) + work_min
	uuid := string(sender.UUID())
	text := fmt.Sprintf("<insert funny message here>, and gain %s%d!", DEFAULT_CURRENCY, roll)

	_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, uuid, roll) // FIXME replace second uuid with sender's display name
	if err != nil {
		onelib.Error.Println(err)
	}

	sender.Location().SendText(text)
}

func checkBal(msg onelib.Message, sender onelib.Sender) {
	uuid := string(sender.UUID())
	cObj := onecurrency.Currency[DEFAULT_CURRENCY].Get(uuid)
	if cObj == nil {
		cObj = onecurrency.Currency[DEFAULT_CURRENCY].New(uuid, uuid) // FIXME replace second parameter with sender's display name
	}
	text := fmt.Sprintf("%s's balance:\n    - Bal: %s%d\n    - Bank bal: %s%d", cObj.DisplayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity)
	formattedText := fmt.Sprintf("<p><strong>%s's balance:</strong><ul><li>Bal: %s%d</li>\n<li>Bank bal: %s%d</li></ul></p>", cObj.DisplayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity)
	sender.Location().SendFormattedText(text, formattedText)
}

// MoneyPlugin is an object for satisfying the Plugin interface.
type MoneyPlugin int

// Name returns the name of the plugin, usually the filename.
func (mp *MoneyPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (mp *MoneyPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (mp *MoneyPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (mp *MoneyPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"bal": checkBal, "balance": checkBal, "work": work}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (mp *MoneyPlugin) Remove() {
}
