package onelib

import (
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

func UnloadPlugins() {
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

func UnloadProtocols() {
	Protocols.DeleteAll()
}
