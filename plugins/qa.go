// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"bytes"

	"github.com/TheDiscordian/onebot/libs/discord"
	"github.com/TheDiscordian/onebot/libs/missioncontrol"
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "qa"
	// LONGNAME is what's presented to the user
	LONGNAME = "Question & Answer Plugin"
	// VERSION of the plugin
	VERSION = "v0.1.0"
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

	// Check if openai_key is set
	if onelib.GetTextConfig(NAME, "openai_key") == "" {
		onelib.Error.Println("[qa] openai_key can't be blank.")
		return nil
	}

	// channelsJson is in the format: {"protocol": ["channel1", "channel2"]}
	channelsJson := onelib.GetTextConfig(NAME, "channels")
	qa.channels = make(map[string][]string)
	err = json.Unmarshal([]byte(channelsJson), &qa.channels)
	if err != nil {
		onelib.Error.Println("[qa] Error decoding channels:", err)
		return nil
	}

	updateDbs := onelib.GetBoolConfig(NAME, "update_databases")
	if updateDbs {
		// TODO Currently these run every time, which takes a long time, and costs some money. We should
		// instead have a goroutine check if the file is updated, if so, then do some updates instead of
		// the whole thing.
		_, err = runqa("db")
		if err != nil {
			onelib.Error.Println("Error downloading db:", err)
			return nil
		}
		_, err = runqa("aidb")
		if err != nil {
			onelib.Error.Println("Error updating aidb:", err)
			return nil
		}
	}

	qa.monitor = &onelib.Monitor{
		OnMessageWithText: qa.OnMessageWithText,
		OnMessageUpdate:   qa.OnMessageUpdate,
	}

	qa.DbLock = new(sync.RWMutex)

	missioncontrol.Plugins.Set(LONGNAME, new(QAMissionControlPlugin))
	return qa
}

func getChannelsMap() map[string][]string {
	channelsJson := onelib.GetTextConfig(NAME, "channels")
	channels := make(map[string][]string)
	err := json.Unmarshal([]byte(channelsJson), &channels)
	if err != nil {
		onelib.Error.Println("[qa] Error decoding channels:", err)
		return nil
	}
	return channels
}

type QAMissionControlPlugin struct { }

