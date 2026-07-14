package dignity

import (
	"math"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

func TestFindStarConjunctions(t *testing.T) {
	// Star at 15°, planet at 15.5° — conjunction at 0.5° orb
	starPositions := map[string]float64{
		"Aldebaran": 15.0,
		"Sirius":    120.0,
	}
	planetPositions := map[string]float64{
		"Sun":   15.5,
		"Moon":  200.0,
		"Mars":  119.2,
	}

	conj := FindStarConjunctions(starPositions, planetPositions, 2.0)

	// Aldebaran conjunct Sun at 0.5°
	found := false
	for _, c := range conj {
		if c.Star == "Aldebaran" && c.Planet == "Sun" {
			found = true
			if math.Abs(c.Orb-0.5) > 0.01 {
				t.Errorf("Aldebaran-Sun orb: expected 0.5, got %.2f", c.Orb)
			}
		}
	}
	if !found {
		t.Error("missing Aldebaran-Sun conjunction")
	}

	// Sirius conjunct Mars at 0.8°
	found = false
	for _, c := range conj {
		if c.Star == "Sirius" && c.Planet == "Mars" {
			found = true
			if math.Abs(c.Orb-0.8) > 0.01 {
				t.Errorf("Sirius-Mars orb: expected 0.8, got %.2f", c.Orb)
			}
		}
	}
	if !found {
		t.Error("missing Sirius-Mars conjunction")
	}

	// Moon at 200° — no star near it
	for _, c := range conj {
		if c.Planet == "Moon" {
			t.Errorf("Moon should have no conjunctions, got %s", c.Star)
		}
	}

	// Should be sorted by orb
	if len(conj) >= 2 && conj[0].Orb > conj[1].Orb {
		t.Error("conjunctions not sorted by orb")
	}
}

func TestFindStarConjunctions_OrbBoundary(t *testing.T) {
	starPositions := map[string]float64{"Aldebaran": 15.0}
	planetPositions := map[string]float64{"Sun": 17.1} // 2.1° — outside 2.0° orb

	conj := FindStarConjunctions(starPositions, planetPositions, 2.0)
	if len(conj) != 0 {
		t.Errorf("expected 0 conjunctions at 2.1°, got %d", len(conj))
	}

	// At 2.1° orb threshold, it should appear
	conj = FindStarConjunctions(starPositions, planetPositions, 2.1)
	if len(conj) != 1 {
		t.Errorf("expected 1 conjunction at 2.1° orb, got %d", len(conj))
	}
}

func TestFindStarConjunctions_Wraparound(t *testing.T) {
	// Star at 358°, planet at 2° — orb is 4° across 0° boundary
	starPositions := map[string]float64{"Scheat": 358.0}
	planetPositions := map[string]float64{"Chiron": 2.0}

	conj := FindStarConjunctions(starPositions, planetPositions, 5.0)
	if len(conj) != 1 {
		t.Fatalf("expected 1 conjunction across 0°, got %d", len(conj))
	}
	if math.Abs(conj[0].Orb-4.0) > 0.01 {
		t.Errorf("orb across 0°: expected 4.0, got %.2f", conj[0].Orb)
	}
}

func TestStarCatalog(t *testing.T) {
	// Verify the catalog has the expected stars
	if _, ok := StarMeanings["Aldebaran"]; !ok {
		t.Error("Aldebaran missing from StarMeanings")
	}
	if _, ok := StarMeanings["Sirius"]; !ok {
		t.Error("Sirius missing from StarMeanings")
	}
	if _, ok := StarMeanings["Regulus"]; !ok {
		t.Error("Regulus missing from StarMeanings")
	}
	if _, ok := StarMeanings["Spica"]; !ok {
		t.Error("Spica missing from StarMeanings")
	}
	if _, ok := StarMeanings["Vega"]; !ok {
		t.Error("Vega missing from StarMeanings")
	}

	// StarNames should match StarMeanings keys
	if len(StarNames) != len(StarMeanings) {
		t.Errorf("StarNames (%d) != StarMeanings (%d)", len(StarNames), len(StarMeanings))
	}
	for _, name := range StarNames {
		if _, ok := StarMeanings[name]; !ok {
			t.Errorf("StarNames has %s but StarMeanings doesn't", name)
		}
	}
}

// ── Cross-System Star Conjunction Tests ───────────────────────────────────

func TestCompareStarConjunctionsCrossSystem_AllSurvive(t *testing.T) {
	// Star and planet positions shift by the same ayanamsa.
	// Angular distances are preserved. All conjunctions should survive.
	starPos := map[string]float64{
		"Spica":   203.84,
		"Regulus": 149.83,
		"Aldebaran": 69.79,
	}
	planetPos := map[string]float64{
		"Sun":    204.00, // 0.16° from Spica
		"Moon":   150.00, // 0.17° from Regulus
		"Mars":   70.00,  // 0.21° from Aldebaran
	}
	ayanamsa := 24.0
	orb := 2.0

	result := CompareStarConjunctionsCrossSystem("test", starPos, planetPos, ayanamsa, orb)

	if result.TotalTrop != 3 {
		t.Errorf("Expected 3 tropical conjunctions, got %d", result.TotalTrop)
	}
	if result.TotalSid != 3 {
		t.Errorf("Expected 3 sidereal conjunctions, got %d", result.TotalSid)
	}
	if result.TotalSurvivors != 3 {
		t.Errorf("Expected 3 survivors (100%% invariance), got %d", result.TotalSurvivors)
	}
}

func TestCompareStarConjunctionsCrossSystem_OrbBoundary(t *testing.T) {
	// A conjunction exactly at the orb boundary should survive in both frames
	starPos := map[string]float64{
		"Vega": 15.00,
	}
	planetPos := map[string]float64{
		"Sun": 17.00, // 2.00° orb
	}
	ayanamsa := 24.0
	orb := 2.0

	result := CompareStarConjunctionsCrossSystem("test", starPos, planetPos, ayanamsa, orb)

	if result.TotalSurvivors != 1 {
		t.Errorf("Expected 1 survivor at orb boundary, got %d", result.TotalSurvivors)
	}
}

func TestCompareStarConjunctionsCrossSystem_NoConjunctions(t *testing.T) {
	starPos := map[string]float64{
		"Spica": 203.84,
	}
	planetPos := map[string]float64{
		"Sun": 10.00, // far from Spica
	}
	ayanamsa := 24.0
	orb := 2.0

	result := CompareStarConjunctionsCrossSystem("test", starPos, planetPos, ayanamsa, orb)

	if result.TotalTrop != 0 {
		t.Errorf("Expected 0 tropical conjunctions, got %d", result.TotalTrop)
	}
	if result.TotalSid != 0 {
		t.Errorf("Expected 0 sidereal conjunctions, got %d", result.TotalSid)
	}
}

// ── Golden Test: AJ's Natal Chart + Regulus ────────────────────────────────

func TestFindStarAspects_AJ_Regulus(t *testing.T) {
	initEphe(t)

	bc, err := ComputeBaseChart(BirthData{
		Name:     "AJ",
		Year:     1969,
		Month:    2,
		Day:      15,
		Hour:     23,
		Minute:   10,
		Second:   0,
		TZOffset: -8.0,
		Lat:      47.038,
		Lng:      -122.901,
	})
	if err != nil {
		t.Fatalf("ComputeBaseChart: %v", err)
	}

	// Regulus position via Swiss Ephemeris
	regulusLon, _, _, _ := swe.Fixstar("Regulus", bc.JD)
	regulusLon = math.Mod(regulusLon+360, 360)

	// Golden: Regulus at ~149.41° (Leo)
	if math.Abs(regulusLon-149.41) > 0.05 {
		t.Errorf("Regulus tropical: got %.2f°, want ~149.41°", regulusLon)
	}

	// Extract planet longitudes from BaseChart (traditional 10 only, matching
	// the original check_regulus tool's planet set)
	planetLons := make(map[string]float64)
	for _, name := range []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"} {
		if pos, ok := bc.Tropical[name]; ok {
			planetLons[name] = pos.Lon
		}
	}

	// Find all aspects within 5° orb
	hits := FindStarAspects(regulusLon, "Regulus", planetLons, DefaultAspects(), 5.0)

	// Golden: 3 aspects — Sun opposition, Neptune square, Mars square
	if len(hits) != 3 {
		t.Fatalf("expected 3 aspects, got %d: %+v", len(hits), hits)
	}

	// Tightest: Neptune square Regulus at 0.74°
	if hits[0].Planet != "Neptune" || hits[0].Aspect != "square" {
		t.Errorf("tightest: want Neptune square, got %s %s", hits[0].Planet, hits[0].Aspect)
	}
	if math.Abs(hits[0].Orb-0.74) > 0.02 {
		t.Errorf("Neptune square orb: want 0.74, got %.2f", hits[0].Orb)
	}

	// Second: Sun opposition Regulus at 1.96°
	if hits[1].Planet != "Sun" || hits[1].Aspect != "opposition" {
		t.Errorf("second: want Sun opposition, got %s %s", hits[1].Planet, hits[1].Aspect)
	}
	if math.Abs(hits[1].Orb-1.96) > 0.02 {
		t.Errorf("Sun opposition orb: want 1.96, got %.2f", hits[1].Orb)
	}

	// Third: Mars square Regulus at 3.60°
	if hits[2].Planet != "Mars" || hits[2].Aspect != "square" {
		t.Errorf("third: want Mars square, got %s %s", hits[2].Planet, hits[2].Aspect)
	}
	if math.Abs(hits[2].Orb-3.60) > 0.02 {
		t.Errorf("Mars square orb: want 3.60, got %.2f", hits[2].Orb)
	}
}

