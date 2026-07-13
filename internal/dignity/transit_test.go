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

func TestFindTransitPeriods_SaturnConjunctSaturn(t *testing.T) {
	// AJ's natal Saturn at 21.48° — Saturn conjunct natal Saturn in 2027
	natalLongs := map[string]float64{"Saturn": 21.48}
	natalPlanets := []string{"Saturn"}
	aspects := []AspectDef{{0, "conjunction"}}

	periods, err := FindTransitPeriods(
		natalLongs, []string{"Saturn"}, natalPlanets,
		"2026-01-01", "2028-01-01",
		aspects, 3.0, 1.0,
	)
	if err != nil {
		t.Fatalf("FindTransitPeriods error: %v", err)
	}
	if len(periods) == 0 {
		t.Fatal("expected at least 1 Saturn-Saturn conjunction period, got 0")
	}
	// Each period should have 3 contacts: ingress, peak, egress
	for i, p := range periods {
		if len(p.Contacts) != 3 {
			t.Errorf("period %d: want 3 contacts, got %d", i, len(p.Contacts))
		}
		if p.Contacts[0].Type != ContactIngress {
			t.Errorf("period %d: first contact should be ingress, got %v", i, p.Contacts[0].Type)
		}
		if p.Contacts[1].Type != ContactPeak {
			t.Errorf("period %d: second contact should be peak, got %v", i, p.Contacts[1].Type)
		}
		if p.Contacts[2].Type != ContactEgress {
			t.Errorf("period %d: third contact should be egress, got %v", i, p.Contacts[2].Type)
		}
		if p.TransitPlanet != "Saturn" || p.NatalPlanet != "Saturn" || p.Aspect != "conjunction" {
			t.Errorf("period %d: unexpected planet/aspect: %s/%s/%s", i, p.TransitPlanet, p.NatalPlanet, p.Aspect)
		}
		// Peak orb should be very small for a conjunction
		if p.Contacts[1].Orb > 0.5 {
			t.Errorf("period %d: peak orb %.4f > 0.5", i, p.Contacts[1].Orb)
		}
	}
	t.Logf("Found %d Saturn-Saturn conjunction periods", len(periods))
}

func TestFindTransitPeriods_SaturnSquareMoon_Zero(t *testing.T) {
	// AJ's natal Moon at 322.29° — Saturn never squares it in 2026-2027
	natalLongs := map[string]float64{"Moon": 322.29}
	aspects := []AspectDef{{90, "square"}}

	periods, err := FindTransitPeriods(
		natalLongs, []string{"Saturn"}, []string{"Moon"},
		"2026-01-01", "2028-01-01",
		aspects, 3.0, 1.0,
	)
	if err != nil {
		t.Fatalf("FindTransitPeriods error: %v", err)
	}
	if len(periods) != 0 {
		t.Errorf("expected 0 Saturn-Moon square periods, got %d", len(periods))
		for _, p := range periods {
			t.Logf("  unexpected: %s %s %s, ingress=%v", p.TransitPlanet, p.Aspect, p.NatalPlanet, p.Contacts[0].JD)
		}
	}
}

func TestFindTransitPeriods_MercurySquareMoon(t *testing.T) {
	// AJ's natal Moon at 322.29° — Mercury squares it multiple times in 2026
	natalLongs := map[string]float64{"Moon": 322.29}
	aspects := []AspectDef{{90, "square"}}

	periods, err := FindTransitPeriods(
		natalLongs, []string{"Mercury"}, []string{"Moon"},
		"2026-01-01", "2027-01-01",
		aspects, 3.0, 1.0,
	)
	if err != nil {
		t.Fatalf("FindTransitPeriods error: %v", err)
	}
	if len(periods) < 2 {
		t.Fatalf("expected at least 2 Mercury-Moon square periods, got %d", len(periods))
	}
	for i, p := range periods {
		if len(p.Contacts) != 3 {
			t.Errorf("period %d: want 3 contacts, got %d", i, len(p.Contacts))
		}
		if p.TransitPlanet != "Mercury" || p.NatalPlanet != "Moon" || p.Aspect != "square" {
			t.Errorf("period %d: unexpected planet/aspect: %s/%s/%s", i, p.TransitPlanet, p.NatalPlanet, p.Aspect)
		}
	}
	t.Logf("Found %d Mercury-Moon square periods", len(periods))
}

func TestFindTransitPeriods_InvalidDate(t *testing.T) {
	_, err := FindTransitPeriods(
		map[string]float64{"Sun": 0}, []string{"Sun"}, []string{"Sun"},
		"bad-date", "2026-01-01",
		HardAspectsOnly(), 3.0, 1.0,
	)
	if err == nil {
		t.Error("expected error for invalid start date")
	}
}

func TestFindTransitPeriods_UnknownPlanet(t *testing.T) {
	_, err := FindTransitPeriods(
		map[string]float64{"Sun": 0}, []string{"NotAPlanet"}, []string{"Sun"},
		"2026-01-01", "2026-01-02",
		HardAspectsOnly(), 3.0, 1.0,
	)
	if err == nil {
		t.Error("expected error for unknown planet")
	}
}
