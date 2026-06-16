package dignity

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// Progressed Cross-System Comparison
// ═══════════════════════════════════════════════════════════════════════
//
// Secondary progressions use day-for-a-year: progressedJD = birthJD + age.
// The progressed planet positions are computed from the ephemeris at that JD.
// When comparing tropical vs sidereal, both natal AND progressed positions
// shift by the same ayanamsa. Angular distances are preserved.
//
// This is the same geometry as fixed star invariance (Phase 11, 100% survival)
// and Arabic Parts aspect invariance (Phase 9, 100% survival).
// The prediction: near-100% survival at any orb.
//
// The only source of non-survival is the ayanamsa drift over the progressed
// interval. At age 90, the ayanamsa has drifted ~0.003° — negligible at 3° orb.

func TestCompareCrossSystemProgressed_AllSurvivors(t *testing.T) {
	// Natal positions: Sun at 0°, Moon at 120°
	natal := map[string]float64{
		"Sun":  0.0,
		"Moon": 120.0,
	}

	// Progressed positions: Sun at 30°, Moon at 150° (both moved 30°)
	prog := map[string]float64{
		"Sun":  30.0,
		"Moon": 150.0,
	}

	// Ayanamsa ~24° (Lahiri)
	ayan := 24.0

	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemProgressed(natal, prog, ayan, aspects, orb)

	// Sun-Sun trine (30° → 120°): tropical: 30° dist → trine at 120°? No.
	// Let me think about what aspects actually form.
	// Sun(0) to progSun(30): dist=30°, sextile at 60° → no match at 3° orb
	// Sun(0) to progMoon(150): dist=150°, opposition at 180° → no (30° off)
	// Moon(120) to progSun(30): dist=90°, square at 90° → match! orb=0
	// Moon(120) to progMoon(150): dist=30°, sextile at 60° → no

	// So we get one hit: Moon-progSun square at 0° orb.
	// In sidereal: natal Moon=96°, progSun=6°, dist=90° → same square.
	// Should be a survivor.

	if len(result.Survivors) != 1 {
		t.Errorf("Expected 1 survivor, got %d", len(result.Survivors))
	}
	if result.Survivors[0].NatalPlanet != "Moon" || result.Survivors[0].ProgressedPlanet != "Sun" {
		t.Errorf("Survivor: want Moon-Sun, got %s-%s", result.Survivors[0].NatalPlanet, result.Survivors[0].ProgressedPlanet)
	}
	if result.Survivors[0].Aspect != "square" {
		t.Errorf("Aspect: want square, got %s", result.Survivors[0].Aspect)
	}

	if len(result.TropicalOnly) != 0 {
		t.Errorf("Expected 0 tropical-only, got %d", len(result.TropicalOnly))
	}
	if len(result.SiderealOnly) != 0 {
		t.Errorf("Expected 0 sidereal-only, got %d", len(result.SiderealOnly))
	}
}

func TestCompareCrossSystemProgressed_GeometryPreserved(t *testing.T) {
	// The key geometric claim: angular distance between natal and progressed
	// planet is identical in both zodiacs because both shift by the same ayanamsa.
	// This test verifies that with a larger ayanamsa and multiple aspects.

	natal := map[string]float64{
		"Sun":     0.0,
		"Moon":    90.0,
		"Mercury": 180.0,
	}
	prog := map[string]float64{
		"Sun":     0.0,  // conjunction with natal Sun
		"Moon":    180.0, // opposition with natal Moon (90→180 = 90° diff? No: natal Moon 90, prog Moon 180, dist=90 → square)
		"Mercury": 0.0,  // opposition with natal Mercury (180→0 = 180° dist)
	}
	// Actually let me be more careful:
	// natal Sun=0, prog Sun=0 → conjunction (0°)
	// natal Moon=90, prog Moon=180 → dist=90 → square
	// natal Mercury=180, prog Mercury=0 → dist=180 → opposition

	ayan := 24.0
	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemProgressed(natal, prog, ayan, aspects, orb)

	// All 9 pairs (3×3) form aspects. All should survive — geometry preserved.
	if len(result.Survivors) != 9 {
		t.Errorf("Expected 9 survivors, got %d (tropical-only=%d, sidereal-only=%d)",
			len(result.Survivors), len(result.TropicalOnly), len(result.SiderealOnly))
	}
	if len(result.TropicalOnly) != 0 {
		t.Errorf("Expected 0 tropical-only, got %d", len(result.TropicalOnly))
	}
	if len(result.SiderealOnly) != 0 {
		t.Errorf("Expected 0 sidereal-only, got %d", len(result.SiderealOnly))
	}
}

