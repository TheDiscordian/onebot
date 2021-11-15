// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/libs/onecurrency"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// NAME is same as filename, minus extension
	NAME = "money"
	// LONGNAME is what's presented to the user
	LONGNAME = "Currency Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.1"

	// DEFAULT_CURRENCY is the default currency symbol
	DEFAULT_CURRENCY = "★"
)

// TODO command to assign a location to a currency location uuid. Allow it to only be unset by whoever set it.

// Load returns the Plugin object.
func Load() onelib.Plugin {
	rand.Seed(time.Now().UnixNano())
	UserActionMap = new(userMap)
	UserActionMap.uMap = make(map[onelib.UUID]lastAction, 1)
	UserActionMap.lock = new(sync.RWMutex)
	AliasConfirmMap = new(aliasConfirmMap)
	AliasConfirmMap.uMap = make(map[onelib.UUID]onelib.UUID, 1)
	AliasConfirmMap.lock = new(sync.RWMutex)
	return new(MoneyPlugin)
}

type aliasConfirmMap struct {
	uMap map[onelib.UUID]onelib.UUID
	lock *sync.RWMutex
}

func (acm *aliasConfirmMap) Get(requester onelib.UUID) (target onelib.UUID) {
	acm.lock.RLock()
	target = acm.uMap[requester]
	acm.lock.RUnlock()
	return
}

func (acm *aliasConfirmMap) Set(requester, target onelib.UUID) {
	acm.lock.Lock()
	acm.uMap[requester] = target
	acm.lock.Unlock()
}

func (acm *aliasConfirmMap) Delete(requester onelib.UUID) {
	acm.lock.Lock()
	delete(acm.uMap, requester)
	acm.lock.Unlock()
}

var (
	// UserActionMap used for storing the last time an action was called
	UserActionMap *userMap
	// AliasConfirmMap stores aliases waiting to be confirmed, key is requester, val is target
	AliasConfirmMap *aliasConfirmMap
)

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
// TODO support location and currency parameters
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

