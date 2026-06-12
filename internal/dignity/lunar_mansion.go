package dignity

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════════
// Lunar Mansion Anchor Star Comparison
// ═══════════════════════════════════════════════════════════════════════════
//
// Comparison of 27 Indian nakshatra and 28 Chinese xiu determinative stars.
// Identifies shared anchor stars and tests overlap against null models.
//
// Data sources: Wikipedia "Nakshatra" and "Twenty-Eight Mansions" pages.
// Magnitudes from Swiss Ephemeris sefstars.txt catalog where available,
// SIMBAD/Hipparcos otherwise.
//
// All 28 xiu have documented single-star determinants per Wikipedia.
// The paper previously claimed 5 xiu use asterisms and excluded them;
// this was incorrect. All 28 are included.

// ── Star Data ──────────────────────────────────────────────────────────────

// StarEntry holds a determinative star's data.
type StarEntry struct {
	Key            string  // Bayer designation if available, else star name
	StarName       string  // Common name
	Bayer          string  // Bayer designation (may be empty)
	Magnitude      float64 // Visual magnitude
	EclipticLon    float64 // Ecliptic longitude (degrees, J2000)
	EclipticLat    float64 // Ecliptic latitude (degrees, J2000)
	NakshatraName  string  // Which nakshatra this anchors (empty if xiu-only)
	NakshatraNum   int     // Nakshatra number 1-27 (0 if xiu-only)
	XiuName        string  // Which xiu this anchors (empty if nakshatra-only)
	XiuNum         int     // Xiu number 1-28 (0 if nakshatra-only)
	XiuPinyin      string  // Pinyin for xiu name
}

// NakshatraStars returns the 27 nakshatra determinative stars.
func NakshatraStars() []StarEntry {
	return []StarEntry{
		{Key: "Beta Arietis", StarName: "Sheratan", Bayer: "Beta Arietis", Magnitude: 2.65, EclipticLon: 33.97, EclipticLat: 8.49, NakshatraName: "Ashwini", NakshatraNum: 1},
		{Key: "35 Arietis", StarName: "35 Arietis", Magnitude: 4.60, EclipticLon: 46.93, EclipticLat: 11.31, NakshatraName: "Bharani", NakshatraNum: 2},
		{Key: "Eta Tauri", StarName: "Alcyone", Bayer: "Eta Tauri", Magnitude: 2.87, EclipticLon: 59.99, EclipticLat: 4.05, NakshatraName: "Krittika", NakshatraNum: 3},
		{Key: "Alpha Tauri", StarName: "Aldebaran", Bayer: "Alpha Tauri", Magnitude: 0.87, EclipticLon: 69.79, EclipticLat: -5.47, NakshatraName: "Rohini", NakshatraNum: 4},
		{Key: "Lambda Orionis", StarName: "Meissa", Bayer: "Lambda Orionis", Magnitude: 3.66, EclipticLon: 83.71, EclipticLat: -13.37, NakshatraName: "Mrigashira", NakshatraNum: 5},
		{Key: "Alpha Orionis", StarName: "Betelgeuse", Bayer: "Alpha Orionis", Magnitude: 0.42, EclipticLon: 88.75, EclipticLat: -16.03, NakshatraName: "Ardra", NakshatraNum: 6},
		{Key: "Alpha Geminorum", StarName: "Castor", Bayer: "Alpha Geminorum", Magnitude: 1.58, EclipticLon: 110.24, EclipticLat: 10.10, NakshatraName: "Punarvasu", NakshatraNum: 7},
		{Key: "Gamma Cancri", StarName: "Asellus Borealis", Bayer: "Gamma Cancri", Magnitude: 4.65, EclipticLon: 127.54, EclipticLat: 3.19, NakshatraName: "Pushya", NakshatraNum: 8},
		{Key: "Delta Hydrae", StarName: "Delta Hydrae", Magnitude: 4.14, EclipticLon: 130.30, EclipticLat: -12.39, NakshatraName: "Ashlesha", NakshatraNum: 9},
		{Key: "Alpha Leonis", StarName: "Regulus", Bayer: "Alpha Leonis", Magnitude: 1.40, EclipticLon: 149.83, EclipticLat: 0.46, NakshatraName: "Magha", NakshatraNum: 10},
		{Key: "Delta Leonis", StarName: "Zosma", Bayer: "Delta Leonis", Magnitude: 2.53, EclipticLon: 161.32, EclipticLat: 14.33, NakshatraName: "Purva Phalguni", NakshatraNum: 11},
		{Key: "Beta Leonis", StarName: "Denebola", Bayer: "Beta Leonis", Magnitude: 2.13, EclipticLon: 171.62, EclipticLat: 12.27, NakshatraName: "Uttara Phalguni", NakshatraNum: 12},
		{Key: "Alpha Corvi", StarName: "Alchiba", Bayer: "Alpha Corvi", Magnitude: 4.00, EclipticLon: 172.17, EclipticLat: 22.08, NakshatraName: "Hasta", NakshatraNum: 13},
		{Key: "Alpha Virginis", StarName: "Spica", Bayer: "Alpha Virginis", Magnitude: 0.97, EclipticLon: 195.43, EclipticLat: 18.33, NakshatraName: "Chitra", NakshatraNum: 14},
		{Key: "Alpha Bootis", StarName: "Arcturus", Bayer: "Alpha Bootis", Magnitude: -0.05, EclipticLon: 204.23, EclipticLat: 30.74, NakshatraName: "Swati", NakshatraNum: 15},
		{Key: "Alpha Librae", StarName: "Zubenelgenubi", Bayer: "Alpha Librae", Magnitude: 2.75, EclipticLon: 214.70, EclipticLat: 30.78, NakshatraName: "Vishakha", NakshatraNum: 16},
		{Key: "Beta Scorpii", StarName: "Acrab", Bayer: "Beta Scorpii", Magnitude: 2.62, EclipticLon: 234.60, EclipticLat: 38.18, NakshatraName: "Anuradha", NakshatraNum: 17},
		{Key: "Alpha Scorpii", StarName: "Antares", Bayer: "Alpha Scorpii", Magnitude: 0.91, EclipticLon: 239.60, EclipticLat: 46.65, NakshatraName: "Jyeshtha", NakshatraNum: 18},
		{Key: "Epsilon Scorpii", StarName: "Epsilon Scorpii", Magnitude: 2.29, EclipticLon: 243.80, EclipticLat: 55.57, NakshatraName: "Mula", NakshatraNum: 19},
		{Key: "Delta Sagittarii", StarName: "Kaus Media", Bayer: "Delta Sagittarii", Magnitude: 2.67, EclipticLon: 277.44, EclipticLat: 51.48, NakshatraName: "Purva Ashadha", NakshatraNum: 20},
		{Key: "Zeta Sagittarii", StarName: "Ascella", Bayer: "Zeta Sagittarii", Magnitude: 2.59, EclipticLon: 291.91, EclipticLat: 50.38, NakshatraName: "Uttara Ashadha", NakshatraNum: 21},
		{Key: "Alpha Aquilae", StarName: "Altair", Bayer: "Alpha Aquilae", Magnitude: 0.76, EclipticLon: 301.78, EclipticLat: 29.30, NakshatraName: "Shravana", NakshatraNum: 22},
		{Key: "Alpha Delphini", StarName: "Sualocin", Bayer: "Alpha Delphini", Magnitude: 3.80, EclipticLon: 317.38, EclipticLat: 33.02, NakshatraName: "Dhanishta", NakshatraNum: 23},
		{Key: "Lambda Aquarii", StarName: "Lambda Aquarii", Magnitude: 3.73, EclipticLon: 346.99, EclipticLat: 12.54, NakshatraName: "Shatabhisha", NakshatraNum: 24},
		{Key: "Alpha Pegasi", StarName: "Markab", Bayer: "Alpha Pegasi", Magnitude: 2.48, EclipticLon: 353.49, EclipticLat: 19.41, NakshatraName: "Purva Bhadrapada", NakshatraNum: 25},
		{Key: "Gamma Pegasi", StarName: "Algenib", Bayer: "Gamma Pegasi", Magnitude: 2.84, EclipticLon: 9.16, EclipticLat: 12.60, NakshatraName: "Uttara Bhadrapada", NakshatraNum: 26},
		{Key: "Zeta Piscium", StarName: "Zeta Piscium", Magnitude: 5.21, EclipticLon: 19.88, EclipticLat: -0.21, NakshatraName: "Revati", NakshatraNum: 27},
	}
}

