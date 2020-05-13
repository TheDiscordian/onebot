// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package onelib

import (
	"fmt"
	"plugin"
	"strings"
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

	commands, _ := plug.Implements()
	for trigger, command := range commands {
		Commands.Put(trigger, command)
	}

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
		return fmt.Errorf("Plugin '%s' not loaded.", name)
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

// getcommand returns the command using the line of text containing the command and the expected prefix (doesn't verify
// prefix presence).
func getcommand(prefix, line string) string {
	text := line[len(prefix):]
	i := strings.Index(text, " ")
	if i == -1 {
		return text
	}
	return text[:i]
}

// ProcessMessage processes command and monitor triggers, spawning a new goroutine for every trigger.
func ProcessMessage(prefix string, msg Message, sender Sender) {
	text := msg.Text()
	Debug.Println("Message text:", text)
	if len(text) > len(prefix) && string(text[:len(prefix)]) == prefix {
		Debug.Println("Attempting command:", getcommand(prefix, text))
		Debug.Println(Commands.Get("say"))
		if command := Commands.Get(getcommand(prefix, text)); command != nil {
			// Call command as goroutine, passing a copy of the message without the command call
			go command(msg.StripPrefix(), sender)
		}
	}

	// TODO process monitors
}
