// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"strings"
)

const (
	NAME     = "8ball"
	LONGNAME = "8Ball Plugin"
	VERSION  = "v0.0.1"
)

var eightballAnswers = []string{
	"As I see it, yes", "It is certain", "It is decidedly so", "Most likely",
	"Outlook good", "Signs point to yes", "Without a doubt", "Yes", "Yes, definitely",
	"You may rely on it", "Reply hazy, try again", "Ask again later",
	"Better not tell you now", "Cannot predict now", "Concentrate and ask again",
	"Don't count on it", "My reply is no", "My sources say no", "Outlook not so good",
	"Very doubtful", "Never ask that question again"
}

func eightball(msg onelib.Message, sender onelib.Sender) {
	var formattedText string
	text := msg.Text()

	if strings.HasPrefix(text, "add ") {
		newAnswer := strings.TrimPrefix(text, "add ")
		eightballAnswers = append(eightballAnswers, newAnswer)
		text = "Added new response."
	} else if strings.HasPrefix(text, "remove ") {
		remAnswer := strings.TrimPrefix(text, "remove ")
		for i, answer := range eightballAnswers {
			if answer == remAnswer {
				eightballAnswers = append(eightballAnswers[:i], eightballAnswers[i+1:]...)
				text = "Removed response."
				break
			}
		}
	} else if len(text) < 3 {
		text = fmt.Sprintf("Predicts the future. Usage: %s8ball `<y/n question>`", onelib.DefaultPrefix)
		formattedText = fmt.Sprintf("Predicts the future. Usage: <code>%s8ball &lt;y/n question&gt;</code>", onelib.DefaultPrefix)
	} else {
		randn := rand.Intn(len(eightballAnswers))
		text = eightballAnswers[randn] + "."
	}

	formattedText = text
	sender.Location().SendFormattedText(text, formattedText)
}

type EightBallPlugin int

func (eb *EightBallPlugin) Name() string {
	return NAME
}

func (eb *EightBallPlugin) LongName() string {
	return LONGNAME
}

func (eb *EightBallPlugin) Version() string {
	return VERSION
}

func (eb *EightBallPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"8b": eightball, "8ball": eightball}, nil
}

func (eb *EightBallPlugin) Remove() {
}