func performAction(uuid onelib.UUID, actionName, actionText string, actionMinPayout, actionMaxPayout, actionMinFine, actionMaxFine, actionFailRate int, positiveResponses, negativeResponses [][2]string, actionCallRate time.Duration) (text string, formattedText string) {
	tuuid, _, _ := onecurrency.Currency.Get(DEFAULT_CURRENCY, onelib.UUID("global"), uuid)
	if tuuid != onelib.UUID("") {
		uuid = tuuid
	}
	if storedTime := UserActionMap.Update(uuid, actionName, actionCallRate); time.Since(storedTime) >= actionCallRate {
		if actionFailRate > 1 && rand.Intn(actionFailRate) == 0 {
			roll := rand.Intn(actionMaxFine-actionMinFine) + actionMinFine
			negativeResponse := negativeResponses[rand.Intn(len(negativeResponses))]
			text = fmt.Sprintf(negativeResponse[0], DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf(negativeResponse[1], DEFAULT_CURRENCY, roll)
			onecurrency.Currency.Add(DEFAULT_CURRENCY, onelib.UUID("global"), uuid, roll*-1, 0)
		} else {
			roll := rand.Intn(actionMaxPayout-actionMinPayout) + actionMinPayout
			positiveResponse := positiveResponses[rand.Intn(len(positiveResponses))]
			text = fmt.Sprintf(positiveResponse[0], DEFAULT_CURRENCY, roll)
			formattedText = fmt.Sprintf(positiveResponse[1], DEFAULT_CURRENCY, roll)
			onecurrency.Currency.Add(DEFAULT_CURRENCY, onelib.UUID("global"), uuid, roll, 0)
		}
	} else {
		timeUntil := time.Until(storedTime.Add(actionCallRate))
		text, formattedText = formatDuration(actionText, timeUntil)
	}
	return
}

func cute(msg onelib.Message, sender onelib.Sender) {
	const (
		cuteMax  = 245
		cuteMin  = 5
		cuteTime = time.Minute * 150 // time until command can be called again
	)
	cuteResponses := [][2]string{
		{"You help Miles build robots and gain **%s%d**!", "You help Miles build robots and gain <strong>%s%d</strong>!"},
		{"You go to a comedy show with Devin and gain **%s%d**!", "You go to a comedy show with Devin and gain <strong>%s%d</strong>!"},
		{"You spend all day painting chicken nuggets while eating chicken nuggets with Paulita and gain **%s%d**.", "You spend all day painting chicken nuggets while eating chicken nuggets with Paulita and gain <strong>%s%d</strong>."},
		{"You join Autumn on one of her nightly travels, gaining **%s%d**.", "You join Autumn on one of her nightly travels, gaining <strong>%s%d</strong>."},
		{"You go rock climbing with Randy the trash panda and gain **%s%d**!", "You go rock climbing with Randy the trash panda and gain <strong>%s%d</strong>!"},
		{"You watch movies with Lola and gain **%s%d**.", "You watch movies with Lola and gain <strong>%s%d</strong>."},
		{"Lucille shows you her rock collection and you gain **%s%d** from the experience!", "Lucille shows you her rock collection and you gain <strong>%s%d</strong> from the experience!"},
		{"You go surfing with Perry, gaining **%s%d**.", "You go surfing with Perry, gaining <strong>%s%d</strong>."},
		{"You join Francis the roar and go firefighting! You gained **%s%d**.", "You join Francis the roar and go firefighting! You gained <strong>%s%d</strong>."},
		{"You join Henry in doing commentating on sports games, you're paid **%s%d**!", "You join Henry in doing commentating on sports games, you're paid <strong>%s%d</strong>!"},
		{"You spend the day reading with Veronica and gain **%s%d**.", "You spend the day reading with Veronica and gain <strong>%s%d</strong>."},
		{"You help Aimee plan a red carpet event and get paid **%s%d**.", "You help Aimee plan a red carpet event and get paid <strong>%s%d</strong>."},
		{"You help Winston make a cullinary delight and gain **%s%d**.", "You help Winston make a cullinary delight and gain <strong>%s%d</strong>."},
		{"You spend the day doing artwork with Esmeralda and gain **%s%d**.", "You spend the day doing artwork with Esmeralda and gain <strong>%s%d</strong>."},
		{"You play soccer with Paco and gain **%s%d**.", "You play soccer with Paco and gain <strong>%s%d</strong>."},
		{"You study medicine with Jessica, surprisingly you're paid **%s%d** to do this!", "You study medicine with Jessica, surprisingly you're paid <strong>%s%d</strong> to do this!"},
		{"Hang out with Candy during the annual Squishmallow Egg Hunt, you gain **%s%d**!", "Hang out with Candy during the annual Squishmallow Egg Hunt, you gain <strong>%s%d</strong>!"},
		{"You go hiking with Rodrigo and gain **%s%d**.", "You go hiking with Rodrigo and gain <strong>%s%d</strong>."},
		{"You help design fashion with Matt the manitee and get paid **%s%d** for your designs!", "You help design fashion with Matt the manitee and get paid <strong>%s%d</strong> for your designs!"},
		{"You find and help identify treasure with Violet, your cut is **%s%d**.", "You find and help identify treasure with Violet, your cut is <strong>%s%d</strong>."},
		{"You help Zoe with restaurant planning. She pays you **%s%d** for your time.", "You help Zoe with restaurant planning. She pays you <strong>%s%d</strong> for your time."},
		{"You cuddle a cat-girl and gain **%s%d**.", "You cuddle a cat-girl and gain <strong>%s%d</strong>."},
		{"Karina teaches you how to hack! You hack a bank, and earn **%s%d**!", "Karina teaches you how to hack! You hack a bank, and earn <strong>%s%d</strong>!"},
		{"You plant tulips with Roxy and gain **%s%d**.", "You plant tulips with Roxy and gain <strong>%s%d</strong>."},
		{"Your favourite Animal Crossing villager gives you **%s%d**!", "Your favourite Animal Crossing villager gives you <strong>%s%d</strong>!"},
	}
	text, formattedText := performAction(sender.UUID(), "cute", "be cute", cuteMin, cuteMax, 0, 0, 0, cuteResponses, nil, cuteTime)
	sender.Location().SendFormattedText(text, formattedText)
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
}

func chill(msg onelib.Message, sender onelib.Sender) {
	const (
		chillMax     = 205
		chillMin     = 40
		chillFineMax = 145
		chillFineMin = 1
		chillFail    = 3                // 1 in x of failure
		chillTime    = time.Minute * 30 // time until command can be called again
	)
	chillResponses := [][2]string{
		{"Smoke weed with Snoop Dogg and gain **%s%d**!", "Smoke weed with Snoop Dogg and gain <strong>%s%d</strong>!"},
		{"You play some 100%% Orange Juice with your friends. It's a good time, you gain **%s%d**.", "You play some 100%% Orange Juice with your friends. It's a good time, you gain <strong>%s%d</strong>."},
		{"You roll a bad blunt, but Snoop Dogg is too high to notice! Good job, have **%s%d**!", "You roll a bad blunt, but Snoop Dogg is too high to notice! Good job, have <strong>%s%d</strong>!"},
		{"You drop acid and experience another reality. You gain **%s%d**.", "You drop acid and experience another reality. You gain <strong>%s%d</strong>."},
		{"You fuck your friend's mother and she pays YOU **%s%d**!", "You fuck your friend's mother and she pays YOU <strong>%s%d</strong>!"},
		{"The weed hit you just right, gain **%s%d**!", "The weed hit you just right, gain <strong>%s%d</strong>!"},
		{"You shoot up some scrubs in an FPS and gain **%s%d**.", "You shoot up some scrubs in an FPS and gain <strong>%s%d</strong>."},
		{"You vibe out and listen to some music. After a while you gain **%s%d**.", "You vibe out and listen to some music. After a while you gain <strong>%s%d</strong>."},
		{"You finish a round of gaming and your score ends in 420! You gain **%s%d** from the experience.", "You finish a round of gaming and your score ends in 420! You gain <strong>%s%d</strong> from the experience."},
		{"Your Second Life empire is booming, you gain **%s%d**.", "Your Second Life empire is booming, you gain <strong>%s%d</strong>."},
		{"You made a nice vape cloud, gain **%s%d**.", "You made a nice vape cloud, gain <strong>%s%d</strong>."},
		{"You blew some killer smoke rings and gained **%s%d**.", "You blew some killer smoke rings and gained <strong>%s%d</strong>."},
		{"You meet your favourite band, and have a good time. You gain **%s%d** from selling the signed merch you got afterwards.", "You meet your favourite band, and have a good time. You gain <strong>%s%d</strong> from selling the signed merch you got afterwards."},
		{"You get the fastest speedrun out of all your friends in Mario. It feels good, and you gain **%s%d**.", "You get the fastest speedrun out of all your friends in Mario. It feels good, and you gain <strong>%s%d</strong>."},
		{"You get the highest score on a game with your friends. You gain pride and **%s%d**.", "You get the highest score on a game with your friends. You gain pride and <strong>%s%d</strong>."},
		{"You roll a fat blunt and smoke it, gain **%s%d**.", "You roll a fat blunt and smoke it, gain <strong>%s%d</strong>."},
		{"You have a nap and gain **%s%d**.", "You have a nap and gain <strong>%s%d</strong>."},
		{"You enter a gaming tournament and win **%s%d**, good job!", "You enter a gaming tournament and win <strong>%s%d</strong>, good job!"},
		{"It's 420. Smoke a bowl and gain **%s%d**.", "It's 420. Smoke a bowl and gain <strong>%s%d</strong>."},
		{"You found a nugget of weed you forgot about, awesome! Gain **%s%d**.", "You found a nugget of weed you forgot about, awesome! Gain <strong>%s%d</strong>."},
		{"Turns out you had more beer than you thought in the back of the fridge, cool! Gain **%s%d**.", "Turns out you had more beer than you thought in the back of the fridge, cool! Gain <strong>%s%d</strong>."},
		{"You find some spare coils for your vape you forgot about, score! Gain **%s%d**.", "You find some spare coils for your vape you forgot about, score! Gain <strong>%s%d</strong>."},
	}
	chillNegativeResponses := [][2]string{
		{"You roll a bad blunt and Snoop Dogg notices, pay a fine of **%s%d**...", "You roll a bad blunt and Snoop Dogg notices, pay a fine of <strong>%s%d</strong>..."},
		{"You play some 100%% Orange Juice with your friends. You come in last and lose **%s%d**.", "You play some 100%% Orange Juice with your friends. You come in last and lose <strong>%s%d</strong>."},
		{"You have a bad trip and end up fucking your friend's mother. Pay a fine of **%s%d**.", "You have a bad trip and end up fucking your friend's mother. Pay a fine of <strong>%s%d</strong>."},
		{"You fuck your friend's mother, but you don't perform very well and she charges you **%s%d** for the inconvenience.", "You fuck your friend's mother, but you don't perform very well and she charges you <strong>%s%d</strong> for the inconvenience."},
		{"You play some Monopoly and the cold hard reality of capitalism sets in... you lose **%s%d**.", "You play some Monopoly and the cold hard reality of capitalism sets in... you lose <strong>%s%d</strong>."},
		{"You ruin a chill time and turn it into an unchill time, pay a fine of **%s%d**.", "You ruin a chill time and turn it into an unchill time, pay a fine of <strong>%s%d</strong>."},
		{"You go to fill your bowl, but spill all your weed costing you **%s%d** 💔.", "You go to fill your bowl, but spill all your weed costing you <strong>%s%d</strong> 💔."},
		{"You're in a good mood so you buy your friends some games costing you **%s%d**.", "You're in a good mood so you buy your friends some games costing you <strong>%s%d</strong>."},
		{"Your joint unravels, pay **%s%d**.", "Your joint unravels, pay <strong>%s%d</strong>."},
		{"You meet your favourite band, and have a good time. Unfortunately they board up in your place for a while, costing you **%s%d**.", "You meet your favourite band, and have a good time. Unfortunately they board up in your place for a while, costing you <strong>%s%d</strong>."},
		{"You try to re-toast your sub, but burn it and lose **%s%d**.", "You try to re-toast your sub, but burn it and lose <strong>%s%d</strong>."},
		{"You fall asleep on the job and lose **%s%d**!", "You fall asleep on the job and lose <strong>%s%d</strong>!"},
		{"You enter a gaming tournament and lose! You paid **%s%d** to enter, so you're down that.", "You enter a gaming tournament and lose! You paid <strong>%s%d</strong> to enter, so you're down that."},
		{"You run out of weed! You pay in sadness and **%s%d** 😞.", "You run out of weed! You pay in sadness and <strong>%s%d</strong> 😞."},
		{"You run out of beer! You throw up and pay **%s%d** to people for taking care of your drunk ass.", "You run out of beer! You throw up and pay <strong>%s%d</strong> to people for taking care of your drunk ass."},
		{"You take a dry hit from your vape and lose **%s%d**.", "You take a dry hit from your vape and lose <strong>%s%d</strong>."},
	}
	text, formattedText := performAction(sender.UUID(), "chill", "chill", chillMin, chillMax, chillFineMin, chillFineMax, chillFail, chillResponses, chillNegativeResponses, chillTime)
	sender.Location().SendFormattedText(text, formattedText)
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
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
		{"Dab on all them haters and gain **%s%d**!", "Dab on all them haters and gain <strong>%s%d</strong>!"},
		{"You pull your dick out for Harambe and gain **%s%d** for your service.", "You pull your dick out for Harambe and gain <strong>%s%d</strong> for your service."},
		{"You did it for the Vine and got **%s%d**.", "You did it for the Vine and got <strong>%s%d</strong>."},
		{"You say \"bork\" in a large crowd. Many assume you're homeless and donate **%s%d** to you.", "You say \"bork\" in a large crowd. Many assume you're homeless and donate <strong>%s%d</strong> to you."},
		{"You vibe out to some penis music. An agent is so impressed he calls you at home and offers you a **%s%d** contract!", "You vibe out to some penis music. An agent is so impressed he calls you at home and offers you a <strong>%s%d</strong> contract!"},
		{"You score 69420 on your favourite game, nice, have **%s%d**!", "You score 69420 on your favourite game, nice, have <strong>%s%d</strong>!"},
		{"You find millions of peaches. Wow! Gain **%s%d**.", "You find millions of peaches. Wow! Gain <strong>%s%d</strong>."},
		{"You spot a government surveilance drone and protect your privacy. Gain **%s%d** for your service.", "You spot a government surveilance drone and protect your privacy. Gain <strong>%s%d</strong> for your service."},
		{"You convince someone that birds aren't real, doing the world a service, and gain **%s%d**.", "You convince someone that birds aren't real, doing the world a service, and gain <strong>%s%d</strong>."},
		{"You make an original Steamed Hams video and gain **%s%d**.", "You make an original Steamed Hams video and gain <strong>%s%d</strong>."},
		{"You make a nice meme, and it hits the Reddit frontpage! Your upvotes are worth **%s%d**.", "You make a nice meme, and it hits the Reddit frontpage! Your upvotes are worth <strong>%s%d</strong>."},
		{"Someone on r/okaybuddyretard thinks you're genuinely retarded! Good job, have **%s%d**.", "Someone on r/okaybuddyretard thinks you're genuinely retarded! Good job, have <strong>%s%d</strong>."},
	}
	memeNegativeResponses := [][2]string{
		{"You talked shit, and got hit, pay a fine of **%s%d**.", "You talked shit, and got hit, pay a fine of <strong>%s%d</strong>."},
		{"You talked shit about Harambe and were forced to pay a fine of **%s%d**.", "You talked shit about Harambe and were forced to pay a fine of <strong>%s%d</strong>."},
		{"You didn't do it for the Vine, in fact, you didn't do it at all! Pay a fine of **%s%d**...", "You didn't do it for the Vine, in fact, you didn't do it at all! Pay a fine of <strong>%s%d</strong>..."},
		{"You catch the covids and have to pay **%s%d** in medical expenses.", "You catch the covids and have to pay <strong>%s%d</strong> in medical expenses."},
		{"You plank in public and take pictures, you pay in shame and **%s%d**.", "You plank in public and take pictures, you pay in shame and <strong>%s%d</strong>."},
		{"A crowd of people gang up on you and claim that birds are in fact \"real\". You're beaten, and lose **%s%d**.", "A crowd of people gang up on you and claim that birds are in fact \"real\". You're beaten, and lose <strong>%s%d</strong>."},
		{"You shout the N Word in an urban environment and get robbed for **%s%d**.", "You shout the N Word in an urban environment and get robbed for <strong>%s%d</strong>."},
		{"You let your memes be dreams and lost **%s%d**.", "You let your memes be dreams and lost <strong>%s%d</strong>."},
		{"You think you made a decent meme, but the mods delete it and take **%s%d** from you 😰.", "You think you made a decent meme, but the mods delete it and take <strong>%s%d</strong> from you 😰."},
		{"You're caught being untarded on r/okaybuddyretard and are forced to pay **%s%d** to Big Chungus.", "You're caught being untarded on r/okaybuddyretard and are forced to pay <strong>%s%d</strong> to Big Chungus."},
	}
	text, formattedText := performAction(sender.UUID(), "meme", "meme", memeMin, memeMax, memeFineMin, memeFineMax, memeFail, memeResponses, memeNegativeResponses, memeTime)
	sender.Location().SendFormattedText(text, formattedText)
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
}

