// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	NAME     = "comics"
	LONGNAME = "The XKCD comic plugin"
	VERSION  = "v0.0.0"
)

func Load() onelib.Plugin {
	rand.Seed(time.Now().UnixNano())
	return new(ComicPlugin)
}

type ComicPlugin int

func (cp *ComicPlugin) Name() string {
	return NAME
}

func (cp *ComicPlugin) LongName() string {
	return LONGNAME
}

func (cp *ComicPlugin) Version() string {
	return VERSION
}

func getComicInfo(url string) (title string, imageURL string, extraText string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	var rawText []byte
	rawText, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	text := string(rawText)
	text = text[strings.Index(text, "<title>")+7:]
	title = text[:strings.Index(text, "</title>")]
	iText := text[strings.Index(text, "Image URL (for hotlinking/embedding): ")+38:]
	imageURL = iText[:strings.Index(iText, "\n")]

	altText := "{{Title text: "
	altTextEnd := "}}"
	altTextPos := strings.Index(text, altText)
	if altTextPos == -1 {
		altText = "{{title text: "
		altTextPos = strings.Index(text, altText)
		if altTextPos == -1 {
			altText = "{{alt-text: "
			altTextPos = strings.Index(text, altText)
			if altTextPos == -1 {
				altText = "g\" title=\""
				altTextEnd = "\""
				altTextPos = strings.Index(text, altText)
			}
		}
	}
	if altTextPos == -1 {
		return
	}

	text = text[altTextPos+len(altText):]
	extraText = text[:strings.Index(text, altTextEnd)]
	return
}

func comic(msg onelib.Message, sender onelib.Sender) {
	min := 200
	max := 2307 //TODO give link to main page for newest updates
	url := fmt.Sprintf("https://www.xkcd.com/%d", rand.Intn(max-min+1)+min)
	title, imageURL, extraText := getComicInfo(url)
	text := fmt.Sprintf("Your comic: \"%s\": %s\n*%s*", title, url, extraText)
	formattedText := fmt.Sprintf("Your comic: <a href=%s>%s</a> (<a href=%s>Web</a>)<br />\n<i>%s</i>", imageURL, title, url, extraText)
	sender.Location().SendFormattedText(text, formattedText)
}

func (cp *ComicPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"comic": comic}, nil
}

func (cp *ComicPlugin) Remove() {
}
