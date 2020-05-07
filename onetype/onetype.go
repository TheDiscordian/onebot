package onetype

// TODO How hard would it be for a plugin with this spec to tag a user on both Matrix and Discord?

type UUID string // A unique identifier

var (
	// TODO Make these concurrent-safe
	Protocols map[string]Protocol // Key is protocol name (ex: "discord")
	Plugins   map[string]Plugin   // Key is plugin name (ex: "admin_tools")
	Commands  map[string]Command  // Key is command trigger (ex: "help")
	Monitors  []Monitor

	Db Database // Db is configured via config file only

	DbEngine    string
	PluginDir   string
	ProtocolDir string
)

type Database interface {
	Get(table, key string) ([]byte, error)           // Retrieves value by key directly
	Search(table, field, key string) ([]byte, error) // Searches for key in field, containing key (IE: field:'username', key:'admin'), using an index if exists.
	Put(table, key string, value []byte) error       // Inserts value into key, erasing any potential previous value.
	SetIndex(table, key string) error                // Sets an index on key.
	Close() error                                    // Terminate a database session (only run if nothing is using the database).
}

type Location interface {
	DisplayName() string // Display name of the location
	Topic() string       // The topic of the location
	// Picture // TODO The avatar of the location
	UUID() UUID           // Unique identifier for the location
	Send(msg Message)     // Sends a message to the location
	SendText(text string) // Sends text to the location
	Protocol() string     // Returns the name of the protocol the location is in
}

// Message contains information either being sent or received
type Message interface {
	Text() string // The unformatted text being received (minus the trigger word for commands)
	// Reactions() []Reaction // TODO The reactions on the message
	Raw() []byte // The raw data received
}

// Sender contains information about who and where a message came from
type Sender interface {
	DisplayName() string // Display name of the sender
	Username() string    // Username of the sender (often unknown, should return an empty string)
	UUID() UUID          // Unique identifier for the sender
	// Picture // TODO The avatar of the location
	Location() Location   // The location where this sender sent the message from
	Protocol() string     // Returns the protocol name responsible for the sender
	Send(msg Message)     // Sends a Message to the sender
	SendText(text string) // Sends text to the sender
}

/* PROTOCOL SPEC

Plugins should contain a function named "Load() Protocol".

*/

// Protocol contains information about a protocol plugin
type Protocol interface {
	Name() string                  // The name of the protocol, used in the protocol map (should be same as filename, minus extension)
	LongName() string              // The display name of the protocol
	Version() int                  // The version of the protocol
	NewMessage(raw []byte) Message // Returns a new Message object built from []byte
	Send(to UUID, msg Message)     // Sends a Message to a location
	SendText(to UUID, text string) // Sends text to a location
	Remove()                       // Called when the protocol is about to be terminated
}

/* PLUGIN SPEC

Plugins should contain a function named "Load() Plugin"

*/

type Plugin interface {
	Name() string                     // The name of the plugin, used in the plugin map (should be same as filename, minus extension)
	LongName() string                 // The display name of the plugin
	Version() int                     // The version of the plugin
	Implements() ([]Command, Monitor) // Returns lists of commands and monitor the plugin implements
	Remove()                          // Called when the plugin is about to be terminated
}

type Monitor interface {
	OnMessage(from Sender, msg Message)         // Called on every message
	OnMessageWithText(from Sender, msg Message) // Called on every message containing text
	//    OnPresenceUpdate(from Sender, update UserPresence) // Called on user presence update
	//    OnLocationUpdate(from Location, update LocationPresence) // Called on location update
}

type Command func(msg Message, sender Sender)
