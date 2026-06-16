package parans

import (
	"testing"
)

// ── FindParans ───────────────────────────────────────────────────────────

func TestFindParans(t *testing.T) {
	// Star on ASC, planet on ASC → paran
	stars := map[string]float64{"Sirius": 0}
	planets := map[string]float64{"Mars": 2}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 1 {
		t.Fatalf("expected 1 paran, got %d", len(parans))
	}
	if parans[0].Star != "Sirius" || parans[0].Planet != "Mars" || parans[0].Angle != "ASC" {
		t.Errorf("unexpected paran: %+v", parans[0])
	}
}

func TestFindParansDifferentAngles(t *testing.T) {
	// Star on ASC, planet on MC → no paran (different angles)
	stars := map[string]float64{"Sirius": 0}
	planets := map[string]float64{"Mars": 90}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 0 {
		t.Errorf("expected 0 parans (different angles), got %d", len(parans))
	}
}

func TestFindParansMultiple(t *testing.T) {
	stars := map[string]float64{
		"Sirius":   0,
		"Aldebaran": 90,
	}
	planets := map[string]float64{
		"Mars":  1,
		"Venus": 91,
	}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	// Sirius+Mars on ASC, Aldebaran+Venus on MC
	if len(parans) != 2 {
		t.Fatalf("expected 2 parans, got %d", len(parans))
	}
	// Should be sorted by angle then star
	if parans[0].Angle != "ASC" {
		t.Errorf("first paran should be ASC, got %s", parans[0].Angle)
	}
}

func TestFindParansDSC(t *testing.T) {
	// Star and planet both on DSC (ASC+180)
	stars := map[string]float64{"Sirius": 180}
	planets := map[string]float64{"Mars": 181}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 1 {
		t.Fatalf("expected 1 paran on DSC, got %d", len(parans))
	}
	if parans[0].Angle != "DSC" {
		t.Errorf("expected DSC, got %s", parans[0].Angle)
	}
}

func TestFindParansIC(t *testing.T) {
	// Star and planet both on IC (MC+180)
	stars := map[string]float64{"Sirius": 270}
	planets := map[string]float64{"Mars": 271}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 1 {
		t.Fatalf("expected 1 paran on IC, got %d", len(parans))
	}
	if parans[0].Angle != "IC" {
		t.Errorf("expected IC, got %s", parans[0].Angle)
	}
}

func TestFindParansOrbBoundary(t *testing.T) {
	// Star at 0, planet at 3.1 → outside 3° orb
	stars := map[string]float64{"Sirius": 0}
	planets := map[string]float64{"Mars": 3.1}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 0 {
		t.Errorf("expected 0 parans (outside orb), got %d", len(parans))
	}
}

// ── ComputeParansReport ──────────────────────────────────────────────────

func TestComputeParansReport(t *testing.T) {
	stars := map[string]float64{"Sirius": 0, "Aldebaran": 90}
	planets := map[string]float64{"Mars": 1, "Venus": 91}
	report := ComputeParansReport("test", stars, planets, 0, 90, 3.0)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if len(report.Angles) != 4 {
		t.Errorf("expected 4 angles, got %d", len(report.Angles))
	}
	if report.Angles["ASC"] != 0 || report.Angles["DSC"] != 180 {
		t.Errorf("angles wrong: %v", report.Angles)
	}
	if len(report.StarsOnAngles) < 2 {
		t.Errorf("expected at least 2 stars on angles, got %d", len(report.StarsOnAngles))
	}
	if len(report.PlanetsOnAngles) < 2 {
		t.Errorf("expected at least 2 planets on angles, got %d", len(report.PlanetsOnAngles))
	}
	if len(report.Parans) != 2 {
		t.Errorf("expected 2 parans, got %d", len(report.Parans))
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────

func TestParansEmptyInput(t *testing.T) {
	report := ComputeParansReport("empty", map[string]float64{}, map[string]float64{}, 0, 90, 3.0)
	if len(report.StarsOnAngles) != 0 {
		t.Errorf("empty stars should produce 0 stars on angles")
	}
	if len(report.PlanetsOnAngles) != 0 {
		t.Errorf("empty planets should produce 0 planets on angles")
	}
	if len(report.Parans) != 0 {
		t.Errorf("empty input should produce 0 parans")
	}

	parans := FindParans(map[string]float64{}, map[string]float64{}, 0, 90, 3.0)
	if len(parans) != 0 {
		t.Errorf("empty maps should produce 0 parans")
	}
}

func TestParansNoAngleContact(t *testing.T) {
	// Star at 50°, planet at 130° — neither on any angle (ASC=0, MC=90)
	stars := map[string]float64{"Sirius": 50}
	planets := map[string]float64{"Mars": 130}
	report := ComputeParansReport("test", stars, planets, 0, 90, 3.0)
	if len(report.StarsOnAngles) != 0 {
		t.Errorf("star at 50° should not be on any angle")
	}
	if len(report.Parans) != 0 {
		t.Errorf("no angle contacts → no parans")
	}
}

// ── Angle wraparound ─────────────────────────────────────────────────────

func TestParansWraparound(t *testing.T) {
	// Star at 359°, planet at 1° — both near ASC (0°), within 2° orb
	stars := map[string]float64{"Sirius": 359}
	planets := map[string]float64{"Mars": 1}
	parans := FindParans(stars, planets, 0, 90, 3.0)
	if len(parans) != 1 {
		t.Fatalf("expected 1 paran across 0° boundary, got %d", len(parans))
	}
	if parans[0].Angle != "ASC" {
		t.Errorf("expected ASC, got %s", parans[0].Angle)
	}
}