// XiuStars returns the 28 Chinese xiu determinative stars.
func XiuStars() []StarEntry {
	return []StarEntry{
		{Key: "Alpha Virginis", StarName: "Spica", Bayer: "Alpha Virginis", Magnitude: 0.97, EclipticLon: 195.43, EclipticLat: 18.33, XiuName: "Horn", XiuNum: 1, XiuPinyin: "Jiao"},
		{Key: "Kappa Virginis", StarName: "Kappa Virginis", Magnitude: 4.18, EclipticLon: 207.45, EclipticLat: 21.70, XiuName: "Neck", XiuNum: 2, XiuPinyin: "Kang"},
		{Key: "Alpha Librae", StarName: "Zubenelgenubi", Bayer: "Alpha Librae", Magnitude: 2.75, EclipticLon: 214.70, EclipticLat: 30.78, XiuName: "Root", XiuNum: 3, XiuPinyin: "Di"},
		{Key: "Pi Scorpii", StarName: "Pi Scorpii", Magnitude: 2.89, EclipticLon: 229.91, EclipticLat: 45.20, XiuName: "Room", XiuNum: 4, XiuPinyin: "Fang"},
		{Key: "Alpha Scorpii", StarName: "Antares", Bayer: "Alpha Scorpii", Magnitude: 0.91, EclipticLon: 239.60, EclipticLat: 46.65, XiuName: "Heart", XiuNum: 5, XiuPinyin: "Xin"},
		{Key: "Mu-1 Scorpii", StarName: "Mu-1 Scorpii", Magnitude: 2.98, EclipticLon: 242.68, EclipticLat: 59.79, XiuName: "Tail", XiuNum: 6, XiuPinyin: "Wei"},
		{Key: "Gamma Sagittarii", StarName: "Gamma Sagittarii", Magnitude: 2.98, EclipticLon: 272.10, EclipticLat: 53.00, XiuName: "Winnowing Basket", XiuNum: 7, XiuPinyin: "Ji"},
		{Key: "Phi Sagittarii", StarName: "Phi Sagittarii", Magnitude: 3.17, EclipticLon: 285.50, EclipticLat: 47.84, XiuName: "Dipper", XiuNum: 8, XiuPinyin: "Dou"},
		{Key: "Beta Capricorni", StarName: "Dabih", Bayer: "Beta Capricorni", Magnitude: 3.08, EclipticLon: 311.35, EclipticLat: 31.74, XiuName: "Ox", XiuNum: 9, XiuPinyin: "Niu"},
		{Key: "Epsilon Aquarii", StarName: "Epsilon Aquarii", Magnitude: 3.78, EclipticLon: 316.99, EclipticLat: 25.37, XiuName: "Girl", XiuNum: 10, XiuPinyin: "Nu"},
		{Key: "Beta Aquarii", StarName: "Beta Aquarii", Magnitude: 2.89, EclipticLon: 326.75, EclipticLat: 18.07, XiuName: "Emptiness", XiuNum: 11, XiuPinyin: "Xu"},
		{Key: "Alpha Aquarii", StarName: "Alpha Aquarii", Magnitude: 2.94, EclipticLon: 333.35, EclipticLat: 10.66, XiuName: "Rooftop", XiuNum: 12, XiuPinyin: "Wei"},
		{Key: "Alpha Pegasi", StarName: "Markab", Bayer: "Alpha Pegasi", Magnitude: 2.48, EclipticLon: 353.49, EclipticLat: 19.41, XiuName: "Encampment", XiuNum: 13, XiuPinyin: "Shi"},
		{Key: "Gamma Pegasi", StarName: "Algenib", Bayer: "Gamma Pegasi", Magnitude: 2.84, EclipticLon: 9.16, EclipticLat: 12.60, XiuName: "Wall", XiuNum: 14, XiuPinyin: "Bi"},
		{Key: "Eta Andromedae", StarName: "Eta Andromedae", Magnitude: 4.40, EclipticLon: 22.38, EclipticLat: 15.93, XiuName: "Legs", XiuNum: 15, XiuPinyin: "Kui"},
		{Key: "Beta Arietis", StarName: "Sheratan", Bayer: "Beta Arietis", Magnitude: 2.65, EclipticLon: 33.97, EclipticLat: 8.49, XiuName: "Bond", XiuNum: 16, XiuPinyin: "Lou"},
		{Key: "35 Arietis", StarName: "35 Arietis", Magnitude: 4.60, EclipticLon: 46.93, EclipticLat: 11.31, XiuName: "Stomach", XiuNum: 17, XiuPinyin: "Wei"},
		{Key: "17 Tauri", StarName: "Electra", Bayer: "17 Tauri", Magnitude: 3.70, EclipticLon: 59.41, EclipticLat: 4.19, XiuName: "Hairy Head", XiuNum: 18, XiuPinyin: "Mao"},
		{Key: "Epsilon Tauri", StarName: "Ain", Bayer: "Epsilon Tauri", Magnitude: 3.53, EclipticLon: 68.47, EclipticLat: -2.57, XiuName: "Net", XiuNum: 19, XiuPinyin: "Bi"},
		{Key: "Lambda Orionis", StarName: "Meissa", Bayer: "Lambda Orionis", Magnitude: 3.66, EclipticLon: 83.71, EclipticLat: -13.37, XiuName: "Turtle Beak", XiuNum: 20, XiuPinyin: "Zi"},
		{Key: "Zeta Orionis", StarName: "Alnitak", Bayer: "Zeta Orionis", Magnitude: 1.79, EclipticLon: 84.76, EclipticLat: -23.29, XiuName: "Three Stars", XiuNum: 21, XiuPinyin: "Shen"},
		{Key: "Mu Geminorum", StarName: "Tejat Posterior", Bayer: "Mu Geminorum", Magnitude: 2.87, EclipticLon: 95.30, EclipticLat: -0.82, XiuName: "Well", XiuNum: 22, XiuPinyin: "Jing"},
		{Key: "Theta Cancri", StarName: "Theta Cancri", Magnitude: 5.33, EclipticLon: 125.73, EclipticLat: -0.77, XiuName: "Ghost", XiuNum: 23, XiuPinyin: "Gui"},
		{Key: "Delta Hydrae", StarName: "Delta Hydrae", Magnitude: 4.14, EclipticLon: 130.30, EclipticLat: -12.39, XiuName: "Willow", XiuNum: 24, XiuPinyin: "Liu"},
		{Key: "Alpha Hydrae", StarName: "Alphard", Bayer: "Alpha Hydrae", Magnitude: 1.97, EclipticLon: 141.88, EclipticLat: -7.25, XiuName: "Star", XiuNum: 25, XiuPinyin: "Xing"},
		{Key: "Upsilon-1 Hydrae", StarName: "Upsilon-1 Hydrae", Magnitude: 4.11, EclipticLon: 145.53, EclipticLat: 0.22, XiuName: "Extended Net", XiuNum: 26, XiuPinyin: "Zhang"},
		{Key: "Alpha Crateris", StarName: "Alkes", Bayer: "Alpha Crateris", Magnitude: 4.07, EclipticLon: 159.28, EclipticLat: 10.40, XiuName: "Wings", XiuNum: 27, XiuPinyin: "Yi"},
		{Key: "Gamma Corvi", StarName: "Gienah", Bayer: "Gamma Corvi", Magnitude: 2.58, EclipticLon: 176.89, EclipticLat: 16.63, XiuName: "Chariot", XiuNum: 28, XiuPinyin: "Zhen"},
	}
}

