package plugin

import (
	"testing"
)

// mockPlugin implements Interpreter for testing.
type mockPlugin struct {
	name    string
	version string
	desc    string
}

func (m *mockPlugin) Name() string                                { return m.name }
func (m *mockPlugin) Version() string                             { return m.version }
func (m *mockPlugin) Description() string                         { return m.desc }
func (m *mockPlugin) PlanetInSign(planet, sign string) string     { return "" }
func (m *mockPlugin) AspectInterpretation(p1, p2, aspectType string, orb float64) string { return "" }
func (m *mockPlugin) HousePlacement(planet string, house int) string { return "" }

func TestRegisterAndGet(t *testing.T) {
	t.Parallel()

	p := &mockPlugin{name: "test-plugin", version: "1.0.0", desc: "test"}
	Register(p)

	got, ok := Get("test-plugin")
	if !ok {
		t.Fatal("expected plugin to be registered")
	}
	if got.Name() != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %q", got.Name())
	}
	if got.Version() != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", got.Version())
	}
}

func TestGetMissing(t *testing.T) {
	t.Parallel()

	_, ok := Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for missing plugin")
	}
}

func TestList(t *testing.T) {
	// TestList uses its own registry to avoid interference with parallel tests.
	// It does NOT call t.Parallel() because it mutates the global registry.
	registryMu.Lock()
	registry = make(map[string]Interpreter)
	registryMu.Unlock()

	p1 := &mockPlugin{name: "a", version: "1.0.0"}
	p2 := &mockPlugin{name: "b", version: "2.0.0"}
	Register(p1)
	Register(p2)

	list := List()
	if len(list) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(list))
	}
}

func TestLoadDirNoDir(t *testing.T) {
	t.Parallel()

	n, err := LoadDir("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Errorf("expected no error for missing dir, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 plugins loaded, got %d", n)
	}
}
