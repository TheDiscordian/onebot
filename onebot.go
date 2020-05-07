package main

/* CONFIG SPEC

Bot user should only have read permissions to config file (unless you'd like plugins to be able to edit the config file)

Config values should use those set in config file until overriden via database entry (IE: value changed via command in
chat).

Configuration file will be in TOML, whatever version is most convenient.

*/

/* TODO

- DBs (LevelDB / MongoDB)
- Plugin system
    - Setup default plugin folder
- Protocol system
    - Setup default protocol folder
- Scan DB for requests (IE: plugin loads)
*/

import (
	. "github.com/TheDiscordian/onebot/loggers"
	. "github.com/TheDiscordian/onebot/onetype"
	"github.com/pelletier/go-toml"
)

const (
	NAME    = "OneBot"
	VERSION = "v1.0.0"
)

var (
	Protocols map[string]Protocol // Key is protocol name (ex: "discord")
	Plugins   map[string]Plugin   // Key is plugin name (ex: "admin_tools")
	Commands  map[string]Command  // Key is command trigger (ex: "help")
	Monitors  []Monitor
	Db        Database // Db is configured via config file only
)

// TODO set path
// Should set all variables unless DB overrides. This also inits DB.
func LoadConfig() {
	config, err := toml.LoadFile("onebot.toml")

	if err != nil {
		Error.Panicln("Error loading config", err.Error())
	}

	dbEngine := config.Get("database.engine").(string)
	if dbEngine == "leveldb" {
		Db = OpenLevelDB(config.Get("database.leveldb_path").(string))
	} else {
		Error.Panicln("database.engine = '%s', only 'leveldb' implemented.", dbEngine)
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

// RemoteLoadPlugin loads the plugin 'name' from plugin 'plugin' and location 'location', saving the new list.
func RemoteLoadPlugin(name string, plugin string, location UUID) {
	// TODO
}

// LoadPlugin loads a plugin by filename (minus extension)
func LoadPlugin(name string) {
	// TODO
}

// LoadPlugins loads all plugins in the plugin directory
func LoadPlugins() {
	// TODO
}

// RemoteLoadProtocol loads the protocol 'name' from protocol 'protocol' and location 'location', saving the new list.
func RemoteLoadProtocol(name string, protocol string, location UUID) {
	// TODO
}

// LoadProtocol loads a protocol by filename (minus extension)
func LoadProtocol(name string) {
	// TODO
}

// LoadProtocols loads all protocols in the protocol directory
func LoadProtocols() {
	// TODO
}

func main() {
	LogFile = "onebot.log"
	InitLoggers()
	Info.Printf("Starting up %s %s...\n", NAME, VERSION)
	LoadConfig()

	Info.Println("Shutting down...")
	Db.Close()
}