func risk(msg onelib.Message, sender onelib.Sender) {
	const (
		riskMax     = 499
		riskMin     = 80
		riskFineMax = 500
		riskFineMin = 81
		riskFail    = 2                // 1 in x of failure
		riskTime    = time.Second * 61 // time until command can be called again
	)
	riskResponses := [][2]string{
		{"You bet your life savings on a horse race ... and win **%s%d**!", "You bet your life savings on a horse race ... and win <strong>%s%d</strong>!"},
		{"You buy a lottery ticket and win **%s%d**!", "You buy a lottery ticket and win <strong>%s%d</strong>!"},
		{"You drive as fast as your vehicle will go, and time is money, so you gain **%s%d**!", "You drive as fast as your vehicle will go, and time is money, so you gain <strong>%s%d</strong>!"},
		{"You sacrifice to the gambling gods, they smile upon you, and bless you with **%s%d**!", "You sacrifice to the gambling gods, they smile upon you, and bless you with <strong>%s%d</strong>!"},
	}
	riskNegativeResponses := [][2]string{
		{"You dangle your child off a roof! That was a bad idea, you lose **%s%d**...", "You dangle your child off a roof! That was a bad idea, you lose <strong>%s%d</strong>..."},
		{"You bet your life savings on a horse race ... and lose **%s%d** 😞", "You bet your life savings on a horse race ... and lose <strong>%s%d</strong> 😞"},
		{"You buy several lottery tickets and have a total net loss of **%s%d** 😞", "You buy several lottery tickets and have a total net loss of <strong>%s%d</strong> 😞"},
		{"You get caught speeding and pay a fine of **%s%d**.", "You get caught speeding and pay a fine of <strong>%s%d</strong>."},
		{"You have to pay interest to your loanshark, you pay **%s%d**.", "You have to pay interest to your loanshark, you pay <strong>%s%d</strong>."},
		{"You sacrifice to the gambling gods and lose **%s%d**.", "You sacrifice to the gambling gods and lose <strong>%s%d</strong>."},
	}
	text, formattedText := performAction(sender.UUID(), "risk", "risk", riskMin, riskMax, riskFineMin, riskFineMax, riskFail, riskResponses, riskNegativeResponses, riskTime)
	sender.Location().SendFormattedText(text, formattedText)
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
}

