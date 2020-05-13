// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package onelib

import "github.com/pelletier/go-toml"

// TODO set default config path
// TODO implement DB override
// Should set all variables unless DB overrides. This also inits DB.
func LoadConfig() {
	config, err := toml.LoadFile("onebot.toml")

	if err != nil {
		Error.Panicln("Error loading config", err.Error())
	}

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
