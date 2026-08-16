package dignity

import (
	"testing"

)

func initEpheForWFB(t *testing.T) {
	t.Helper()
	initEphe(t)
}

func TestWesternFromBase_KnownChart(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)

	if report == nil {
		t.Fatal("WesternFromBase returned nil")
	}
	if report.Name != "AJ" {
		t.Errorf("Name = %q, want %q", report.Name, "AJ")
	}
	if len(report.PlanetSigns) == 0 {
		t.Error("expected non-empty PlanetSigns")
	}
	if len(report.PlanetHouses) == 0 {
		t.Error("expected non-empty PlanetHouses")
	}
	if len(report.Aspects) == 0 {
		t.Error("expected non-empty Aspects")
	}
	if len(report.Patterns) == 0 {
		t.Error("expected non-empty Patterns (Western planet set)")
	}
	if len(report.Stars) == 0 {
		t.Error("expected non-empty Stars (star conjunctions)")
	}
}

func TestWesternFromBase_IncludesOuterPlanets(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)

	// Verify outer planets appear in PlanetSigns
	outerPlanets := map[string]bool{"Uranus": false, "Neptune": false, "Pluto": false}
	for _, ps := range report.PlanetSigns {
		for planet := range outerPlanets {
			if len(ps) > len(planet) && ps[:len(planet)] == planet {
				outerPlanets[planet] = true
			}
		}
	}
	for planet, found := range outerPlanets {
		if !found {
			t.Errorf("outer planet %s not found in PlanetSigns", planet)
		}
	}
}

func TestWesternFromBase_EmptyName(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	if report == nil {
		t.Fatal("WesternFromBase returned nil for empty name")
	}
	if len(report.PlanetSigns) == 0 {
		t.Error("expected non-empty PlanetSigns for empty name chart")
	}
}

func TestWesternFromBase_CustomOrb(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	tight := WesternFromBase(bc, 1.0, false)
	wide := WesternFromBase(bc, 10.0, false)

	if tight == nil || wide == nil {
		t.Fatal("WesternFromBase returned nil")
	}
	if len(tight.Aspects) > len(wide.Aspects) {
		t.Errorf("tight orb (%d aspects) should not exceed wide orb (%d aspects)",
			len(tight.Aspects), len(wide.Aspects))
	}
}

func TestWesternFromBase_JSON(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	data, err := report.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("JSON() returned empty bytes")
	}
}

func TestFindDirectMidpoints_Basic(t *testing.T) {
	// Sun at 10°, Moon at 50°, Mars at 30° — Mars sits at Sun/Moon midpoint
	objects := map[string]float64{
		"Sun":  10.0,
		"Moon": 50.0,
		"Mars": 30.0,
	}
	hits := FindDirectMidpoints(objects, 1.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].PairA != "Moon" || hits[0].PairB != "Sun" || hits[0].Planet != "Mars" {
		t.Errorf("unexpected hit: %+v", hits[0])
	}
	if hits[0].Orb > 0.01 {
		t.Errorf("orb should be ~0, got %.4f", hits[0].Orb)
	}
}