func alias(msg onelib.Message, sender onelib.Sender) {
	text := strings.ReplaceAll(msg.Text(), "`", "")
	if text == "" {
		txt := fmt.Sprintf("Turns current UUID into target of another UUID. Usage: %salias <UUID>", onelib.DefaultPrefix)
		formattedTxt := fmt.Sprintf("Turns current UUID into target of another UUID. Usage: <code>%salias &lt;UUID&gt;</code>", onelib.DefaultPrefix)
		sender.Location().SendFormattedText(txt, formattedTxt)
		return
	}
	AliasConfirmMap.Set(sender.UUID(), onelib.UUID(text))
	txt := fmt.Sprintf("Almost done, just type '%sconfirmalias `%s`' in a room the bot can see on the target account, and you're set!", onelib.DefaultPrefix, sender.UUID())
	formattedTxt := fmt.Sprintf("Almost done, just type <code>%sconfirmalias `%s`</code> in a room the bot can see on the target account, and you're set!", onelib.DefaultPrefix, sender.UUID())
	sender.Location().SendFormattedText(txt, formattedTxt)
}

func confirmalias(msg onelib.Message, sender onelib.Sender) {
	text := strings.ReplaceAll(msg.Text(), "`", "")
	if text == "" {
		txt := fmt.Sprintf("Confirms an alias. Usage: '%sconfirmalias `<UUID>`'", onelib.DefaultPrefix)
		formattedTxt := fmt.Sprintf("Confirms an alias. Usage: <code>%sconfirmalias &lt;UUID&gt;</code>", onelib.DefaultPrefix)
		sender.Location().SendFormattedText(txt, formattedTxt)
		return
	}
	alias := AliasConfirmMap.Get(onelib.UUID(text))
	if alias == onelib.UUID("") {
		txt := "That account isn't trying to alias with anyone."
		sender.Location().SendText(txt)
		return
	}
	if alias != sender.UUID() {
		txt := "That account isn't trying to alias with you."
		sender.Location().SendText(txt)
		return
	}
	err := onelib.Alias.Set(onelib.UUID(text), sender.UUID())
	if err != nil {
		sender.Location().SendText(fmt.Sprintf("Alias failed (%s): %s", text, err))
		return
	}
	AliasConfirmMap.Delete(onelib.UUID(text))
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
	onelib.Alias.UnAlias(sender.UUID())
	sender.Location().SendText("Alias removed!")
}

