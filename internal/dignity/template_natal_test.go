
package dignity

import (
	"strings"
	"testing"

)

func initEpheForTemplate(t *testing.T) {
	t.Helper()
	initEphe(t)
}

func TestRenderKoinéNatal_KnownChart(t *testing.T) {
	initEpheForTemplate(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := KoinéFromBase(bc, 5.0)
	html, err := RenderKoinéNatal(report)
	if err != nil {
		t.Fatalf("RenderKoinéNatal failed: %v", err)
	}

	// Must be valid HTML with expected sections
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("output missing DOCTYPE")
	}
	if !strings.Contains(html, "Koiné Natal Chart") {
		t.Error("output missing title")
	}
	if !strings.Contains(html, report.Name) {
		t.Error("output missing chart name")
	}
	if !strings.Contains(html, "Planets in Signs") {
		t.Error("output missing PlanetSigns section")
	}
	if !strings.Contains(html, "Planets in Houses") {
		t.Error("output missing PlanetHouses section")
	}
	if !strings.Contains(html, "Aspects") {
		t.Error("output missing Aspects section")
	}
	// Must contain at least one planet interpretation
	if !strings.Contains(html, "Sun") {
		t.Error("output missing Sun interpretation")
	}
}

func TestRenderKoinéNatal_EmptyName(t *testing.T) {
	initEpheForTemplate(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := KoinéFromBase(bc, 5.0)
	html, err := RenderKoinéNatal(report)
	if err != nil {
		t.Fatalf("RenderKoinéNatal failed for empty name: %v", err)
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("output missing DOCTYPE for empty name")
	}
}

func TestRenderKoinéNatal_NightChart(t *testing.T) {
	initEpheForTemplate(t)

	// Midnight birth = night chart
	bc, err := ComputeBaseChart(BirthData{Name: "Night", Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := KoinéFromBase(bc, 5.0)
	html, err := RenderKoinéNatal(report)
	if err != nil {
		t.Fatalf("RenderKoinéNatal failed: %v", err)
	}
	if !strings.Contains(html, "night") {
		t.Error("output missing night sect indicator")
	}
}