// ── Star Catalog Validation ─────────────────────────────────────────────

func TestCheckStarCatalog_Valid(t *testing.T) {
	// Should not panic with consistent catalog
	names := []string{"Sirius", "Vega", "Spica"}
	meanings := map[string]string{
		"Sirius": "bright",
		"Vega":   "harp star",
		"Spica":  "wheat",
	}
	// This should not panic
	checkStarCatalog(names, meanings)
}

func TestCheckStarCatalog_LengthMismatch(t *testing.T) {
	names := []string{"Sirius", "Vega"}
	meanings := map[string]string{"Sirius": "bright"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for length mismatch")
		}
	}()
	checkStarCatalog(names, meanings)
}

func TestCheckStarCatalog_MissingMeaning(t *testing.T) {
	names := []string{"Sirius", "Vega"}
	meanings := map[string]string{"Sirius": "bright"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing meaning")
		}
	}()
	checkStarCatalog(names, meanings)
}

func TestCheckStarCatalog_ExtraMeaning(t *testing.T) {
	names := []string{"Sirius"}
	meanings := map[string]string{"Sirius": "bright", "Vega": "harp star"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for extra meaning")
		}
	}()
	checkStarCatalog(names, meanings)
}

func TestValidateStarCatalog_DoesNotPanic(t *testing.T) {
	// The real catalog should validate without panicking
	validateStarCatalog()
}
