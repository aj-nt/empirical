package dignity

import (
	"encoding/json"
	"testing"
)

// ── Phase 1: Sign For Longitude ──────────────────────────────────────────

func TestSignForLongitude(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0.0, "Aries"},
		{29.9, "Aries"},
		{30.0, "Taurus"},
		{359.9, "Pisces"},
		{360.0, "Aries"},
		{720.0, "Aries"},
		{-30.0, "Pisces"}, // Negative longitude wraps correctly
	}
	for _, tt := range tests {
		got := SignForLongitude(tt.lon)
		if got != tt.want {
			t.Errorf("SignForLongitude(%f) = %q, want %q", tt.lon, got, tt.want)
		}
	}
}

// ── Phase 1: Western Dignity ─────────────────────────────────────────────

func TestWesternDignity(t *testing.T) {
	// Mars: test all five dignity categories
	t.Run("Mars domicile", func(t *testing.T) {
		if got := WesternDignity("Mars", "Aries"); got != "domicile" {
			t.Errorf("Mars in Aries = domicile, got %q", got)
		}
		if got := WesternDignity("Mars", "Scorpio"); got != "domicile" {
			t.Errorf("Mars in Scorpio = domicile, got %q", got)
		}
	})
	t.Run("Mars exaltation", func(t *testing.T) {
		if got := WesternDignity("Mars", "Capricorn"); got != "exaltation" {
			t.Errorf("Mars in Capricorn = exaltation, got %q", got)
		}
	})
	t.Run("Mars fall", func(t *testing.T) {
		if got := WesternDignity("Mars", "Cancer"); got != "fall" {
			t.Errorf("Mars in Cancer = fall, got %q", got)
		}
	})
	t.Run("Mars detriment", func(t *testing.T) {
		if got := WesternDignity("Mars", "Libra"); got != "detriment" {
			t.Errorf("Mars in Libra = detriment, got %q", got)
		}
		if got := WesternDignity("Mars", "Taurus"); got != "detriment" {
			t.Errorf("Mars in Taurus = detriment, got %q", got)
		}
	})
	t.Run("Mars peregrine", func(t *testing.T) {
		if got := WesternDignity("Mars", "Leo"); got != "peregrine" {
			t.Errorf("Mars in Leo = peregrine, got %q", got)
		}
		if got := WesternDignity("Mars", "Sagittarius"); got != "peregrine" {
			t.Errorf("Mars in Sagittarius = peregrine, got %q", got)
		}
	})
	t.Run("unknown planet", func(t *testing.T) {
		if got := WesternDignity("Chiron", "Leo"); got != "peregrine" {
			t.Errorf("Chiron in Leo = peregrine, got %q", got)
		}
	})
	t.Run("all classical planets have rules", func(t *testing.T) {
		for _, planet := range ClassicalPlanets {
			// Just verify they return something, not panic
			_ = WesternDignity(planet, "Leo")
		}
	})
}

// ── Phase 1: Vedic Dignity ───────────────────────────────────────────────

func TestVedicDignity(t *testing.T) {
	t.Run("swakshetra", func(t *testing.T) {
		if got := VedicDignity("Mars", "Aries"); got != "swakshetra" {
			t.Errorf("Mars in Aries = swakshetra, got %q", got)
		}
		if got := VedicDignity("Sun", "Leo"); got != "swakshetra" {
			t.Errorf("Sun in Leo = swakshetra, got %q", got)
		}
	})
	t.Run("uchcha", func(t *testing.T) {
		if got := VedicDignity("Sun", "Aries"); got != "uchcha" {
			t.Errorf("Sun in Aries = uchcha, got %q", got)
		}
		if got := VedicDignity("Mars", "Capricorn"); got != "uchcha" {
			t.Errorf("Mars in Capricorn = uchcha, got %q", got)
		}
	})
	t.Run("neecha", func(t *testing.T) {
		if got := VedicDignity("Sun", "Libra"); got != "neecha" {
			t.Errorf("Sun in Libra = neecha, got %q", got)
		}
		if got := VedicDignity("Mars", "Cancer"); got != "neecha" {
			t.Errorf("Mars in Cancer = neecha, got %q", got)
		}
	})
	t.Run("peregrine", func(t *testing.T) {
		if got := VedicDignity("Mars", "Leo"); got != "peregrine" {
			t.Errorf("Mars in Leo = peregrine, got %q", got)
		}
		if got := VedicDignity("Sun", "Taurus"); got != "peregrine" {
			t.Errorf("Sun in Taurus = peregrine, got %q", got)
		}
	})
	t.Run("no detriment category", func(t *testing.T) {
		// Vedic has no detriment — Venus in Scorpio is peregrine, not debilitated
		if got := VedicDignity("Venus", "Scorpio"); got != "peregrine" {
			t.Errorf("Venus in Scorpio = peregrine (no detriment in Vedic), got %q", got)
		}
	})
}

