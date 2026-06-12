package dignity

import (
	"testing"
)

// ── Star Data Integrity ───────────────────────────────────────────────────

func TestNakshatraStarsCount(t *testing.T) {
	stars := NakshatraStars()
	if len(stars) != 27 {
		t.Errorf("NakshatraStars() = %d stars, want 27", len(stars))
	}
}

func TestXiuStarsCount(t *testing.T) {
	stars := XiuStars()
	if len(stars) != 28 {
		t.Errorf("XiuStars() = %d stars, want 28", len(stars))
	}
}

func TestCombinedPoolNoDuplicates(t *testing.T) {
	pool := CombinedPool()
	seen := make(map[string]bool)
	for _, s := range pool {
		if seen[s.Key] {
			t.Errorf("Duplicate key in combined pool: %s", s.Key)
		}
		seen[s.Key] = true
	}
}

func TestCombinedPoolSize(t *testing.T) {
	pool := CombinedPool()
	// 27 nakshatra + 28 xiu, with overlap
	if len(pool) < 27 || len(pool) > 55 {
		t.Errorf("CombinedPool() = %d stars, expected between 27 and 55", len(pool))
	}
}

// ── Shared Star Identification ────────────────────────────────────────────

func TestFindSharedStarsCount(t *testing.T) {
	shared := FindSharedStars()
	if len(shared) != 9 {
		t.Errorf("FindSharedStars() = %d shared, want 9", len(shared))
	}
}

func TestFindSharedStarsKnownStars(t *testing.T) {
	shared := FindSharedStars()
	sharedKeys := make(map[string]bool)
	for _, s := range shared {
		sharedKeys[s.Key] = true
	}

	// All 9 expected shared stars
	expected := []string{
		"Beta Arietis", "35 Arietis", "Alpha Virginis", "Alpha Scorpii",
		"Alpha Pegasi", "Gamma Pegasi", "Lambda Orionis", "Alpha Librae",
		"Delta Hydrae",
	}
	for _, key := range expected {
		if !sharedKeys[key] {
			t.Errorf("Expected shared star %s not found", key)
		}
	}
}

func TestFindSharedStarsFaintCount(t *testing.T) {
	shared := FindSharedStars()
	faint := 0
	bright := 0
	for _, s := range shared {
		if s.IsFaint {
			faint++
		} else {
			bright++
		}
	}
	if faint != 6 {
		t.Errorf("Faint shared stars = %d, want 6", faint)
	}
	if bright != 3 {
		t.Errorf("Bright shared stars = %d, want 3", bright)
	}
}

func TestFindSharedStarsDeltaHydrae(t *testing.T) {
	shared := FindSharedStars()
	found := false
	for _, s := range shared {
		if s.Key == "Delta Hydrae" {
			found = true
			if s.Nakshatra != "Ashlesha" {
				t.Errorf("Delta Hydrae nakshatra = %s, want Ashlesha", s.Nakshatra)
			}
			if s.Xiu != "Willow" {
				t.Errorf("Delta Hydrae xiu = %s, want Willow", s.Xiu)
			}
			if s.Magnitude < 4.0 || s.Magnitude > 4.2 {
				t.Errorf("Delta Hydrae magnitude = %.2f, want ~4.14", s.Magnitude)
			}
		}
	}
	if !found {
		t.Error("Delta Hydrae not found in shared stars")
	}
}

func TestFindSharedStarsNoPleiades(t *testing.T) {
	// The paper previously claimed Pleiades as shared.
	// Wikipedia lists Mao's determinant as 17 Tauri (Electra), not the cluster.
	// Krittika uses Alcyone (Eta Tauri). These are different stars.
	shared := FindSharedStars()
	for _, s := range shared {
		if s.Key == "Pleiades" || s.Key == "Eta Tauri" {
			t.Errorf("Pleiades/Eta Tauri should not be shared: Krittika uses Alcyone (Eta Tauri), Mao uses Electra (17 Tauri)")
		}
	}
}

func TestFindSharedStarsNoFalsePositives(t *testing.T) {
	shared := FindSharedStars()
	sharedKeys := make(map[string]bool)
	for _, s := range shared {
		sharedKeys[s.Key] = true
	}

	// Known false positives from earlier drafts
	falsePositives := []string{
		"Alpha Tauri",    // Aldebaran: Rohini uses it, Bi (Net) uses Epsilon Tauri
		"Alpha Geminorum", // Castor: Punarvasu uses it, Jing (Well) uses Mu Geminorum
	}
	for _, key := range falsePositives {
		if sharedKeys[key] {
			t.Errorf("False positive %s incorrectly identified as shared", key)
		}
	}
}

// ── Magnitude Classification ──────────────────────────────────────────────

func TestFaintThreshold(t *testing.T) {
	if FaintThreshold != 2.5 {
		t.Errorf("FaintThreshold = %.1f, want 2.5", FaintThreshold)
	}
}

