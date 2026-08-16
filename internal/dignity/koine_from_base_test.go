package dignity

import (
	"testing"

)

func initEpheForKFB(t *testing.T) {
	t.Helper()
	initEphe(t)
}

func TestKoinéFromBase_KnownChart(t *testing.T) {
	initEpheForKFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := KoinéFromBase(bc, 5.0)

	if report == nil {
		t.Fatal("KoinéFromBase returned nil")
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
	// Patterns may be empty with only 7 classical planets — not an error
	if len(report.Patterns) == 0 {
		t.Log("Patterns empty (expected with classical planets only)")
	}
}

func TestKoinéFromBase_EmptyName(t *testing.T) {
	initEpheForKFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := KoinéFromBase(bc, 5.0)
	if report == nil {
		t.Fatal("KoinéFromBase returned nil for empty name")
	}
	// Should still produce valid output
	if len(report.PlanetSigns) == 0 {
		t.Error("expected non-empty PlanetSigns for empty name chart")
	}
}

func TestKoinéFromBase_CustomOrb(t *testing.T) {
	initEpheForKFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	// Tight orb should produce fewer aspects
	tight := KoinéFromBase(bc, 1.0)
	wide := KoinéFromBase(bc, 10.0)

	if tight == nil || wide == nil {
		t.Fatal("KoinéFromBase returned nil")
	}
	if len(tight.Aspects) > len(wide.Aspects) {
		t.Errorf("tight orb (%d aspects) should not exceed wide orb (%d aspects)",
			len(tight.Aspects), len(wide.Aspects))
	}
}
