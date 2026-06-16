package declination

import (
	"math"
	"testing"
)

// ── EclipticToDeclination ───────────────────────────────────────────────

func TestEclipticToDeclination(t *testing.T) {
	// Known values: Sun at 0° Aries (vernal equinox), 0° latitude → declination 0°
	decl := EclipticToDeclination(0, 0)
	if math.Abs(decl) > 0.01 {
		t.Errorf("Sun at 0° Aries, 0° lat: declination = %.4f, want 0", decl)
	}

	// Sun at 0° Cancer (summer solstice), 0° latitude → declination ~23.44° (obliquity)
	decl = EclipticToDeclination(90, 0)
	if math.Abs(decl-23.439) > 0.1 {
		t.Errorf("Sun at 0° Cancer, 0° lat: declination = %.4f, want ~23.44", decl)
	}

	// Sun at 0° Capricorn (winter solstice), 0° latitude → declination ~-23.44°
	decl = EclipticToDeclination(270, 0)
	if math.Abs(decl+23.439) > 0.1 {
		t.Errorf("Sun at 0° Capricorn, 0° lat: declination = %.4f, want ~-23.44", decl)
	}

	// Sun at 0° Libra (autumnal equinox), 0° latitude → declination 0°
	decl = EclipticToDeclination(180, 0)
	if math.Abs(decl) > 0.01 {
		t.Errorf("Sun at 0° Libra, 0° lat: declination = %.4f, want 0", decl)
	}
}

func TestEclipticToDeclinationWithLatitude(t *testing.T) {
	// Moon can have up to ~5° latitude. At 0° Cancer with +5° lat → declination > 23.44
	decl := EclipticToDeclination(90, 5)
	if decl < 23.44 {
		t.Errorf("Moon at 0° Cancer, +5° lat: declination = %.4f, should be > 23.44", decl)
	}

	// At 0° Cancer with -5° lat → declination < 23.44
	decl = EclipticToDeclination(90, -5)
	if decl > 23.44 {
		t.Errorf("Moon at 0° Cancer, -5° lat: declination = %.4f, should be < 23.44", decl)
	}
}

// ── ComputeDeclinations ──────────────────────────────────────────────────

func TestComputeDeclinations(t *testing.T) {
	positions := map[string][2]float64{
		"Sun":   {0, 0},
		"Moon":  {90, 0},
		"Mars":  {270, 0},
	}
	decls := ComputeDeclinations(positions)
	if len(decls) != 3 {
		t.Fatalf("expected 3 declinations, got %d", len(decls))
	}
	// Should be sorted by body name
	if decls[0].Body != "Mars" || decls[1].Body != "Moon" || decls[2].Body != "Sun" {
		t.Errorf("declinations not sorted by body name: %v", decls)
	}
	// Sun at 0° → 0° declination, North
	if math.Abs(decls[2].Declination) > 0.01 {
		t.Errorf("Sun declination = %.4f, want 0", decls[2].Declination)
	}
	if decls[2].Hemisphere != "North" {
		t.Errorf("Sun at 0° declination should be North, got %s", decls[2].Hemisphere)
	}
	// Moon at 90° → ~23.44° North
	if math.Abs(decls[1].Declination-23.44) > 0.1 {
		t.Errorf("Moon declination = %.4f, want ~23.44", decls[1].Declination)
	}
	if decls[1].Hemisphere != "North" {
		t.Errorf("Moon hemisphere = %s, want North", decls[1].Hemisphere)
	}
	// Mars at 270° → ~23.44° South
	if math.Abs(decls[0].Declination+23.44) > 0.1 {
		t.Errorf("Mars declination = %.4f, want ~-23.44", decls[0].Declination)
	}
	if decls[0].Hemisphere != "South" {
		t.Errorf("Mars hemisphere = %s, want South", decls[0].Hemisphere)
	}
}

// ── FindParallels ────────────────────────────────────────────────────────

