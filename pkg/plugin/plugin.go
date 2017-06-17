package plugin

import (
	"sync"

	"github.com/hellofresh/janus/pkg/api"
	"github.com/hellofresh/janus/pkg/router"
)

// Plugin defines basic methods for plugins
type Plugin interface {
	GetName() string
	GetMiddlewares(rawConfig map[string]interface{}, referenceSpec *api.Spec) ([]router.Constructor, error)
}

// Loader holds all availables plugins
type Loader struct {
	sync.RWMutex
	plugins map[string]Plugin
}

// NewLoader creates a new instance of Loader
func NewLoader() *Loader {
	return &Loader{
		plugins: make(map[string]Plugin),
	}
}

// Add a new plugin to the loader
func (l *Loader) Add(plugins ...Plugin) {
	l.Lock()
	defer l.Unlock()

	for _, p := range plugins {
		l.plugins[p.GetName()] = p
	}
}

// Get a plugin by name
func (l *Loader) Get(name string) Plugin {
	l.RLock()
	defer l.RUnlock()

	return l.plugins[name]
}
