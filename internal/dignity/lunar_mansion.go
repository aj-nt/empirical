package dignity

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
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
	ManazilName    string  // Which manazil this anchors (empty if not manazil)
	ManazilNum     int     // Manazil number 1-28 (0 if not manazil)
}

//go:embed data/nakshatra_stars.json
var nakshatraJSON []byte

//go:embed data/xiu_stars.json
var xiuJSON []byte

//go:embed data/manazil_stars.json
var manazilJSON []byte

var (
	nakshatraOnce sync.Once
	nakshatraCache []StarEntry

	xiuOnce sync.Once
	xiuCache []StarEntry

	manazilOnce sync.Once
	manazilCache []StarEntry
)

// NakshatraStars returns the 27 nakshatra determinative stars, loaded from embedded JSON.
// Results are cached after the first call via sync.Once.
func NakshatraStars() []StarEntry {
	nakshatraOnce.Do(func() {
		nakshatraCache = mustLoadStars(nakshatraJSON)
	})
	return nakshatraCache
}

// XiuStars returns the 28 Chinese xiu determinative stars, loaded from embedded JSON.
// Results are cached after the first call via sync.Once.
func XiuStars() []StarEntry {
	xiuOnce.Do(func() {
		xiuCache = mustLoadStars(xiuJSON)
	})
	return xiuCache
}

// ManazilStars returns the 28 Arabic manazil al-qamar determinative stars, loaded from embedded JSON.
// Results are cached after the first call via sync.Once.
func ManazilStars() []StarEntry {
	manazilOnce.Do(func() {
		manazilCache = mustLoadStars(manazilJSON)
	})
	return manazilCache
}

