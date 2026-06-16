package dignity

import (
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// Interpretation Engine — Tests
// ═══════════════════════════════════════════════════════════════════════

func TestInterpretPlanetInSign(t *testing.T) {
	// Sun in Leo: domicile, strong
	result := InterpretPlanetInSign("Sun", "Leo")
	if !strings.Contains(result, "domicile") {
		t.Errorf("Sun in Leo should mention domicile, got: %s", result)
	}
	if len(result) < 20 {
		t.Errorf("Interpretation too short: %s", result)
	}

	// Moon in Capricorn: detriment
	result = InterpretPlanetInSign("Moon", "Capricorn")
	if !strings.Contains(result, "detriment") {
		t.Errorf("Moon in Capricorn should mention detriment, got: %s", result)
	}

	// Mars in Aries: domicile
	result = InterpretPlanetInSign("Mars", "Aries")
	if !strings.Contains(result, "domicile") {
		t.Errorf("Mars in Aries should mention domicile, got: %s", result)
	}

	// Venus in Virgo: fall
	result = InterpretPlanetInSign("Venus", "Virgo")
	if !strings.Contains(result, "fall") {
		t.Errorf("Venus in Virgo should mention fall, got: %s", result)
	}

	// Jupiter in Cancer: exaltation
	result = InterpretPlanetInSign("Jupiter", "Cancer")
	if !strings.Contains(result, "exalted") {
		t.Errorf("Jupiter in Cancer should mention exaltation, got: %s", result)
	}

	// Saturn in Aries: fall
	result = InterpretPlanetInSign("Saturn", "Aries")
	if !strings.Contains(result, "fall") {
		t.Errorf("Saturn in Aries should mention fall, got: %s", result)
	}

	// Unknown planet
	result = InterpretPlanetInSign("FakePlanet", "Leo")
	if result == "" {
		t.Error("Unknown planet should still produce text")
	}
}

func TestInterpretPlanetInHouse(t *testing.T) {
	// Sun in H10: career, visibility
	result := InterpretPlanetInHouse("Sun", 10)
	if !strings.Contains(strings.ToLower(result), "career") && !strings.Contains(strings.ToLower(result), "visible") {
		t.Errorf("Sun in H10 should mention career/visibility, got: %s", result)
	}

	// Moon in H4: home, roots
	result = InterpretPlanetInHouse("Moon", 4)
	if !strings.Contains(strings.ToLower(result), "home") && !strings.Contains(strings.ToLower(result), "root") {
		t.Errorf("Moon in H4 should mention home/roots, got: %s", result)
	}

	// Mars in H12: hidden, contained
	result = InterpretPlanetInHouse("Mars", 12)
	if !strings.Contains(strings.ToLower(result), "hidden") && !strings.Contains(strings.ToLower(result), "contain") {
		t.Errorf("Mars in H12 should mention hidden/contained, got: %s", result)
	}

	// Invalid house
	result = InterpretPlanetInHouse("Sun", 13)
	if result == "" {
		t.Error("Invalid house should still produce text")
	}
}

func TestInterpretAspect(t *testing.T) {
	// Sun conjunction Mercury
	result := InterpretAspect("Sun", "Mercury", "conjunction", 0.5)
	if !strings.Contains(strings.ToLower(result), "merge") && !strings.Contains(strings.ToLower(result), "identity") {
		t.Errorf("Sun-Mercury conjunction should mention merge/identity, got: %s", result)
	}
	if !strings.Contains(result, "0.5") {
		t.Errorf("Should include orb: %s", result)
	}

	// Mars square Saturn
	result = InterpretAspect("Mars", "Saturn", "square", 1.2)
	if !strings.Contains(strings.ToLower(result), "friction") && !strings.Contains(strings.ToLower(result), "conflict") {
		t.Errorf("Mars-Saturn square should mention friction/conflict, got: %s", result)
	}

	// Venus trine Jupiter
	result = InterpretAspect("Venus", "Jupiter", "trine", 0.3)
	if !strings.Contains(strings.ToLower(result), "flow") && !strings.Contains(strings.ToLower(result), "ease") {
		t.Errorf("Venus-Jupiter trine should mention flow/ease, got: %s", result)
	}

	// Moon opposition Pluto
	result = InterpretAspect("Moon", "Pluto", "opposition", 2.0)
	if !strings.Contains(strings.ToLower(result), "polarity") && !strings.Contains(strings.ToLower(result), "tension") {
		t.Errorf("Moon-Pluto opposition should mention polarity/tension, got: %s", result)
	}

	// Unknown aspect type
	result = InterpretAspect("Sun", "Moon", "quintile", 0.5)
	if result == "" {
		t.Error("Unknown aspect type should still produce text")
	}
}

func TestInterpretPattern(t *testing.T) {
	// T-Square
	result := InterpretPattern("T-Square", []string{"Mars", "Saturn", "Uranus"})
	if !strings.Contains(strings.ToLower(result), "tension") && !strings.Contains(strings.ToLower(result), "dynamic") {
		t.Errorf("T-Square should mention tension/dynamic, got: %s", result)
	}
	if !strings.Contains(result, "Mars") || !strings.Contains(result, "Saturn") || !strings.Contains(result, "Uranus") {
		t.Errorf("T-Square should name all planets, got: %s", result)
	}

	// Grand Trine
	result = InterpretPattern("Grand Trine", []string{"Sun", "Moon", "Jupiter"})
	if !strings.Contains(strings.ToLower(result), "flow") && !strings.Contains(strings.ToLower(result), "talent") {
		t.Errorf("Grand Trine should mention flow/talent, got: %s", result)
	}

	// Yod
	result = InterpretPattern("Yod", []string{"Venus", "Neptune", "Pluto"})
	if !strings.Contains(strings.ToLower(result), "finger") && !strings.Contains(strings.ToLower(result), "destiny") {
		t.Errorf("Yod should mention finger of god/destiny, got: %s", result)
	}

	// Unknown pattern
	result = InterpretPattern("Mystic Rectangle", []string{"Sun", "Moon"})
	if result == "" {
		t.Error("Unknown pattern should still produce text")
	}
}

func TestInterpretChart_Empty(t *testing.T) {
	// Empty chart should produce valid JSON
	report := InterpretChart("Test", nil, nil, nil, nil, nil)
	if report.Name != "Test" {
		t.Errorf("Name should be Test, got %s", report.Name)
	}
	if len(report.PlanetSigns) != 0 {
		t.Errorf("Empty chart should have 0 planet-sign entries")
	}
}

func TestInterpretChart_Full(t *testing.T) {
	planets := map[string]float64{
		"Sun":     320.0, // Aquarius
		"Moon":    95.0,  // Cancer
		"Mercury": 300.0, // Aquarius
		"Venus":   10.0,  // Aries
		"Mars":    185.0, // Libra
		"Jupiter": 120.0, // Leo
		"Saturn":  45.0,  // Taurus
	}

	houses := map[string]int{
		"Sun": 4, "Moon": 10, "Mercury": 3, "Venus": 5,
		"Mars": 12, "Jupiter": 9, "Saturn": 6,
	}

	aspects := []AspectHit{
		{Planet1: "Sun", Planet2: "Moon", Aspect: "trine", Orb: 0.5},
		{Planet1: "Mars", Planet2: "Saturn", Aspect: "opposition", Orb: 1.2},
		{Planet1: "Venus", Planet2: "Jupiter", Aspect: "trine", Orb: 0.3},
	}

	patterns := []PatternHit{
		{Name: "T-Square", Planets: []string{"Mars", "Saturn", "Uranus"}},
	}

	report := InterpretChart("Test Chart", planets, houses, aspects, patterns, nil)

	if len(report.PlanetSigns) != 7 {
		t.Errorf("Expected 7 planet-sign entries, got %d", len(report.PlanetSigns))
	}
	if len(report.PlanetHouses) != 7 {
		t.Errorf("Expected 7 planet-house entries, got %d", len(report.PlanetHouses))
	}
	if len(report.Aspects) != 3 {
		t.Errorf("Expected 3 aspect entries, got %d", len(report.Aspects))
	}
	if len(report.Patterns) != 1 {
		t.Errorf("Expected 1 pattern entry, got %d", len(report.Patterns))
	}

	// Verify JSON round-trip
	js, err := report.JSON()
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	if len(js) < 100 {
		t.Errorf("JSON output too short: %s", string(js))
	}
}
