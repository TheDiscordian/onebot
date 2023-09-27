// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package missioncontrol

import (
	"html/template"
	"sync"
)

var Plugins *plugins = &plugins {
		plugins: make(map[string]Plugin),
		lock: &sync.RWMutex{},
}

type plugins struct {
	plugins map[string]Plugin
	lock *sync.RWMutex
}

func (p *plugins) Set(name string, plugin Plugin) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.plugins[name] = plugin
}

func (p *plugins) Get(name string) Plugin {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.plugins[name]
}

func (p *plugins) Del(name string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.plugins, name)
}

func (p *plugins) List() []string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	var plugins []string
	for name := range p.plugins {
		plugins = append(plugins, name)
	}
	return plugins
}

type Plugin interface {
	HTML() template.HTML
	Functions() map[string]func(map[string]any) (string, error)
}
