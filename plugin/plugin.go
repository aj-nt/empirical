// Package plugin defines the interface and registry for interpretation plugins.
// Plugins are compiled as .so files and loaded at runtime via the --plugin-dir flag.
package plugin

// Interpreter is the interface for interpretation plugins.
// Each plugin provides interpretations for a specific domain
// (e.g., planet-in-sign, aspect, house placement).
type Interpreter interface {
	// Name returns the plugin's unique identifier.
	Name() string
	// Version returns the plugin's semantic version.
	Version() string
	// Description returns a human-readable description.
	Description() string
	// PlanetInSign returns an interpretation for a planet in a sign.
	// Returns empty string if this plugin doesn't cover this combination.
	PlanetInSign(planet, sign string) string
	// AspectInterpretation returns an interpretation for an aspect between two planets.
	AspectInterpretation(p1, p2, aspectType string, orb float64) string
	// HousePlacement returns an interpretation for a planet in a house.
	HousePlacement(planet string, house int) string
}
