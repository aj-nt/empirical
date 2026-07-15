package geocode

import (
	"math"
	"testing"
)

func TestNearestCity_ExactMatch(t *testing.T) {
	t.Parallel()
	cities := []City{
		{Name: "Bangkok", Country: "TH", Lat: 13.7563, Lon: 100.5018},
		{Name: "Phuket", Country: "TH", Lat: 7.8804, Lon: 98.3923},
		{Name: "Chiang Mai", Country: "TH", Lat: 18.7883, Lon: 98.9853},
	}
	result, ok := NearestCity(7.8804, 98.3923, cities)
	if !ok {
		t.Fatal("expected to find a city")
	}
	if result.Name != "Phuket" {
		t.Errorf("expected Phuket, got %s", result.Name)
	}
	if result.Country != "TH" {
		t.Errorf("expected TH, got %s", result.Country)
	}
}

func TestNearestCity_Approximate(t *testing.T) {
	t.Parallel()
	cities := []City{
		{Name: "Bangkok", Country: "TH", Lat: 13.7563, Lon: 100.5018},
		{Name: "Phuket", Country: "TH", Lat: 7.8804, Lon: 98.3923},
		{Name: "Chiang Mai", Country: "TH", Lat: 18.7883, Lon: 98.9853},
	}
	// Krabi (~8.0863, 98.9063) is closer to Phuket than to any other
	result, ok := NearestCity(8.0863, 98.9063, cities)
	if !ok {
		t.Fatal("expected to find a city")
	}
	if result.Name != "Phuket" {
		t.Errorf("expected Phuket (nearest to Krabi), got %s", result.Name)
	}
}

func TestNearestCity_EmptyList(t *testing.T) {
	t.Parallel()
	_, ok := NearestCity(0, 0, nil)
	if ok {
		t.Error("expected false for empty city list")
	}
}

func TestNearestCity_SingleCity(t *testing.T) {
	t.Parallel()
	cities := []City{
		{Name: "Reykjavik", Country: "IS", Lat: 64.1466, Lon: -21.9426},
	}
	result, ok := NearestCity(0, 0, cities)
	if !ok {
		t.Fatal("expected to find a city")
	}
	if result.Name != "Reykjavik" {
		t.Errorf("expected Reykjavik, got %s", result.Name)
	}
}

func TestHaversine(t *testing.T) {
	t.Parallel()
	// Bangkok to Phuket: ~690 km
	dist := haversine(13.7563, 100.5018, 7.8804, 98.3923)
	expected := 690.0
	if math.Abs(dist-expected) > 20 {
		t.Errorf("Bangkok-Phuket distance: got %.1f km, expected ~%.1f km", dist, expected)
	}

	// Same point: zero distance
	dist = haversine(0, 0, 0, 0)
	if dist != 0 {
		t.Errorf("same point: got %.1f, expected 0", dist)
	}

	// North Pole to South Pole: ~20015 km (half circumference)
	dist = haversine(90, 0, -90, 0)
	expected = 20015.0
	if math.Abs(dist-expected) > 50 {
		t.Errorf("pole-to-pole: got %.1f km, expected ~%.1f km", dist, expected)
	}
}

func TestLoadCities(t *testing.T) {
	cities, err := LoadCities()
	if err != nil {
		t.Fatalf("LoadCities failed: %v", err)
	}
	if len(cities) == 0 {
		t.Fatal("expected non-empty city list")
	}
	// Verify a known city exists
	found := false
	for _, c := range cities {
		if c.Name == "Bangkok" && c.Country == "TH" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Bangkok, TH in city list")
	}
}

func TestNearestCity_Integration(t *testing.T) {
	t.Parallel()
	cities, err := LoadCities()
	if err != nil {
		t.Fatalf("LoadCities failed: %v", err)
	}
	// Phuket should find itself
	result, ok := NearestCity(7.8804, 98.3923, cities)
	if !ok {
		t.Fatal("expected to find a city")
	}
	if result.Name != "Phuket" {
		t.Errorf("expected Phuket, got %s", result.Name)
	}

	// Olympia, WA (~47.038, -122.901) should find Olympia or nearby
	result, ok = NearestCity(47.038, -122.901, cities)
	if !ok {
		t.Fatal("expected to find a city near Olympia WA")
	}
	t.Logf("Olympia WA → %s, %s (%.4f, %.4f)", result.Name, result.Country, result.Lat, result.Lon)
}
