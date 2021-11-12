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

	commands, mon := plug.Implements()
	for trigger, command := range commands {
		Commands.Put(trigger, command)
	}
	// TODO unload
	if mon != nil {
		Monitors.Put(mon)
	}

	Info.Printf("Loaded '%s' version %s.\n", plug.LongName(), plug.Version())
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

// UnloadPlugin removes a plugin from the active plugins map, returning an error if not loaded, calling the related
// Delete methods.
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

// UnloadPlugins unloads every plugin, calling their unload routines.
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
	Info.Printf("Loaded '%s' version %s.\n", proto.LongName(), proto.Version())
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

// UnloadProtocols unloads every protocol, calling their unload routines.
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
	if len(text) > len(prefix) && string(text[:len(prefix)]) == prefix {
		commandName := getcommand(prefix, text)
		if command := Commands.Get(commandName); command != nil {
			// Call command as goroutine, passing a copy of the message without the command call
			go command(msg.StripPrefix(prefix+commandName), sender)

			return // TODO once command outputs are bridged, this line needs to be removed so the bridge can still bridge the call itself
		}
	}

	mons := Monitors.Get()
	for _, mon := range mons {
		if mon.OnMessage != nil {
			go mon.OnMessage(sender, msg)
		}
		if len(text) > 0 && mon.OnMessageWithText != nil {
			go mon.OnMessageWithText(sender, msg)
		}
	}

}

// ProcessUpdate processes monitor trigger "mon.OnMessageUpdate"
func ProcessUpdate(msg Message, sender Sender) {
	mons := Monitors.Get()
	for _, mon := range mons {
		if mon.OnMessageUpdate != nil {
			mon.OnMessageUpdate(sender, msg)
		}
	}
}
