package dignity

import (
	"math"
	"testing"
)

func TestAspectDetection_Conjunction(t *testing.T) {
	natalLongs := map[string]float64{"Sun": 0.0}
	natalPlanets := []string{"Sun"}

	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 0.5, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}

	hits, err := ScanTransits(natalLongs, natalPlanets,
		"2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if err != nil { t.Fatalf("error: %v", err) }
	if len(hits) != 1 { t.Fatalf("expected 1 hit, got %d", len(hits)) }
	h := hits[0]
	if h.Aspect != "conjunction" { t.Errorf("want conjunction, got %s", h.Aspect) }
	if math.Abs(h.Orb-0.5) > 0.01 { t.Errorf("want orb 0.5, got %.2f", h.Orb) }
}

func TestAspectDetection_Square(t *testing.T) {
	natalLongs := map[string]float64{"Sun": 0.0}
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 89.5, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(natalLongs, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(hits) != 1 { t.Fatalf("want 1, got %d", len(hits)) }
	if hits[0].Aspect != "square" { t.Errorf("want square, got %s", hits[0].Aspect) }
}

func TestAspectDetection_Opposition(t *testing.T) {
	natalLongs := map[string]float64{"Sun": 0.0}
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 179.7, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(natalLongs, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(hits) != 1 { t.Fatalf("want 1, got %d", len(hits)) }
	if hits[0].Aspect != "opposition" { t.Errorf("want opp, got %s", hits[0].Aspect) }
}

func TestAspectDetection_OrbBoundary(t *testing.T) {
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 3.0, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(map[string]float64{"Sun": 0.0}, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(hits) != 1 { t.Fatalf("want 1 at boundary, got %d", len(hits)) }
}

func TestAspectDetection_OrbExcluded(t *testing.T) {
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 3.1, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(map[string]float64{"Sun": 0.0}, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(hits) != 0 { t.Fatalf("want 0, got %d", len(hits)) }
}

func TestAspectDetection_MultipleNatal(t *testing.T) {
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 0.5, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(
		map[string]float64{"Sun": 0.0, "Moon": 90.0},
		[]string{"Sun", "Moon"},
		"2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn,
	)
	if len(hits) != 2 { t.Fatalf("want 2, got %d", len(hits)) }
}

func TestAspectDetection_MultiDay(t *testing.T) {
	idx := 0; orbit := []float64{0.5, 0.3, 0.8}
	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { o := orbit[idx]; idx++; return o, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}
	hits, _ := ScanTransits(
		map[string]float64{"Sun": 0.0}, []string{"Sun"},
		"2026-06-09", "2026-06-11", HardAspectsOnly(), 3.0, computeFn,
	)
	if len(hits) != 3 { t.Fatalf("want 3 over 3 days, got %d", len(hits)) }
}

func TestTransit_InvalidDate(t *testing.T) {
	_, err := ScanTransits(nil, nil, "bad", "2026-06-09", HardAspectsOnly(), 3.0, nil)
	if err == nil { t.Error("want error for bad start") }
}

func TestTransit_HardAspectsCount(t *testing.T) {
	if len(HardAspectsOnly()) != 3 { t.Fatal("want 3 hard aspects") }
}

func TestTransit_CompactTransits(t *testing.T) {
	hits := []TransitHit{
		{Date: "2026-06-09", TransitPlanet: "Saturn", NatalPlanet: "Saturn", Aspect: "opposition", Orb: 0.50},
		{Date: "2026-06-10", TransitPlanet: "Saturn", NatalPlanet: "Saturn", Aspect: "opposition", Orb: 0.00},
		{Date: "2026-06-11", TransitPlanet: "Saturn", NatalPlanet: "Saturn", Aspect: "opposition", Orb: 0.52},
	}
	c := CompactTransitsWithRange(hits)
	if len(c) != 1 { t.Fatalf("want 1 compact, got %d", len(c)) }
	if c[0].MinOrb != 0.0 { t.Errorf("want 0.0, got %.2f", c[0].MinOrb) }
}
