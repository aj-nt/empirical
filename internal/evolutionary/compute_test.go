package evolutionary

import (
	"strings"
	"testing"
)

func TestComputeEvolutionary_AssemblesAllIndicators(t *testing.T) {
	t.Parallel()

	// AJ's chart (1969-02-15, 23:10, -8, 47.038, -122.901) — verified against empirical engine
	planets := map[string]float64{
		"Sun":     327.45, // Aquarius
		"Moon":    322.29, // Aquarius
		"Mercury": 302.10, // Aquarius
		"Venus":   12.59,  // Aries
		"Mars":    235.80, // Scorpio
		"Jupiter": 184.94, // Libra
		"Saturn":  21.48,  // Aries
		"Uranus":  183.34, // Libra
		"Neptune": 238.67, // Scorpio
		"Pluto":   174.46, // Virgo
	}
	houses := map[string]int{
		"Sun": 4, "Moon": 4, "Mercury": 4, "Venus": 6,
		"Mars": 1, "Jupiter": 12, "Saturn": 6,
		"Uranus": 12, "Neptune": 1, "Pluto": 11,
	}
	nn := 2.0 // North Node at 2° Aries

	result := ComputeEvolutionary("Test", planets, houses, nn, 6, 12, 5, 5.0)

	// Name
	if result.Name != "Test" {
		t.Errorf("expected Name 'Test', got %q", result.Name)
	}

	// Pluto
	if result.Pluto.Sign != "Virgo" {
		t.Errorf("expected Pluto in Virgo, got %q", result.Pluto.Sign)
	}
	if result.Pluto.House != 11 {
		t.Errorf("expected Pluto in 11th house, got %d", result.Pluto.House)
	}

	// North Node
	if result.NorthNode.Sign != "Aries" {
		t.Errorf("expected NN in Aries, got %q", result.NorthNode.Sign)
	}
	if result.NorthNode.House != 6 {
		t.Errorf("expected NN in 6th house, got %d", result.NorthNode.House)
	}

	// South Node (opposite NN)
	if result.SouthNode.Sign != "Libra" {
		t.Errorf("expected SN in Libra, got %q", result.SouthNode.Sign)
	}
	if result.SouthNode.House != 12 {
		t.Errorf("expected SN in 12th house, got %d", result.SouthNode.House)
	}

	// Saturn
	if result.Saturn.Sign != "Aries" {
		t.Errorf("expected Saturn in Aries, got %q", result.Saturn.Sign)
	}
	if result.Saturn.House != 6 {
		t.Errorf("expected Saturn in 6th house, got %d", result.Saturn.House)
	}

	// Pluto polarity point (opposite Pluto)
	if result.PlutoPolarity.Sign != "Pisces" {
		t.Errorf("expected Pluto polarity in Pisces, got %q", result.PlutoPolarity.Sign)
	}
	if result.PlutoPolarity.House != 5 {
		t.Errorf("expected Pluto polarity in 5th house, got %d", result.PlutoPolarity.House)
	}

	// South Node ruler: SN in Libra → Venus rules Libra
	if result.SouthNodeRuler.Planet != "Venus" {
		t.Errorf("expected SN ruler Venus, got %q", result.SouthNodeRuler.Planet)
	}

	// Skipped steps: planets square or opposite the nodal axis
	// Jupiter at 185° Libra — opposite NN at 2° Aries? 185-2=183, 183-180=3° orb — yes, opposition
	// Uranus at 185° Libra — same
	foundJupiter := false
	foundUranus := false
	for _, ss := range result.SkippedSteps {
		if ss.Planet == "Jupiter" {
			foundJupiter = true
			if ss.Aspect != "opposition" {
				t.Errorf("expected Jupiter opposition to nodes, got %q", ss.Aspect)
			}
		}
		if ss.Planet == "Uranus" {
			foundUranus = true
		}
	}
	if !foundJupiter {
		t.Error("expected Jupiter as skipped step (opposition to nodes)")
	}
	if !foundUranus {
		t.Error("expected Uranus as skipped step (opposition to nodes)")
	}

	// Narrative should not be empty
	if result.Narrative == "" {
		t.Error("Narrative should not be empty")
	}
	if !strings.Contains(result.Narrative, "Pluto") {
		t.Error("Narrative should mention Pluto")
	}
	if !strings.Contains(result.Narrative, "Virgo") {
		t.Error("Narrative should mention Pluto's sign")
	}
}

func TestComputeEvolutionary_NoSkippedSteps(t *testing.T) {
	t.Parallel()

	// All planets far from the nodal axis
	planets := map[string]float64{
		"Sun": 120.0, "Moon": 150.0, "Mercury": 80.0,
		"Venus": 100.0, "Mars": 200.0, "Jupiter": 300.0,
		"Saturn": 15.0, "Uranus": 250.0, "Neptune": 260.0,
		"Pluto": 170.0,
	}
	houses := map[string]int{
		"Sun": 5, "Moon": 6, "Mercury": 5, "Venus": 4,
		"Mars": 7, "Jupiter": 10, "Saturn": 1,
		"Uranus": 9, "Neptune": 9, "Pluto": 6,
	}
	nn := 45.0 // NN at 15° Taurus — no planet within 5° of square/opposition

	result := ComputeEvolutionary("Test", planets, houses, nn, 2, 8, 8, 5.0)

	if len(result.SkippedSteps) != 0 {
		t.Errorf("expected 0 skipped steps, got %d: %+v", len(result.SkippedSteps), result.SkippedSteps)
	}
}

func TestComputeEvolutionary_SouthNodeRuler(t *testing.T) {
	t.Parallel()

	// SN in Aries → Mars rules
	planets := map[string]float64{
		"Sun": 100.0, "Moon": 200.0, "Pluto": 300.0,
	}
	houses := map[string]int{
		"Sun": 4, "Moon": 7, "Pluto": 10,
	}
	nn := 190.0 // NN at 10° Libra → SN at 10° Aries → Mars rules

	result := ComputeEvolutionary("Test", planets, houses, nn, 7, 1, 7, 5.0)

	if result.SouthNodeRuler.Planet != "Mars" {
		t.Errorf("expected SN ruler Mars (SN in Aries), got %q", result.SouthNodeRuler.Planet)
	}
}
