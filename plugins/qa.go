// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"strings"
	"encoding/json"
	"os"
	"io"
	"os/exec"

	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "qa"
	// LONGNAME is what's presented to the user
	LONGNAME = "Question & Answer Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	qa := new(QAPlugin)
	// Load expertise from expertise.json, which contains a string map where the values are arrays of strings. Store just the keys in qa.expertise
	expertise_file, err := os.Open("plugins/qa/expertise.json")
	if err != nil {
		onelib.Error.Println("Error opening expertise.json:", err)
		return nil
	}
	defer expertise_file.Close()
	qa.expertise = make([]string, 0)
	dec := json.NewDecoder(expertise_file)
	for {
		var m map[string][]string
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			onelib.Error.Println("Error decoding expertise.json:", err)
			return nil
		}
		for k := range m {
			qa.expertise = append(qa.expertise, k)
		}
	}

	qa.openaiKey = onelib.GetTextConfig(NAME, "openai_key")
	if qa.openaiKey == "" {
		onelib.Error.Println("[qa] openai_key can't be blank.")
		return nil
	}

	qa.prompt = onelib.GetTextConfig(NAME, "prompt")

	// channelsJson is in the format: {"protocol": ["channel1", "channel2"]}
	channelsJson := onelib.GetTextConfig(NAME, "channels")
	qa.channels = make(map[string][]string)
	err = json.Unmarshal([]byte(channelsJson), &qa.channels)
	if err != nil {
		onelib.Error.Println("[qa] Error decoding channels:", err)
		return nil
	}

	// TODO Currently these run every time, which takes a long time, and costs some money. We should
	// instead have a goroutine check if the file is updated, if so, then do some updates instead of
	// the whole thing.
	_, err = qa.runqa("db")
	if err != nil {
		onelib.Error.Println("Error downloading db:", err)
		return nil
	}
	_, err = qa.runqa("aidb")
	if err != nil {
		onelib.Error.Println("Error updating aidb:", err)
		return nil
	}

	qa.monitor = &onelib.Monitor{
		OnMessageWithText: qa.OnMessageWithText,
	}
	return qa
}

// QAPlugin is an object for satisfying the Plugin interface.
type QAPlugin struct {
	monitor *onelib.Monitor
	expertise []string
	openaiKey string
	prompt string
	channels map[string][]string
}

func (qa *QAPlugin) runqa(args ...string) (string, error) {
	// Call the python script plugins/qa/qa.py, passing the openai key as an environment variable, and capturing the output.
	_args := []string{"plugins/qa/qa.py", "-e", "plugins/qa/expertise.json", "-mi", "plugins/qa/misinfos.json", "-p", qa.prompt, "-db", "plugins/qa/db-noembed.csv", "-edb", "plugins/qa/db.csv"}
	cmd := exec.Command("python3", append(_args, args...)...)
	cmd.Env = append(os.Environ(), "OPENAI_API_KEY="+qa.openaiKey)
	out, err := cmd.CombinedOutput()
	if err != nil {
		onelib.Debug.Println("qa.py output:", string(out))
		return string(out), err
	}
	return string(out), nil
}

func (qa *QAPlugin) ask_question(msg onelib.Message, sender onelib.Sender) {
	txt, err := qa.runqa("-q", msg.Text(), "question")
	if err != nil {
		onelib.Error.Println("Error running qa.py:", err)
		return
	}
	onelib.Debug.Println("Replying:", txt)
	sender.Location().SendText(txt)
}

// Name returns the name of the plugin, usually the filename.
func (qa *QAPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (qa *QAPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (qa *QAPlugin) Version() string {
	return VERSION
}

func (qa *QAPlugin) OnMessageWithText(from onelib.Sender, msg onelib.Message) {
	// Check if the message is in a channel we're monitoring
	channel := from.Location().UUID()
	proto := from.Protocol()
	if _, ok := qa.channels[proto]; !ok {
		return
	}
	found := false
	for _, v := range qa.channels[proto] {
		if v == "*" || onelib.UUID(v) == channel {
			found = true
			break
		}
	}
	if !found {
		return
	}

	txt := strings.ToLower(msg.Text())
	// Check if txt contains any of the strings in qa.expertise, and ends in a question mark
	for _, v := range qa.expertise {
		if strings.Contains(txt, v) && strings.HasSuffix(txt, "?") {
			qa.ask_question(msg, from)
			break
		}
	}
}

// Implements returns a map of commands and monitor the plugin implements.
func (qa *QAPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"q": qa.ask_question, "question": qa.ask_question}, qa.monitor
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (qa *QAPlugin) Remove() {
}