// ── Shared Star Identification ────────────────────────────────────────────

// SharedStar is a star that anchors both a nakshatra and a xiu.
type SharedStar struct {
	Key       string
	StarName  string
	Bayer     string
	Magnitude float64
	EclipticLat float64
	Nakshatra string
	NakshatraNum int
	Xiu       string
	XiuNum    int
	XiuPinyin string
	IsFaint   bool // true if magnitude >= FaintThreshold
}

// FaintThreshold is the magnitude boundary between bright and faint stars.
// Stars with magnitude >= this value are classified as faint.
const FaintThreshold = 2.5

// FindSharedStars returns all stars that appear in both nakshatra and xiu pools.
func FindSharedStars() []SharedStar {
	naks := NakshatraStars()
	xius := XiuStars()

	nMap := make(map[string]StarEntry)
	for _, s := range naks {
		nMap[s.Key] = s
	}

	var shared []SharedStar
	for _, x := range xius {
		if n, ok := nMap[x.Key]; ok {
			shared = append(shared, SharedStar{
				Key:          x.Key,
				StarName:     x.StarName,
				Bayer:        x.Bayer,
				Magnitude:    n.Magnitude,
				EclipticLat:  n.EclipticLat,
				Nakshatra:    n.NakshatraName,
				NakshatraNum: n.NakshatraNum,
				Xiu:          x.XiuName,
				XiuNum:       x.XiuNum,
				XiuPinyin:    x.XiuPinyin,
				IsFaint:      n.Magnitude >= FaintThreshold,
			})
		}
	}

	sort.Slice(shared, func(i, j int) bool {
		return shared[i].Key < shared[j].Key
	})
	return shared
}