// ── Phase 1: Classify Convergence ────────────────────────────────────────

func TestClassifyConvergence(t *testing.T) {
	tests := []struct {
		western, vedic, want string
	}{
		// agree cases
		{"domicile", "swakshetra", "agree"},
		{"exaltation", "uchcha", "agree"},
		{"fall", "neecha", "agree"},
		{"peregrine", "peregrine", "agree"},
		// western_only (detriment has no Vedic parallel)
		{"detriment", "peregrine", "western_only"},
		// vedic_only
		{"peregrine", "swakshetra", "vedic_only"},
		{"peregrine", "neecha", "vedic_only"},
		// diverge
		{"domicile", "neecha", "diverge"},
		{"fall", "uchcha", "diverge"},
		{"detriment", "swakshetra", "diverge"},
	}
	for _, tt := range tests {
		got := ClassifyConvergence(tt.western, tt.vedic)
		if got != tt.want {
			t.Errorf("ClassifyConvergence(%q, %q) = %q, want %q",
				tt.western, tt.vedic, got, tt.want)
		}
	}
}

// ── Phase 1: Compute Dignity Convergence ─────────────────────────────────

func TestComputeDignityConvergence(t *testing.T) {
	t.Run("structural", func(t *testing.T) {
		tropical := map[string]float64{
			"Sun":     127.0,
			"Moon":    307.0,
			"Mercury": 67.0,
			"Venus":   247.0,
			"Mars":    235.0,
			"Jupiter": 277.0,
			"Saturn":  7.0,
		}
		report := ComputeDignityConvergence(tropical, 24.0, "test")
		if len(report.Planets) != 7 {
			t.Errorf("expected 7 planets, got %d", len(report.Planets))
		}
		if report.AyanamsaDegrees != 24.0 {
			t.Errorf("ayanamsa = %f, want 24.0", report.AyanamsaDegrees)
		}
	})

	t.Run("mars converges in Scorpio", func(t *testing.T) {
		tropical := map[string]float64{
			"Sun":     127.0,
			"Moon":    307.0,
			"Mercury": 67.0,
			"Venus":   247.0,
			"Mars":    235.0, // Scorpio (domicile) → sidereal Scorpio (swakshetra) → agree
			"Jupiter": 277.0,
			"Saturn":  7.0,
		}
		report := ComputeDignityConvergence(tropical, 24.0, "test")
		mars := findPlanet(report, "Mars")
		if mars.Convergence != "agree" {
			t.Errorf("Mars convergence = %q, want agree", mars.Convergence)
		}
		if !mars.IsSignal() {
			t.Error("Mars should be signal")
		}
		if mars.Western != "domicile" {
			t.Errorf("Mars western = %q, want domicile", mars.Western)
		}
		if mars.Vedic != "swakshetra" {
			t.Errorf("Mars vedic = %q, want swakshetra", mars.Vedic)
		}
	})

	t.Run("convergence rate computed", func(t *testing.T) {
		tropical := map[string]float64{
			"Sun":     127.0,
			"Moon":    307.0,
			"Mercury": 67.0,
			"Venus":   247.0,
			"Mars":    235.0,
			"Jupiter": 277.0,
			"Saturn":  7.0,
		}
		report := ComputeDignityConvergence(tropical, 24.0, "test")
		if report.ConvergenceRate() <= 0 || report.ConvergenceRate() > 1.0 {
			t.Errorf("convergence rate should be between 0 and 1, got %f", report.ConvergenceRate())
		}
	})
}

