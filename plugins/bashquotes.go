// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// NAME is same as filename, minus extension
	NAME = "bashquotes"
	// LONGNAME is what's presented to the user
	LONGNAME = "Bash Quotes"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	rand.Seed(time.Now().UnixNano())
	return new(BashPlugin)
}

// BashPlugin is an object for satisfying the Plugin interface.
type BashPlugin int

// Name returns the name of the plugin, usually the filename.
func (bp *BashPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (bp *BashPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (bp *BashPlugin) Version() string {
	return VERSION
}

func seekN(s, substr string, count int) string {
	for c := count; c > 0; c-- {
		s = s[strings.Index(s, substr)+len(substr):]
	}
	return s
}

// TODO error checking for parser
func getRandomQuoteInfo() (quoteNumber int, score int, quote string) {
	resp, err := http.Get("http://bash.org/?random1") // site literally doesn't support HTTPS (???)
	if err != nil {
		return
	}
	var rawText []byte
	rawText, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	text := string(rawText)
	quoten := rand.Intn(50)
	text = seekN(text, `<p class="quote">`, quoten)
	text = text[strings.Index(text, "\"?")+2:]          // seek to quoteNumber
	quoteNumberText := text[:strings.Index(text, "\"")] // grab number, as string
	quoteNumber, _ = strconv.Atoi(quoteNumberText)      // convert number to integer
	text = text[strings.Index(text, "n\">")+3:]
	scoreText := text[:strings.Index(text, "<")]
	score, _ = strconv.Atoi(scoreText) // convert number to indeger
	text = text[strings.Index(text, "t\">")+3:]
	quote = text[:strings.Index(text, "</")]

	return
}

func bashQuote(msg onelib.Message, sender onelib.Sender) {
	quoteNumber, score, quote := getRandomQuoteInfo()
	//text := fmt.Sprintf("Your comic: \"%s\": %s\n*%s*", title, url, extraText)
	formattedText := fmt.Sprintf("Quote <a href=http://bash.org/?%d>#%d</a> (<span data-mx-color=\"#00FF00\">%d</span>):<br />\n%s", quoteNumber, quoteNumber, score, quote)
	sender.Location().SendFormattedText(formattedText, formattedText)
}

// Implements returns a map of commands and monitor the plugin implements.
func (bp *BashPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"bash": bashQuote}, nil
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (bp *BashPlugin) Remove() {
}
