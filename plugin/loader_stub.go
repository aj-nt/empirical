//go:build !linux || !cgo

package plugin

import "fmt"

func openPlugin(path string) (Interpreter, error) {
	return nil, fmt.Errorf("plugin loading not supported on this platform")
}