// mustLoadStars unmarshals embedded star JSON data, panicking on any error.
// Validation happens at first call time (via sync.Once), not on every invocation.
func mustLoadStars(data []byte) []StarEntry {
	var stars []StarEntry
	if err := json.Unmarshal(data, &stars); err != nil {
		panic("failed to unmarshal embedded star data: " + err.Error())
	}
	return stars
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

// ── Per-Chart Mansion Placement ───────────────────────────────────────────

// MansionPlacement records which mansion a planet falls in.
type MansionPlacement struct {
	Planet       string  `json:"planet"`
	TropicalLon  float64 `json:"tropical_lon"`
	SiderealLon  float64 `json:"sidereal_lon"`
	Nakshatra    string  `json:"nakshatra"`
	NakshatraNum int     `json:"nakshatra_num"`
	Xiu          string  `json:"xiu"`
	XiuNum       int     `json:"xiu_num"`
	XiuPinyin    string  `json:"xiu_pinyin"`
	Manazil      string  `json:"manazil,omitempty"`
	ManazilNum   int     `json:"manazil_num,omitempty"`
	Converges    bool    `json:"converges"` // true if nakshatra and xiu share the same anchor star
}

// MansionConvergence holds per-chart mansion placement results.
type MansionConvergence struct {
	Name       string             `json:"name"`
	Ayanamsa   float64            `json:"ayanamsa"`
	Planets    []MansionPlacement `json:"planets"`
	Converging int                `json:"converging"`
	Total      int                `json:"total"`
}

// NakshatraForLongitude returns the nakshatra for a sidereal longitude.
// Nakshatras are 13°20' (13.333°) sectors starting at 0° Aries (sidereal).
func NakshatraForLongitude(sidLon float64) (name string, num int) {
	lon := normalizeLon(sidLon)
	idx := int(lon / (360.0 / 27.0))
	if idx >= 27 {
		idx = 26
	}
	stars := NakshatraStars()
	return stars[idx].NakshatraName, stars[idx].NakshatraNum
}

// XiuForLongitude returns the Chinese xiu for a sidereal longitude.
// Xiu are unequal sectors anchored to determinative stars.
// The sector boundary is the midpoint between consecutive anchor stars.
func XiuForLongitude(sidLon float64) (name string, num int, pinyin string) {
	lon := normalizeLon(sidLon)
	stars := XiuStars()

	// Build boundary array: midpoint between consecutive stars
	// Xiu 1 starts at the midpoint between star 28 and star 1
	for i := 0; i < 28; i++ {
		curr := stars[i]
		next := stars[(i+1)%28]

		// Midpoint between current and next star
		boundary := (curr.EclipticLon + next.EclipticLon) / 2.0
		if next.EclipticLon < curr.EclipticLon {
			// Wrap-around case
			boundary = (curr.EclipticLon + next.EclipticLon + 360) / 2.0
			if boundary >= 360 {
				boundary -= 360
			}
		}

		// Check if lon falls in this sector
		prev := stars[(i+27)%28]
		prevBoundary := (prev.EclipticLon + curr.EclipticLon) / 2.0
		if curr.EclipticLon < prev.EclipticLon {
			prevBoundary = (prev.EclipticLon + curr.EclipticLon + 360) / 2.0
			if prevBoundary >= 360 {
				prevBoundary -= 360
			}
		}

		if inSector(lon, prevBoundary, boundary) {
			return curr.XiuName, curr.XiuNum, curr.XiuPinyin
		}
	}

	// Fallback: return last xiu
	return stars[27].XiuName, stars[27].XiuNum, stars[27].XiuPinyin
}

// inSector checks if lon falls in the sector from start to end (inclusive of start, exclusive of end).
func inSector(lon, start, end float64) bool {
	if start <= end {
		return lon >= start && lon < end
	}
	// Wrap-around sector
	return lon >= start || lon < end
}

// ComputeMansionConvergence computes nakshatra and xiu placements for all planets
// in a natal chart, using sidereal positions (tropical - ayanamsa).
func ComputeMansionConvergence(name string, tropical map[string]float64, ayanamsa float64) *MansionConvergence {
	// Build shared star lookup
	shared := FindSharedStars()
	sharedKeys := make(map[string]bool)
	for _, s := range shared {
		sharedKeys[s.Key] = true
	}

	// Build nakshatra star key lookup
	naks := NakshatraStars()
	naksKeyToName := make(map[string]string)
	for _, s := range naks {
		naksKeyToName[s.Key] = s.NakshatraName
	}

	// Build xiu star key lookup
	xius := XiuStars()
	xiuKeyToName := make(map[string]string)
	xiuKeyToPinyin := make(map[string]string)
	for _, s := range xius {
		xiuKeyToName[s.Key] = s.XiuName
		xiuKeyToPinyin[s.Key] = s.XiuPinyin
	}

	// Classical planets only for empirical verification
	planets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}

	var placements []MansionPlacement
	converging := 0

	for _, p := range planets {
		tropLon, ok := tropical[p]
		if !ok {
			continue
		}
		sidLon := normalizeLon(tropLon - ayanamsa)

		nakName, nakNum := NakshatraForLongitude(sidLon)
		xiuName, xiuNum, xiuPinyin := XiuForLongitude(sidLon)

		// Check convergence: do the nakshatra and xiu share the same anchor star?
		converges := false
		for _, s := range naks {
			if s.NakshatraName == nakName {
				if sharedKeys[s.Key] {
					// This nakshatra's star is shared — check if the xiu also uses it
					for _, x := range xius {
						if x.Key == s.Key && x.XiuName == xiuName {
							converges = true
							break
						}
					}
				}
				break
			}
		}

		if converges {
			converging++
		}

		placements = append(placements, MansionPlacement{
			Planet:       p,
			TropicalLon:  tropLon,
			SiderealLon:  sidLon,
			Nakshatra:    nakName,
			NakshatraNum: nakNum,
			Xiu:          xiuName,
			XiuNum:       xiuNum,
			XiuPinyin:    xiuPinyin,
			Converges:    converges,
		})
	}

	return &MansionConvergence{
		Name:       name,
		Ayanamsa:   ayanamsa,
		Planets:    placements,
		Converging: converging,
		Total:      len(placements),
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Three-Way Mansion Comparison (Nakshatra / Xiu / Manazil)
// ═══════════════════════════════════════════════════════════════════════════

// ThreePoolCombined returns all unique stars from nakshatra, xiu, and manazil.
func ThreePoolCombined() []StarEntry {
	naks := NakshatraStars()
	xius := XiuStars()
	manazils := ManazilStars()

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
	for _, s := range manazils {
		if !seen[s.Key] {
			seen[s.Key] = true
			pool = append(pool, s)
		}
	}
	return pool
}

// ThreeWayShared holds stars that appear in multiple mansion systems.
type ThreeWayShared struct {
	Key       string
	StarName  string
	Bayer     string
	Magnitude float64
	EclipticLat float64
	IsFaint   bool
	InNakshatra bool
	InXiu       bool
	InManazil   bool
	NakshatraName string
	XiuName       string
	ManazilName   string
}

// FindThreeWayShared returns all stars and their membership across the three systems.
func FindThreeWayShared() []ThreeWayShared {
	naks := NakshatraStars()
	xius := XiuStars()
	manazils := ManazilStars()

	nMap := make(map[string]StarEntry)
	for _, s := range naks {
		nMap[s.Key] = s
	}
	xMap := make(map[string]StarEntry)
	for _, s := range xius {
		xMap[s.Key] = s
	}
	mMap := make(map[string]StarEntry)
	for _, s := range manazils {
		mMap[s.Key] = s
	}

	// Collect all unique keys
	allKeys := make(map[string]bool)
	for _, s := range naks {
		allKeys[s.Key] = true
	}
	for _, s := range xius {
		allKeys[s.Key] = true
	}
	for _, s := range manazils {
		allKeys[s.Key] = true
	}

	var result []ThreeWayShared
	for key := range allKeys {
		n, inN := nMap[key]
		x, inX := xMap[key]
		m, inM := mMap[key]

		// Get magnitude from whichever system has it
		mag := 0.0
		starName := key
		bayer := ""
		lat := 0.0
		if inN {
			mag = n.Magnitude
			starName = n.StarName
			bayer = n.Bayer
			lat = n.EclipticLat
		} else if inX {
			mag = x.Magnitude
			starName = x.StarName
			bayer = x.Bayer
			lat = x.EclipticLat
		} else if inM {
			mag = m.Magnitude
			starName = m.StarName
			bayer = m.Bayer
			lat = m.EclipticLat
		}

		ts := ThreeWayShared{
			Key:        key,
			StarName:   starName,
			Bayer:      bayer,
			Magnitude:  mag,
			EclipticLat: lat,
			IsFaint:    mag >= FaintThreshold,
			InNakshatra: inN,
			InXiu:       inX,
			InManazil:   inM,
		}
		if inN {
			ts.NakshatraName = n.NakshatraName
		}
		if inX {
			ts.XiuName = x.XiuName
		}
		if inM {
			ts.ManazilName = m.ManazilName
		}
		result = append(result, ts)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// ThreeWayNullResult holds the output of a three-way null model run.
type ThreeWayNullResult struct {
	ModelName          string
	Iterations         int
	PoolSize           int
	FaintThreshold     float64
	ObservedAllThree   int
	ObservedAllThreeFaint int
	ObservedNakXiu     int
	ObservedNakMan     int
	ObservedXiuMan     int
	NullMeanAllThree   float64
	NullCIAllThreeLow  float64
	NullCIAllThreeHigh float64
	NullMeanAllThreeFaint float64
	NullCIAllThreeFaintLow float64
	NullCIAllThreeFaintHigh float64
	NullMeanNakXiu     float64
	NullMeanNakMan     float64
	NullMeanXiuMan     float64
	PAllThreeGE        float64
	PAllThreeFaintGE   float64
}

// RunThreeWayNullBrightness runs a three-way brightness-weighted null model.
// All three cultures independently select stars from the combined three-system pool,
// with selection probability proportional to exp(-magnitude).
func RunThreeWayNullBrightness(cfg NullModelConfig) ThreeWayNullResult {
	pool := ThreePoolCombined()
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
	probs := make([]float64, len(pool))
	for i, w := range weights {
		probs[i] = w / totalWeight
	}

	rng := rand.New(rand.NewSource(cfg.Seed))

	// Observed values
	allStars := FindThreeWayShared()
	obsAllThree := 0
	obsAllThreeFaint := 0
	obsNakXiu := 0
	obsNakMan := 0
	obsXiuMan := 0
	for _, s := range allStars {
		if s.InNakshatra && s.InXiu && s.InManazil {
			obsAllThree++
			if s.IsFaint {
				obsAllThreeFaint++
			}
		}
		if s.InNakshatra && s.InXiu {
			obsNakXiu++
		}
		if s.InNakshatra && s.InManazil {
			obsNakMan++
		}
		if s.InXiu && s.InManazil {
			obsXiuMan++
		}
	}

	allThreeOverlaps := make([]int, cfg.Iterations)
	allThreeFaintOverlaps := make([]int, cfg.Iterations)
	nakXiuOverlaps := make([]int, cfg.Iterations)
	nakManOverlaps := make([]int, cfg.Iterations)
	xiuManOverlaps := make([]int, cfg.Iterations)

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

		// Manazil draws
		manSet := make(map[string]bool)
		for j := 0; j < 28; j++ {
			idx := weightedRandom(rng, probs)
			manSet[keys[idx]] = true
		}

		// Three-way overlap
		allThree := 0
		allThreeFaint := 0
		for k := range nakSet {
			if xiuSet[k] && manSet[k] {
				allThree++
				for pi, pk := range keys {
					if pk == k {
						if mags[pi] >= cfg.FaintThreshold {
							allThreeFaint++
						}
						break
					}
				}
			}
		}

		// Pairwise overlaps
		nakXiu := 0
		for k := range nakSet {
			if xiuSet[k] {
				nakXiu++
			}
		}
		nakMan := 0
		for k := range nakSet {
			if manSet[k] {
				nakMan++
			}
		}
		xiuMan := 0
		for k := range xiuSet {
			if manSet[k] {
				xiuMan++
			}
		}

		allThreeOverlaps[i] = allThree
		allThreeFaintOverlaps[i] = allThreeFaint
		nakXiuOverlaps[i] = nakXiu
		nakManOverlaps[i] = nakMan
		xiuManOverlaps[i] = xiuMan
	}

	// Statistics
	meanAllThree := meanInt(allThreeOverlaps)
	sdAllThree := stddev(allThreeOverlaps, meanAllThree)
	seAllThree := sdAllThree / math.Sqrt(float64(cfg.Iterations))

	meanAllThreeFaint := meanInt(allThreeFaintOverlaps)
	sdAllThreeFaint := stddev(allThreeFaintOverlaps, meanAllThreeFaint)
	seAllThreeFaint := sdAllThreeFaint / math.Sqrt(float64(cfg.Iterations))

	meanNakXiu := meanInt(nakXiuOverlaps)
	meanNakMan := meanInt(nakManOverlaps)
	meanXiuMan := meanInt(xiuManOverlaps)

	pAllThreeGE := float64(countGE(allThreeOverlaps, obsAllThree)) / float64(cfg.Iterations)
	pAllThreeFaintGE := float64(countGE(allThreeFaintOverlaps, obsAllThreeFaint)) / float64(cfg.Iterations)

	return ThreeWayNullResult{
		ModelName:            cfg.Name,
		Iterations:           cfg.Iterations,
		PoolSize:             len(pool),
		FaintThreshold:       cfg.FaintThreshold,
		ObservedAllThree:     obsAllThree,
		ObservedAllThreeFaint: obsAllThreeFaint,
		ObservedNakXiu:       obsNakXiu,
		ObservedNakMan:       obsNakMan,
		ObservedXiuMan:       obsXiuMan,
		NullMeanAllThree:     meanAllThree,
		NullCIAllThreeLow:    meanAllThree - 1.96*seAllThree,
		NullCIAllThreeHigh:   meanAllThree + 1.96*seAllThree,
		NullMeanAllThreeFaint: meanAllThreeFaint,
		NullCIAllThreeFaintLow: meanAllThreeFaint - 1.96*seAllThreeFaint,
		NullCIAllThreeFaintHigh: meanAllThreeFaint + 1.96*seAllThreeFaint,
		NullMeanNakXiu:       meanNakXiu,
		NullMeanNakMan:       meanNakMan,
		NullMeanXiuMan:       meanXiuMan,
		PAllThreeGE:          pAllThreeGE,
		PAllThreeFaintGE:     pAllThreeFaintGE,
	}
}


// ComputeMansionThreeWayConvergence computes mansion convergence across
// all three systems: nakshatra, xiu, and manazil.
// A planet converges if its nakshatra and xiu share a star AND its manazil
// also shares that same star (three-way convergence).
func ComputeMansionThreeWayConvergence(name string, tropical map[string]float64, ayanamsa float64) *MansionConvergence {
	// Build three-way shared star lookup
	allStars := FindThreeWayShared()
	threeWayKeys := make(map[string]bool)
	for _, s := range allStars {
		if s.InNakshatra && s.InXiu && s.InManazil {
			threeWayKeys[s.Key] = true
		}
	}

	// Build nakshatra star key lookup
	naks := NakshatraStars()
	naksKeyToName := make(map[string]string)
	for _, s := range naks {
		naksKeyToName[s.Key] = s.NakshatraName
	}

	// Build xiu star key lookup
	xius := XiuStars()
	xiuKeyToName := make(map[string]string)
	xiuKeyToPinyin := make(map[string]string)
	for _, s := range xius {
		xiuKeyToName[s.Key] = s.XiuName
		xiuKeyToPinyin[s.Key] = s.XiuPinyin
	}

	// Build manazil star key lookup
	manazils := ManazilStars()
	manKeyToName := make(map[string]string)
	for _, s := range manazils {
		manKeyToName[s.Key] = s.ManazilName
	}

	// Classical planets only
	planets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}

	var placements []MansionPlacement
	converging := 0

	for _, p := range planets {
		tropLon, ok := tropical[p]
		if !ok {
			continue
		}
		sidLon := normalizeLon(tropLon - ayanamsa)

		nakName, nakNum := NakshatraForLongitude(sidLon)
		xiuName, xiuNum, xiuPinyin := XiuForLongitude(sidLon)
		manName, manNum := ManazilForLongitude(sidLon)

		// Check three-way convergence
		converges := false
		for _, s := range naks {
			if s.NakshatraName == nakName {
				if threeWayKeys[s.Key] {
					// This nakshatra's star is three-way shared — check xiu and manazil
					for _, x := range xius {
						if x.Key == s.Key && x.XiuName == xiuName {
							for _, m := range manazils {
								if m.Key == s.Key && m.ManazilName == manName {
									converges = true
									break
								}
							}
							break
						}
					}
				}
				break
			}
		}

		placement := MansionPlacement{
			Planet:        p,
			TropicalLon:   tropLon,
			SiderealLon:   sidLon,
			Nakshatra:     nakName,
			NakshatraNum:  nakNum,
			Xiu:           xiuName,
			XiuNum:        xiuNum,
			XiuPinyin:     xiuPinyin,
			Manazil:       manName,
			ManazilNum:    manNum,
			Converges:     converges,
		}
		placements = append(placements, placement)
		if converges {
			converging++
		}
	}

	return &MansionConvergence{
		Name:       name,
		Planets:    placements,
		Converging: converging,
		Total:      len(placements),
	}
}

// ManazilForLongitude returns the manazil name and number for a given sidereal longitude.
// Manazil boundaries are midpoints between consecutive determinative stars,
// following the same convention as XiuForLongitude.
func ManazilForLongitude(sidLon float64) (string, int) {
	lon := normalizeLon(sidLon)
	stars := ManazilStars()

	for i := 0; i < 28; i++ {
		curr := stars[i]
		next := stars[(i+1)%28]

		// Midpoint between current and next star
		boundary := (curr.EclipticLon + next.EclipticLon) / 2.0
		if next.EclipticLon < curr.EclipticLon {
			boundary = (curr.EclipticLon + next.EclipticLon + 360) / 2.0
			if boundary >= 360 {
				boundary -= 360
			}
		}

		prev := stars[(i+27)%28]
		prevBoundary := (prev.EclipticLon + curr.EclipticLon) / 2.0
		if curr.EclipticLon < prev.EclipticLon {
			prevBoundary = (prev.EclipticLon + curr.EclipticLon + 360) / 2.0
			if prevBoundary >= 360 {
				prevBoundary -= 360
			}
		}

		if inSector(lon, prevBoundary, boundary) {
			return curr.ManazilName, curr.ManazilNum
		}
	}

	return stars[27].ManazilName, stars[27].ManazilNum
}
