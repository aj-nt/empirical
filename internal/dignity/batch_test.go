package dignity

import (
	"testing"
)

func TestBatchTransits_SinglePerson(t *testing.T) {
	// One person, fake transit engine that returns a fixed hit.
	// Batch should collect it under that person's name.
	people := []BatchPerson{
		{Name: "Pete", PlanetLongs: map[string]float64{"Sun": 0.0}},
	}

	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 { return 0.5, 0, 0, 0 }
		return 47.0, 0, 0, 0
	}

	results := BatchTransits(people, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Name != "Pete" {
		t.Errorf("expected Pete, got %s", r.Name)
	}
	if len(r.Hits) != 1 {
		t.Fatalf("expected 1 hit for Pete, got %d", len(r.Hits))
	}
	if r.Hits[0].Aspect != "conjunction" {
		t.Errorf("expected conjunction, got %s", r.Hits[0].Aspect)
	}
}

func TestBatchTransits_TwoPeople(t *testing.T) {
	// Two people with different natal positions.
	// Person A: Sun at 0 (gets conj at 0.5 orb)
	// Person B: Sun at 90 (gets square at 0.5 orb)
	people := []BatchPerson{
		{Name: "AJ", PlanetLongs: map[string]float64{"Sun": 0.0}},
		{Name: "Pete", PlanetLongs: map[string]float64{"Sun": 90.0}},
	}

	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID == 0 {
			return 0.5, 0, 0, 0 // transit Sun at 0.5
		}
		return 200.0, 0, 0, 0 // scatter
	}

	results := BatchTransits(people, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// AJ's Sun at 0 → conj from transit at 0.5
	if results[0].Name != "AJ" {
		t.Errorf("expected AJ first, got %s", results[0].Name)
	}
	if len(results[0].Hits) != 1 || results[0].Hits[0].Aspect != "conjunction" {
		t.Errorf("AJ should have 1 conjunction hit")
	}

	// Pete's Sun at 90 → square from transit at 0.5 (|90-0.5|=89.5, diff from 90 = 0.5)
	if results[1].Name != "Pete" {
		t.Errorf("expected Pete second, got %s", results[1].Name)
	}
	if len(results[1].Hits) != 1 || results[1].Hits[0].Aspect != "square" {
		t.Errorf("Pete should have 1 square hit, got %d hits, aspect=%s",
			len(results[1].Hits), results[1].Hits[0].Aspect)
	}
}

func TestBatchTransits_NoHits(t *testing.T) {
	people := []BatchPerson{
		{Name: "Ghost", PlanetLongs: map[string]float64{"Sun": 0.0}},
	}

	computeFn := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 47.0, 0, 0, 0 // nowhere near 0, 90, or 180
	}

	results := BatchTransits(people, []string{"Sun"}, "2026-06-09", "2026-06-09", HardAspectsOnly(), 3.0, computeFn)
	if len(results) != 1 {
		t.Fatalf("expected 1 result entry, got %d", len(results))
	}
	if len(results[0].Hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(results[0].Hits))
	}
}

func TestBatchSynastry_OnePair(t *testing.T) {
	// Single pair: AJ Sun at 0, Cait Sun at 2 → conjunction at 2°
	people := []BatchPerson{
		{Name: "AJ", PlanetLongs: map[string]float64{"Sun": 0.0}},
		{Name: "Cait", PlanetLongs: map[string]float64{"Sun": 2.0}},
	}

	results := BatchSynastry(people, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}, 3.0)
	if len(results) != 1 {
		t.Fatalf("expected 1 pair result, got %d", len(results))
	}
	r := results[0]
	if r.Name1 != "AJ" || r.Name2 != "Cait" {
		t.Errorf("expected AJ-Cait, got %s-%s", r.Name1, r.Name2)
	}
	if len(r.Hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(r.Hits))
	}
	if r.Hits[0].Aspect != "conjunction" || r.Hits[0].Orb > 2.01 {
		t.Errorf("expected conjunction at ~2.0, got %s at %.2f", r.Hits[0].Aspect, r.Hits[0].Orb)
	}
}

func TestBatchSynastry_AllPairs(t *testing.T) {
	// Three people → 3 pairs: AJ-Cait, AJ-Pete, Cait-Pete
	// AJ Sun 0, Cait Sun 2 (conj), Pete Sun 90 (square to AJ)
	people := []BatchPerson{
		{Name: "AJ", PlanetLongs: map[string]float64{"Sun": 0.0}},
		{Name: "Cait", PlanetLongs: map[string]float64{"Sun": 2.0}},
		{Name: "Pete", PlanetLongs: map[string]float64{"Sun": 90.0}},
	}

	results := BatchSynastry(people, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}, {90, "square"}, {180, "opposition"}}, 3.0)
	if len(results) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(results))
	}
}

func TestBatchSynastry_EmptyList(t *testing.T) {
	results := BatchSynastry(nil, []string{"Sun"},
		[]AspectDef{{0, "conjunction"}}, 3.0)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty list, got %d", len(results))
	}
}