func TestFindParallels(t *testing.T) {
	// Sun at 0° decl, Moon at 0.5° decl → parallel within 1° orb
	decls := []DeclinationData{
		{Body: "Sun", Declination: 0, Hemisphere: "North"},
		{Body: "Moon", Declination: 0.5, Hemisphere: "North"},
		{Body: "Mars", Declination: -0.3, Hemisphere: "South"},
	}
	pars := FindParallels(decls, 1.0)
	if len(pars) < 2 {
		t.Fatalf("expected at least 2 parallels, got %d", len(pars))
	}
	// Sun-Moon: parallel (same hemisphere, 0.5° orb)
	foundParallel := false
	foundContra := false
	for _, p := range pars {
		if (p.BodyA == "Sun" && p.BodyB == "Moon") || (p.BodyA == "Moon" && p.BodyB == "Sun") {
			if p.Type == "parallel" {
				foundParallel = true
			}
		}
		if (p.BodyA == "Sun" && p.BodyB == "Mars") || (p.BodyA == "Mars" && p.BodyB == "Sun") {
			if p.Type == "contraparallel" {
				foundContra = true
			}
		}
	}
	if !foundParallel {
		t.Errorf("Sun-Moon should be parallel: %+v", pars)
	}
	if !foundContra {
		t.Errorf("Sun-Mars should be contraparallel: %+v", pars)
	}
}

func TestFindParallelsNoMatch(t *testing.T) {
	decls := []DeclinationData{
		{Body: "Sun", Declination: 0, Hemisphere: "North"},
		{Body: "Moon", Declination: 20, Hemisphere: "North"},
	}
	pars := FindParallels(decls, 1.0)
	if len(pars) != 0 {
		t.Errorf("expected 0 parallels, got %d", len(pars))
	}
}

func TestFindParallelsSorted(t *testing.T) {
	decls := []DeclinationData{
		{Body: "Sun", Declination: 0, Hemisphere: "North"},
		{Body: "Moon", Declination: 0.8, Hemisphere: "North"},
		{Body: "Mars", Declination: 0.2, Hemisphere: "North"},
	}
	pars := FindParallels(decls, 1.0)
	// Should be sorted by orb (tightest first)
	for i := 1; i < len(pars); i++ {
		if pars[i].Orb < pars[i-1].Orb {
			t.Errorf("parallels not sorted by orb: %+v", pars)
		}
	}
}

// ── ComputeDeclinationReport ─────────────────────────────────────────────

func TestComputeDeclinationReport(t *testing.T) {
	positions := map[string][2]float64{
		"Sun":  {0, 0},
		"Moon": {90, 0},
	}
	report := ComputeDeclinationReport("test", positions, 1.0)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if len(report.Declinations) != 2 {
		t.Errorf("expected 2 declinations, got %d", len(report.Declinations))
	}
	// Sun at 0°, Moon at ~23.44° → not parallel within 1°
	if len(report.Parallels) != 0 {
		t.Errorf("Sun-Moon should not be parallel within 1°: %+v", report.Parallels)
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────

func TestDeclinationEmptyInput(t *testing.T) {
	report := ComputeDeclinationReport("empty", map[string][2]float64{}, 1.0)
	if len(report.Declinations) != 0 {
		t.Errorf("empty input should produce 0 declinations")
	}
	if len(report.Parallels) != 0 {
		t.Errorf("empty input should produce 0 parallels")
	}

	pars := FindParallels([]DeclinationData{}, 1.0)
	if len(pars) != 0 {
		t.Errorf("empty declinations should produce 0 parallels")
	}
}

func TestDeclinationSingleBody(t *testing.T) {
	positions := map[string][2]float64{"Sun": {0, 0}}
	report := ComputeDeclinationReport("single", positions, 1.0)
	if len(report.Declinations) != 1 {
		t.Errorf("expected 1 declination, got %d", len(report.Declinations))
	}
	if len(report.Parallels) != 0 {
		t.Errorf("single body should produce 0 parallels")
	}
}
