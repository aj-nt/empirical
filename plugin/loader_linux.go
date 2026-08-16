//go:build linux && cgo

package plugin

import (
	"fmt"
	"plugin"
)

func openPlugin(path string) (Interpreter, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("symbol 'Plugin' not found: %w", err)
	}
	interp, ok := sym.(Interpreter)
	if !ok {
		return nil, fmt.Errorf("symbol 'Plugin' is not an Interpreter")
	}
	return interp, nil
}
