// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package onelib

// TODO How hard would it be for a plugin with this spec to tag a user on both Matrix and Discord?

import (
	"os"
	"sync"
)

// UUID represents a unique identifier, usually for a Location (room) or a Sender (user).
type UUID string

const(
	// VERSION is the displayed version of the bot
	VERSION = "v0.1.1-WIP"
)

var (
	// DefaultPrefix is the default prefix used to trigger commands. (Ex: In ",say Hello" the "," would be the prefix)
	DefaultPrefix string
	// DefaultNickname is the default display name for the bot.
	DefaultNickname string
	// DefaultAvatar is a URL to the default avatar for the bot.
	DefaultAvatar string

	// Protocols is a map of loaded Protocols. Key is protocol name (ex: "discord")
	Protocols *ProtocolMap
	// Plugins is a map of loaded Plugins. Key is plugin name (ex: "admin_tools")
	Plugins *PluginMap
	// Commands is a map of loaded Commands. Key is command trigger (ex: "help")
	Commands *CommandMap
	// Monitors is a slice of loaded Monitors.
	Monitors *MonitorSlice

	// Db is configured via config file only.
	Db Database

	// DbEngine is the database type currently in use.
	DbEngine string
	// PluginDir is the directory plugins are stored.
	PluginDir string
	// PluginLoadList is only used for loading the default plugins.
	PluginLoadList []string
	// ProtocolDir is the directory protocols are stored.
	ProtocolDir string
	// ProtocolLoadList is only used for loading the default protocols.
	ProtocolLoadList []string

	// Quit - the main thread will watch this to terminate the process.
	Quit chan os.Signal
)

func init() {
	Protocols = NewProtocolMap()
	Plugins = NewPluginMap()
	Commands = NewCommandMap()
	Monitors = NewMonitorSlice()
	Quit = make(chan os.Signal)
}

// ProtocolMap is a concurrent-safe map of protocols.
type ProtocolMap struct {
	protocols map[string]Protocol
	lock      *sync.RWMutex
}

// NewProtocolMap returns a new concurrent-safe ProtocolMap
func NewProtocolMap() *ProtocolMap {
	pm := &ProtocolMap{lock: new(sync.RWMutex)}
	pm.protocols = make(map[string]Protocol, 1)
	return pm
}

// Get a protocol from the ProtocolMap
func (pm *ProtocolMap) Get(protocolName string) Protocol {
	pm.lock.RLock()
	protocol := pm.protocols[protocolName]
	pm.lock.RUnlock()
	return protocol
}

// Put an already loaded protocol into the ProtocolMap
func (pm *ProtocolMap) Put(protocolName string, protocol Protocol) {
	pm.lock.Lock()
	pm.protocols[protocolName] = protocol
	pm.lock.Unlock()
}

// Delete removes the protocol from the active protocol list, calling the protocol's unload method via goroutine
func (pm *ProtocolMap) Delete(protocolName string) {
	pm.lock.Lock()
	go pm.protocols[protocolName].Remove()
	delete(pm.protocols, protocolName)
	pm.lock.Unlock()
}

// DeleteAll removes all protocols from the active protocol list, calling the protocol's unload method via goroutine
func (pm *ProtocolMap) DeleteAll() {
	pm.lock.Lock()
	for protoName, proto := range pm.protocols {
		go proto.Remove()
		delete(pm.protocols, protoName)
	}
	pm.lock.Unlock()
}

// List returns a list of all protocols in the ProtocolMap.
func (pm *ProtocolMap) List() []string {
	pm.lock.RLock()
	list := make([]string, 0, len(pm.protocols))
	for protocolName := range pm.protocols {
		list = append(list, protocolName)
	}
	pm.lock.RUnlock()
	return list
}

// PluginMap is a concurrent-safe map of plugins.
type PluginMap struct {
	plugins map[string]Plugin
	lock    *sync.RWMutex
}

// NewPluginMap returns a new concurrent-safe PluginMap.
func NewPluginMap() *PluginMap {
	pm := &PluginMap{lock: new(sync.RWMutex)}
	pm.plugins = make(map[string]Plugin, 2)
	return pm
}

// Get a plugin from the PluginMap
func (pm *PluginMap) Get(pluginName string) Plugin {
	pm.lock.RLock()
	plugin := pm.plugins[pluginName]
	pm.lock.RUnlock()
	return plugin
}

// Put an already loaded plugin into the PluginMap
func (pm *PluginMap) Put(pluginName string, plugin Plugin) {
	pm.lock.Lock()
	pm.plugins[pluginName] = plugin
	pm.lock.Unlock()
}