func TestMarkabMagnitude(t *testing.T) {
	// Markab (Alpha Pegasi) is 2.48 in Swiss Ephemeris, not 2.5.
	// This flips it from FAINT to BRIGHT compared to the paper's original claim.
	naks := NakshatraStars()
	for _, s := range naks {
		if s.Key == "Alpha Pegasi" {
			if s.Magnitude >= 2.5 {
				t.Errorf("Markab magnitude = %.2f, should be < 2.5 (BRIGHT)", s.Magnitude)
			}
			if s.Magnitude < 2.4 || s.Magnitude > 2.6 {
				t.Errorf("Markab magnitude = %.2f, want ~2.48", s.Magnitude)
			}
		}
	}
}

func TestElectraMagnitude(t *testing.T) {
	// Electra (17 Tauri) is 3.70, not 1.6 (Pleiades cluster magnitude).
	// This flips it from BRIGHT to FAINT compared to the paper's original claim.
	xius := XiuStars()
	for _, s := range xius {
		if s.Key == "17 Tauri" {
			if s.Magnitude < 2.5 {
				t.Errorf("Electra magnitude = %.2f, should be >= 2.5 (FAINT)", s.Magnitude)
			}
		}
	}
}

// ── Null Model ─────────────────────────────────────────────────────────────

func TestNullModelBrightnessDeterministic(t *testing.T) {
	// Same seed should produce same results
	cfg := NullModelConfig{
		Name:            "test",
		Iterations:      100,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  2.5,
		Seed:            42,
	}
	r1 := RunNullModelBrightness(cfg)
	r2 := RunNullModelBrightness(cfg)

	if r1.NullMeanTotal != r2.NullMeanTotal {
		t.Errorf("Deterministic null model: mean total %.1f != %.1f", r1.NullMeanTotal, r2.NullMeanTotal)
	}
	if r1.NullMeanFaint != r2.NullMeanFaint {
		t.Errorf("Deterministic null model: mean faint %.1f != %.1f", r1.NullMeanFaint, r2.NullMeanFaint)
	}
}

func TestNullModelBrightnessObservedValues(t *testing.T) {
	cfg := NullModelConfig{
		Name:            "test",
		Iterations:      100,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  2.5,
		Seed:            42,
	}
	r := RunNullModelBrightness(cfg)

	if r.ObservedTotal != 9 {
		t.Errorf("Observed total = %d, want 9", r.ObservedTotal)
	}
	if r.ObservedFaint != 6 {
		t.Errorf("Observed faint = %d, want 6", r.ObservedFaint)
	}
	if r.ObservedBright != 3 {
		t.Errorf("Observed bright = %d, want 3", r.ObservedBright)
	}
}

func TestNullModelBrightnessPoolSize(t *testing.T) {
	cfg := NullModelConfig{
		Name:            "test",
		Iterations:      10,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  2.5,
		Seed:            42,
	}
	r := RunNullModelBrightness(cfg)

	if r.PoolSize != 46 {
		t.Errorf("Pool size = %d, want 46 (27 + 28 - 9 shared)", r.PoolSize)
	}
}

// ── Ecliptic Confound Test ─────────────────────────────────────────────────

func TestEclipticConfoundDeterministic(t *testing.T) {
	r1 := RunEclipticConfoundTest(1000, 42)
	r2 := RunEclipticConfoundTest(1000, 42)

	if r1.SharedFaintMeanLat != r2.SharedFaintMeanLat {
		t.Errorf("Deterministic confound test: shared faint mean lat %.1f != %.1f",
			r1.SharedFaintMeanLat, r2.SharedFaintMeanLat)
	}
}

func TestEclipticConfoundSharedFaintCount(t *testing.T) {
	r := RunEclipticConfoundTest(100, 42)

	if r.SharedFaintCount != 6 {
		t.Errorf("Shared faint count = %d, want 6", r.SharedFaintCount)
	}
}

// ── Full Report ────────────────────────────────────────────────────────────

func TestComputeLunarMansionReport(t *testing.T) {
	report := ComputeLunarMansionReport(42)

	if len(report.SharedStars) != 9 {
		t.Errorf("Report shared stars = %d, want 9", len(report.SharedStars))
	}

	if report.NullModelBrightness.ObservedTotal != 9 {
		t.Errorf("Report observed total = %d, want 9", report.NullModelBrightness.ObservedTotal)
	}

	if report.NullModelBrightness.ObservedFaint != 6 {
		t.Errorf("Report observed faint = %d, want 6", report.NullModelBrightness.ObservedFaint)
	}
}

func TestLunarMansionReportJSON(t *testing.T) {
	report := ComputeLunarMansionReport(42)
	jsonBytes, err := report.LunarMansionReportJSON()
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}
	if len(jsonBytes) < 100 {
		t.Errorf("JSON output too short: %d bytes", len(jsonBytes))
	}
}
