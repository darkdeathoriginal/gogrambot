package handler

import (
	"sync"

	"github.com/amarnathcjd/gogram/telegram"
)

type PluginDetails struct {
	Name        string
	Description string
	Category    string
	Usage       string
	On          string
	Filter      *telegram.Filter
	AllowAll    bool
}

type Plugin struct {
	PluginDetails
	Handler func(*telegram.NewMessage) error
}

var (
	Plugins      []Plugin
	seenPlugins  = map[string]bool{}
	pluginsMutex sync.Mutex
)

// --- Builder API ---
type PluginBuilder struct {
	p Plugin
}

func NewPlugin(name string) *PluginBuilder {
	return &PluginBuilder{
		p: Plugin{
			PluginDetails: PluginDetails{Name: name},
		},
	}
}

func (b *PluginBuilder) Description(desc string) *PluginBuilder {
	b.p.Description = desc
	return b
}
func (b *PluginBuilder) Category(cat string) *PluginBuilder {
	b.p.Category = cat
	return b
}
func (b *PluginBuilder) Usage(usage string) *PluginBuilder {
	b.p.Usage = usage
	return b
}
func (b *PluginBuilder) Handle(fn func(*telegram.NewMessage) error) {
	b.p.Handler = fn
	RegisterPlugin(b.p)
}

func (b *PluginBuilder) On(event string) *PluginBuilder {
	b.p.On = event
	return b
}
func (b *PluginBuilder) Filter(filter telegram.Filter) *PluginBuilder {
	b.p.Filter = &filter
	return b
}
func (b *PluginBuilder) AllowAll(allow bool) *PluginBuilder {
	b.p.AllowAll = allow
	return b
}

// --- Core registry ---
func RegisterPlugin(p Plugin) {
	pluginsMutex.Lock()
	defer pluginsMutex.Unlock()

	if seenPlugins[p.Name] {
		panic("handler: duplicate plugin registered: " + p.Name)
	}
	seenPlugins[p.Name] = true
	Plugins = append(Plugins, p)
}
