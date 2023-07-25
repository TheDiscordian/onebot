// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"fmt"
	"strings"
	"encoding/json"
	"os"
	"time"
	"io"
	"os/exec"
	"sync"

	"github.com/TheDiscordian/onebot/libs/discord"
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
	qa.replyToQuestions = onelib.GetBoolConfig(NAME, "reply_to_questions")
	qa.replyToMentions = onelib.GetBoolConfig(NAME, "reply_to_mentions")
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

	updateDbs := onelib.GetBoolConfig(NAME, "update_databases")
	if updateDbs {
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
	}

	qa.monitor = &onelib.Monitor{
		OnMessageWithText: qa.OnMessageWithText,
		OnMessageUpdate: qa.OnMessageUpdate,
	}

	qa.DbLock = new(sync.RWMutex)
	return qa
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
	Date      int64	      `bson:"D"`   // The date the answer was posted
}

// QAPlugin is an object for satisfying the Plugin interface.
type QAPlugin struct {
	monitor *onelib.Monitor
	expertise []string
	openaiKey string
	prompt string
	channels map[string][]string

	replyToQuestions bool
	replyToMentions bool

	lastMsg string

	DbLock *sync.RWMutex
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

	ask := false
	txt := strings.ToLower(msg.Text())
	if qa.replyToMentions && msg.Mentioned() {
		ask = true
	}
	if !ask && qa.replyToQuestions {
		// Check if txt contains any of the strings in qa.expertise, and ends in a question mark
		for _, v := range qa.expertise {
			if strings.Contains(txt, v) && strings.HasSuffix(txt, "?") {
				ask = true
				break
			}
		}
	}

	if ask {
		qa.ask_question(msg, from)
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
