package main

/* CONFIG SPEC

Bot user should only have read permissions to config file (unless you'd like plugins to be able to edit the config file)

Config values should use those set in config file until overriden via database entry (IE: value changed via command in
chat).

Configuration file will be in TOML, whatever version is most convenient.

*/

/* TODO

- DBs (LevelDB(x) / MongoDB)
- Plugin system
    - Setup default plugin folder
- Protocol system
    - Setup default protocol folder

(x) == completed
*/

import (
	. "github.com/TheDiscordian/onebot/loggers"
	. "github.com/TheDiscordian/onebot/onetool"
	. "github.com/TheDiscordian/onebot/onetype"
	"github.com/pelletier/go-toml"
)

const (
	NAME    = "OneBot"
	VERSION = "v1.0.0"
)

// TODO set path
// Should set all variables unless DB overrides. This also inits DB.
func LoadConfig() {
	config, err := toml.LoadFile("onebot.toml")

	if err != nil {
		Error.Panicln("Error loading config", err.Error())
	}

	PluginDir = config.Get("general.plugin_path").(string)
	ProtocolDir = config.Get("general.protocol_path").(string)

	DbEngine = config.Get("database.engine").(string)
	if DbEngine == "leveldb" {
		Db = OpenLevelDB(config.Get("database.leveldb_path").(string))
	} else {
		Error.Panicln("database.engine = '%s', only 'leveldb' implemented.", DbEngine)
	}
}

/* DATABASE SPEC

DBs should perhaps support conversion to other DBs for portability.

Plugins may not include a "." or "~" in key names.

LevelDB indexes will be stored as "tableName.indexKey.indexValue", the value contains the key to get the value
    - First index key will be "tableName.indexKey.." (value is nil)
    - Last index key will be "tableName.indexKey.~" (value is nil)
LevelDB values will be stored as "tableName.key", key will be the ID of the object
LevelDB keys will be generated as regular MongoDB ObjectIDs in bytes, unless explicitly specified

*/

func main() {
	LogFile = "onebot.log"
	InitLoggers()
	Info.Printf("Starting up %s %s...\n", NAME, VERSION)
	LoadConfig()
	LoadPlugin("firstplugin")

	Info.Println("Shutting down...")
	Db.Close()
}