// ── Combined Pool ─────────────────────────────────────────────────────────

// CombinedPool returns all unique stars from both systems.
func CombinedPool() []StarEntry {
	naks := NakshatraStars()
	xius := XiuStars()

	seen := make(map[string]bool)
	var pool []StarEntry
	for _, s := range naks {
		if !seen[s.Key] {
			seen[s.Key] = true
			pool = append(pool, s)
		}
	}
	for _, s := range xius {
		if !seen[s.Key] {
			seen[s.Key] = true
			pool = append(pool, s)
		}
	}
	return pool
}

// ── Null Models ───────────────────────────────────────────────────────────

// NullModelResult holds the output of a null model bootstrap run.
type NullModelResult struct {
	ModelName       string
	Iterations      int
	PoolSize        int
	NakshatraDraws  int
	XiuDraws        int
	FaintThreshold  float64
	ObservedTotal   int
	ObservedFaint   int
	ObservedBright  int
	NullMeanTotal   float64
	NullCITotalLow  float64
	NullCITotalHigh float64
	NullMeanFaint   float64
	NullCIFaintLow  float64
	NullCIFaintHigh float64
	NullMeanBright  float64
	PTotalGE        float64 // p(overlap >= observed)
	PTotalLE        float64 // p(overlap <= observed)
	PFaintGE        float64 // p(faint >= observed)
	FaintDist       map[int]int
}

// NullModelConfig holds parameters for a null model run.
type NullModelConfig struct {
	Name            string
	Iterations      int
	NakshatraDraws  int
	XiuDraws        int
	FaintThreshold  float64
	Seed            int64
	UseEclipticWeight bool   // if true, weight by exp(-|lat|/EclipticScale)
	EclipticScale   float64 // scale factor for ecliptic proximity weight
}