func TestFindDirectMidpoints_Wraparound(t *testing.T) {
	// Sun at 350°, Moon at 30°, Mars at 10° — midpoint across 0° is 10°
	objects := map[string]float64{
		"Sun":  350.0,
		"Moon": 30.0,
		"Mars": 10.0,
	}
	hits := FindDirectMidpoints(objects, 1.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Planet != "Mars" {
		t.Errorf("expected Mars at midpoint, got %s", hits[0].Planet)
	}
}

func TestFindDirectMidpoints_NoHit(t *testing.T) {
	objects := map[string]float64{
		"Sun":   10.0,
		"Moon":  50.0,
		"Mars":  100.0,
		"Venus": 200.0,
	}
	hits := FindDirectMidpoints(objects, 1.0)
	if len(hits) != 0 {
		for _, h := range hits {
			t.Logf("unexpected hit: %s/%s = %s (orb %.2f)", h.PairA, h.PairB, h.Planet, h.Orb)
		}
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}

func TestFindDirectMidpoints_ExcludesSelf(t *testing.T) {
	// Sun at 10°, Moon at 10° — midpoint is 10°, but Sun and Moon are the pair
	objects := map[string]float64{
		"Sun":  10.0,
		"Moon": 10.0,
	}
	hits := FindDirectMidpoints(objects, 1.0)
	if len(hits) != 0 {
		t.Errorf("expected 0 hits (no third planet), got %d", len(hits))
	}
}

func TestWesternFromBase_HasMidpoints(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	if len(report.Midpoints) == 0 {
		t.Error("expected non-empty Midpoints in WesternFromBase output")
	}
	t.Logf("Midpoints: %d", len(report.Midpoints))
}

func TestElementBalance_Basic(t *testing.T) {
	// Sun in Aries (Fire), Moon in Cancer (Water), Mars in Capricorn (Earth)
	planets := map[string]float64{
		"Sun":  5.0,   // Aries
		"Moon": 100.0, // Cancer
		"Mars": 280.0, // Capricorn
	}
	eb := ComputeElementBalance(planets)
	if eb["Fire"] != 1 {
		t.Errorf("Fire = %d, want 1", eb["Fire"])
	}
	if eb["Water"] != 1 {
		t.Errorf("Water = %d, want 1", eb["Water"])
	}
	if eb["Earth"] != 1 {
		t.Errorf("Earth = %d, want 1", eb["Earth"])
	}
	if eb["Air"] != 0 {
		t.Errorf("Air = %d, want 0", eb["Air"])
	}
}

func TestModalityBalance_Basic(t *testing.T) {
	planets := map[string]float64{
		"Sun":  5.0,   // Aries = Cardinal
		"Moon": 55.0,  // Taurus = Fixed
		"Mars": 100.0, // Cancer = Cardinal
	}
	mb := ComputeModalityBalance(planets)
	if mb["Cardinal"] != 2 {
		t.Errorf("Cardinal = %d, want 2", mb["Cardinal"])
	}
	if mb["Fixed"] != 1 {
		t.Errorf("Fixed = %d, want 1", mb["Fixed"])
	}
	if mb["Mutable"] != 0 {
		t.Errorf("Mutable = %d, want 0", mb["Mutable"])
	}
}

func TestWesternFromBase_HasElementBalance(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	if len(report.ElementBalance) == 0 {
		t.Error("expected non-empty ElementBalance")
	}
	if len(report.ModalityBalance) == 0 {
		t.Error("expected non-empty ModalityBalance")
	}
	t.Logf("Elements: %v", report.ElementBalance)
	t.Logf("Modalities: %v", report.ModalityBalance)
}

func TestHemisphereEmphasis_Basic(t *testing.T) {
	// ASC at 0° Aries → houses 1-6 below, 7-12 above
	// East: 10,11,12,1,2,3  West: 4,5,6,7,8,9
	asc := 0.0
	// Sun at 15° Aries = house 1 = below, east
	// Moon at 15° Libra = house 7 = above, west
	// Mars at 15° Cancer = house 4 = below, west
	planets := map[string]float64{
		"Sun":  15.0,  // Aries, house 1
		"Moon": 195.0, // Libra, house 7
		"Mars": 105.0, // Cancer, house 4
	}
	he := ComputeHemisphereEmphasis(planets, asc)
	if he.Above != 1 {
		t.Errorf("Above = %d, want 1", he.Above)
	}
	if he.Below != 2 {
		t.Errorf("Below = %d, want 2", he.Below)
	}
	if he.East != 1 {
		t.Errorf("East = %d, want 1", he.East)
	}
	if he.West != 2 {
		t.Errorf("West = %d, want 2", he.West)
	}
}

func TestWesternFromBase_HasHemisphereEmphasis(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	if report.Hemisphere == nil {
		t.Fatal("expected non-nil Hemisphere")
	}
	t.Logf("Hemisphere: Above=%d Below=%d East=%d West=%d",
		report.Hemisphere.Above, report.Hemisphere.Below,
		report.Hemisphere.East, report.Hemisphere.West)
}

func TestRulershipChain_Basic(t *testing.T) {
	// ASC at 0° Aries → house 1 cusp = 0° Aries, ruler = Mars
	// Mars at 105° Cancer → house 4 (whole-sign from ASC 0°)
	// House 4 cusp = 90° Cancer, ruler = Moon
	// Moon at 15° Aries → house 1
	// Chain: 1: Aries→Mars in 4→Cancer→Moon in 1→Aries (loop)
	asc := 0.0
	houseCusps := []float64{0, 0, 30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330}
	planets := map[string]float64{
		"Sun":  45.0,  // Taurus, house 2
		"Moon": 15.0,  // Aries, house 1
		"Mars": 105.0, // Cancer, house 4
	}
	chains := ComputeRulershipChains(houseCusps, planets, asc)
	if len(chains) == 0 {
		t.Fatal("expected non-empty chains")
	}
	t.Logf("Chains: %v", chains)
	// House 1 chain should involve Aries→Mars
	if chain, ok := chains[1]; !ok {
		t.Error("expected chain for house 1")
	} else if len(chain) == 0 {
		t.Error("house 1 chain is empty")
	}
}

func TestDispositorTree_Basic(t *testing.T) {
	// Sun at 5° Leo → ruler = Sun (in its own sign — final dispositor)
	// Moon at 5° Cancer → ruler = Moon (in its own sign — final dispositor)
	// Mars at 5° Aries → ruler = Mars (in its own sign — final dispositor)
	planets := map[string]float64{
		"Sun":  125.0, // Leo
		"Moon": 95.0,  // Cancer
		"Mars": 5.0,   // Aries
	}
	trees := ComputeDispositorTrees(planets)
	if len(trees) != 3 {
		t.Fatalf("expected 3 trees, got %d", len(trees))
	}
	for planet, tree := range trees {
		if len(tree) == 0 {
			t.Errorf("empty tree for %s", planet)
		}
		t.Logf("%s: %s", planet, tree)
	}
}

func TestDispositorTree_MutualReception(t *testing.T) {
	// Mars at 5° Libra (Venus-ruled), Venus at 5° Aries (Mars-ruled)
	// Mutual reception — chain should show the loop
	planets := map[string]float64{
		"Mars":  185.0, // Libra
		"Venus": 5.0,   // Aries
	}
	trees := ComputeDispositorTrees(planets)
	if len(trees) != 2 {
		t.Fatalf("expected 2 trees, got %d", len(trees))
	}
	for planet, tree := range trees {
		t.Logf("%s: %s", planet, tree)
	}
}

func TestWesternFromBase_HasRulershipChains(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0, false)
	if len(report.RulershipChains) == 0 {
		t.Error("expected non-empty RulershipChains")
	}
	if len(report.DispositorTrees) == 0 {
		t.Error("expected non-empty DispositorTrees")
	}
	t.Logf("RulershipChains: %d houses", len(report.RulershipChains))
	t.Logf("DispositorTrees: %d planets", len(report.DispositorTrees))
}
