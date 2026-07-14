package mundane

import (
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
)

type HousesFunc func(jd, lat, lon float64, hsys byte) (cusps [13]float64, ascmc [10]float64)

// MundaneChart is a full astrological chart cast for a specific time and location.
type MundaneChart struct {
	Time    time.Time         `json:"time"`
	Lat     float64           `json:"lat"`
	Lon     float64           `json:"lon"`
	Planets map[string]float64 `json:"planets"`
	Houses  [12]float64       `json:"houses"` // cusps 1-12
	ASC     float64           `json:"asc"`
	MC      float64           `json:"mc"`
}

// DefaultMundanePlanets returns the planet set used for mundane charts.
// 10 bodies: Sun through Pluto plus North Node.
func DefaultMundanePlanets() []dignity.PlanetID {
	return []dignity.PlanetID{
		{"Sun", 0},
		{"Moon", 1},
		{"Mercury", 2},
		{"Venus", 3},
		{"Mars", 4},
		{"Jupiter", 5},
		{"Saturn", 6},
		{"Uranus", 7},
		{"Neptune", 8},
		{"Pluto", 9},
		{"Node", 10},
	}
}

// julianDay computes the Julian Day from a time.Time using the standard
// astronomical formula (Fliegel & Van Flandern). Pure Go, no CGo dependency.
// Note: divisional and firdaria packages use a different algorithm (Meeus)
// paired with their own jdToDate — the two algorithms produce different JDs
// and cannot be unified without breaking roundtrip tests.
func julianDay(t time.Time) float64 {
	y, m, day := t.Year(), int(t.Month()), t.Day()
	h := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0

	if m <= 2 {
		y--
		m += 12
	}
	a := y / 100
	b := 2 - a + a/4
	jd := float64(int(365.25*float64(y+4716))) +
		float64(int(30.6001*float64(m+1))) +
		float64(day) + h/24.0 + float64(b) - 1524.5
	return jd
}

// CastChart computes a full mundane chart for the given time and location.
// compute provides planet positions. houses provides house cusps and angles.
// hsys is the house system (e.g., 'W' for Whole Sign, 'P' for Placidus).
func CastChart(tm time.Time, lat, lon float64, compute ComputeFunc, houses HousesFunc, hsys byte) (*MundaneChart, error) {
	jd := julianDay(tm)

	// Compute planet positions
	planets := make(map[string]float64)
	for _, p := range DefaultMundanePlanets() {
		plon, _, _, _ := compute(tm.Year(), int(tm.Month()), tm.Day(),
			float64(tm.Hour())+float64(tm.Minute())/60.0+float64(tm.Second())/3600.0,
			p.ID)
		planets[p.Name] = plon
	}

	// Compute houses
	cusps, ascmc := houses(jd, lat, lon, hsys)

	// Copy cusps 1-12
	var houseCusps [12]float64
	for i := 0; i < 12; i++ {
		houseCusps[i] = cusps[i+1]
	}

	return &MundaneChart{
		Time:    tm,
		Lat:     lat,
		Lon:     lon,
		Planets: planets,
		Houses:  houseCusps,
		ASC:     ascmc[0],
		MC:      ascmc[1],
	}, nil
}

// CastIngressChart casts a chart for an ingress event at the given location.
// Uses Whole Sign houses by default.
func CastIngressChart(event IngressEvent, lat, lon float64, compute ComputeFunc, houses HousesFunc) (*MundaneChart, error) {
	return CastChart(event.Time, lat, lon, compute, houses, 'W')
}

// CastLunationChart casts a chart for a lunation event at the given location.
// Uses Whole Sign houses by default.
func CastLunationChart(event LunationEvent, lat, lon float64, compute ComputeFunc, houses HousesFunc) (*MundaneChart, error) {
	return CastChart(event.Time, lat, lon, compute, houses, 'W')
}