func (qamc *QAMissionControlPlugin) HTML() template.HTML {
	templateString := `<h2>{{ .Name}}</h2>
<h3>Settings</h3>
<h4>OpenAI Key:<h4>
<input size="48" type="password" id="openai_key" name="openai_key" value="{{ .OpenAIKey}}"><button onclick="var x = document.getElementById('openai_key');if (x.type === 'password') {x.type = 'text';} else {x.type = 'password';}">👁️‍🗨️</button><button onclick="doAction('set_openai_key', document.getElementById('openai_key').value)">Save</button><br>
<h4>Prompt:</h4>
<textarea cols="80" rows="5" id="prompt" name="prompt">{{ .Prompt}}</textarea><button onclick="doAction('set_prompt', document.getElementById('prompt').value)">Save</button><br>
<h4>Channels</h4>
<span>Channels to monitor:</span><br>
{{ range $k, $v := .Channels}}<details><summary><b>{{ $k }} <button onclick="doAction('delete_protocol', '{{$k}}').then(() => {window.location.reload();});">❌</button></b></summary>
	{{ range $v }}
		<input class="list-input-box" type="text" value="{{ .}}"></input><button onclick="doAction('delete_channel', {p: '{{$k}}', c: '{{.}}'}).then(() => {window.location.reload();});">❌</button><br>
	{{ end }}<br>
	Add {{$k}} channel: <input type="text" id="add_channel_{{$k}}" name="add_channel_{{$k}}"></input><button onclick="doAction('add_channel', {p: '{{$k}}', c: document.getElementById('add_channel_{{$k}}').value}).then(() => {window.location.reload();});">➕</button><br>
	</details>
{{ end }}<br>
Add protocol: <input type="text" id="add_protocol" name="add_protocol"></input><button onclick="doAction('add_protocol', document.getElementById('add_protocol').value).then(() => {window.location.reload();});">➕</button><br>
<h4>Reply Settings</h4>
<span>Reply to questions:</span><input type="checkbox" id="reply_to_questions" name="reply_to_questions" {{ if .ReplyToQuestions}}checked{{ end }}><br>
<span>Reply to mentions:</span><input type="checkbox" id="reply_to_mentions" name="reply_to_mentions" {{ if .ReplyToMentions}}checked{{ end }}><br>
<button onclick="doAction('set_replies', {qs: document.getElementById('reply_to_questions').checked, ms: document.getElementById('reply_to_mentions').checked})">Save</button><br>
<h4>Expertise:</h4>
<textarea cols="80" rows="5" id="expertise" name="expertise" disabled>{{ .Expertise}}</textarea><button onclick="doAction('set_expertise', document.getElementById('expertise').value)" disabled>Save</button><br>
<h4>Misinfos:</h4>
<textarea cols="80" rows="5" id="misinfos" name="misinfos" disabled>{{ .Misinfos}}</textarea><button onclick="doAction('set_misinfos', document.getElementById('misinfos').value)" disabled>Save</button><br>
<h3>Tools</h3>
<!-- Tools are collapsed to avoid clutter -->
<details><summary>Ask Question</summary>
<!-- Prompt that doesn't save -->
<h4>Prompt (doesn't save):</h4>
<textarea cols="80" rows="5" id="temp_prompt" name="prompt">{{ .Prompt}}</textarea><br>
<h4>Question:</h4>
<textarea cols="80" rows="5" id="question" name="question"></textarea><button onclick="doAction('question', {q: document.getElementById('question').value, p: document.getElementById('temp_prompt').value})">Ask</button><br>
</details>
`
	tmpl, err := template.New("qa").Parse(templateString)
	if err != nil {
		onelib.Error.Println("Error parsing template:", err)
		return ""
	}

	// Expertise and Misinfos are stored in files, so we need to read them in.
	// Read the entire contents of the file into expertise
	expertiseFile, err := os.Open("plugins/qa/expertise.json")
	if err != nil {
		onelib.Error.Println("Error opening expertise.json:", err)
		return ""
	}
	defer expertiseFile.Close()
	expertiseBytes, err := io.ReadAll(expertiseFile)
	if err != nil {
		onelib.Error.Println("Error reading expertise.json:", err)
		return ""
	}

	// Read the entire contents of the file into misinfos
	misinfosFile, err := os.Open("plugins/qa/misinfos.json")
	if err != nil {
		onelib.Error.Println("Error opening misinfos.json:", err)
		return ""
	}
	defer misinfosFile.Close()
	misinfosBytes, err := io.ReadAll(misinfosFile)
	if err != nil {
		onelib.Error.Println("Error reading misinfos.json:", err)
		return ""
	}

	templateVars := struct {
		Name string
		Prompt string
		OpenAIKey string
		Channels map[string][]string
		ReplyToQuestions bool
		ReplyToMentions bool
		Expertise string
		Misinfos string
	}{
		Name: LONGNAME,
		Prompt: onelib.GetTextConfig(NAME, "prompt"),
		OpenAIKey: onelib.GetTextConfig(NAME, "openai_key"),
		Channels: getChannelsMap(),
		ReplyToQuestions: onelib.GetBoolConfig(NAME, "reply_to_questions"),
		ReplyToMentions: onelib.GetBoolConfig(NAME, "reply_to_mentions"),
		Expertise: string(expertiseBytes),
		Misinfos: string(misinfosBytes),
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, templateVars)
	if err != nil {
		onelib.Error.Println("Error executing template:", err)
		return ""
	}
	return template.HTML(output.String())
}

