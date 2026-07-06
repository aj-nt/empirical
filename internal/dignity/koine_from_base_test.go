package dignity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

func initEpheForKFB(t *testing.T) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot find home dir: %v", err)
	}
	cacheDir := filepath.Join(home, ".cache", "empirical", "ephe")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Skipf("ephemeris cache not found at %s", cacheDir)
	}
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)
}

func TestKoinéFromBase_KnownChart(t *testing.T) {
	initEpheForKFB(t)

	bc, err := ComputeBaseChart("AJ", 1969, 2, 15, 23, 10, 0, -8, 47.038, -122.901)
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

	bc, err := ComputeBaseChart("", 2000, 1, 1, 12, 0, 0, 0, 51.5, -0.12)
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

	bc, err := ComputeBaseChart("Test", 2000, 1, 1, 12, 0, 0, 0, 51.5, -0.12)
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