// Delete removes the plugin from the active plugin list, calling the plugin's unload method via goroutine
func (pm *PluginMap) Delete(pluginName string) {
	pm.lock.Lock()
	go pm.plugins[pluginName].Remove()
	delete(pm.plugins, pluginName)
	pm.lock.Unlock()
}

// DeleteAll removes all plugins from the active plugin list, calling the plugin's unload method via goroutine
func (pm *PluginMap) DeleteAll() {
	pm.lock.Lock()
	for plugName, plug := range pm.plugins {
		go plug.Remove()
		delete(pm.plugins, plugName)
	}
	pm.lock.Unlock()
}

// List returns a list of all plugins in the PluginMap.
func (pm *PluginMap) List() []string {
	pm.lock.RLock()
	list := make([]string, 0, len(pm.plugins))
	for pluginName := range pm.plugins {
		list = append(list, pluginName)
	}
	pm.lock.RUnlock()
	return list
}

// CommandMap is a concurrent-safe map of commands
type CommandMap struct {
	commands map[string]Command
	lock     *sync.RWMutex
}

// NewCommandMap returns a new concurrent-safe CommandMap.
func NewCommandMap() *CommandMap {
	cm := &CommandMap{lock: new(sync.RWMutex)}
	cm.commands = make(map[string]Command, 4)
	return cm
}

// Get a command from the CommandMap.
func (cm *CommandMap) Get(commandName string) Command {
	cm.lock.RLock()
	command := cm.commands[commandName]
	cm.lock.RUnlock()
	return command
}

// Put a command into the CommandMap.
func (cm *CommandMap) Put(commandName string, command Command) {
	cm.lock.Lock()
	cm.commands[commandName] = command
	cm.lock.Unlock()
}

// Delete removes the command from the active command list, calling the command's unload method via goroutine.
func (cm *CommandMap) Delete(commandName string) {
	cm.lock.Lock()
	delete(cm.commands, commandName)
	cm.lock.Unlock()
}

// DeleteSet removes a map of commands from the active command list, calling the command's unload method via goroutine.
func (cm *CommandMap) DeleteSet(set map[string]Command) {
	if set == nil {
		return
	}
	cm.lock.Lock()
	for commandName := range set {
		delete(cm.commands, commandName)
	}
	cm.lock.Unlock()
}

// DeleteAll removes all commands from the active command list.
func (cm *CommandMap) DeleteAll() {
	cm.lock.Lock()
	cm.commands = make(map[string]Command)
	cm.lock.Unlock()
}

// MonitorSlice is a concurrent-safe slice of monitors.
type MonitorSlice struct {
	monitors []*Monitor
	lock     *sync.RWMutex
}

// NewMonitorSlice returns a new concurrent-safe MonitorSlice.
func NewMonitorSlice() *MonitorSlice {
	ms := &MonitorSlice{lock: new(sync.RWMutex)}
	ms.monitors = make([]*Monitor, 0)
	return ms
}

// Get copy of monitor slice for reading.
func (ms *MonitorSlice) Get() []*Monitor {
	ms.lock.RLock()
	slice := ms.monitors
	ms.lock.RUnlock()
	return slice
}

// Put a monitor into the MonitorSlice.
func (ms *MonitorSlice) Put(monitor *Monitor) {
	ms.lock.Lock()
	ms.monitors = append(ms.monitors, monitor)
	ms.lock.Unlock()
}

// Delete removes the monitor from the active monitor list.
func (ms *MonitorSlice) Delete(monitor *Monitor) {
	if monitor == nil {
		return
	}
	ms.lock.Lock()
	for i, mon := range ms.monitors {
		if mon == monitor {
			ms.monitors = append(ms.monitors[:i], ms.monitors[i+1:]...)
			break
		}
	}
	ms.lock.Unlock()
}

// DeleteAll removes all monitors from the active monitor list.
func (ms MonitorSlice) DeleteAll() {
	ms.lock.Lock()
	ms.monitors = make([]*Monitor, 0)
	ms.lock.Unlock()
}

// Database represents a database connection. It's meant to be simple, to work for most general usage.
type Database interface {
	Get(table, key string) (map[string]interface{}, error)           // Retrieves value by key directly
	GetString(table, key string) (string, error)                     // Retrieve a string stored with PutString.
	GetInt(table, key string) (int, error)                           // Retrieve an int stored with PutInt.
	GetObj(table, key string, obj interface{}) error                 // Retrieve an object stored with PutObj.
	Search(table, field, key string) (map[string]interface{}, error) // Searches for key in field, containing key (IE: field:'username', key:'admin'), using an index if exists.
	Put(table string, data map[string]interface{}) ([]byte, error)   // Inserts data into database, using "_id" field as key, generating one if none exists. Returns key.
	PutString(table, key, text string) error                         // Inserts text at location "key" for retrieval via GetString
	PutInt(table, key string, i int) error                           // Inserts an integer at location "key" for retrieval via GetInt
	PutObj(table, key string, obj interface{}) error                 // Inserts an object at location "key" for retrieval via GetObj
	Remove(table, key string) error                                  // Removes an object at location "key"
	SetIndex(table, field string) error                              // Sets an index on field. If using LevelDB, values in this field must be unique.
	Close() error                                                    // Terminate a database session (only run if nothing is using the database).
}

