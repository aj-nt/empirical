package dignity

// ═══════════════════════════════════════════════════════════════════════════
// Ashtakavarga — Eight-Fold Point System (Brihat Parashara Hora Shastra)
// ═══════════════════════════════════════════════════════════════════════════
//
// Ashtakavarga assigns benefic points (bindus) to each house from the
// perspective of each planet + the ascendant (8 contributors = ashta-varga).
//
// Each contributor gives 0 or 1 bindu per house based on its position.
// Total bindus in a house indicate its auspiciousness:
//   > 28: very auspicious, 25-28: auspicious, 20-24: neutral, < 20: inauspicious
//
// Sarvashtakavarga = sum of all 8 individual ashtakavargas.
// Bhinnashtakavarga = individual contributor's bindu pattern.

// AshtakavargaReport holds the full ashtakavarga analysis.
type AshtakavargaReport struct {
	Bhinnashtakavarga map[string][]int `json:"bhinnashtakavarga"`
	Sarvashtakavarga  []int            `json:"sarvashtakavarga"`
	HouseQuality      map[int]string   `json:"house_quality"`
}

// ComputeAshtakavarga computes the full ashtakavarga for a chart.
func ComputeAshtakavarga(planetHouses map[string]int) *AshtakavargaReport {
	contributors := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Ascendant"}

	bhinnashtakavarga := make(map[string][]int, 8)
	sarvashtakavarga := make([]int, 12)

	for _, contributor := range contributors {
		var contribHouse int
		if contributor == "Ascendant" {
			contribHouse = 1
		} else {
			contribHouse = planetHouses[contributor]
		}

		bindus := computeBinduPattern(contributor, contribHouse)
		bhinnashtakavarga[contributor] = bindus
		for h := 0; h < 12; h++ {
			sarvashtakavarga[h] += bindus[h]
		}
	}

	quality := make(map[int]string, 12)
	for h := 0; h < 12; h++ {
		total := sarvashtakavarga[h]
		switch {
		case total > 28:
			quality[h+1] = "very auspicious"
		case total >= 25:
			quality[h+1] = "auspicious"
		case total >= 20:
			quality[h+1] = "neutral"
		default:
			quality[h+1] = "inauspicious"
		}
	}

	return &AshtakavargaReport{
		Bhinnashtakavarga: bhinnashtakavarga,
		Sarvashtakavarga:  sarvashtakavarga,
		HouseQuality:      quality,
	}
}

// computeBinduPattern returns a 12-element slice (houses 1-12, 0-indexed)
// where 1 = bindu (benefic point), 0 = no bindu.
// Based on classical Parashari ashtakavarga tables (BPHS Ch. 65-72).
func computeBinduPattern(contributor string, fromHouse int) []int {
	table, ok := ashtakavargaTables[contributor]
	if !ok {
		// Fallback: all houses get bindus
		b := make([]int, 12)
		for i := range b {
			b[i] = 1
		}
		return b
	}

	beneficSet, ok := table[fromHouse]
	if !ok {
		b := make([]int, 12)
		for i := range b {
			b[i] = 1
		}
		return b
	}

	bindus := make([]int, 12)
	for _, h := range beneficSet {
		if h >= 1 && h <= 12 {
			bindus[h-1] = 1
		}
	}
	return bindus
}

// ashtakavargaTables: classical Parashari bindu tables.
// Each entry: from house N, give bindus to these houses.
// Source: BPHS Ch. 65-72, standardized tables.
// Houses NOT listed receive 0 bindus (rekha).
var ashtakavargaTables = map[string]map[int][]int{
	"Sun": {
		1:  {1, 2, 4, 7, 8, 9, 10, 11},
		2:  {1, 2, 4, 5, 7, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 8, 9, 11, 12},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 8, 9, 11},
		7:  {1, 2, 3, 4, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Moon": {
		1:  {1, 3, 6, 7, 10, 11},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 3, 4, 5, 6, 7, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 3, 4, 5, 6, 7, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Mars": {
		1:  {1, 2, 4, 7, 8, 9, 10, 11},
		2:  {1, 2, 4, 5, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 8, 9, 11, 12},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 8, 9, 11},
		7:  {1, 2, 3, 4, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Mercury": {
		1:  {1, 2, 3, 4, 5, 6, 8, 9, 10, 11, 12},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Jupiter": {
		1:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Venus": {
		1:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Saturn": {
		1:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"Ascendant": {
		1:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		2:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		3:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		4:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		5:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		6:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		7:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		8:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		9:  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		11: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		12: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
}
