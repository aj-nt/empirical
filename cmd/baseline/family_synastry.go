package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

// familyBirth holds birth data for one family member.
type familyBirth struct {
	Name   string
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	TzOff  float64
	Lat    float64
	Lng    float64
}

// familyData is the 17-person Flinton/Bucci/Coffey family dataset.
var familyData = []familyBirth{
	// ── AJ's generation ──
	{"AJ", 1969, 2, 15, 23, 10, -8, 47.038, -122.901},
	{"Cait", 1986, 4, 29, 3, 0, -4, 41.034, -73.763},
	{"Ex-wife", 1969, 11, 12, 12, 0, -5, 41.700, -73.921},
	{"Beth (Cait sister)", 1988, 3, 19, 12, 0, -4, 41.034, -73.763},

	// ── Parents ──
	{"Bill (AJ father)", 1949, 9, 2, 12, 0, -8, 47.038, -122.901},
	{"Mary (AJ mother)", 1950, 2, 10, 12, 0, -8, 37.804, -122.271},
	{"Ed (AJ stepfather)", 1950, 9, 17, 12, 0, -6, 41.878, -87.630},
	{"Cait father", 1950, 1, 1, 12, 0, -4, 41.034, -73.763},
	{"Cait mother", 1952, 1, 1, 12, 0, -4, 41.034, -73.763},

	// ── AJ Grandparents ──
	{"George (patGF)", 1905, 12, 5, 12, 0, -5, 46.126, -67.840},
	{"Gerry (patGM)", 1917, 8, 8, 12, 0, -8, 48.535, -123.079},
	{"Bill (matGF)", 1919, 6, 30, 12, 0, -8, 47.606, -122.332},
	{"Margaret (matGM)", 1918, 11, 11, 12, 0, -8, 37.765, -122.242},

	// ── Cait Grandparents ──
	{"George Coffey (matGF)", 1918, 10, 29, 12, 0, -5, 40.931, -73.899},
	{"Salvatore Bucci (patGF)", 1919, 4, 8, 12, 0, 1, 41.117, 16.867},
	{"Muriel Coffey (matGM)", 1924, 3, 1, 12, 0, -5, 40.931, -73.899},
	{"Amy Bucci (patGM)", 1924, 6, 2, 12, 0, -5, 41.137, -73.808},
}

// familyPairs defines the relationship pairs for synastry analysis.
type familyPair struct {
	Name1 string
	Name2 string
	Cat   string // "couple", "parent_child", "grandparent_grandchild"
}