// NationalChartEntry holds the birth data for a nation's chart.
type NationalChartEntry struct {
	Name      string  `json:"name"`
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      float64 `json:"hour"`      // UT
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timezone  string  `json:"timezone"`  // for display only
	Note      string  `json:"note"`      // source or variant info
}

// nationalCharts is the database of national charts.
// Times are in UT. Sources are noted where multiple charts exist.
var nationalCharts = []NationalChartEntry{
	{
		Name: "United States",
		Year: 1776, Month: 7, Day: 4,
		Hour: 22.2167, // ~5:13 PM local Philadelphia = 22:13 UT (Sibly chart)
		Lat: 39.95, Lon: -75.15,
		Timezone: "LMT",
		Note: "Sibly chart (Declaration of Independence). Alternative: Gemini rising chart (2:21 AM).",
	},
	{
		Name: "United Kingdom",
		Year: 1801, Month: 1, Day: 1,
		Hour: 0.0, // midnight, Act of Union
		Lat: 51.5, Lon: -0.1,
		Timezone: "LMT",
		Note: "Act of Union 1801. Alternative: 1066 Norman Conquest chart.",
	},
	{
		Name: "China",
		Year: 1949, Month: 10, Day: 1,
		Hour: 7.0, // ~3 PM Beijing = 7:00 UT
		Lat: 39.9, Lon: 116.4,
		Timezone: "CST",
		Note: "PRC founding. Time approximate (ceremony ~3 PM).",
	},
	{
		Name: "Russia",
		Year: 1991, Month: 12, Day: 25,
		Hour: 16.0, // ~7 PM Moscow = 16:00 UT
		Lat: 55.75, Lon: 37.6,
		Timezone: "MSK",
		Note: "Dissolution of USSR / Russian Federation. Alternative: 1990 declaration of sovereignty.",
	},
	{
		Name: "India",
		Year: 1947, Month: 8, Day: 15,
		Hour: 0.0, // midnight
		Lat: 28.6, Lon: 77.2,
		Timezone: "IST",
		Note: "Independence. Midnight chart is standard.",
	},
	{
		Name: "Japan",
		Year: 1947, Month: 5, Day: 3,
		Hour: 0.0, // midnight, constitution effective
		Lat: 35.7, Lon: 139.7,
		Timezone: "JST",
		Note: "Post-war constitution effective date.",
	},
	{
		Name: "Germany",
		Year: 1990, Month: 10, Day: 3,
		Hour: 0.0, // midnight, reunification
		Lat: 52.5, Lon: 13.4,
		Timezone: "CET",
		Note: "Reunification. Alternative: 1949 Federal Republic chart.",
	},
	{
		Name: "France",
		Year: 1958, Month: 10, Day: 5,
		Hour: 0.0, // midnight, Fifth Republic
		Lat: 48.85, Lon: 2.35,
		Timezone: "CET",
		Note: "Fifth Republic. Alternative: 1789 Revolution chart.",
	},
	{
		Name: "European Union",
		Year: 1993, Month: 11, Day: 1,
		Hour: 0.0, // midnight, Maastricht Treaty
		Lat: 50.85, Lon: 4.35,
		Timezone: "CET",
		Note: "Maastricht Treaty effective. Chart cast for Brussels.",
	},
	{
		Name: "Israel",
		Year: 1948, Month: 5, Day: 14,
		Hour: 14.0, // ~4 PM Tel Aviv = 14:00 UT
		Lat: 32.1, Lon: 34.8,
		Timezone: "IST",
		Note: "Declaration of Independence. Time approximate (~4 PM).",
	},
	{
		Name: "Brazil",
		Year: 1822, Month: 9, Day: 7,
		Hour: 16.0, // ~1 PM local = ~16:00 UT
		Lat: -23.55, Lon: -46.63,
		Timezone: "LMT",
		Note: "Independence declared. Time approximate.",
	},
	{
		Name: "United Nations",
		Year: 1945, Month: 10, Day: 24,
		Hour: 16.0, // charter ratified ~4 PM UT
		Lat: 40.75, Lon: -73.97,
		Timezone: "EST",
		Note: "UN Charter entered into force. Chart for NYC.",
	},
	{
		Name: "Australia",
		Year: 1901, Month: 1, Day: 1,
		Hour: 0.0, // midnight, Federation
		Lat: -35.3, Lon: 149.1,
		Timezone: "AEST",
		Note: "Federation. Chart for Canberra (Sydney at federation, Canberra for modern).",
	},
	{
		Name: "Canada",
		Year: 1867, Month: 7, Day: 1,
		Hour: 0.0, // midnight, Confederation
		Lat: 45.4, Lon: -75.7,
		Timezone: "LMT",
		Note: "Confederation / Dominion of Canada. Chart for Ottawa.",
	},
	{
		Name: "Iran",
		Year: 1979, Month: 4, Day: 1,
		Hour: 0.0, // Islamic Republic referendum result
		Lat: 35.7, Lon: 51.4,
		Timezone: "IRST",
		Note: "Islamic Republic established. Alternative: 1979 Feb 11 revolution victory.",
	},
	{
		Name: "Saudi Arabia",
		Year: 1932, Month: 9, Day: 23,
		Hour: 0.0, // unification
		Lat: 24.7, Lon: 46.7,
		Timezone: "AST",
		Note: "Kingdom unified. Chart for Riyadh.",
	},
	{
		Name: "Turkey",
		Year: 1923, Month: 10, Day: 29,
		Hour: 0.0, // Republic proclaimed
		Lat: 39.9, Lon: 32.9,
		Timezone: "EET",
		Note: "Republic of Turkey proclaimed. Chart for Ankara.",
	},
	{
		Name: "South Korea",
		Year: 1948, Month: 8, Day: 15,
		Hour: 0.0, // Republic established
		Lat: 37.6, Lon: 127.0,
		Timezone: "KST",
		Note: "Republic of Korea established. Chart for Seoul.",
	},
	{
		Name: "North Korea",
		Year: 1948, Month: 9, Day: 9,
		Hour: 0.0, // DPRK founded
		Lat: 39.0, Lon: 125.8,
		Timezone: "KST",
		Note: "DPRK founded. Chart for Pyongyang.",
	},
	{
		Name: "Mexico",
		Year: 1821, Month: 9, Day: 27,
		Hour: 0.0, // independence recognized
		Lat: 19.4, Lon: -99.1,
		Timezone: "LMT",
		Note: "Independence recognized. Chart for Mexico City.",
	},
	{
		Name: "Italy",
		Year: 1946, Month: 6, Day: 2,
		Hour: 0.0, // Republic referendum
		Lat: 41.9, Lon: 12.5,
		Timezone: "CET",
		Note: "Republic established. Chart for Rome.",
	},
	{
		Name: "Spain",
		Year: 1978, Month: 12, Day: 29,
		Hour: 0.0, // constitution effective
		Lat: 40.4, Lon: -3.7,
		Timezone: "CET",
		Note: "Constitution of 1978 effective. Chart for Madrid.",
	},
	{
		Name: "Ukraine",
		Year: 1991, Month: 8, Day: 24,
		Hour: 0.0, // independence declared
		Lat: 50.5, Lon: 30.5,
		Timezone: "EET",
		Note: "Independence declared. Chart for Kyiv.",
	},
	{
		Name: "Pakistan",
		Year: 1947, Month: 8, Day: 14,
		Hour: 0.0, // independence
		Lat: 33.7, Lon: 73.0,
		Timezone: "PKT",
		Note: "Independence. Chart for Islamabad (Karachi at independence).",
	},
	{
		Name: "Indonesia",
		Year: 1945, Month: 8, Day: 17,
		Hour: 0.0, // independence proclaimed
		Lat: -6.2, Lon: 106.8,
		Timezone: "WIB",
		Note: "Independence proclaimed. Chart for Jakarta.",
	},
	{
		Name: "Nigeria",
		Year: 1960, Month: 10, Day: 1,
		Hour: 0.0, // independence
		Lat: 9.1, Lon: 7.5,
		Timezone: "WAT",
		Note: "Independence. Chart for Abuja (Lagos at independence).",
	},
	{
		Name: "South Africa",
		Year: 1994, Month: 4, Day: 27,
		Hour: 0.0, // first democratic election
		Lat: -25.7, Lon: 28.2,
		Timezone: "SAST",
		Note: "Post-apartheid democracy. Chart for Pretoria.",
	},
	{
		Name: "Egypt",
		Year: 1952, Month: 7, Day: 23,
		Hour: 0.0, // revolution
		Lat: 30.0, Lon: 31.2,
		Timezone: "EET",
		Note: "1952 Revolution / Republic. Chart for Cairo.",
	},
	{
		Name: "Argentina",
		Year: 1816, Month: 7, Day: 9,
		Hour: 0.0, // independence declared
		Lat: -34.6, Lon: -58.4,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Buenos Aires.",
	},
	{
		Name: "Switzerland",
		Year: 1848, Month: 9, Day: 12,
		Hour: 0.0, // federal constitution
		Lat: 46.9, Lon: 7.4,
		Timezone: "LMT",
		Note: "Federal constitution. Chart for Bern.",
	},
	{
		Name: "Sweden",
		Year: 1523, Month: 6, Day: 6,
		Hour: 0.0, // Gustav Vasa elected king
		Lat: 59.3, Lon: 18.1,
		Timezone: "LMT",
		Note: "Gustav Vasa elected king — foundation of modern Sweden. Chart for Stockholm.",
	},
	{
		Name: "Vatican City",
		Year: 1929, Month: 2, Day: 11,
		Hour: 0.0, // Lateran Treaty
		Lat: 41.9, Lon: 12.5,
		Timezone: "CET",
		Note: "Lateran Treaty — Vatican City State established. Chart for Vatican City.",
	},
	{
		Name: "Greece",
		Year: 1822, Month: 1, Day: 1,
		Hour: 0.0, // independence declared
		Lat: 38.0, Lon: 23.7,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Athens.",
	},
	{
		Name: "Poland",
		Year: 1989, Month: 6, Day: 4,
		Hour: 0.0, // first free elections
		Lat: 52.2, Lon: 21.0,
		Timezone: "CET",
		Note: "First free elections — end of communist rule. Chart for Warsaw.",
	},
	{
		Name: "Taiwan",
		Year: 1912, Month: 1, Day: 1,
		Hour: 0.0, // Republic of China founding
		Lat: 25.0, Lon: 121.5,
		Timezone: "CST",
		Note: "Republic of China founding (same date as PRC claims). Chart for Taipei.",
	},
	{
		Name: "Venezuela",
		Year: 1811, Month: 7, Day: 5,
		Hour: 0.0, // independence declared
		Lat: 10.5, Lon: -66.9,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Caracas.",
	},
	{
		Name: "Philippines",
		Year: 1946, Month: 7, Day: 4,
		Hour: 0.0, // independence from US
		Lat: 14.6, Lon: 121.0,
		Timezone: "PHT",
		Note: "Independence from US. Chart for Manila.",
	},
	{
		Name: "Thailand",
		Year: 1932, Month: 6, Day: 24,
		Hour: 0.0, // constitutional monarchy
		Lat: 13.8, Lon: 100.5,
		Timezone: "ICT",
		Note: "Constitutional monarchy established. Chart for Bangkok.",
	},
	{
		Name: "Vietnam",
		Year: 1976, Month: 7, Day: 2,
		Hour: 0.0, // reunification
		Lat: 21.0, Lon: 105.8,
		Timezone: "ICT",
		Note: "Reunification. Chart for Hanoi.",
	},
	{
		Name: "Ireland",
		Year: 1922, Month: 12, Day: 6,
		Hour: 0.0, // Irish Free State
		Lat: 53.3, Lon: -6.3,
		Timezone: "GMT",
		Note: "Irish Free State established. Chart for Dublin.",
	},
	{
		Name: "Netherlands",
		Year: 1815, Month: 3, Day: 16,
		Hour: 0.0, // Kingdom established
		Lat: 52.1, Lon: 4.3,
		Timezone: "LMT",
		Note: "Kingdom of the Netherlands. Chart for The Hague.",
	},
	{
		Name: "Belgium",
		Year: 1830, Month: 10, Day: 4,
		Hour: 0.0, // independence declared
		Lat: 50.9, Lon: 4.4,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Brussels.",
	},
	{
		Name: "Norway",
		Year: 1905, Month: 6, Day: 7,
		Hour: 0.0, // independence from Sweden
		Lat: 59.9, Lon: 10.8,
		Timezone: "CET",
		Note: "Independence from Sweden. Chart for Oslo.",
	},
	{
		Name: "Denmark",
		Year: 1849, Month: 6, Day: 5,
		Hour: 0.0, // constitution signed
		Lat: 55.7, Lon: 12.6,
		Timezone: "LMT",
		Note: "Constitutional monarchy. Chart for Copenhagen.",
	},
	{
		Name: "Finland",
		Year: 1917, Month: 12, Day: 6,
		Hour: 0.0, // independence declared
		Lat: 60.2, Lon: 24.9,
		Timezone: "EET",
		Note: "Independence declared. Chart for Helsinki.",
	},
	{
		Name: "Austria",
		Year: 1955, Month: 5, Day: 15,
		Hour: 0.0, // State Treaty
		Lat: 48.2, Lon: 16.4,
		Timezone: "CET",
		Note: "State Treaty — restored sovereignty. Chart for Vienna.",
	},
	{
		Name: "Portugal",
		Year: 1974, Month: 4, Day: 25,
		Hour: 0.0, // Carnation Revolution
		Lat: 38.7, Lon: -9.1,
		Timezone: "WET",
		Note: "Carnation Revolution — end of Estado Novo. Chart for Lisbon.",
	},
	{
		Name: "Colombia",
		Year: 1810, Month: 7, Day: 20,
		Hour: 0.0, // independence declared
		Lat: 4.7, Lon: -74.1,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Bogotá.",
	},
	{
		Name: "Chile",
		Year: 1818, Month: 2, Day: 12,
		Hour: 0.0, // independence declared
		Lat: -33.4, Lon: -70.7,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Santiago.",
	},
	{
		Name: "Peru",
		Year: 1821, Month: 7, Day: 28,
		Hour: 0.0, // independence declared
		Lat: -12.0, Lon: -77.0,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Lima.",
	},
	{
		Name: "New Zealand",
		Year: 1907, Month: 9, Day: 26,
		Hour: 0.0, // Dominion status
		Lat: -41.3, Lon: 174.8,
		Timezone: "NZST",
		Note: "Dominion status. Chart for Wellington.",
	},
	{
		Name: "Singapore",
		Year: 1965, Month: 8, Day: 9,
		Hour: 0.0, // independence
		Lat: 1.3, Lon: 103.8,
		Timezone: "SGT",
		Note: "Independence from Malaysia. Chart for Singapore.",
	},
	{
		Name: "Malaysia",
		Year: 1957, Month: 8, Day: 31,
		Hour: 0.0, // independence
		Lat: 3.1, Lon: 101.7,
		Timezone: "MYT",
		Note: "Independence from UK. Chart for Kuala Lumpur.",
	},
	{
		Name: "Bangladesh",
		Year: 1971, Month: 3, Day: 26,
		Hour: 0.0, // independence declared
		Lat: 23.8, Lon: 90.4,
		Timezone: "BST",
		Note: "Independence declared. Chart for Dhaka.",
	},
	{
		Name: "Ethiopia",
		Year: 1991, Month: 5, Day: 28,
		Hour: 0.0, // Derg regime falls
		Lat: 9.0, Lon: 38.7,
		Timezone: "EAT",
		Note: "Derg regime falls — modern Ethiopia. Chart for Addis Ababa.",
	},
	{
		Name: "Kenya",
		Year: 1963, Month: 12, Day: 12,
		Hour: 0.0, // independence
		Lat: -1.3, Lon: 36.8,
		Timezone: "EAT",
		Note: "Independence. Chart for Nairobi.",
	},
	{
		Name: "Cuba",
		Year: 1959, Month: 1, Day: 1,
		Hour: 0.0, // revolution victory
		Lat: 23.1, Lon: -82.4,
		Timezone: "CST",
		Note: "Revolution victory. Chart for Havana.",
	},
	{
		Name: "Iraq",
		Year: 1932, Month: 10, Day: 3,
		Hour: 0.0, // independence
		Lat: 33.3, Lon: 44.4,
		Timezone: "AST",
		Note: "Independence from British mandate. Chart for Baghdad.",
	},
	{
		Name: "Afghanistan",
		Year: 1919, Month: 8, Day: 19,
		Hour: 0.0, // independence
		Lat: 34.5, Lon: 69.2,
		Timezone: "AFT",
		Note: "Independence from British control. Chart for Kabul.",
	},
	{
		Name: "Syria",
		Year: 1946, Month: 4, Day: 17,
		Hour: 0.0, // independence
		Lat: 33.5, Lon: 36.3,
		Timezone: "EET",
		Note: "Independence from French mandate. Chart for Damascus.",
	},
	{
		Name: "Myanmar",
		Year: 1948, Month: 1, Day: 4,
		Hour: 0.0, // independence
		Lat: 16.8, Lon: 96.2,
		Timezone: "MMT",
		Note: "Independence from UK. Chart for Yangon.",
	},
	{
		Name: "Sudan",
		Year: 1956, Month: 1, Day: 1,
		Hour: 0.0, // independence
		Lat: 15.5, Lon: 32.6,
		Timezone: "CAT",
		Note: "Independence. Chart for Khartoum.",
	},
	{
		Name: "Ghana",
		Year: 1957, Month: 3, Day: 6,
		Hour: 0.0, // independence
		Lat: 5.6, Lon: -0.2,
		Timezone: "GMT",
		Note: "First sub-Saharan African independence. Chart for Accra.",
	},
	{
		Name: "Algeria",
		Year: 1962, Month: 7, Day: 5,
		Hour: 0.0, // independence
		Lat: 36.8, Lon: 3.0,
		Timezone: "CET",
		Note: "Independence from France. Chart for Algiers.",
	},
	{
		Name: "Morocco",
		Year: 1956, Month: 3, Day: 2,
		Hour: 0.0, // independence
		Lat: 34.0, Lon: -6.8,
		Timezone: "WET",
		Note: "Independence from France. Chart for Rabat.",
	},
	{
		Name: "Kazakhstan",
		Year: 1991, Month: 12, Day: 16,
		Hour: 0.0, // independence
		Lat: 51.2, Lon: 71.4,
		Timezone: "ALMT",
		Note: "Independence from USSR. Chart for Astana.",
	},
	{
		Name: "Uzbekistan",
		Year: 1991, Month: 9, Day: 1,
		Hour: 0.0, // independence
		Lat: 41.3, Lon: 69.3,
		Timezone: "UZT",
		Note: "Independence from USSR. Chart for Tashkent.",
	},
}

// NationalChart returns the chart entry for a given nation by name.
// Returns false if not found.
func NationalChart(name string) (NationalChartEntry, bool) {
	for _, c := range nationalCharts {
		if c.Name == name {
			return c, true
		}
	}
	return NationalChartEntry{}, false
}

// NationalCharts returns all national chart entries.
func NationalCharts() []NationalChartEntry {
	result := make([]NationalChartEntry, len(nationalCharts))
	copy(result, nationalCharts)
	return result
}

// ChartAspect records an aspect between two planets in a chart.