// RunNullModelUniform runs the uniform null model (Null 1).
// Each system independently selects stars from the combined pool
// without replacement — each mansion within a system gets a unique star.
// The two systems draw independently from the same pool.
func RunNullModelUniform(cfg NullModelConfig) NullModelResult {
	pool := CombinedPool()
	keys := make([]string, len(pool))
	mags := make([]float64, len(pool))
	for i, s := range pool {
		keys[i] = s.Key
		mags[i] = s.Magnitude
	}

	rng := rand.New(rand.NewSource(cfg.Seed))

	shared := FindSharedStars()
	observedTotal := len(shared)
	observedFaint := 0
	for _, s := range shared {
		if s.IsFaint {
			observedFaint++
		}
	}
	observedBright := observedTotal - observedFaint

	totalOverlaps := make([]int, cfg.Iterations)
	faintOverlaps := make([]int, cfg.Iterations)
	brightOverlaps := make([]int, cfg.Iterations)

	// Pre-allocate indices for Fisher-Yates shuffle
	indices := make([]int, len(keys))

	for i := 0; i < cfg.Iterations; i++ {
		// Shuffle for nakshatra draw
		for j := range indices {
			indices[j] = j
		}
		for j := len(indices) - 1; j > 0; j-- {
			k := rng.Intn(j + 1)
			indices[j], indices[k] = indices[k], indices[j]
		}
		nakSet := make(map[string]bool)
		for j := 0; j < cfg.NakshatraDraws; j++ {
			nakSet[keys[indices[j]]] = true
		}

		// Shuffle for xiu draw (independent)
		for j := range indices {
			indices[j] = j
		}
		for j := len(indices) - 1; j > 0; j-- {
			k := rng.Intn(j + 1)
			indices[j], indices[k] = indices[k], indices[j]
		}
		xiuSet := make(map[string]bool)
		for j := 0; j < cfg.XiuDraws; j++ {
			xiuSet[keys[indices[j]]] = true
		}

		overlap := 0
		faint := 0
		for k := range nakSet {
			if xiuSet[k] {
				overlap++
				for pi, pk := range keys {
					if pk == k {
						if mags[pi] >= cfg.FaintThreshold {
							faint++
						}
						break
					}
				}
			}
		}
		totalOverlaps[i] = overlap
		faintOverlaps[i] = faint
		brightOverlaps[i] = overlap - faint
	}

	meanTotal := meanInt(totalOverlaps)
	sdTotal := stddev(totalOverlaps, meanTotal)
	seTotal := sdTotal / math.Sqrt(float64(cfg.Iterations))
	ciTotalLow := meanTotal - 1.96*seTotal
	ciTotalHigh := meanTotal + 1.96*seTotal

	meanFaint := meanInt(faintOverlaps)
	sdFaint := stddev(faintOverlaps, meanFaint)
	seFaint := sdFaint / math.Sqrt(float64(cfg.Iterations))
	ciFaintLow := meanFaint - 1.96*seFaint
	ciFaintHigh := meanFaint + 1.96*seFaint

	meanBright := meanInt(brightOverlaps)

	pTotalGE := float64(countGE(totalOverlaps, observedTotal)) / float64(cfg.Iterations)
	pTotalLE := float64(countLE(totalOverlaps, observedTotal)) / float64(cfg.Iterations)
	pFaintGE := float64(countGE(faintOverlaps, observedFaint)) / float64(cfg.Iterations)

	faintDist := make(map[int]int)
	for _, f := range faintOverlaps {
		faintDist[f]++
	}

	return NullModelResult{
		ModelName:       cfg.Name,
		Iterations:      cfg.Iterations,
		PoolSize:        len(pool),
		NakshatraDraws:  cfg.NakshatraDraws,
		XiuDraws:        cfg.XiuDraws,
		FaintThreshold:  cfg.FaintThreshold,
		ObservedTotal:   observedTotal,
		ObservedFaint:   observedFaint,
		ObservedBright:  observedBright,
		NullMeanTotal:   meanTotal,
		NullCITotalLow:  ciTotalLow,
		NullCITotalHigh: ciTotalHigh,
		NullMeanFaint:   meanFaint,
		NullCIFaintLow:  ciFaintLow,
		NullCIFaintHigh: ciFaintHigh,
		NullMeanBright:  meanBright,
		PTotalGE:        pTotalGE,
		PTotalLE:        pTotalLE,
		PFaintGE:        pFaintGE,
		FaintDist:       faintDist,
	}
}