func (qamc *QAMissionControlPlugin) Functions() map[string]func(map[string]any) (string, error) {
	return map[string]func(map[string]any) (string, error) {
		"set_prompt": func(args map[string]any) (string, error) {
			onelib.SetTextConfig(NAME, "prompt", args["v"].(string))
			return "Prompt saved!", nil
		},
		"set_openai_key": func(args map[string]any) (string, error) {
			onelib.SetTextConfig(NAME, "openai_key", args["v"].(string))
			return "OpenAI Key saved!", nil
		},
		"set_replies": func(args map[string]any) (string, error) {
			onelib.SetBoolConfig(NAME, "reply_to_questions", args["qs"].(bool))
			onelib.SetBoolConfig(NAME, "reply_to_mentions", args["ms"].(bool))
			return "Reply settings saved!", nil
		},
		"question": func(args map[string]any) (string, error) {
			txt, err := runqa("-q", args["q"].(string), "question", "-p", args["p"].(string))
			if err != nil {
				onelib.Error.Println("Error running qa.py:", err)
				return "", err
			}
			return txt, nil
		},
		"add_channel": func(args map[string]any) (string, error) {
			channels := getChannelsMap()
			channels[args["p"].(string)] = append(channels[args["p"].(string)], args["c"].(string))
			channelsJson, err := json.Marshal(channels)
			if err != nil {
				onelib.Error.Println("[qa] Error encoding channels:", err)
				return "", err
			}
			onelib.SetTextConfig(NAME, "channels", string(channelsJson))
			return "", nil
		},
		"delete_channel": func(args map[string]any) (string, error) {
			channels := getChannelsMap()
			for i, v := range channels[args["p"].(string)] {
				if v == args["c"].(string) {
					channels[args["p"].(string)] = append(channels[args["p"].(string)][:i], channels[args["p"].(string)][i+1:]...)
					break
				}
			}
			channelsJson, err := json.Marshal(channels)
			if err != nil {
				onelib.Error.Println("[qa] Error encoding channels:", err)
				return "", err
			}
			onelib.SetTextConfig(NAME, "channels", string(channelsJson))
			return "", nil
		},
		"add_protocol": func(args map[string]any) (string, error) {
			channels := getChannelsMap()
			channels[args["v"].(string)] = make([]string, 0)
			channelsJson, err := json.Marshal(channels)
			if err != nil {
				onelib.Error.Println("[qa] Error encoding channels:", err)
				return "", err
			}
			onelib.SetTextConfig(NAME, "channels", string(channelsJson))
			return "", nil
		},
		"delete_protocol": func(args map[string]any) (string, error) {
			channels := getChannelsMap()
			delete(channels, args["v"].(string))
			channelsJson, err := json.Marshal(channels)
			if err != nil {
				onelib.Error.Println("[qa] Error encoding channels:", err)
				return "", err
			}
			onelib.SetTextConfig(NAME, "channels", string(channelsJson))
			return "", nil
		},
	}
}

type QuestionIndex struct {
	Ids []onelib.UUID // The IDs of the responses
}

type QuestionAnswer struct {
	Id        onelib.UUID `bson:"_id"` // The ID of the response
	Answer    string      `bson:"a"`   // The answer to the question
	Question  string      `bson:"q"`   // The question asked (can be blank)
	UpVotes   int         `bson:"u"`   // The number of upvotes the answer has
	DownVotes int         `bson:"d"`   // The number of downvotes the answer has
	Date      int64       `bson:"D"`   // The date the answer was posted
}

// QAPlugin is an object for satisfying the Plugin interface.
type QAPlugin struct {
	monitor   *onelib.Monitor
	expertise []string
	channels  map[string][]string

	lastMsg string

	DbLock *sync.RWMutex
}

