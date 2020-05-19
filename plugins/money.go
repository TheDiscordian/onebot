// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/libs/onecurrency"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"strings"
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
	UserActionMap.uMap = make(map[onelib.UUID]lastAction, 1)
	UserActionMap.lock = new(sync.RWMutex)
	return new(MoneyPlugin)
}

var UserActionMap *userMap

type userMap struct {
	uMap map[onelib.UUID]lastAction
	lock *sync.RWMutex
}

func (um *userMap) Get(user onelib.UUID, action string) time.Time {
	um.lock.RLock()
	if um.uMap[user] == nil {
		um.lock.RUnlock()
		return time.Time{}
	}
	lastTime := um.uMap[user][action]
	um.lock.RUnlock()
	return lastTime
}

func (um *userMap) Set(user onelib.UUID, action string, lastTime time.Time) {
	um.lock.Lock()
	if um.uMap[user] == nil {
		um.uMap[user] = make(lastAction, 3)
	}
	um.uMap[user][action] = lastTime
	um.lock.Unlock()
}

// Update returns the time stored, and updates the stored time if longer than duration.
func (um *userMap) Update(user onelib.UUID, action string, duration time.Duration) (lastTime time.Time) {
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
	if days := int(d.Hours() / 24); days > 0 {
		text = fmt.Sprintf("You cannot %s for another %d days, %d hours, and %d minutes.", action, days, int(d.Hours())%24, int(d.Minutes())%60)
		formattedText = fmt.Sprintf("You cannot %s for another <strong>%d days</strong>, <strong>%d hours</strong>, and <strong>%d minutes</strong>.", action, days, int(d.Hours())%24, int(d.Minutes())%60)
	} else if int(d.Hours()) > 0 {
		text = fmt.Sprintf("You cannot %s for another %d hours, and %d minutes.", action, int(d.Hours()), int(d.Minutes())%60)
		formattedText = fmt.Sprintf("You cannot %s for another <strong>%d hours</strong>, and <strong>%d minutes</strong>.", action, int(d.Hours()), int(d.Minutes())%60)
	} else {
		text = fmt.Sprintf("You cannot %s for another %d minutes, and %d seconds.", action, int(d.Minutes()), int(d.Seconds())%60)
		formattedText = fmt.Sprintf("You cannot %s for another <strong>%d minutes</strong>, and <strong>%d seconds</strong>.", action, int(d.Minutes()), int(d.Seconds())%60)
	}
	return
}