// RunNullModelBrightness runs the brightness-weighted null model (Null 2).
// Both cultures independently select stars from the combined pool,
// with selection probability proportional to exp(-magnitude).
func RunNullModelBrightness(cfg NullModelConfig) NullModelResult {
	pool := CombinedPool()
	keys := make([]string, len(pool))
	mags := make([]float64, len(pool))
	for i, s := range pool {
		keys[i] = s.Key
		mags[i] = s.Magnitude
	}

	// Compute weights
	weights := make([]float64, len(pool))
	totalWeight := 0.0
	for i, m := range mags {
		w := math.Exp(-m)
		if cfg.UseEclipticWeight {
			w *= math.Exp(-math.Abs(pool[i].EclipticLat) / cfg.EclipticScale)
		}
		weights[i] = w
		totalWeight += w
	}

	// Normalize to probabilities
	probs := make([]float64, len(pool))
	for i, w := range weights {
		probs[i] = w / totalWeight
	}

	rng := rand.New(rand.NewSource(cfg.Seed))

	shared := FindSharedStars()
	observedTotal := len(shared)
	observedFaint := 0
	for _, s := range shared {
		if s.IsFaint {
			observedFaint++
		}
	}
	observedBright := observedTotal - observedFaint

	totalOverlaps := make([]int, cfg.Iterations)
	faintOverlaps := make([]int, cfg.Iterations)
	brightOverlaps := make([]int, cfg.Iterations)

	for i := 0; i < cfg.Iterations; i++ {
		// Nakshatra draws
		nakSet := make(map[string]bool)
		for j := 0; j < cfg.NakshatraDraws; j++ {
			idx := weightedRandom(rng, probs)
			nakSet[keys[idx]] = true
		}

		// Xiu draws
		xiuSet := make(map[string]bool)
		for j := 0; j < cfg.XiuDraws; j++ {
			idx := weightedRandom(rng, probs)
			xiuSet[keys[idx]] = true
		}

		// Overlap
		overlap := 0
		faint := 0
		for k := range nakSet {
			if xiuSet[k] {
				overlap++
				// Find magnitude
				for pi, pk := range keys {
					if pk == k {
						if mags[pi] >= cfg.FaintThreshold {
							faint++
						}
						break
					}
				}
			}
		}
		totalOverlaps[i] = overlap
		faintOverlaps[i] = faint
		brightOverlaps[i] = overlap - faint
	}

	// Statistics
	meanTotal := meanInt(totalOverlaps)
	sdTotal := stddev(totalOverlaps, meanTotal)
	seTotal := sdTotal / math.Sqrt(float64(cfg.Iterations))
	ciTotalLow := meanTotal - 1.96*seTotal
	ciTotalHigh := meanTotal + 1.96*seTotal

	meanFaint := meanInt(faintOverlaps)
	sdFaint := stddev(faintOverlaps, meanFaint)
	seFaint := sdFaint / math.Sqrt(float64(cfg.Iterations))
	ciFaintLow := meanFaint - 1.96*seFaint
	ciFaintHigh := meanFaint + 1.96*seFaint

	meanBright := meanInt(brightOverlaps)

	// P-values
	pTotalGE := float64(countGE(totalOverlaps, observedTotal)) / float64(cfg.Iterations)
	pTotalLE := float64(countLE(totalOverlaps, observedTotal)) / float64(cfg.Iterations)
	pFaintGE := float64(countGE(faintOverlaps, observedFaint)) / float64(cfg.Iterations)

	// Faint distribution
	faintDist := make(map[int]int)
	for _, f := range faintOverlaps {
		faintDist[f]++
	}

	return NullModelResult{
		ModelName:       cfg.Name,
		Iterations:      cfg.Iterations,
		PoolSize:        len(pool),
		NakshatraDraws:  cfg.NakshatraDraws,
		XiuDraws:        cfg.XiuDraws,
		FaintThreshold:  cfg.FaintThreshold,
		ObservedTotal:   observedTotal,
		ObservedFaint:   observedFaint,
		ObservedBright:  observedBright,
		NullMeanTotal:   meanTotal,
		NullCITotalLow:  ciTotalLow,
		NullCITotalHigh: ciTotalHigh,
		NullMeanFaint:   meanFaint,
		NullCIFaintLow:  ciFaintLow,
		NullCIFaintHigh: ciFaintHigh,
		NullMeanBright:  meanBright,
		PTotalGE:        pTotalGE,
		PTotalLE:        pTotalLE,
		PFaintGE:        pFaintGE,
		FaintDist:       faintDist,
	}
}

// ── Ecliptic Position Confound Test ───────────────────────────────────────

// EclipticConfoundResult holds the output of the ecliptic position confound test.
type EclipticConfoundResult struct {
	SharedFaintMeanLat    float64
	NonSharedFaintMeanLat float64
	AllFaintMeanLat       float64
	SharedFaintCount      int
	NonSharedFaintCount   int
	PermutationN          int
	PLower                float64 // p(random subset has <= mean |lat|)
	PHigher               float64 // p(random subset has >= mean |lat|)
	SharedAllMeanLat      float64
	NonSharedAllMeanLat   float64
	PAllLower             float64
}