func runqa(args ...string) (string, error) {
	// Call the python script plugins/qa/qa.py, passing the openai key as an environment variable, and capturing the output.
	_args := []string{"plugins/qa/qa.py", "-e", "plugins/qa/expertise.json", "-mi", "plugins/qa/misinfos.json", "-db", "plugins/qa/db-noembed.csv", "-edb", "plugins/qa/db.csv"}
	cmd := exec.Command("python3", append(_args, args...)...)
	cmd.Env = append(os.Environ(), "OPENAI_API_KEY="+onelib.GetTextConfig(NAME, "openai_key"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		onelib.Debug.Println("qa.py output:", string(out))
		return string(out), err
	}
	return string(out), nil
}

func (qa *QAPlugin) ask_question(msg onelib.Message, sender onelib.Sender) {
	txt, err := runqa("-q", msg.Text(), "question", "-p", onelib.GetTextConfig(NAME, "prompt"))
	if err != nil {
		onelib.Error.Println("Error running qa.py:", err)
		return
	}
	onelib.Debug.Println("Replying:", txt)
	qa.lastMsg = txt
	sender.Location().SendText(txt)
}

func (qa *QAPlugin) stats(msg onelib.Message, sender onelib.Sender) {
	now := time.Now()
	var yearMonth string
	if msg.Text() != "" {
		yearMonth = strings.TrimSpace(msg.Text())
	} else {
		yearMonth = fmt.Sprintf("%d-%d", now.Year(), now.Month())
	}
	indexKey := fmt.Sprintf("%s-index", yearMonth)
	qa.DbLock.RLock()

	// Get the index
	var index QuestionIndex
	err := onelib.Db.GetObj(NAME, indexKey, &index)
	if err != nil {
		qa.DbLock.RUnlock()
		sender.Location().SendText(fmt.Sprintf("No stats found for %s.", yearMonth))
		return
	}
	answers := 0
	totalUpvotes := 0
	totalDownvotes := 0
	for _, id := range index.Ids {
		answers++
		var answer QuestionAnswer
		err := onelib.Db.GetObj(NAME, string(id), &answer)
		if err != nil {
			onelib.Error.Println("Error getting answer:", err)
			continue
		}
		totalUpvotes += answer.UpVotes
		totalDownvotes += answer.DownVotes
	}
	qa.DbLock.RUnlock()
	sender.Location().SendText(fmt.Sprintf("Stats for %s: %d answers, %d upvotes, %d downvotes.", yearMonth, answers, totalUpvotes, totalDownvotes))
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
	if from.Self() {
		if strings.TrimSpace(msg.Text()) == strings.TrimSpace(qa.lastMsg) {
			qa.DbLock.Lock()
			// Add DB entries for the response (FIXME: Question should be logged too)
			now := time.Now()
			indexKey := fmt.Sprintf("%d-%d-index", now.Year(), now.Month())
			questionIndex := new(QuestionIndex)
			err := onelib.Db.GetObj(NAME, indexKey, questionIndex)
			if err != nil {
				questionIndex = new(QuestionIndex)
				questionIndex.Ids = make([]onelib.UUID, 0)
			}
			questionIndex.Ids = append(questionIndex.Ids, msg.UUID())
			onelib.Db.PutObj(NAME, indexKey, questionIndex)
			questionAnswer := new(QuestionAnswer)
			questionAnswer.Id = msg.UUID()
			questionAnswer.Answer = msg.Text()
			questionAnswer.Date = now.Unix()
			onelib.Db.PutObj(NAME, string(msg.UUID()), questionAnswer)
			qa.DbLock.Unlock()
			if from.Protocol() == "discord" { // Add reactions to encourage feedback
				disClient := from.Location().(*discord.DiscordLocation).Client
				disClient.MessageReactionAdd(string(from.Location().UUID()), string(msg.UUID()), "👍")
				disClient.MessageReactionAdd(string(from.Location().UUID()), string(msg.UUID()), "👎")
			}
		}
		return
	}

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
	if onelib.GetBoolConfig(NAME, "reply_to_mentions") && msg.Mentioned() {
		qa.ask_question(msg, from)
	} else if onelib.GetBoolConfig(NAME, "reply_to_questions") {
		// Check if txt contains any of the strings in qa.expertise, and ends in a question mark
		for _, v := range qa.expertise {
			if strings.Contains(txt, v) && strings.HasSuffix(txt, "?") {
				qa.ask_question(msg, from)
				break
			}
		}
	}
}

func (qa *QAPlugin) OnMessageUpdate(from onelib.Sender, update onelib.Message) {
	if from.Self() {
		return // Don't process msg if it's from ourselves
	}
	reaction := update.Reaction()
	if reaction == nil {
		return // We're checking for reactions, let's ignore messages not about those
	}

	record := new(QuestionAnswer)
	qa.DbLock.Lock()
	err := onelib.Db.GetObj(NAME, string(update.UUID()), record)
	if err != nil {
		qa.DbLock.Unlock()
		return // We don't know about this message, let's ignore it
	}
	if reaction.Name == "👍" {
		if reaction.Added {
			// Add to upvote count
			record.UpVotes++
		} else {
			// Remove from upvote count
			record.UpVotes--
		}
		onelib.Db.PutObj(NAME, string(update.UUID()), record)
	} else if reaction.Name == "👎" {
		if reaction.Added {
			// Add to downvote count
			record.DownVotes++
		} else {
			// Remove from downvote count
			record.DownVotes--
		}
		onelib.Db.PutObj(NAME, string(update.UUID()), record)
	}
	qa.DbLock.Unlock()
}

// Implements returns a map of commands and monitor the plugin implements.
func (qa *QAPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"q": qa.ask_question, "question": qa.ask_question, "stats": qa.stats}, qa.monitor
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (qa *QAPlugin) Remove() {
}