func leaderboard(msg onelib.Message, sender onelib.Sender) {
	rtext := []rune("Leaderboard:\n")
	all := onecurrency.Currency.GetAll(DEFAULT_CURRENCY, onelib.UUID("global"))
	if len(all) > 10 {
		all = all[:10]
	}
	for i, uco := range all {
		var displayName string
		if uco.DisplayName != "" {
			displayName = uco.DisplayName
		} else {
			displayName = string(uco.UUID)
		}
		rtext = append(rtext, []rune(fmt.Sprintf("%d. %s (%s%d)\n", i+1, displayName, DEFAULT_CURRENCY, uco.Quantity+uco.BankQuantity))...)
	}
	sender.Location().SendText(string(rtext))
}

func checkBal(msg onelib.Message, sender onelib.Sender) {
	var (
		uuid        onelib.UUID
		displayName string
	)
	if text := msg.Text(); text == "" {
		uuid = sender.UUID()
		displayName = sender.DisplayName()
		onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), uuid, displayName)
	} else {
		displayName = text[:len(text)]
		if sender.Protocol() == "matrix" {
			formattedText := msg.FormattedText()
			if len(formattedText) > 29 && string(formattedText[:29]) == "<a href=\"https://matrix.to/#/" {
				tuuid := formattedText[29:]
				if end := strings.Index(tuuid, "\""); end > 0 {
					uuid = onelib.UUID(tuuid[:end])
				}
			} else {
				uuid = onelib.UUID(displayName)
			}
		} else {
			uuid = onelib.UUID(displayName)
		}

	}

	if uuid == onelib.UUID("") {
		sender.Location().SendFormattedText(fmt.Sprintf("Check balance. Usage: `%sbal [uuid]`", onelib.DefaultPrefix), fmt.Sprintf("Check balance. Usage: <code>%sbal [uuid]</code>", onelib.DefaultPrefix))
		return
	}

	_, cObj, err := onecurrency.Currency.Get(DEFAULT_CURRENCY, onelib.UUID("global"), uuid)
	if cObj == nil {
		cObj = new(onecurrency.CurrencyObject)
	}
	if err != nil {
		onelib.Error.Printf("(UUID: %s) %s\n", uuid, err)
	}
	text := fmt.Sprintf("%s's balance:\n\nOn-hand | Bank | Net\n**%s%d**    | **%s%d**   | **%s%d**", displayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity, DEFAULT_CURRENCY, cObj.Quantity+cObj.BankQuantity)
	formattedText := fmt.Sprintf("<strong>%s's balance:</strong><br /><table><tr><th> On-hand </th><th> Bank </th><th> Net </th></tr><br /><tr><th> <strong>%s%d</strong>  </th><th>  <strong>%s%d</strong>  </th><th> <strong>%s%d</strong></th></tr></table>", displayName, DEFAULT_CURRENCY, cObj.Quantity, DEFAULT_CURRENCY, cObj.BankQuantity, DEFAULT_CURRENCY, cObj.Quantity+cObj.BankQuantity)
	sender.Location().SendFormattedText(text, formattedText)
}

