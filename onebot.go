// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package main

/* CONFIG SPEC

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
	. "github.com/TheDiscordian/onebot/onelib"
	"os"
	"os/signal"
	"syscall"
)

const (
    // NAME is the default display name of the bot
	NAME    = "OneBot"
    // VERSION is the displayed version of the bot
	VERSION = "v0.0.0"
)

/* DATABASE SPEC

DBs should perhaps support conversion to other DBs for portability.

Plugins may not include a "." or "~" in key names.

LevelDB indexes will be stored as "tableName.indexKey.indexValue", the value contains the key to get the value
LevelDB values will be stored as "tableName.key", key will be the ID of the object
LevelDB keys will be generated as regular MongoDB ObjectIDs in bytes, unless explicitly specified

*/

func main() {
	LogFile = "onebot.log"
	InitLoggers()
	Info.Printf("Starting up %s %s...\n", NAME, VERSION)
	LoadConfig()
	Info.Println("Loading protocols...")
	LoadProtocols()
	Info.Println("Loading plugins...")
	LoadPlugins()

	defer func() {
		Info.Println("Shutting down...")
		UnloadPlugins()
		UnloadProtocols()
		Db.Close()
	}()

	signal.Notify(Quit, os.Interrupt, syscall.SIGTERM)
	<-Quit
}
