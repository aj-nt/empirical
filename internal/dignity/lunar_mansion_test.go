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

// ── Per-Chart Mansion Placement ───────────────────────────────────────────

func TestNakshatraForLongitude(t *testing.T) {
	// Ashwini (nakshatra 1) starts at 0° sidereal
	name, num := NakshatraForLongitude(0.0)
	if name != "Ashwini" || num != 1 {
		t.Errorf("0° → %s (#%d), want Ashwini (#1)", name, num)
	}

	// Mid-Ashwini
	name, num = NakshatraForLongitude(6.66)
	if name != "Ashwini" {
		t.Errorf("6.66° → %s, want Ashwini", name)
	}

	// Bharani starts at 13.33°
	name, num = NakshatraForLongitude(13.34)
	if name != "Bharani" || num != 2 {
		t.Errorf("13.34° → %s (#%d), want Bharani (#2)", name, num)
	}

	// Last nakshatra: Revati (27) at 346.67-360°
	name, num = NakshatraForLongitude(350.0)
	if name != "Revati" || num != 27 {
		t.Errorf("350° → %s (#%d), want Revati (#27)", name, num)
	}

	// Wrap-around
	name, num = NakshatraForLongitude(360.0)
	if name != "Ashwini" || num != 1 {
		t.Errorf("360° → %s (#%d), want Ashwini (#1)", name, num)
	}
}

func TestXiuForLongitude(t *testing.T) {
	// Jiao (Horn, xiu 1) anchored by Spica at 195.43°
	// Previous star: Zhen (Gienah, Gamma Corvi) at 176.89°
	// Boundary: (176.89 + 195.43)/2 = 186.16°
	// Next star: Kang (Kappa Virginis) at 207.45°
	// Boundary: (195.43 + 207.45)/2 = 201.44°
	// So Jiao spans 186.16° to 201.44°

	name, num, pinyin := XiuForLongitude(195.0)
	if name != "Horn" || num != 1 || pinyin != "Jiao" {
		t.Errorf("195° → %s (#%d, %s), want Horn (#1, Jiao)", name, num, pinyin)
	}

	// Kang (Neck, xiu 2) anchored by Kappa Virginis at 207.45°
	name, num, pinyin = XiuForLongitude(205.0)
	if name != "Neck" || num != 2 || pinyin != "Kang" {
		t.Errorf("205° → %s (#%d, %s), want Neck (#2, Kang)", name, num, pinyin)
	}

	// Wrap-around: Bi (Wall, xiu 14) anchored by Algenib at 9.16°
	// Previous: Shi (Markab) at 353.49°, boundary = (353.49+9.16)/2 = 1.325°
	// Next: Kui (Eta And) at 22.38°, boundary = (9.16+22.38)/2 = 15.77°
	name, num, pinyin = XiuForLongitude(5.0)
	if name != "Wall" || num != 14 || pinyin != "Bi" {
		t.Errorf("5° → %s (#%d, %s), want Wall (#14, Bi)", name, num, pinyin)
	}

	// Wrap-around: 0.5° falls in Shi (Encampment, xiu 13) sector [343.42°, 1.325°)
	name, num, pinyin = XiuForLongitude(0.5)
	if name != "Encampment" || num != 13 || pinyin != "Shi" {
		t.Errorf("0.5° → %s (#%d, %s), want Encampment (#13, Shi)", name, num, pinyin)
	}
}

func TestComputeMansionConvergence_KnownChart(t *testing.T) {
	// AJ's chart: tropical positions, Lahiri ayanamsa ~24.23°
	tropical := map[string]float64{
		"Sun":     327.27,
		"Moon":    329.58,
		"Mercury": 299.94,
		"Venus":   350.12,
		"Mars":    220.15,
		"Jupiter": 182.78,
		"Saturn":  50.42,
	}
	ayanamsa := 24.23

	result := ComputeMansionConvergence("AJ", tropical, ayanamsa)
	if result.Total != 7 {
		t.Errorf("Total = %d, want 7", result.Total)
	}
	if result.Ayanamsa != ayanamsa {
		t.Errorf("Ayanamsa = %f, want %f", result.Ayanamsa, ayanamsa)
	}

	// Verify each planet has both mansion assignments
	for _, p := range result.Planets {
		if p.Nakshatra == "" {
			t.Errorf("%s: empty nakshatra", p.Planet)
		}
		if p.Xiu == "" {
			t.Errorf("%s: empty xiu", p.Planet)
		}
		if p.NakshatraNum < 1 || p.NakshatraNum > 27 {
			t.Errorf("%s: nakshatra num %d out of range", p.Planet, p.NakshatraNum)
		}
		if p.XiuNum < 1 || p.XiuNum > 28 {
			t.Errorf("%s: xiu num %d out of range", p.Planet, p.XiuNum)
		}
	}

	// Convergence should be low — only 9 of 55 mansion pairs share stars
	// Most planets won't land in a shared pair
	if result.Converging > result.Total {
		t.Errorf("Converging %d > Total %d", result.Converging, result.Total)
	}
}

func TestComputeMansionConvergence_NoConvergence(t *testing.T) {
	// Chart where all planets fall in non-shared mansion pairs
	// Use longitudes that avoid the 9 shared anchor stars
	tropical := map[string]float64{
		"Sun":     100.0,
		"Moon":    120.0,
		"Mercury": 140.0,
		"Venus":   160.0,
		"Mars":    250.0,
		"Jupiter": 270.0,
		"Saturn":  290.0,
	}
	ayanamsa := 24.0

	result := ComputeMansionConvergence("test", tropical, ayanamsa)
	if result.Converging != 0 {
		t.Errorf("Converging = %d, want 0 for non-shared positions", result.Converging)
	}
}
