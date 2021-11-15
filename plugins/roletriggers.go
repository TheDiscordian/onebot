// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package main

import (
	"strings"

	"github.com/TheDiscordian/onebot/libs/discord"
	"github.com/TheDiscordian/onebot/onelib"
)

const (
	// NAME is same as filename, minus extension
	NAME = "roletriggers"
	// LONGNAME is what's presented to the user
	LONGNAME = "Role Triggers Plugin"
	// VERSION of the plugin
	VERSION = "v0.0.0"

	DB_TABLE = "roletriggers"
)

// Load returns the Plugin object.
func Load() onelib.Plugin {
	return &RoleTriggersPlugin{
		monitor: &onelib.Monitor{
			OnMessageUpdate: handlereact,
		},
	}
}

func handlereact(sender onelib.Sender, msg onelib.Message) {
	if sender.Protocol() != "discord" {
		return
	}
	reaction := msg.Reaction()
	if reaction == nil {
		return
	}

	roleId, err := onelib.Db.GetString(DB_TABLE, string(msg.UUID())+"_"+reaction.Name)
	if err != nil {
		return // not found, just return
	}

	disLoc := sender.Location().(*discord.DiscordLocation)
	if reaction.Added {
		// add role
		err = disLoc.Client.GuildMemberRoleAdd(string(disLoc.GuildID), string(sender.UUID()), roleId)
	} else {
		// remove role
		err = disLoc.Client.GuildMemberRoleRemove(string(disLoc.GuildID), string(sender.UUID()), roleId)
	}
	if err != nil {
		onelib.Error.Println("["+NAME+"] Error adding/removing role:", err)
	}
}

func roleid(msg onelib.Message, sender onelib.Sender) {
	if sender.Protocol() != "discord" {
		return
	}
	txt := msg.Text()
	if txt == "" || len(txt) < 5 || txt[:2] != "<@" {
		return
	}
	// disLoc := sender.Location().(*discord.DiscordLocation)
	roleId := txt[3 : len(txt)-1]
	sender.Location().SendText("RoleID: " + roleId)
}

// !addtrigger msgID emojiName @role
func addtrigger(msg onelib.Message, sender onelib.Sender) {
	if sender.Protocol() != "discord" {
		return
	}
	if sender.UUID() != discord.DiscordAdminId {
		return
	}
	txt := msg.Text()
	if txt == "" || len(txt) < 9 {
		return
	}
	splitTxt := strings.Split(txt, " ")
	if len(splitTxt) != 3 {
		return
	}
	// Get msg id
	msgId := splitTxt[0]
	// Get emoji name
	if len(splitTxt[1]) < 1 {
		return
	}
	var emojiName string
	if splitTxt[1][0] == ':' {
		emojiName = splitTxt[1][1 : len(splitTxt[1])-1]
	} else if len(splitTxt[1]) > 5 && splitTxt[1][0] == '<' {
		emojiName = strings.Split(splitTxt[1], ":")[1]
	} else {
		emojiName = splitTxt[1]
	}
	// Get role id
	if len(splitTxt[2]) < 5 || splitTxt[2][:2] != "<@" {
		return
	}
	roleId := splitTxt[2][3 : len(splitTxt[2])-1]

	err := onelib.Db.PutString(DB_TABLE, msgId+"_"+emojiName, roleId)
	if err != nil {
		sender.Location().SendText("Failed to add trigger: " + err.Error())
		return
	}
	sender.Location().SendText("Trigger added successfully!")
}

func removetrigger(msg onelib.Message, sender onelib.Sender) {
	if sender.Protocol() != "discord" {
		return
	}
	if sender.UUID() != discord.DiscordAdminId {
		return
	}
	txt := msg.Text()
	if txt == "" || len(txt) < 9 {
		return
	}
	splitTxt := strings.Split(txt, " ")
	if len(splitTxt) != 2 {
		return
	}
	// Get msg id
	msgId := splitTxt[0]
	// Get emoji name
	if len(splitTxt[1]) < 1 {
		return
	}
	var emojiName string
	if splitTxt[1][0] == ':' {
		emojiName = splitTxt[1][1 : len(splitTxt[1])-1]
	} else if len(splitTxt[1]) > 5 && splitTxt[1][0] == '<' {
		emojiName = strings.Split(splitTxt[1], ":")[1]
	} else {
		emojiName = splitTxt[1]
	}

	_, err := onelib.Db.GetString(DB_TABLE, msgId+"_"+emojiName)
	if err != nil {
		sender.Location().SendText("Failed to find trigger: " + err.Error())
		return
	}
	err = onelib.Db.Remove(DB_TABLE, msgId+"_"+emojiName)
	if err != nil {
		sender.Location().SendText("Failed to remove trigger: " + err.Error())
		return
	}
	sender.Location().SendText("Trigger removed successfully!")
}

// RoleTriggersPlugin is an object for satisfying the Plugin interface.
type RoleTriggersPlugin struct {
	monitor *onelib.Monitor
}

// Name returns the name of the plugin, usually the filename.
func (rt *RoleTriggersPlugin) Name() string {
	return NAME
}

// LongName returns the display name of the plugin.
func (rt *RoleTriggersPlugin) LongName() string {
	return LONGNAME
}

// Version returns the version of the plugin, usually in the format of "v0.0.0".
func (rt *RoleTriggersPlugin) Version() string {
	return VERSION
}

// Implements returns a map of commands and monitor the plugin implements.
func (rt *RoleTriggersPlugin) Implements() (map[string]onelib.Command, *onelib.Monitor) {
	return map[string]onelib.Command{"roleid": roleid, "addtrigger": addtrigger, "at": addtrigger, "removetrigger": removetrigger, "rt": removetrigger}, rt.monitor
}

// Remove is necessary to satisfy the Plugin interface, it does nothing.
func (rt *RoleTriggersPlugin) Remove() {
}
