package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LoadDir loads all .so plugins from a directory.
// Plugins must export a symbol "Plugin" of type Interpreter.
// NOTE: Go's plugin package only works on Linux with CGO_ENABLED=1.
// On other platforms, LoadDir returns 0, nil (no error — just no plugins loaded).
func LoadDir(dir string) (int, error) {
	if runtime.GOOS != "linux" {
		return 0, nil // plugin loading only supported on Linux
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No plugins dir is fine
		}
		return 0, err
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".so" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		p, err := openPlugin(path)
		if err != nil {
			return loaded, fmt.Errorf("loading %s: %w", entry.Name(), err)
		}
		Register(p)
		loaded++
	}
	return loaded, nil
}