// TODO handle alias
func performAction(uuid onelib.UUID, actionName, actionText string, actionMinPayout, actionMaxPayout, actionMinFine, actionMaxFine, actionFailRate int, positiveResponses, negativeResponses [][2]string, actionCallRate time.Duration) (text string, formattedText string) {
	tuuid, _ := onecurrency.Currency[DEFAULT_CURRENCY].Get(uuid)
	if tuuid != onelib.UUID("") {
		uuid = tuuid
	}
	if storedTime := UserActionMap.Update(uuid, actionName, actionCallRate); time.Since(storedTime) >= actionCallRate {
		if actionFailRate > 1 && rand.Intn(actionFailRate) == 0 {
			roll := rand.Intn(actionMaxFine-actionMinFine) + actionMinFine
			negativeResponse := negativeResponses[rand.Intn(len(negativeResponses))]
			text = fmt.Sprintf(negativeResponse[0], DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf(negativeResponse[1], DEFAULT_CURRENCY, roll)
			_, _, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, roll*-1, 0)
			if err != nil {
				onelib.Error.Println(err)
			}
		} else {
			roll := rand.Intn(actionMaxPayout-actionMinPayout) + actionMinPayout
			positiveResponse := positiveResponses[rand.Intn(len(positiveResponses))]
			text = fmt.Sprintf(positiveResponse[0], DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf(positiveResponse[1], DEFAULT_CURRENCY, roll)
			_, _, err := onecurrency.Currency[DEFAULT_CURRENCY].Add(uuid, roll, 0)
			if err != nil {
				onelib.Error.Println(err)
			}
		}
	} else {
		timeUntil := time.Until(storedTime.Add(actionCallRate))
		text, formattedText = formatDuration(actionText, timeUntil)
	}
	return
}

func cute(msg onelib.Message, sender onelib.Sender) {
	const (
		cuteMax  = 300
		cuteMin  = 5
		cuteTime = time.Hour * 3 // time until command can be called again
	)
	cuteResponses := [][2]string{
		{"You help Miles build robots and gain %s%d!", "You help Miles build robots and gain <strong>%s%d</strong>!"},
		{"You go to a comedy show with Devin and gain %s%d!", "You go to a comedy show with Devin and gain <strong>%s%d</strong>!"},
		{"You spend all day painting chicken nuggets while eating chicken nuggets with Paulita and gain %s%d.", "You spend all day painting chicken nuggets while eating chicken nuggets with Paulita and gain <strong>%s%d</strong>."},
		{"You join Autumn on one of her nightly travels, gaining %s%d.", "You join Autumn on one of her nightly travels, gaining <strong>%s%d</strong>."},
		{"You go rock climbing with Randy the trash panda and gain %s%d!", "You go rock climbing with Randy the trash panda and gain <strong>%s%d</strong>!"},
		{"You watch movies with Lola and gain %s%d.", "You watch movies with Lola and gain <strong>%s%d</strong>."},
		{"Lucille shows you her rock collection and you gain %s%d from the experience!", "Lucille shows you her rock collection and you gain <strong>%s%d</strong> from the experience!"},
		{"You go surfing with Perry, gaining %s%d.", "You go surfing with Perry, gaining <strong>%s%d</strong>."},
		{"You join Francis the roar and go firefighting! You gained %s%d.", "You join Francis the roar and go firefighting! You gained <strong>%s%d</strong>."},
		{"You join Henry in doing commentating on sports games, you're paid %s%d!", "You join Henry in doing commentating on sports games, you're paid <strong>%s%d</strong>!"},
		{"You spend the day reading with Veronica and gain %s%d.", "You spend the day reading with Veronica and gain <strong>%s%d</strong>."},
		{"You help Aimee plan a red carpet even and get paid %s%d.", "You help Aimee plan a red carpet even and get paid <strong>%s%d</strong>."},
		{"You help Winston make a cullinary delight and gain %s%d.", "You help Winston make a cullinary delight and gain <strong>%s%d</strong>."},
		{"You spend the day doing artwork with Esmeralda and gain %s%d.", "You spend the day doing artwork with Esmeralda and gain <strong>%s%d</strong>."},
		{"You play soccer with Paco and gain %s%d.", "You play soccer with Paco and gain <strong>%s%d</strong>."},
		{"You study medicine with Jessica, surprisingly you're paid %s%d to do this!", "You study medicine with Jessica, surprisingly you're paid <strong>%s%d</strong> to do this!"},
		{"Hang out with Candy during the annual Squishmallow Egg Hunt, you gain %s%d!", "Hang out with Candy during the annual Squishmallow Egg Hunt, you gain <strong>%s%d</strong>!"},
		{"You go hiking with Rodrigo and gain %s%d.", "You go hiking with Rodrigo and gain <strong>%s%d</strong>."},
		{"You help design fashion with Matt the manitee and get paid %s%d for your designs!", "You help design fashion with Matt the manitee and get paid <strong>%s%d</strong> for your designs!"},
		{"You find and help identify treasure with Violet, your cut is %s%d.", "You find and help identify treasure with Violet, your cut is <strong>%s%d</strong>."},
		{"You help Zoe with restaurant planning. She pays you %s%d for your time.", "You help Zoe with restaurant planning. She pays you <strong>%s%d</strong> for your time."},
	}
	text, formattedText := performAction(sender.UUID(), "cute", "be cute", cuteMin, cuteMax, 0, 0, 0, cuteResponses, nil, cuteTime)
	sender.Location().SendFormattedText(text, formattedText)
}

func chill(msg onelib.Message, sender onelib.Sender) {
	const (
		chillMax     = 250
		chillMin     = 50
		chillFineMax = 150
		chillFineMin = 5
		chillFail    = 3                // 1 in x of failure
		chillTime    = time.Minute * 42 // time until command can be called again
	)
	chillResponses := [][2]string{
		{"Smoke weed with Snoop Dogg and gain %s%d!", "Smoke weed with Snoop Dogg and gain <strong>%s%d</strong>!"},
		{"You play some 100%% Orange Juice with your friends. It's a good time, you gain %s%d.", "You play some 100%% Orange Juice with your friends. It's a good time, you gain <strong>%s%d</strong>."},
		{"You roll a bad blunt, but Snoop Dogg is too high to notice! Good job, have %s%d!", "You roll a bad blunt, but Snoop Dogg is too high to notice! Good job, have <strong>%s%d</strong>!"},
		{"You drop acid and experience another reality. You gain %s%d.", "You drop acid and experience another reality. You gain <strong>%s%d</strong>."},
		{"You fuck your friend's mother and she pays YOU %s%d!", "You fuck your friend's mother and she pays YOU <strong>%s%d</strong>!"},
		{"The weed hit you just right, gain %s%d!", "The weed hit you just right, gain <strong>%s%d</strong>!"},
	}
	chillNegativeResponses := [][2]string{
		{"You roll a bad blunt and Snoop Dogg notices, pay a fine of %s%d...", "You roll a bad blunt and Snoop Dogg notices, pay a fine of <strong>%s%d</strong>..."},
		{"You play some 100%% Orange Juice with your friends. You come in last and lose %s%d.", "You play some 100%% Orange Juice with your friends. You come in last and lose <strong>%s%d</strong>."},
		{"You have a bad trip and end up fucking your friend's mother. Pay a fine of %s%d.", "You have a bad trip and end up fucking your friend's mother. Pay a fine of <strong>%s%d</strong>."},
		{"You fuck your friend's mother, but you don't perform very well and she charges you %s%d for the inconvenience.", "You fuck your friend's mother, but you don't perform very well and she charges you <strong>%s%d</strong> for the inconvenience."},
		{"You play some Monopoly and the cold hard reality of capitalism sets in... you lose %s%d.", "You play some Monopoly and the cold hard reality of capitalism sets in... you lose <strong>%s%d</strong>."},
		{"You ruin a chill time and turn it into an unchill time, pay a fine of %s%d.", "You ruin a chill time and turn it into an unchill time, pay a fine of <strong>%s%d</strong>."},
	}
	text, formattedText := performAction(sender.UUID(), "chill", "chill", chillMin, chillMax, chillFineMin, chillFineMax, chillFail, chillResponses, chillNegativeResponses, chillTime)
	sender.Location().SendFormattedText(text, formattedText)
}

func meme(msg onelib.Message, sender onelib.Sender) {
	const (
		memeMax     = 25
		memeMin     = 5
		memeFineMax = 250
		memeFineMin = 5
		memeFail    = 20                // 1 in x of failure
		memeTime    = time.Second * 260 // time until command can be called again
	)
	memeResponses := [][2]string{
		{"Dab on all them haters and gain %s%d!", "Dab on all them haters and gain <strong>%s%d</strong>!"},
		{"You pull your dick out for Harambe and gain %s%d for your service.", "You pull your dick out for Harambe and gain <strong>%s%d</strong> for your service."},
		{"You did it for the Vine and got %s%d.", "You did it for the Vine and got <strong>%s%d</strong>."},
	}
	memeNegativeResponses := [][2]string{
		{"You talked shit, and got hit, pay a fine of %s%d.", "You talked shit, and got hit, pay a fine of <strong>%s%d</strong>."},
		{"You talked shit about Harambe and were forced to pay a fine of %s%d.", "You talked shit about Harambe and were forced to pay a fine of <strong>%s%d</strong>."},
		{"You didn't do it for the Vine, in fact, you didn't do it at all! Pay a fine of %s%d...", "You didn't do it for the Vine, in fact, you didn't do it at all! Pay a fine of <strong>%s%d</strong>..."},
	}
	text, formattedText := performAction(sender.UUID(), "meme", "meme", memeMin, memeMax, memeFineMin, memeFineMax, memeFail, memeResponses, memeNegativeResponses, memeTime)
	sender.Location().SendFormattedText(text, formattedText)
}

func alias(msg onelib.Message, sender onelib.Sender) {
	text := strings.ReplaceAll(msg.Text(), "`", "")
	if text == "" {
		txt := fmt.Sprintf("Turns current UUID into target of another UUID. Usage: %salias <UUID>", onelib.DefaultPrefix)
		formattedTxt := fmt.Sprintf("Turns current UUID into target of another UUID. Usage: <code>%salias &lt;UUID&gt;</code>", onelib.DefaultPrefix)
		sender.Location().SendFormattedText(txt, formattedTxt)
		return
	}
	err := onecurrency.Currency[DEFAULT_CURRENCY].Alias(sender.UUID(), onelib.UUID(text))
	if err != nil {
		sender.Location().SendText(fmt.Sprintf("Alias failed (%s): %s", text, err))
		return
	}
	sender.Location().SendText("Alias succeeded!")
}

func unalias(msg onelib.Message, sender onelib.Sender) {
	text := msg.Text()
	if text != "" {
		txt := fmt.Sprintf("Removes alias on current UUID. Usage: %sunalias", onelib.DefaultPrefix)
		formattedTxt := fmt.Sprintf("Removes alias on current UUID. Usage: <code>%sunalias</code>", onelib.DefaultPrefix)
		sender.Location().SendFormattedText(txt, formattedTxt)
		return
	}
	onecurrency.Currency[DEFAULT_CURRENCY].UnAlias(sender.UUID())
	sender.Location().SendText("Alias removed!")
}

// TODO support extracting uuid from formatted text example: <a href=\"https://matrix.to/#/@kittypeach:matrix.thedisco.zone\">KittyPeach</a>
func checkBal(msg onelib.Message, sender onelib.Sender) {
	var (
		uuid        onelib.UUID
		displayName string
	)
	if text := msg.Text(); text == "" {
		uuid = sender.UUID()
		displayName = sender.DisplayName()
	} else {
		displayName = text[:len(text)-1]
		if sender.Protocol() == "matrix" {
			formattedText := msg.FormattedText()
			if len(formattedText) > 29 && string(formattedText[:29]) == "<a href=\"https://matrix.to/#/" {
				tuuid := formattedText[29:]
				if end := strings.Index(tuuid, "\""); end > 0 {
					uuid = onelib.UUID(tuuid[:end])
				}
			} else {
				uuid = onelib.UUID("")
			}
		} else {
			uuid = onelib.UUID(displayName)
		}

	}

	if uuid == onelib.UUID("") {
		sender.Location().SendFormattedText(fmt.Sprintf("Check balance. Usage: `%sbal [uuid]`", onelib.DefaultPrefix), fmt.Sprintf("Check balance. Usage: <code>%sbal [uuid]</code>", onelib.DefaultPrefix))
		return
	}

	_, cObj := onecurrency.Currency[DEFAULT_CURRENCY].Get(uuid)
	if cObj == nil {
		cObj = onecurrency.Currency[DEFAULT_CURRENCY].New(uuid)
	}
	text := fmt.Sprintf("%s's balance:\n\nOn-hand | Bank | Net\n%s%d    | %s%d   | %s%d", displayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity, DEFAULT_CURRENCY, cObj.Quantity+cObj.BankQuantity)
	formattedText := fmt.Sprintf("<strong>%s's balance:</strong><br /><table><tr><th> On-hand </th><th> Bank </th><th> Net </th></tr><br /><tr><th> <strong>%s%d</strong>  </th><th>  <strong>%s%d</strong>  </th><th> <strong>%s%d</strong></th></tr></table>", displayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity, DEFAULT_CURRENCY, cObj.Quantity+cObj.BankQuantity)
	sender.Location().SendFormattedText(text, formattedText)
}

func deposit(msg onelib.Message, sender onelib.Sender) {
	text := msg.Text()
	if text == "all" {
		q, err := onecurrency.Currency[DEFAULT_CURRENCY].DepositAll(sender.UUID())
		if err != nil {
			sender.Location().SendText(fmt.Sprintf("Nothing to deposit!"))
			return
		}
		sender.Location().SendText(fmt.Sprintf("Deposited all %s%d!", DEFAULT_CURRENCY, q))
		return
	}
	sender.Location().SendText("Not implemented.")
}

func withdraw(msg onelib.Message, sender onelib.Sender) {
	sender.Location().SendText("Not implemented.")
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
	return map[string]onelib.Command{"bal": checkBal, "balance": checkBal, "cute": cute, "chill": chill, "meme": meme, "dep": deposit, "deposit": deposit, "withdraw": withdraw, "alias": alias, "unalias": unalias}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (mp *MoneyPlugin) Remove() {
}
