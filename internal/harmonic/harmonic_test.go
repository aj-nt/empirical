package harmonic

import (
	"math"
	"testing"
)

// ── HarmonicLongitude ────────────────────────────────────────────────────

func TestHarmonicLongitude(t *testing.T) {
	tests := []struct {
		lon     float64
		harmonic int
		want    float64
	}{
		{0, 1, 0},
		{0, 5, 0},
		{72, 5, 0},     // quintile → conjunction in H5
		{144, 5, 0},    // biquintile → conjunction in H5
		{90, 4, 0},     // square → conjunction in H4
		{180, 4, 0},    // opposition → conjunction in H4
		{120, 3, 0},    // trine → conjunction in H3
		{60, 6, 0},     // sextile → conjunction in H6
		{51.428, 7, 0}, // septile → conjunction in H7
		{40, 9, 0},     // novile → conjunction in H9
		{10, 5, 50},    // 10*5 = 50
		{100, 5, 140},  // 100*5 = 500 → 140
		{300, 5, 60},   // 300*5 = 1500 → 60
	}
	for _, tc := range tests {
		got := HarmonicLongitude(tc.lon, tc.harmonic)
		diff := math.Abs(math.Mod(got-tc.want+540, 360) - 180)
		if diff > 0.1 {
			t.Errorf("HarmonicLongitude(%.3f, %d) = %.4f, want %.1f (angular diff=%.4f)", tc.lon, tc.harmonic, got, tc.want, diff)
		}
	}
}

// ── HarmonicChart ────────────────────────────────────────────────────────

func TestHarmonicChart(t *testing.T) {
	planets := map[string]float64{
		"Sun": 0, "Moon": 72, "Mars": 144,
	}
	h5 := HarmonicChart(planets, 5)
	// All should be at 0 in H5 (quintile/biquintile → conjunction)
	for name, lon := range h5 {
		if math.Abs(lon) > 0.1 {
			t.Errorf("%s in H5 = %.4f, want ~0", name, lon)
		}
	}
}

// ── FindHarmonicConjunctions ─────────────────────────────────────────────

func TestFindHarmonicConjunctions(t *testing.T) {
	// Sun at 0, Moon at 72 (quintile), Mars at 144 (biquintile)
	// In H5, all three should be conjunct at 0
	planets := map[string]float64{
		"Sun": 0, "Moon": 72, "Mars": 144,
	}
	conj := FindHarmonicConjunctions(planets, 5, 2.0)
	if len(conj) < 2 {
		t.Fatalf("expected at least 2 conjunctions in H5, got %d", len(conj))
	}
	// All orbs should be near 0
	for _, c := range conj {
		if c.Orb > 0.1 {
			t.Errorf("H5 conjunction %s/%s orb = %.4f, want ~0", c.PlanetA, c.PlanetB, c.Orb)
		}
	}
}

func TestFindHarmonicConjunctionsNoMatch(t *testing.T) {
	// Random positions that don't form quintiles
	planets := map[string]float64{
		"Sun": 0, "Moon": 50, "Mars": 100,
	}
	conj := FindHarmonicConjunctions(planets, 5, 1.0)
	if len(conj) != 0 {
		t.Errorf("expected 0 conjunctions, got %d", len(conj))
	}
}

func TestFindHarmonicConjunctionsH4(t *testing.T) {
	// Sun at 0, Moon at 90 (square), Mars at 180 (opposition)
	// In H4, all three should be conjunct at 0
	planets := map[string]float64{
		"Sun": 0, "Moon": 90, "Mars": 180,
	}
	conj := FindHarmonicConjunctions(planets, 4, 2.0)
	if len(conj) < 2 {
		t.Fatalf("expected at least 2 conjunctions in H4, got %d", len(conj))
	}
}

// ── HarmonicAspectName ───────────────────────────────────────────────────

func TestHarmonicAspectName(t *testing.T) {
	tests := []struct {
		h    int
		want string
	}{
		{1, "conjunction"},
		{2, "opposition"},
		{3, "trine"},
		{4, "square"},
		{5, "quintile"},
		{6, "sextile"},
		{7, "septile"},
		{8, "semisquare"},
		{9, "novile"},
		{10, "decile"},
		{11, "undecile"},
		{12, "semisextile"},
		{13, "unknown"},
	}
	for _, tc := range tests {
		got := HarmonicAspectName(tc.h)
		if got != tc.want {
			t.Errorf("HarmonicAspectName(%d) = %q, want %q", tc.h, got, tc.want)
		}
	}
}

// ── ComputeHarmonicReport ────────────────────────────────────────────────

func TestComputeHarmonicReport(t *testing.T) {
	planets := map[string]float64{
		"Sun": 0, "Moon": 72, "Mars": 144,
	}
	report := ComputeHarmonicReport("test", planets, []int{4, 5, 7, 9}, 2.0)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if len(report.Harmonics) != 4 {
		t.Errorf("expected 4 harmonics, got %d", len(report.Harmonics))
	}
	// H5 should have conjunctions (quintile pattern)
	for _, h := range report.Harmonics {
		if h.Harmonic == 5 && len(h.Conjunctions) == 0 {
			t.Error("H5 should have conjunctions for quintile pattern")
		}
		if h.Conjunctions == nil {
			t.Errorf("H%d Conjunctions should be non-nil (empty slice, not nil)", h.Harmonic)
		}
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────

func TestHarmonicEmptyInput(t *testing.T) {
	report := ComputeHarmonicReport("empty", map[string]float64{}, []int{5}, 2.0)
	if len(report.Harmonics) != 1 {
		t.Errorf("expected 1 harmonic, got %d", len(report.Harmonics))
	}
	if len(report.Harmonics[0].Conjunctions) != 0 {
		t.Errorf("empty planets should produce 0 conjunctions")
	}

	conj := FindHarmonicConjunctions(map[string]float64{}, 5, 2.0)
	if len(conj) != 0 {
		t.Errorf("empty planets should produce 0 conjunctions")
	}
}

func TestHarmonicSinglePlanet(t *testing.T) {
	planets := map[string]float64{"Sun": 0}
	conj := FindHarmonicConjunctions(planets, 5, 2.0)
	if len(conj) != 0 {
		t.Errorf("single planet should produce 0 conjunctions")
	}
}
