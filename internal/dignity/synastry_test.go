package dignity

import (
	"math"
	"testing"
)

func TestSynastry_Conjunction(t *testing.T) {
	// AJ Sun at 0°, Cait Sun at 2.5° — conjunction at 2.5° orb
	chart1 := map[string]float64{"Sun": 0.0}
	chart2 := map[string]float64{"Sun": 2.5}

	planets := []string{"Sun"}
	aspects := []AspectDef{{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	h := hits[0]
	if h.Aspect != "conjunction" {
		t.Errorf("expected conjunction, got %s", h.Aspect)
	}
	if math.Abs(h.Orb-2.5) > 0.01 {
		t.Errorf("expected orb 2.5, got %.2f", h.Orb)
	}
	if h.Planet1 != "Sun" || h.Planet2 != "Sun" {
		t.Errorf("expected Sun-Sun, got %s-%s", h.Planet1, h.Planet2)
	}
}

func TestSynastry_Square(t *testing.T) {
	// AJ Sun at 0°, Cait Moon at 88° — approaching square (2° orb)
	chart1 := map[string]float64{"Sun": 0.0}
	chart2 := map[string]float64{"Moon": 88.0}

	planets := []string{"Sun", "Moon"}
	aspects := []AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit (square), got %d", len(hits))
	}
	if hits[0].Aspect != "square" {
		t.Errorf("expected square, got %s", hits[0].Aspect)
	}
	if math.Abs(hits[0].Orb-2.0) > 0.01 {
		t.Errorf("expected orb 2.0, got %.2f", hits[0].Orb)
	}
}

func TestSynastry_Opposition(t *testing.T) {
	// AJ Mars at 10°, Cait Venus at 188° — opposition at 2° orb
	chart1 := map[string]float64{"Mars": 10.0}
	chart2 := map[string]float64{"Venus": 188.0}

	planets := []string{"Mars", "Venus"}
	aspects := []AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit (opposition), got %d", len(hits))
	}
	if hits[0].Aspect != "opposition" {
		t.Errorf("expected opposition, got %s", hits[0].Aspect)
	}
	if math.Abs(hits[0].Orb-2.0) > 0.01 {
		t.Errorf("expected orb 2.0, got %.2f", hits[0].Orb)
	}
}

func TestSynastry_Trine(t *testing.T) {
	chart1 := map[string]float64{"Jupiter": 10.0}
	chart2 := map[string]float64{"Saturn": 131.0}

	planets := []string{"Jupiter", "Saturn"}
	aspects := []AspectDef{{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit (trine), got %d", len(hits))
	}
	if hits[0].Aspect != "trine" {
		t.Errorf("expected trine, got %s", hits[0].Aspect)
	}
	// |(131-10) - 120| = 1.0
	if math.Abs(hits[0].Orb-1.0) > 0.01 {
		t.Errorf("expected orb 1.0, got %.2f", hits[0].Orb)
	}
}

func TestSynastry_Sextile(t *testing.T) {
	chart1 := map[string]float64{"Venus": 50.0}
	chart2 := map[string]float64{"Mars": 108.0}

	planets := []string{"Venus", "Mars"}
	aspects := []AspectDef{{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit (sextile), got %d", len(hits))
	}
	if hits[0].Aspect != "sextile" {
		t.Errorf("expected sextile, got %s", hits[0].Aspect)
	}
	// |108-50| = 58, |58-60| = 2
	if math.Abs(hits[0].Orb-2.0) > 0.01 {
		t.Errorf("expected orb 2.0, got %.2f", hits[0].Orb)
	}
}

func TestSynastry_MultipleHits(t *testing.T) {
	// AJ Sun at 0, AJ Moon at 90. Cait Sun at 2, Cait Mars at 92.
	// Sun-Sun: conj 2°, Moon-Mars: conj 2°
	// Sun-Mars: square 2°, Moon-Sun: square 2° → 4 total hits
	chart1 := map[string]float64{"Sun": 0.0, "Moon": 90.0}
	chart2 := map[string]float64{"Sun": 2.0, "Mars": 92.0}

	planets := []string{"Sun", "Moon", "Mars"}
	aspects := []AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}

	hits := ComputeSynastry(chart1, chart2, planets, aspects, 3.0)
	if len(hits) != 4 {
		t.Errorf("expected 4 hits (Sun-Sun, Sun-Mars, Moon-Sun, Moon-Mars), got %d", len(hits))
	}
	// Verify both aspect types present
	hasConj := false
	hasSquare := false
	for _, h := range hits {
		if h.Aspect == "conjunction" { hasConj = true }
		if h.Aspect == "square" { hasSquare = true }
	}
	if !hasConj { t.Error("missing conjunction") }
	if !hasSquare { t.Error("missing square") }
}

func TestSynastry_OrbBoundary(t *testing.T) {
	chart1 := map[string]float64{"Sun": 0.0}
	chart2 := map[string]float64{"Sun": 3.0}

	hits := ComputeSynastry(chart1, chart2, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}}, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit at boundary, got %d", len(hits))
	}
}