func deposit(msg onelib.Message, sender onelib.Sender) {
	text := msg.Text()
	if text == "all" {
		q, err := onecurrency.Currency.DepositAll(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID())
		if err != nil {
			sender.Location().SendText("Nothing to deposit!")
			return
		}
		sender.Location().SendFormattedText(fmt.Sprintf("Deposited all **%s%d**!", DEFAULT_CURRENCY, q), fmt.Sprintf("Deposited all <strong>%s%d</strong>!", DEFAULT_CURRENCY, q))
		return
	}
	q, err := strconv.Atoi(text)
	if err != nil {
		sender.Location().SendText("Quantity must be an integer.")
		return
	}
	err = onecurrency.Currency.Deposit(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), q)
	if err != nil {
		sender.Location().SendText(err.Error())
		return
	}
	sender.Location().SendFormattedText(fmt.Sprintf("Deposited **%s%d**!", DEFAULT_CURRENCY, q), fmt.Sprintf("Deposited <strong>%s%d</strong>!", DEFAULT_CURRENCY, q))
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
}

func withdraw(msg onelib.Message, sender onelib.Sender) {
	text := msg.Text()
	if text == "all" {
		q, err := onecurrency.Currency.WithdrawAll(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID())
		if err != nil {
			sender.Location().SendText("Nothing to withdraw!")
			return
		}
		sender.Location().SendFormattedText(fmt.Sprintf("Withdrew all **%s%d**!", DEFAULT_CURRENCY, q), fmt.Sprintf("Withdrew all <strong>%s%d</strong>!", DEFAULT_CURRENCY, q))
		return
	}
	q, err := strconv.Atoi(text)
	if err != nil {
		sender.Location().SendText("Quantity must be an integer.")
		return
	}
	err = onecurrency.Currency.Withdraw(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), q)
	if err != nil {
		sender.Location().SendText(err.Error())
		return
	}
	sender.Location().SendFormattedText(fmt.Sprintf("Withdrew **%s%d**!", DEFAULT_CURRENCY, q), fmt.Sprintf("Withdrew <strong>%s%d</strong>!", DEFAULT_CURRENCY, q))
	onecurrency.Currency.UpdateDisplayName(DEFAULT_CURRENCY, onelib.UUID("global"), sender.UUID(), sender.DisplayName())
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
	return map[string]onelib.Command{"bal": checkBal, "balance": checkBal, "cute": cute, "chill": chill, "meme": meme, "risk": risk, "dep": deposit, "deposit": deposit, "withdraw": withdraw, "alias": alias, "unalias": unalias, "confirmalias": confirmalias, "leaderboard": leaderboard, "lb": leaderboard}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (mp *MoneyPlugin) Remove() {
}
