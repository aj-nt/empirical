package geocode

import (
	_ "embed"
	"encoding/json"
	"math"
)

// City represents a named location with country code.
type City struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

//go:embed cities.json
var citiesJSON []byte

// LoadCities loads the embedded cities database.
func LoadCities() ([]City, error) {
	var cities []City
	if err := json.Unmarshal(citiesJSON, &cities); err != nil {
		return nil, err
	}
	return cities, nil
}

// NearestCity finds the nearest city to (lat, lon) using Haversine distance.
// Returns false if the city list is empty.
func NearestCity(lat, lon float64, cities []City) (*City, bool) {
	if len(cities) == 0 {
		return nil, false
	}
	best := &cities[0]
	bestDist := haversine(lat, lon, best.Lat, best.Lon)
	for i := 1; i < len(cities); i++ {
		d := haversine(lat, lon, cities[i].Lat, cities[i].Lon)
		if d < bestDist {
			bestDist = d
			best = &cities[i]
		}
	}
	return best, true
}

// haversine returns the great-circle distance in kilometers between two points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth's mean radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