// RunEclipticConfoundTest tests whether shared faint stars are systematically
// closer to the ecliptic than non-shared faint stars.
func RunEclipticConfoundTest(permutations int, seed int64) EclipticConfoundResult {
	pool := CombinedPool()
	shared := FindSharedStars()

	sharedKeys := make(map[string]bool)
	for _, s := range shared {
		sharedKeys[s.Key] = true
	}

	// Collect absolute ecliptic latitudes for faint stars
	var sharedFaintLats []float64
	var nonSharedFaintLats []float64
	var allFaintLats []float64
	var allSharedLats []float64
	var allNonSharedLats []float64

	for _, s := range pool {
		absLat := math.Abs(s.EclipticLat)
		if sharedKeys[s.Key] {
			allSharedLats = append(allSharedLats, absLat)
		} else {
			allNonSharedLats = append(allNonSharedLats, absLat)
		}
		if s.Magnitude >= FaintThreshold {
			allFaintLats = append(allFaintLats, absLat)
			if sharedKeys[s.Key] {
				sharedFaintLats = append(sharedFaintLats, absLat)
			} else {
				nonSharedFaintLats = append(nonSharedFaintLats, absLat)
			}
		}
	}

	sharedFaintMean := mean(sharedFaintLats)
	nonSharedFaintMean := mean(nonSharedFaintLats)
	allFaintMean := mean(allFaintLats)
	sharedAllMean := mean(allSharedLats)
	nonSharedAllMean := mean(allNonSharedLats)

	rng := rand.New(rand.NewSource(seed))

	// Permutation test: shared faint vs all faint
	nShared := len(sharedFaintLats)
	countLower := 0
	countHigher := 0
	for i := 0; i < permutations; i++ {
		// Draw random subset of size nShared from allFaintLats
		sample := sampleWithoutReplacement(rng, allFaintLats, nShared)
		sampleMean := mean(sample)
		if sampleMean <= sharedFaintMean {
			countLower++
		}
		if sampleMean >= sharedFaintMean {
			countHigher++
		}
	}

	// Permutation test: all shared vs all non-shared
	nAllShared := len(allSharedLats)
	allPoolLats := make([]float64, len(pool))
	for i, s := range pool {
		allPoolLats[i] = math.Abs(s.EclipticLat)
	}
	countAllLower := 0
	for i := 0; i < permutations; i++ {
		sample := sampleWithoutReplacement(rng, allPoolLats, nAllShared)
		if mean(sample) <= sharedAllMean {
			countAllLower++
		}
	}

	return EclipticConfoundResult{
		SharedFaintMeanLat:    sharedFaintMean,
		NonSharedFaintMeanLat: nonSharedFaintMean,
		AllFaintMeanLat:       allFaintMean,
		SharedFaintCount:      len(sharedFaintLats),
		NonSharedFaintCount:   len(nonSharedFaintLats),
		PermutationN:          permutations,
		PLower:                float64(countLower) / float64(permutations),
		PHigher:               float64(countHigher) / float64(permutations),
		SharedAllMeanLat:      sharedAllMean,
		NonSharedAllMeanLat:   nonSharedAllMean,
		PAllLower:             float64(countAllLower) / float64(permutations),
	}
}

// ── Formatting ─────────────────────────────────────────────────────────────

// FormatSharedStars returns a human-readable table of shared stars.
func FormatSharedStars() string {
	shared := FindSharedStars()
	var b strings.Builder
	b.WriteString("Shared Nakshatra-Xiu Anchor Stars\n")
	b.WriteString(strings.Repeat("=", 72) + "\n\n")
	b.WriteString(fmt.Sprintf("%-22s %-6s %-8s %-18s %-18s\n",
		"Star", "Mag", "Class", "Nakshatra", "Xiu"))
	b.WriteString(strings.Repeat("-", 72) + "\n")

	faintCount := 0
	brightCount := 0
	for _, s := range shared {
		class := "BRIGHT"
		if s.IsFaint {
			class = "FAINT"
			faintCount++
		} else {
			brightCount++
		}
		b.WriteString(fmt.Sprintf("%-22s %-6.2f %-8s %-18s %-18s\n",
			s.StarName, s.Magnitude, class, s.Nakshatra, s.Xiu+" ("+s.XiuPinyin+")"))
	}
	b.WriteString(fmt.Sprintf("\nTotal: %d shared (%d bright, %d faint)\n",
		len(shared), brightCount, faintCount))
	return b.String()
}

// FormatNullModelResult returns a human-readable null model report.
func FormatNullModelResult(r NullModelResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Null Model: %s\n", r.ModelName))
	b.WriteString(strings.Repeat("=", 60) + "\n\n")
	b.WriteString(fmt.Sprintf("Iterations:    %d\n", r.Iterations))
	b.WriteString(fmt.Sprintf("Pool size:     %d stars\n", r.PoolSize))
	b.WriteString(fmt.Sprintf("Draws:         nakshatra=%d, xiu=%d\n", r.NakshatraDraws, r.XiuDraws))
	b.WriteString(fmt.Sprintf("Faint cutoff:  mag >= %.1f\n\n", r.FaintThreshold))

	b.WriteString("TOTAL OVERLAP:\n")
	b.WriteString(fmt.Sprintf("  Null mean:   %.1f (95%% CI: %.1f-%.1f)\n",
		r.NullMeanTotal, r.NullCITotalLow, r.NullCITotalHigh))
	b.WriteString(fmt.Sprintf("  Observed:    %d\n", r.ObservedTotal))
	b.WriteString(fmt.Sprintf("  p (>= obs):  %.4f\n", r.PTotalGE))
	b.WriteString(fmt.Sprintf("  p (<= obs):  %.4f\n\n", r.PTotalLE))

	b.WriteString("FAINT STARS:\n")
	b.WriteString(fmt.Sprintf("  Null mean:   %.1f (95%% CI: %.1f-%.1f)\n",
		r.NullMeanFaint, r.NullCIFaintLow, r.NullCIFaintHigh))
	b.WriteString(fmt.Sprintf("  Observed:    %d\n", r.ObservedFaint))
	b.WriteString(fmt.Sprintf("  p (>= obs):  %.4f\n\n", r.PFaintGE))

	b.WriteString("BRIGHT STARS:\n")
	b.WriteString(fmt.Sprintf("  Null mean:   %.1f\n", r.NullMeanBright))
	b.WriteString(fmt.Sprintf("  Observed:    %d\n\n", r.ObservedBright))

	b.WriteString("Faint overlap distribution:\n")
	keys := make([]int, 0, len(r.FaintDist))
	for k := range r.FaintDist {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		pct := float64(r.FaintDist[k]) / float64(r.Iterations) * 100
		b.WriteString(fmt.Sprintf("  %d: %d (%.1f%%)\n", k, r.FaintDist[k], pct))
	}

	return b.String()
}