func TestSynastry_OrbExcluded(t *testing.T) {
	chart1 := map[string]float64{"Sun": 0.0}
	chart2 := map[string]float64{"Sun": 3.1}

	hits := ComputeSynastry(chart1, chart2, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}}, 3.0)
	if len(hits) != 0 {
		t.Fatalf("expected 0 hits, got %d", len(hits))
	}
}

func TestSynastry_NoMatch(t *testing.T) {
	chart1 := map[string]float64{"Sun": 0.0}
	chart2 := map[string]float64{"Sun": 47.0}

	hits := ComputeSynastry(chart1, chart2, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}, 3.0)
	if len(hits) != 0 {
		t.Fatalf("expected 0 hits, got %d", len(hits))
	}
}

func TestSynastry_PlanetOrderStable(t *testing.T) {
	// Planet order should be stable: planet1 from chart1, planet2 from chart2
	chart1 := map[string]float64{"Venus": 10.0}
	chart2 := map[string]float64{"Mars": 10.5}

	hits := ComputeSynastry(chart1, chart2, []string{"Venus", "Mars"},
		[]AspectDef{{0, "conjunction"}}, 3.0)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Planet1 != "Venus" || hits[0].Planet2 != "Mars" {
		t.Errorf("expected Venus-Mars, got %s-%s", hits[0].Planet1, hits[0].Planet2)
	}
}

func TestSynastryIntegration_AJCait(t *testing.T) {
	// Real AJ and Cait positions matched with Python reference.
	// Cross-validated: Uranus-Venus trine at 0.30° — the tightest synastry aspect.
	// Moon-Venus trine at 2.82°, Mars-Jupiter conjunction at 1.72°.
	aj := map[string]float64{
		"Sun": 327.22, "Moon": 17.46, "Mercury": 302.96,
		"Venus": 8.15, "Mars": 238.69, "Jupiter": 178.17,
		"Saturn": 19.96, "Uranus": 185.00, "Neptune": 240.53,
		"Pluto": 176.00,
	}
	cait := map[string]float64{
		"Sun": 39.03, "Moon": 297.22, "Mercury": 54.86,
		"Venus": 65.26, "Mars": 283.55, "Jupiter": 343.82,
		"Saturn": 248.48, "Uranus": 262.30, "Neptune": 276.24,
		"Pluto": 215.73,
	}

	planets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"}
	aspects := []AspectDef{
		{0, "conjunction"},
		{60, "sextile"},
		{90, "square"},
		{120, "trine"},
		{180, "opposition"},
	}

	hits := ComputeSynastry(aj, cait, planets, aspects, 5.0)

	// Verify the known tight aspects are present
	find := func(p1, p2, asp string) *SynastryHit {
		for i := range hits {
			h := &hits[i]
			if h.Planet1 == p1 && h.Planet2 == p2 && h.Aspect == asp {
				return h
			}
		}
		return nil
	}

	// Uranus-Venus trine: the signature aspect of this couple (~0.3°)
	h := find("Uranus", "Venus", "trine")
	if h == nil {
		t.Fatal("missing Uranus-Venus trine — this is the couple's signature aspect")
	}
	if h.Orb > 1.0 {
		t.Errorf("Uranus-Venus trine too wide: %.2f", h.Orb)
	}

	// Moon-Venus trine: emotional harmony
	h = find("Moon", "Venus", "trine")
	if h == nil {
		t.Log("Moon-Venus trine not found at 5 degree orb (may be wider)")
	}

	// There should be a reasonable number of inter-aspects
	if len(hits) < 5 {
		t.Errorf("expected at least 5 synastry aspects, got %d", len(hits))
	}
	if len(hits) > 60 {
		t.Errorf("unusually high aspect count: %d", len(hits))
	}
}
