package main

import (
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"math/rand"
	"time"
)

const (
	NAME     = "comics"
	LONGNAME = "The XKCD comic plugin"
	VERSION  = "v0.0.0"
)

func Load() onelib.Plugin {
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

func comic(msg onelib.Message, sender onelib.Sender) {
	rand.Seed(time.Now().UnixNano())
	min := 200
	max := 2307 //TODO give link to main page for newest updates
	fmt.Println("Here you go: https://www.xkcd.com/", (rand.Intn(max-min+1) + min))
}

func (cp *ComicPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"comic": comic}, nil
}

func (cp *ComicPlugin) Remove() {
}
