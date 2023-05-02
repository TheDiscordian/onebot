// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package onelib

import (
	"fmt"
	"github.com/pelletier/go-toml"
)

var config *toml.Tree

// GetTextConfig returns a config value, checking the DB first, expecting a string.
func GetTextConfig(plugin, key string) string {
	if txt, _ := Db.GetString(plugin, key); txt != "" {
		return txt
	}
	if cfg := config.Get(fmt.Sprintf("%s.%s", plugin, key)); cfg != nil {
		return cfg.(string)
	}
	return ""
}

// GetIntConfig returns a config value, checking the DB first, expecting an int.
func GetIntConfig(plugin, key string) (int, error) {
	if num, err := Db.GetInt(plugin, key); err == nil {
		return num, nil
	}
	if cfg := config.Get(fmt.Sprintf("%s.%s", plugin, key)); cfg != nil {
		return int(cfg.(int64)), nil
	}
	return 0, fmt.Errorf("config key '%s.%s' not found", plugin, key)
}

// SetTextConfig sets a string config value.
func SetTextConfig(plugin, key, text string) {
	Db.PutString(plugin, key, text)
}

// LoadConfig loads the configuration file and inits the DB. This does not respect locks on config, do not run this
// while any goroutines are running. Ultimately this will check the DB before loading from the config file.
// TODO set default config path
// TODO add an option to check DB before config file
func LoadConfig() {
	var err error
	config, err = toml.LoadFile("onebot.toml")

	if err != nil {
		Error.Panicln("Error loading config", err.Error())
	}
	DefaultPrefix = config.Get("general.default_prefix").(string)
	DefaultNickname = config.Get("general.default_nickname").(string)
	DefaultAvatar = config.Get("general.default_avatar").(string)

	PluginDir = config.Get("general.plugin_path").(string)
	pluginList := config.Get("general.plugins").([]interface{})
	PluginLoadList = make([]string, len(pluginList))
	for i, plugName := range pluginList {
		PluginLoadList[i] = plugName.(string)
	}

	ProtocolDir = config.Get("general.protocol_path").(string)
	protocolList := config.Get("general.protocols").([]interface{})
	ProtocolLoadList = make([]string, len(protocolList))
	for i, protoName := range protocolList {
		ProtocolLoadList[i] = protoName.(string)
	}

	DbEngine = config.Get("database.engine").(string)
	if DbEngine == "leveldb" {
		Db = openLevelDB(config.Get("database.leveldb_path").(string))
	} else {
		Error.Panicf("database.engine = '%s', only 'leveldb' implemented.\n", DbEngine)
	}
}