// FormatEclipticConfoundResult returns a human-readable confound test report.
func FormatEclipticConfoundResult(r EclipticConfoundResult) string {
	var b strings.Builder
	b.WriteString("Ecliptic Position Confound Test\n")
	b.WriteString(strings.Repeat("=", 60) + "\n\n")
	b.WriteString(fmt.Sprintf("Faint stars (mag >= %.1f):\n", FaintThreshold))
	b.WriteString(fmt.Sprintf("  Shared:     %d stars, mean |lat| = %.1f deg\n",
		r.SharedFaintCount, r.SharedFaintMeanLat))
	b.WriteString(fmt.Sprintf("  Non-shared: %d stars, mean |lat| = %.1f deg\n",
		r.NonSharedFaintCount, r.NonSharedFaintMeanLat))
	b.WriteString(fmt.Sprintf("  All faint:  mean |lat| = %.1f deg\n\n", r.AllFaintMeanLat))

	b.WriteString(fmt.Sprintf("Permutation test (N=%d):\n", r.PermutationN))
	b.WriteString(fmt.Sprintf("  p (random subset has <= mean |lat|): %.4f\n", r.PLower))
	b.WriteString(fmt.Sprintf("  p (random subset has >= mean |lat|): %.4f\n\n", r.PHigher))

	b.WriteString("All shared stars:\n")
	b.WriteString(fmt.Sprintf("  Shared mean |lat|:     %.1f deg\n", r.SharedAllMeanLat))
	b.WriteString(fmt.Sprintf("  Non-shared mean |lat|: %.1f deg\n", r.NonSharedAllMeanLat))
	b.WriteString(fmt.Sprintf("  p (<= mean):           %.4f\n", r.PAllLower))

	return b.String()
}

// ── JSON Output ───────────────────────────────────────────────────────────

// LunarMansionReport is the combined output for the lunar mansion analysis.
type LunarMansionReport struct {
	SharedStars         []SharedStar            `json:"shared_stars"`
	NullModelUniform    NullModelResult         `json:"null_model_uniform"`
	NullModelBrightness NullModelResult         `json:"null_model_brightness"`
	NullModelEcliptic   NullModelResult         `json:"null_model_ecliptic_weighted"`
	EclipticConfound    EclipticConfoundResult  `json:"ecliptic_confound"`
}

// LunarMansionReportJSON serializes the full report to JSON.
func (r *LunarMansionReport) LunarMansionReportJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ComputeLunarMansionReport runs the full lunar mansion analysis.
func ComputeLunarMansionReport(seed int64) *LunarMansionReport {
	shared := FindSharedStars()

	brightCfg := NullModelConfig{
		Name:            "Brightness-weighted (combined pool)",
		Iterations:      10000,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  FaintThreshold,
		Seed:            seed,
		UseEclipticWeight: false,
	}

	eclipticCfg := NullModelConfig{
		Name:            "Brightness + Ecliptic proximity weighted",
		Iterations:      10000,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  FaintThreshold,
		Seed:            seed,
		UseEclipticWeight: true,
		EclipticScale:   10.0,
	}

	uniformCfg := NullModelConfig{
		Name:            "Uniform (combined pool, without replacement)",
		Iterations:      10000,
		NakshatraDraws:  27,
		XiuDraws:        28,
		FaintThreshold:  FaintThreshold,
		Seed:            seed,
		UseEclipticWeight: false,
	}

	return &LunarMansionReport{
		SharedStars:        shared,
		NullModelUniform:    RunNullModelUniform(uniformCfg),
		NullModelBrightness: RunNullModelBrightness(brightCfg),
		NullModelEcliptic:   RunNullModelBrightness(eclipticCfg),
		EclipticConfound:    RunEclipticConfoundTest(100000, seed),
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────

func weightedRandom(rng *rand.Rand, probs []float64) int {
	r := rng.Float64()
	cum := 0.0
	for i, p := range probs {
		cum += p
		if r < cum {
			return i
		}
	}
	return len(probs) - 1
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func meanInt(vals []int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0
	for _, v := range vals {
		sum += v
	}
	return float64(sum) / float64(len(vals))
}

func stddev(vals []int, m float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	sumSq := 0.0
	for _, v := range vals {
		d := float64(v) - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(vals)-1))
}

func countGE(vals []int, threshold int) int {
	count := 0
	for _, v := range vals {
		if v >= threshold {
			count++
		}
	}
	return count
}

func countLE(vals []int, threshold int) int {
	count := 0
	for _, v := range vals {
		if v <= threshold {
			count++
		}
	}
	return count
}

func sampleWithoutReplacement(rng *rand.Rand, pool []float64, n int) []float64 {
	// Fisher-Yates shuffle on indices, take first n
	indices := make([]int, len(pool))
	for i := range indices {
		indices[i] = i
	}
	for i := len(indices) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		result[i] = pool[indices[i]]
	}
	return result
}
