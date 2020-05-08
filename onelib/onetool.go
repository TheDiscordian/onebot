package onelib

import (
	"errors"
	"fmt"
	"plugin"
)

// TODO Plugin / protocol list should save when manually changed

// LoadPlugin loads a plugin by filename (minus extension)
func LoadPlugin(name string) error {
	rawPlug, err := plugin.Open(fmt.Sprintf("%s/%s.so", PluginDir, name))
	if err != nil {
		return err
	}
	loadF, err := rawPlug.Lookup("Load")
	if err != nil {
		return err
	}
	plug := loadF.(func() Plugin)()
	Plugins.Put(name, plug)
	Info.Printf("Loaded '%s' version %d.\n", plug.LongName(), plug.Version())
	return nil
}

// LoadPlugins loads all plugins in the plugin directory
func LoadPlugins() {
	for _, pluginName := range PluginLoadList {
		err := LoadPlugin(pluginName)
		if err != nil {
			Error.Printf("Failed to load plugin '%s': %v\n", pluginName, err)
		}
	}
}

func UnloadPlugin(name string) error {
	plug := Plugins.Get(name)
	if plug == nil {
		return errors.New(fmt.Sprintf("Plugin '%s' not loaded.", name))
	}
	Plugins.Delete(name)
	commands, monitor := plug.Implements()
	Monitors.Delete(monitor)
	Commands.DeleteSet(commands)
	return nil
}

// Unloads every plugin, calling their unload routines.
func UnloadPlugins() {
	Monitors.DeleteAll()
	Commands.DeleteAll()
	Plugins.DeleteAll()
}

// LoadProtocol loads a protocol by filename (minus extension)
func LoadProtocol(name string) error {
	rawProto, err := plugin.Open(fmt.Sprintf("%s/%s.so", ProtocolDir, name))
	if err != nil {
		return err
	}
	loadF, err := rawProto.Lookup("Load")
	if err != nil {
		return err
	}
	proto := loadF.(func() Protocol)()
	Protocols.Put(name, proto)
	Info.Printf("Loaded '%s' version %d.\n", proto.LongName(), proto.Version())
	return nil
}

// LoadProtocols loads all protocols in the protocol directory
func LoadProtocols() {
	for _, protocolName := range ProtocolLoadList {
		err := LoadProtocol(protocolName)
		if err != nil {
			Error.Printf("Failed to load protocol '%s': %v\n", protocolName, err)
		}
	}
}

// Unloads every protocol, calling their unload routines.
func UnloadProtocols() {
	Protocols.DeleteAll()
}

// ProcessMessage processes command and monitor triggers, spawning a new goroutine for every trigger.
func ProcessMessage(msg Message, sender Sender) {
	// TODO
}