// Location represents a location in a protocol. Think like a room, or a group.
type Location interface {
	DisplayName() string // Display name of the location
	Nickname() string    // The nickname of the bot in the location
	Topic() string       // The topic of the location
	// Picture // TODO The avatar of the location
	UUID() UUID                                   // Unique identifier for the location
	Send(msg Message)                             // Sends a message to the location
	SendText(text string)                         // Sends text to the location
	SendFormattedText(text, formattedText string) // Sends formatted text to the location (correctness might vary between protocols)
	Protocol() string                             // Returns the name of the protocol the location is in
}

// Message contains information either being sent or received
type Message interface {
	Text() string                      // The unformatted text being received (minus the trigger for commands)
	FormattedText() string             // the formatted text being received (minus trigger for commands)
	StripPrefix(prefix string) Message // Returns a copy of the message with `prefix + commandName + " "` stripped (Ex: "!say Hello" becomes "Hello")
	Raw() []byte                       // The raw data received
	UUID() UUID                        // Unique identifier for the message (can be empty)
	Mentioned() bool                   // True if the bot was mentioned in the message
	Reaction() *Emoji                  // Returns an emoji that was either added or removed, or nil if none
	// Reactions() []Reaction // TODO The reactions on the message
}

// Emoji contains data which should be useful around emojis.
type Emoji struct {
	ID    UUID   // The UUID of the emoji, should never be blank.
	Name  string // Either a unicode representation of the emoji, or a name.
	Added bool   // If true, the emoji was just added as a reaction. If false, it was just removed. Ignore field on reaction lists (See Message.Reactions()).
}

// Sender contains information about who and where a message came from
type Sender interface {
	DisplayName() string // Display name of the sender
	Username() string    // Username of the sender (often unknown, should return an empty string if so)
	UUID() UUID          // Unique identifier for the sender
	// Picture // TODO The avatar of the sender
	Location() Location                           // The location where this sender sent the message from
	Protocol() string                             // Returns the protocol name responsible for the sender
	Self() bool                                   // Returns true if the sender is the bot
	Send(msg Message)                             // Sends a Message to the sender
	SendText(text string)                         // Sends text to the sender
	SendFormattedText(text, formattedText string) // Sends formatted text to the sender (correctness might vary between protocols)
}

/* PROTOCOL SPEC

Plugins should contain a function named "Load() Protocol".

*/

// Protocol contains information about a protocol plugin
type Protocol interface {
	Name() string                                          // The name of the protocol, used in the protocol map (should be same as filename, minus extension)
	LongName() string                                      // The display name of the protocol
	Version() string                                       // The version of the protocol
	NewMessage(raw []byte) Message                         // Returns a new Message object built from []byte (TODO: I hate this)
	Send(to UUID, msg Message)                             // Sends a Message to a location
	SendText(to UUID, text string)                         // Sends text to a location
	SendFormattedText(to UUID, text, formattedText string) // Sends formatted text to a location (correctness might vary between protocols)
	Remove()                                               // Called when the protocol is about to be terminated
}

/* PLUGIN SPEC

Plugins should contain a function named "Load() Plugin"

*/

// Plugin is an object representing a OneBot plugin.
type Plugin interface {
	Name() string                               // The name of the plugin, used in the plugin map (should be same as filename, minus extension)
	LongName() string                           // The display name of the plugin
	Version() string                            // The version of the plugin
	Implements() (map[string]Command, *Monitor) // Returns a map of commands and monitor the plugin implements
	Remove()                                    // Called when the plugin is about to be terminated
}

// Monitor is a struct containing pointers to functions which are called on certain events (can be nil).
type Monitor struct {
	OnMessage         func(from Sender, msg Message)    // Called on every new message
	OnMessageWithText func(from Sender, msg Message)    // Called on every new message containing text
	OnMessageUpdate   func(from Sender, update Message) // Called on message update (IE: edit, reaction)
	//    OnPresenceUpdate func(from Sender, update UserPresence) // Called on user presence update
	//    OnLocationUpdate func(from Location, update LocationPresence) // Called on location update
}

// Command is a function called when a certain key is triggered.
type Command func(msg Message, sender Sender)
