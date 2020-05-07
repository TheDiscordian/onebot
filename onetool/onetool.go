package onetool

import (
	"fmt"
	. "github.com/TheDiscordian/onebot/loggers"
	. "github.com/TheDiscordian/onebot/onetype"
	"plugin"
)

// TODO Plugin / protocol list should save when manually changed

// LoadPlugin loads a plugin by filename (minus extension)
func LoadPlugin(name string) {
	rawPlug, err := plugin.Open(fmt.Sprintf("%s/%s.so", PluginDir, name))
	if err != nil {
		Error.Println("Failed to open plugin:", err)
	}
	loadF, err := rawPlug.Lookup("Load")
	if err != nil {
		Error.Printf("Failed to locate load function for %s: %s\n", name, err)
	}
	plug := Plugin(loadF.(func() Plugin)())
	Info.Printf("Loaded '%s' version %d.\n", plug.LongName(), plug.Version())
	// TODO
}

// LoadPlugins loads all plugins in the plugin directory
func LoadPlugins() {
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
