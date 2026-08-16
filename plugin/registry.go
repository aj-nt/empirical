package plugin

import "sync"

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Interpreter)
)

// Register adds a plugin to the global registry.
func Register(p Interpreter) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.Name()] = p
}

// Get returns a plugin by name.
func Get(name string) (Interpreter, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[name]
	return p, ok
}

// List returns all registered plugins.
func List() []Interpreter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make([]Interpreter, 0, len(registry))
	for _, p := range registry {
		result = append(result, p)
	}
	return result
}