// ── Phase 1: JSON Output ─────────────────────────────────────────────────

func TestDignityConvergenceJSON(t *testing.T) {
	tropical := map[string]float64{
		"Sun":     127.0,
		"Moon":    307.0,
		"Mercury": 67.0,
		"Venus":   247.0,
		"Mars":    235.0,
		"Jupiter": 277.0,
		"Saturn":  7.0,
	}
	report := ComputeDignityConvergence(tropical, 24.0, "test")
	js, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(js), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["name"] != "test" {
		t.Errorf("name = %v, want test", parsed["name"])
	}
	if parsed["signal_count"].(float64) == 0 && parsed["noise_count"].(float64) == 0 {
		t.Error("signal/noise counts should not both be zero")
	}
}

// ── Phase 2: Branch Angle ────────────────────────────────────────────────

func TestBranchAngle(t *testing.T) {
	t.Run("opposition", func(t *testing.T) {
		if got := BranchAngle("Zi", "Wu"); got != 180 {
			t.Errorf("Zi-Wu = %d, want 180", got)
		}
	})
	t.Run("trine", func(t *testing.T) {
		if got := BranchAngle("Shen", "Zi"); got != 120 {
			t.Errorf("Shen-Zi = %d, want 120", got)
		}
		if got := BranchAngle("Si", "You"); got != 120 {
			t.Errorf("Si-You = %d, want 120", got)
		}
	})
	t.Run("adjacent", func(t *testing.T) {
		if got := BranchAngle("Zi", "Chou"); got != 30 {
			t.Errorf("Zi-Chou = %d, want 30", got)
		}
	})
	t.Run("symmetric", func(t *testing.T) {
		a := BranchAngle("Zi", "Wu")
		b := BranchAngle("Wu", "Zi")
		if a != b {
			t.Errorf("angle not symmetric: %d vs %d", a, b)
		}
	})
	t.Run("all six oppositions", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			if got := BranchAngle(EarthlyBranches[i], EarthlyBranches[i+6]); got != 180 {
				t.Errorf("%s-%s = %d, want 180", EarthlyBranches[i], EarthlyBranches[i+6], got)
			}
		}
	})
}

// ── Phase 2: Aspect Catalog ──────────────────────────────────────────────

func TestAspectCatalog(t *testing.T) {
	catalog := AspectCatalog()

	if len(catalog) != 7 {
		t.Errorf("expected 7 angles, got %d", len(catalog))
	}

	var universal, partial int
	for _, a := range catalog {
		switch a.Universality {
		case "universal":
			universal++
		case "partial":
			partial++
		}
	}

	if universal != 3 {
		t.Errorf("expected 3 universal angles (0, 120, 180), got %d", universal)
	}
	if partial != 4 {
		t.Errorf("expected 4 partial angles, got %d", partial)
	}

	// Verify specific entries
	conj := findAspect(catalog, 0)
	if conj.Universality != "universal" {
		t.Errorf("conjunction universality = %q, want universal", conj.Universality)
	}
	trine := findAspect(catalog, 120)
	if trine.Universality != "universal" {
		t.Errorf("trine universality = %q, want universal", trine.Universality)
	}
	opp := findAspect(catalog, 180)
	if opp.Universality != "universal" {
		t.Errorf("opposition universality = %q, want universal", opp.Universality)
	}
}

func TestFormatAspectCatalog(t *testing.T) {
	out := FormatAspectCatalog()
	if out == "" {
		t.Error("expected non-empty output")
	}
	if !stringsContains(out, "UNIVERSAL") || !stringsContains(out, "conjunction") {
		t.Errorf("missing expected content in catalog output:\n%s", out)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────

func findPlanet(dc *DignityConvergence, name string) PlanetDignity {
	for _, p := range dc.Planets {
		if p.Planet == name {
			return p
		}
	}
	panic("planet not found: " + name)
}

func findAspect(catalog []AspectEntry, angle int) AspectEntry {
	for _, a := range catalog {
		if a.AngleDegrees == angle {
			return a
		}
	}
	panic("aspect not found")
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
