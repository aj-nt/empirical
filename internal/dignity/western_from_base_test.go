package dignity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

func initEpheForWFB(t *testing.T) {
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

func TestWesternFromBase_KnownChart(t *testing.T) {
	initEpheForWFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := WesternFromBase(bc, 5.0)

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
		t.Error("expected non-empty Patterns (all planets, not just classical)")
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

	report := WesternFromBase(bc, 5.0)

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

	report := WesternFromBase(bc, 5.0)
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

	tight := WesternFromBase(bc, 1.0)
	wide := WesternFromBase(bc, 10.0)

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

	report := WesternFromBase(bc, 5.0)
	data, err := report.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("JSON() returned empty bytes")
	}
}
