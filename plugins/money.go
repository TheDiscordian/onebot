// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/libs/onecurrency"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"sync"
	"time"
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
	rand.Seed(time.Now().UnixNano())
	onecurrency.InitCurrency(DEFAULT_CURRENCY)
	UserActionMap = new(userMap)
	UserActionMap.uMap = make(map[string]lastAction, 1)
	UserActionMap.lock = new(sync.RWMutex)
	return new(MoneyPlugin)
}

var UserActionMap *userMap

type userMap struct {
	uMap map[string]lastAction
	lock *sync.RWMutex
}

func (um *userMap) Get(user, action string) time.Time {
	um.lock.RLock()
	if um.uMap[user] == nil {
		um.lock.RUnlock()
		return time.Time{}
	}
	lastTime := um.uMap[user][action]
	um.lock.RUnlock()
	return lastTime
}

func (um *userMap) Set(user, action string, lastTime time.Time) {
	um.lock.Lock()
	if um.uMap[user] == nil {
		um.uMap[user] = make(lastAction, 3)
	}
	um.uMap[user][action] = lastTime
	um.lock.Unlock()
}

// Update returns the time stored, and updates the stored time if longer than duration.
func (um *userMap) Update(user, action string, duration time.Duration) (lastTime time.Time) {
	um.lock.Lock()
	if um.uMap[user] == nil {
		um.uMap[user] = make(lastAction, 3)
	}
	lastTime = um.uMap[user][action]
	if time.Since(lastTime) >= duration {
		um.uMap[user][action] = time.Now()
	}
	um.lock.Unlock()
	return
}

type lastAction map[string]time.Time

func formatDuration(action string, d time.Duration) (text string, formattedText string) {
	if d.Hours() > 0 {
		text = fmt.Sprintf("You cannot %s for another %d hours, and %d minutes.", action, int(d.Hours()), int(d.Minutes())%60)
		formattedText = fmt.Sprintf("You cannot %s for another <strong>%d hours</strong>, and <strong>%d minutes</strong>.", action, int(d.Hours()), int(d.Minutes())%60)
	} else {
		text = fmt.Sprintf("You cannot %s for another %d minutes, and %d seconds.", int(d.Minutes()), action, int(d.Seconds())%60)
		formattedText = fmt.Sprintf("You cannot %s for another <strong>%d minutes</strong>, and <strong>%d seconds</strong>.", action, int(d.Minutes()), int(d.Seconds())%60)
	}
	return
}

func work(msg onelib.Message, sender onelib.Sender) {
	const (
		workMax  = 30
		workMin  = 15
		workTime = time.Hour * 23 // time until command can be called again
	)

	var text, formattedText string
	uuid := string(sender.UUID())
	if storedTime := UserActionMap.Update(uuid, "work", workTime); time.Since(storedTime) >= workTime {
		roll := rand.Intn(workMax-workMin) + workMin
		text = fmt.Sprintf("<insert funny message here>, and gain %s%d!", DEFAULT_CURRENCY, roll)
		formattedText = fmt.Sprintf("&lt;insert funny message here&gt;, and gain <strong>%s%d</strong>!", DEFAULT_CURRENCY, roll)
		_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, sender.DisplayName(), roll)
		if err != nil {
			onelib.Error.Println(err)
		}
	} else {
		timeUntil := time.Until(storedTime.Add(workTime))
		text, formattedText = formatDuration("work", timeUntil)
	}

	sender.Location().SendFormattedText(text, formattedText)
}

func slut(msg onelib.Message, sender onelib.Sender) {
	const (
		slutMax     = 40
		slutMin     = 15
		slutFineMax = 25
		slutFineMin = 5
		slutFail    = 3              // 1 in x of failure
		slutTime    = time.Hour * 12 // time until command can be called again
	)

	var text, formattedText string
	uuid := string(sender.UUID())
	if storedTime := UserActionMap.Update(uuid, "slut", slutTime); time.Since(storedTime) >= slutTime {
		if rand.Intn(slutFail) == 0 {
			roll := rand.Intn(slutFineMax-slutFineMin) + slutFineMin
			text = fmt.Sprintf("<insert funny message here>, and lose %s%d!", DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf("&lt;insert funny message here&gt;, and lose <strong>%s%d</strong>!", DEFAULT_CURRENCY, roll)
			_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, sender.DisplayName(), roll*-1)
			if err != nil {
				onelib.Error.Println(err)
			}
		} else {
			roll := rand.Intn(slutMax-slutMin) + slutMin
			text = fmt.Sprintf("<insert funny message here>, and gain %s%d!", DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf("&lt;insert funny message here&gt;, and gain <strong>%s%d</strong>!", DEFAULT_CURRENCY, roll)
			_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, sender.DisplayName(), roll)
			if err != nil {
				onelib.Error.Println(err)
			}
		}
	} else {
		timeUntil := time.Until(storedTime.Add(slutTime))
		text, formattedText = formatDuration("be a slut", timeUntil)
	}

	sender.Location().SendFormattedText(text, formattedText)
}

func crime(msg onelib.Message, sender onelib.Sender) {
	const (
		crimeMax     = 100
		crimeMin     = 40
		crimeFineMax = 55
		crimeFineMin = 25
		crimeFail    = 2             // 1 in x of failure
		crimeTime    = time.Hour * 8 // time until command can be called again
	)

	var text, formattedText string
	uuid := string(sender.UUID())
	if storedTime := UserActionMap.Update(uuid, "crime", crimeTime); time.Since(storedTime) >= crimeTime {
		if rand.Intn(crimeFail) == 0 {
			roll := rand.Intn(crimeFineMax-crimeFineMin) + crimeFineMin
			text = fmt.Sprintf("<insert funny message here>, and lose %s%d!", DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf("&lt;insert funny message here&gt;, and lose <strong>%s%d</strong>!", DEFAULT_CURRENCY, roll)
			_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, sender.DisplayName(), roll*-1)
			if err != nil {
				onelib.Error.Println(err)
			}
		} else {
			roll := rand.Intn(crimeMax-crimeMin) + crimeMin
			text = fmt.Sprintf("<insert funny message here>, and gain %s%d!", DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf("&lt;insert funny message here&gt;, and gain <strong>%s%d</strong>!", DEFAULT_CURRENCY, roll)
			_, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, sender.DisplayName(), roll)
			if err != nil {
				onelib.Error.Println(err)
			}
		}
	} else {
		timeUntil := time.Until(storedTime.Add(crimeTime))
		text, formattedText = formatDuration("commit a crime", timeUntil)
	}

	sender.Location().SendFormattedText(text, formattedText)
}

func checkBal(msg onelib.Message, sender onelib.Sender) {
	uuid := string(sender.UUID())
	cObj := onecurrency.Currency[DEFAULT_CURRENCY].Get(uuid)
	if cObj == nil {
		cObj = onecurrency.Currency[DEFAULT_CURRENCY].New(uuid, sender.DisplayName())
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
	return map[string]onelib.Command{"bal": checkBal, "balance": checkBal, "work": work, "slut": slut, "crime": crime}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (mp *MoneyPlugin) Remove() {
}
