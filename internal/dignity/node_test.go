package dignity

import (
	"encoding/json"
	"testing"
)

// ── Phase 5: Node Axis Convergence ─────────────────────────────────────────

func TestNodeConvergence_AJ(t *testing.T) {
	// AJ's tropical North Node longitude (from SWE) and ayanamsa
	// NN ~ 5.0 deg Aries, ayanamsa ~ 23.4259 deg Lahiri
	nnLong := 5.0   // approximate tropical NN longitude for AJ
	ayan := 23.425866590270687

	n := ComputeNodeConvergence(nnLong, ayan, "AJ")

	if !n.AxisPreserved() {
		t.Error("AxisPreserved should be true")
	}
	if n.AxisAngle < 179.9 || n.AxisAngle > 180.1 {
		t.Errorf("AxisAngle should be ~180, got %f", n.AxisAngle)
	}
	if n.NNTropSign != "Aries" {
		t.Errorf("NNTropSign = %q, want Aries", n.NNTropSign)
	}
	if n.NNSidSign != "Pisces" {
		t.Errorf("NNSidSign = %q, want Pisces (ayan %f shifts Aries back)", n.NNSidSign, ayan)
	}
	if n.SignMatch {
		t.Error("SignMatch should be false — tropical and sidereal disagree on NN sign")
	}
	if !n.MeaningConverges() {
		t.Error("MeaningConverges should always be true")
	}

	// Verify South Node is opposite
	if n.SNTropSign != "Libra" {
		t.Errorf("SNTropSign = %q, want Libra (opposite Aries)", n.SNTropSign)
	}
	if n.SNSidSign != "Virgo" {
		t.Errorf("SNSidSign = %q, want Virgo (opposite Pisces)", n.SNSidSign)
	}
}

func TestNodeConvergence_AxisAlways180(t *testing.T) {
	// Test with several charts — axis should always be ~180
	tests := []struct {
		nnLong, ayan float64
	}{
		{15.0, 24.0},
		{45.0, 24.0},
		{120.0, 23.0},
		{330.0, 25.0},
		{200.0, 24.0},
	}
	for _, tt := range tests {
		n := ComputeNodeConvergence(tt.nnLong, tt.ayan, "test")
		if !n.AxisPreserved() {
			t.Errorf("axis not preserved for nn=%f ayan=%f: angle=%f", tt.nnLong, tt.ayan, n.AxisAngle)
		}
	}
}

func TestNodeConvergence_Format(t *testing.T) {
	n := &NodeConvergence{
		Name:       "Test",
		NNTropSign: "Aries",
		NNSidSign:  "Pisces",
		SNTropSign: "Libra",
		SNSidSign:  "Virgo",
		SignMatch:  false,
		AxisAngle:  180.0,
	}
	out := FormatNodeConvergence(n)
	if out == "" {
		t.Error("expected non-empty output")
	}
	if !contains(out, "AXIS PRESERVED") || !contains(out, "MEANING CONVERGENCE") {
		t.Errorf("missing expected content in output:\n%s", out)
	}
}

func TestNodeConvergence_JSON(t *testing.T) {
	n := ComputeNodeConvergence(15.0, 24.0, "Test")
	js, err := n.NodeConvergenceJSON()
	if err != nil {
		t.Fatalf("NodeConvergenceJSON() error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["name"] != "Test" {
		t.Errorf("name = %v, want Test", parsed["name"])
	}
	if parsed["axis_preserved"] != true {
		t.Error("axis_preserved should be true")
	}
	if parsed["meaning_converges"] != true {
		t.Error("meaning_converges should be true")
	}
}

// ── Phase 6: Zodiac Comparison ─────────────────────────────────────────────

func TestZodiacComparison_AJ(t *testing.T) {
	// AJ's tropical longitudes (approximate from previous cross-validation)
	tropical := map[string]float64{
		"Sun":     327.0, // Aquarius ~27 deg
		"Moon":    315.0, // Aquarius ~15 deg
		"Mercury": 310.0, // Aquarius ~10 deg
		"Venus":   12.0,  // Aries
		"Mars":    235.0, // Scorpio
		"Jupiter": 187.0, // Libra ~7 deg
		"Saturn":  7.0,   // Aries
	}
	ayan := 23.425866590270687

	zc := ComputeZodiacComparison(tropical, ayan, "AJ")

	if zc.Winner() != "tie" {
		t.Errorf("Winner() = %q, want tie", zc.Winner())
	}
	if zc.Tropical.DignityDensity() != zc.Sidereal.DignityDensity() {
		t.Errorf("Density should be equal: trop=%f sid=%f",
			zc.Tropical.DignityDensity(), zc.Sidereal.DignityDensity())
	}
	if zc.Tropical.TotalPlanets != 7 {
		t.Errorf("TotalPlanets = %d, want 7", zc.Tropical.TotalPlanets)
	}

	// Verify Mars is domicile in both zodiacs (Scorpio in both)
	if zc.Tropical.Placements["Mars"] != "domicile" {
		t.Errorf("Mars trop = %q, want domicile", zc.Tropical.Placements["Mars"])
	}
	if zc.Sidereal.Placements["Mars"] != "domicile" {
		t.Errorf("Mars sid = %q, want domicile", zc.Sidereal.Placements["Mars"])
	}
}

func TestZodiacComparison_DensityBounds(t *testing.T) {
	// Dignity density should be between 0 and 1
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 15.0, "Mars": 130.0, "Jupiter": 250.0, "Saturn": 10.0,
	}
	zc := ComputeZodiacComparison(tropical, 24.0, "test")
	dT := zc.Tropical.DignityDensity()
	dS := zc.Sidereal.DignityDensity()
	if dT < 0 || dT > 1 {
		t.Errorf("Tropical density out of range: %f", dT)
	}
	if dS < 0 || dS > 1 {
		t.Errorf("Sidereal density out of range: %f", dS)
	}
}

func TestZodiacComparison_Format(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 15.0, "Mars": 130.0, "Jupiter": 250.0, "Saturn": 10.0,
	}
	zc := ComputeZodiacComparison(tropical, 24.0, "test")
	out := FormatZodiacComparison(zc)
	if out == "" {
		t.Error("expected non-empty output")
	}
	winner := zc.Winner()
	if winner != "tie" && winner != "tropical" && winner != "sidereal" {
		t.Errorf("unexpected winner: %q", winner)
	}
	if !contains(out, "TIE") && !contains(out, "TROPICAL") && !contains(out, "SIDEREAL") {
		t.Errorf("missing winner in output:\n%s", out)
	}
}

func TestZodiacComparison_JSON(t *testing.T) {
	tropical := map[string]float64{
		"Sun": 0.0, "Moon": 120.0, "Mercury": 240.0,
		"Venus": 15.0, "Mars": 130.0, "Jupiter": 250.0, "Saturn": 10.0,
	}
	zc := ComputeZodiacComparison(tropical, 24.0, "test")
	js, err := zc.ZodiacComparisonJSON()
	if err != nil {
		t.Fatalf("ZodiacComparisonJSON() error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(js, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Errorf("name = %v, want test", parsed["name"])
	}
}

// helpers

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