func TestCompareCrossSystemProgressed_TropicalOnly(t *testing.T) {
	// Create a scenario where a hit appears in tropical but not sidereal.
	// This requires the ayanamsa shift to push one planet pair out of orb
	// while another pair moves into orb. But since both natal and progressed
	// shift by the same amount, angular distances are preserved.
	//
	// The ONLY way to get tropical-only is if the ayanamsa drift over the
	// progressed interval is large enough to matter. At realistic ayanamsa
	// drift rates (~0.00003°/year), this is negligible.
	//
	// For testing the classification logic, we use an artificially large
	// ayanamsa DIFFERENCE between natal and progressed frames to simulate
	// what would happen if the ayanamsa changed significantly over a lifetime.
	// This doesn't happen in reality, but it exercises the code path.

	// Use different ayanamsa values for natal vs progressed to force divergence.
	// We'll test this by calling the function with positions that simulate
	// a scenario where tropical and sidereal disagree.

	// Actually, the function takes a single ayanamsa. The geometry guarantee
	// means tropical-only and sidereal-only should be empty for realistic data.
	// Let me test the classification logic by constructing positions where
	// the tropical and sidereal sets differ.

	// We'll use a helper approach: create positions where tropical forms an
	// aspect but sidereal doesn't (by using different ayanamsa values).
	// Since the real function uses one ayanamsa, we need to pre-shift.

	// Tropical natal: Sun=0, Moon=60
	// Tropical prog: Sun=120, Moon=180
	// → natal Sun(0) to prog Sun(120): dist=120 → trine ✓
	// → natal Moon(60) to prog Moon(180): dist=120 → trine ✓

	// Now sidereal (ayan=30): natal Sun=330, Moon=30, prog Sun=90, prog Moon=150
	// → natal Sun(330) to prog Sun(90): dist=120 → trine ✓ (same!)
	// → natal Moon(30) to prog Moon(150): dist=120 → trine ✓ (same!)

	// Angular distances are preserved. We can't get tropical-only with a single ayanamsa.
	// The classification code paths are exercised by the draconic cross-system tests.
	// For progressed, the test verifies the geometry guarantee: zero non-survivors.

	natal := map[string]float64{
		"Sun":  0.0,
		"Mars": 60.0,
	}
	prog := map[string]float64{
		"Sun":  120.0,
		"Mars": 180.0,
	}

	ayan := 30.0 // large ayanamsa, but same for both natal and prog
	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemProgressed(natal, prog, ayan, aspects, orb)

	// All 4 pairs (2×2) form aspects. All should survive — geometry preserved.
	if len(result.Survivors) != 4 {
		t.Errorf("Expected 4 survivors, got %d", len(result.Survivors))
	}
	if len(result.TropicalOnly) != 0 {
		t.Errorf("Expected 0 tropical-only with preserved geometry, got %d", len(result.TropicalOnly))
	}
	if len(result.SiderealOnly) != 0 {
		t.Errorf("Expected 0 sidereal-only with preserved geometry, got %d", len(result.SiderealOnly))
	}
}

func TestCompareCrossSystemProgressed_Empty(t *testing.T) {
	// No aspects at all
	natal := map[string]float64{"Sun": 0.0}
	prog := map[string]float64{"Sun": 45.0} // 45° — no Ptolemaic aspect at 3° orb

	ayan := 24.0
	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemProgressed(natal, prog, ayan, aspects, orb)

	if len(result.Survivors) != 0 {
		t.Errorf("Expected 0 survivors, got %d", len(result.Survivors))
	}
	if len(result.TropicalOnly) != 0 {
		t.Errorf("Expected 0 tropical-only, got %d", len(result.TropicalOnly))
	}
	if len(result.SiderealOnly) != 0 {
		t.Errorf("Expected 0 sidereal-only, got %d", len(result.SiderealOnly))
	}
}

func TestCompareCrossSystemProgressed_Conjunction(t *testing.T) {
	// Conjunction at 0° — should survive any ayanamsa
	natal := map[string]float64{"Sun": 100.0}
	prog := map[string]float64{"Sun": 100.0} // exact conjunction

	ayan := 24.0
	aspects := DefaultAspects()
	orb := 1.0

	result := CompareCrossSystemProgressed(natal, prog, ayan, aspects, orb)

	if len(result.Survivors) != 1 {
		t.Errorf("Expected 1 survivor (conjunction), got %d", len(result.Survivors))
	}
	if result.Survivors[0].Aspect != "conjunction" {
		t.Errorf("Expected conjunction, got %s", result.Survivors[0].Aspect)
	}
}
