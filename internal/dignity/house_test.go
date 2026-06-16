package dignity

import (
	"encoding/json"
	"testing"
)

// ── Phase 3: House Division Tests ──────────────────────────────────────────

func TestHouseConvergence_AJ(t *testing.T) {
	t.Skip("requires SWE ephemeris — run with go test -count=1")
}

func TestHouseConvergence_TestChart(t *testing.T) {
	// Same test chart as Python reference: London, 2000-01-01 noon
	tropical := map[string]float64{
		"Sun": 127.0, "Moon": 307.0, "Mercury": 67.0,
		"Venus": 247.0, "Mars": 235.0, "Jupiter": 277.0, "Saturn": 7.0,
	}

	hc := ComputeHouseConvergence(tropical, 2000, 1, 1, 12, 0, 0, 0.0, 51.5, -0.1, "Test")

	if hc.Name != "Test" {
		t.Errorf("Name = %q, want Test", hc.Name)
	}
	if len(hc.Planets) != 7 {
		t.Errorf("expected 7 planets, got %d", len(hc.Planets))
	}

	// With 8 systems, 75% threshold = 6+ agreement
	// Verify each planet has placements for all 8 systems
	for _, p := range hc.Planets {
		if len(p.Placements) != 8 {
			t.Errorf("%s has %d placements, want 8", p.Planet, len(p.Placements))
		}
	}

	// Verify specific placements match expectations
	mercury := findPlanetHouse(hc, "Mercury")
	if mercury.Placements["whole_sign"] != 3 {
		t.Errorf("Mercury whole_sign = H%d, want H3", mercury.Placements["whole_sign"])
	}

	saturn := findPlanetHouse(hc, "Saturn")
	if saturn.Placements["whole_sign"] != 1 {
		t.Errorf("Saturn whole_sign = H%d, want H1", saturn.Placements["whole_sign"])
	}

	// Verify JSON round-trip
	js, err := hc.HouseConvergenceJSON()
	if err != nil {
		t.Fatalf("HouseConvergenceJSON() error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestHouseConvergence_AllPlanetsPresent(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 60.0, "Mars": 180.0, "Jupiter": 300.0, "Saturn": 90.0,
	}
	hc := ComputeHouseConvergence(tropical, 2000, 1, 1, 12, 0, 0, 0.0, 40.0, -75.0, "Test")

	names := make(map[string]bool)
	for _, p := range hc.Planets {
		names[p.Planet] = true
	}
	expected := map[string]bool{
		"Sun": true, "Moon": true, "Mercury": true,
		"Venus": true, "Mars": true, "Jupiter": true, "Saturn": true,
	}
	for k := range expected {
		if !names[k] {
			t.Errorf("missing planet %q", k)
		}
	}
	if len(names) != 7 {
		t.Errorf("got %d planets, want 7", len(names))
	}
}

func TestHouseConvergence_NonClassicalSkipped(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 60.0, "Mars": 180.0, "Jupiter": 300.0,
		"Saturn": 90.0, "Chiron": 45.0,
	}
	hc := ComputeHouseConvergence(tropical, 2000, 1, 1, 12, 0, 0, 0.0, 40.0, -75.0, "Test")
	if len(hc.Planets) != 7 {
		t.Errorf("expected 7 classical planets, got %d (Chiron should be skipped)", len(hc.Planets))
	}
}

func TestHouseConvergence_Format(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 127.0, "Moon": 307.0, "Mercury": 67.0,
		"Venus": 247.0, "Mars": 235.0, "Jupiter": 277.0, "Saturn": 7.0,
	}
	hc := ComputeHouseConvergence(tropical, 2000, 1, 1, 12, 0, 0, 0.0, 51.5, -0.1, "Test")
	out := FormatHouseConvergence(hc)
	if out == "" {
		t.Error("expected non-empty output")
	}
	if !contains(out, "Whole-sign vs quadrant") {
		t.Errorf("missing expected content in output")
	}
	if !contains(out, "RECOVERY IMPLICATION") {
		t.Errorf("missing recovery implication section")
	}
}

func TestHouseConvergence_JSON(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 60.0, "Mars": 180.0, "Jupiter": 300.0, "Saturn": 90.0,
	}
	hc := ComputeHouseConvergence(tropical, 2000, 1, 1, 12, 0, 0, 0.0, 40.0, -75.0, "Test")
	js, err := hc.HouseConvergenceJSON()
	if err != nil {
		t.Fatalf("HouseConvergenceJSON() error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["name"] != "Test" {
		t.Errorf("name = %v, want Test", parsed["name"])
	}
	if _, ok := parsed["convergence_rate"]; !ok {
		t.Error("missing convergence_rate")
	}
}

func TestPlanetHouse_Properties(t *testing.T) {
	ph := &PlanetHouse{
		Planet: "Mars",
		Placements: map[string]int{
			"whole_sign": 1,
			"placidus":   1,
			"porphyry":   1,
			"koch":       1,
			"equal":      2,
			"regiomontanus": 1,
			"alcabitius":    1,
			"campanus":      1,
		},
	}
	if ph.ConsensusHouse() != 1 {
		t.Errorf("ConsensusHouse() = %d, want 1", ph.ConsensusHouse())
	}
	// 7 of 8 agree on house 1
	if ph.AgreementCount() != 7 {
		t.Errorf("AgreementCount() = %d, want 7", ph.AgreementCount())
	}
	// 75% threshold = 6+, so 7/8 is unambiguous
	if !ph.IsUnambiguous() {
		t.Error("IsUnambiguous() should be true (7/8)")
	}
	if ph.IsDisputed() {
		t.Error("IsDisputed() should be false (7/8)")
	}
	if ph.AgreementRatio() != 0.875 {
		t.Errorf("AgreementRatio() = %f, want 0.875", ph.AgreementRatio())
	}
}

// helpers

func findPlanetHouse(hc *HouseConvergence, name string) *PlanetHouse {
	for i := range hc.Planets {
		if hc.Planets[i].Planet == name {
			return &hc.Planets[i]
		}
	}
	panic("planet not found: " + name)
}