func buildFamilyPairs() []familyPair {
	var pairs []familyPair

	// Couples
	pairs = append(pairs, familyPair{"AJ", "Cait", "couple"})
	pairs = append(pairs, familyPair{"Bill (AJ father)", "Mary (AJ mother)", "couple"})
	pairs = append(pairs, familyPair{"George (patGF)", "Gerry (patGM)", "couple"})
	pairs = append(pairs, familyPair{"Bill (matGF)", "Margaret (matGM)", "couple"})
	pairs = append(pairs, familyPair{"George Coffey (matGF)", "Muriel Coffey (matGM)", "couple"})
	pairs = append(pairs, familyPair{"Salvatore Bucci (patGF)", "Amy Bucci (patGM)", "couple"})

	// Parent-child
	pairs = append(pairs, familyPair{"Bill (AJ father)", "AJ", "parent_child"})
	pairs = append(pairs, familyPair{"Mary (AJ mother)", "AJ", "parent_child"})
	pairs = append(pairs, familyPair{"Cait father", "Cait", "parent_child"})
	pairs = append(pairs, familyPair{"Cait mother", "Cait", "parent_child"})
	pairs = append(pairs, familyPair{"George (patGF)", "Bill (AJ father)", "parent_child"})
	pairs = append(pairs, familyPair{"Gerry (patGM)", "Bill (AJ father)", "parent_child"})
	pairs = append(pairs, familyPair{"Bill (matGF)", "Mary (AJ mother)", "parent_child"})
	pairs = append(pairs, familyPair{"Margaret (matGM)", "Mary (AJ mother)", "parent_child"})
	pairs = append(pairs, familyPair{"George Coffey (matGF)", "Cait mother", "parent_child"})
	pairs = append(pairs, familyPair{"Muriel Coffey (matGM)", "Cait mother", "parent_child"})
	pairs = append(pairs, familyPair{"Salvatore Bucci (patGF)", "Cait father", "parent_child"})
	pairs = append(pairs, familyPair{"Amy Bucci (patGM)", "Cait father", "parent_child"})

	// Grandparent-grandchild
	pairs = append(pairs, familyPair{"George (patGF)", "AJ", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Gerry (patGM)", "AJ", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Bill (matGF)", "AJ", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Margaret (matGM)", "AJ", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"George Coffey (matGF)", "Cait", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Muriel Coffey (matGM)", "Cait", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Salvatore Bucci (patGF)", "Cait", "grandparent_grandchild"})
	pairs = append(pairs, familyPair{"Amy Bucci (patGM)", "Cait", "grandparent_grandchild"})

	return pairs
}

func runFamilySynastryBaseline(nRandom int, seed int64, rng *rand.Rand) {
	// Compute natal charts for all family members
	familyCharts := make(map[string]map[string]float64)
	for _, fb := range familyData {
		utHour := float64(fb.Hour) + float64(fb.Minute)/60.0 - fb.TzOff
		jd := swe.Julday(fb.Year, fb.Month, fb.Day, utHour, true)

		chart := make(map[string]float64)
		classical := []struct {
			name string
			id   int
		}{
			{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
			{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
			{"Saturn", swe.SATURN},
		}
		for _, p := range classical {
			lon, _, _, _ := swe.CalcUT(jd, p.id)
			for lon < 0 {
				lon += 360
			}
			for lon >= 360 {
				lon -= 360
			}
			chart[p.name] = lon
		}
		// Add Node
		nn, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
		for nn < 0 {
			nn += 360
		}
		for nn >= 360 {
			nn -= 360
		}
		chart["Node"] = nn

		familyCharts[fb.Name] = chart
	}

	// Build family pairs
	fpairs := buildFamilyPairs()

	// Build pair structs for the metric function
	var familyPairStructs []struct {
		Name1  string
		Chart1 map[string]float64
		Name2  string
		Chart2 map[string]float64
	}
	for _, fp := range fpairs {
		c1, ok1 := familyCharts[fp.Name1]
		c2, ok2 := familyCharts[fp.Name2]
		if !ok1 || !ok2 {
			continue
		}
		familyPairStructs = append(familyPairStructs, struct {
			Name1  string
			Chart1 map[string]float64
			Name2  string
			Chart2 map[string]float64
		}{fp.Name1, c1, fp.Name2, c2})
	}

	// Generate random pairs
	var randomPairStructs []struct {
		Name1  string
		Chart1 map[string]float64
		Name2  string
		Chart2 map[string]float64
	}
	for i := 0; i < nRandom; i++ {
		c1 := randomChart(rng)
		c2 := randomChart(rng)
		randomPairStructs = append(randomPairStructs, struct {
			Name1  string
			Chart1 map[string]float64
			Name2  string
			Chart2 map[string]float64
		}{"", c1, "", c2})
	}

	// Planets to check
	planets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}

	// Run metrics
	results := dignity.ComputeSynastryMetrics(familyPairStructs, randomPairStructs, planets)

	output := struct {
		FamilyN     int                            `json:"family_n"`
		RandomN     int                            `json:"random_n"`
		Seed        int64                          `json:"seed"`
		Metrics     []dignity.SynastryMetricResult `json:"metrics"`
	}{
		FamilyN: len(familyPairStructs),
		RandomN: nRandom,
		Seed:    seed,
		Metrics: results,
	}

	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(b))
}

func randomChart(rng *rand.Rand) map[string]float64 {
	year := 1900 + rng.Intn(131)
	month := 1 + rng.Intn(12)
	day := 1 + rng.Intn(28)
	hour := rng.Intn(24)
	minute := rng.Intn(60)
	tzOff := float64(-12 + rng.Intn(25))

	utHour := float64(hour) + float64(minute)/60.0 - tzOff
	jd := swe.Julday(year, month, day, utHour, true)

	chart := make(map[string]float64)
	classical := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN},
	}
	for _, p := range classical {
		lon, _, _, _ := swe.CalcUT(jd, p.id)
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		chart[p.name] = lon
	}
	nn, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	for nn < 0 {
		nn += 360
	}
	for nn >= 360 {
		nn -= 360
	}
	chart["Node"] = nn

	return chart
}
