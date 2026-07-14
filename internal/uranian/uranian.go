package uranian

import (
	"math"
	"sort"

	"github.com/aj-nt/empirical/internal/dignity"
)

// ── Midpoint / Half-Sum ─────────────────────────────────────────────────

// Midpoint calculates the direct half-sum of two ecliptic positions.
// Handles 360° wraparound by taking the shorter arc.
func Midpoint(a, b float64) float64 {
	diff := math.Mod(b-a+360, 360)
	if diff > 180 {
		diff -= 360
	}
	return math.Mod(a+diff/2+360, 360)
}

// IndirectHalfSums returns all 4 sensitive points for a pair:
// direct half-sum + 90°, 180°, 270° offsets.
func IndirectHalfSums(a, b float64) []float64 {
	direct := Midpoint(a, b)
	pts := []float64{
		math.Round(direct*1e4) / 1e4,
		math.Round(math.Mod(direct+90, 360)*1e4) / 1e4,
		math.Round(math.Mod(direct+180, 360)*1e4) / 1e4,
		math.Round(math.Mod(direct+270, 360)*1e4) / 1e4,
	}
	sort.Float64s(pts)
	return pts
}

// PlanetaryPicture computes A + B - C (three-factor combination).
func PlanetaryPicture(a, b, c float64) float64 {
	return math.Round(math.Mod(a+b-c+360, 360)*1e4) / 1e4
}

// ── 90-Degree Dial ──────────────────────────────────────────────────────

// ToDial converts an ecliptic longitude to its 90-degree dial position.
func ToDial(lon float64) float64 {
	return math.Round(math.Mod(lon, 90)*1e4) / 1e4
}

// DialPosition returns detailed dial position with sign context.
type DialPosition struct {
	DialDeg     float64 `json:"dial_deg"`
	Sign        string  `json:"sign"`
	SignDeg     float64 `json:"sign_deg"`
	OriginalLon float64 `json:"original_lon"`
}

func GetDialPosition(lon float64) DialPosition {
	signIdx := int(lon / 30) % 12
	signDeg := math.Mod(lon, 30)
	return DialPosition{
		DialDeg:     ToDial(lon),
		Sign:        dignity.Signs[signIdx],
		SignDeg:     math.Round(signDeg*1e4) / 1e4,
		OriginalLon: math.Round(lon*1e4) / 1e4,
	}
}

// DialEntry is a factor sorted by dial position.
type DialEntry struct {
	Name    string  `json:"name"`
	DialDeg float64 `json:"dial_deg"`
}

// DialSort returns factors sorted by dial position.
func DialSort(factors map[string]float64) []DialEntry {
	entries := make([]DialEntry, 0, len(factors))
	for name, lon := range factors {
		entries = append(entries, DialEntry{Name: name, DialDeg: ToDial(lon)})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].DialDeg != entries[j].DialDeg {
			return entries[i].DialDeg < entries[j].DialDeg
		}
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// ── Harmonic Midpoint Pictures ──────────────────────────────────────────

// HarmonicOffsets maps harmonic names to their dial offsets.
var HarmonicOffsets = map[string][]float64{
	"4th":  {0.0},
	"8th":  {45.0},
	"16th": {22.5, 67.5},
}

// HarmonicNames maps harmonic names to human-readable descriptions.
var HarmonicNames = map[string]string{
	"4th":  "Conjunction/Square (0 offset)",
	"8th":  "Semi-square (45 offset)",
	"16th": "Semi-semi-square (22.5/67.5 offset)",
}

// MidpointPicture represents a midpoint structure found on the 90° dial.
type MidpointPicture struct {
	FactorA   string  `json:"factor_a"`
	FactorB   string  `json:"factor_b"`
	Activator string  `json:"activator"`
	Harmonic  string  `json:"harmonic"`
	Offset    float64 `json:"offset"`
	Orb       float64 `json:"orb"`
}

// FindMidpointPictures finds all midpoint pictures across all harmonics.
// excludeCuspCusp suppresses pictures where all three factors are house cusps.
func FindMidpointPictures(factors map[string]float64, tolerance float64, excludeCuspCusp bool) []MidpointPicture {
	names := sortedKeys(factors)
	var results []MidpointPicture

	for i, nameA := range names {
		for _, nameB := range names[i+1:] {
			for _, nameC := range names {
				if nameC == nameA || nameC == nameB {
					continue
				}
				if excludeCuspCusp && isHouseCusp(nameA) && isHouseCusp(nameB) && isHouseCusp(nameC) {
					continue
				}

				hs := Midpoint(factors[nameA], factors[nameB])
				hsDial := ToDial(hs)
				cDial := ToDial(factors[nameC])

				for harmonic, offsets := range HarmonicOffsets {
					for _, offset := range offsets {
						diff := math.Mod(hsDial-cDial-offset+90, 90)
						if diff > 45 {
							diff = 90 - diff
						}
						if diff <= tolerance {
							results = append(results, MidpointPicture{
								FactorA:   nameA,
								FactorB:   nameB,
								Activator: nameC,
								Harmonic:  harmonic,
								Offset:    offset,
								Orb:       math.Round(diff*1e4) / 1e4,
							})
						}
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Orb < results[j].Orb })
	return results
}

// FindTightPictures returns only midpoint pictures with orb <= maxOrb.
func FindTightPictures(factors map[string]float64, maxOrb float64) []MidpointPicture {
	all := FindMidpointPictures(factors, maxOrb, true)
	var tight []MidpointPicture
	for _, p := range all {
		if p.Orb <= maxOrb {
			tight = append(tight, p)
		}
	}
	return tight
}

// ── Planetary Picture Activations ───────────────────────────────────────

// Activation represents a planetary picture that activates a target position.
type Activation struct {
	FactorA    string  `json:"factor_a"`
	FactorB    string  `json:"factor_b"`
	FactorC    string  `json:"factor_c"`
	PictureLon float64 `json:"picture_lon"`
	TargetLon  float64 `json:"target_lon"`
	Orb        float64 `json:"orb"`
}

// FindActivations finds all A+B-C pictures that match target longitudes on the dial.
func FindActivations(targetLongitudes []float64, factors map[string]float64, tolerance float64) []Activation {
	names := sortedKeys(factors)
	var activations []Activation

	for _, targetLon := range targetLongitudes {
		targetDial := ToDial(targetLon)
		for i, nameA := range names {
			for _, nameB := range names[i+1:] {
				for _, nameC := range names {
					if nameC == nameA || nameC == nameB {
						continue
					}
					picLon := PlanetaryPicture(factors[nameA], factors[nameB], factors[nameC])
					picDial := ToDial(picLon)
					diff := math.Abs(picDial - targetDial)
					orb := math.Min(diff, 90-diff)
					if orb <= tolerance {
						activations = append(activations, Activation{
							FactorA:    nameA,
							FactorB:    nameB,
							FactorC:    nameC,
							PictureLon: picLon,
							TargetLon:  targetLon,
							Orb:        math.Round(orb*1e4) / 1e4,
						})
					}
				}
			}
		}
	}
	return activations
}

// ── All Pictures for Target ─────────────────────────────────────────────

// PictureMatch represents an A+B-C combination matching a target.
type PictureMatch struct {
	FactorA   string  `json:"factor_a"`
	FactorB   string  `json:"factor_b"`
	FactorC   string  `json:"factor_c"`
	ResultLon float64 `json:"result_lon"`
}

// AllPicturesForTarget finds all A+B-C combinations matching a target within tolerance.
func AllPicturesForTarget(target float64, factors map[string]float64, tolerance float64) []PictureMatch {
	names := sortedKeys(factors)
	var matches []PictureMatch

	for i, nameA := range names {
		for _, nameB := range names[i+1:] {
			for _, nameC := range names {
				result := PlanetaryPicture(factors[nameA], factors[nameB], factors[nameC])
				diff := math.Abs(result - target)
				if diff > 180 {
					diff = 360 - diff
				}
				if diff <= tolerance {
					matches = append(matches, PictureMatch{
						FactorA:   nameA,
						FactorB:   nameB,
						FactorC:   nameC,
						ResultLon: result,
					})
				}
			}
		}
	}
	return matches
}

// ── Helpers ─────────────────────────────────────────────────────────────

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isHouseCusp(name string) bool {
	if len(name) < 2 || len(name) > 3 {
		return false
	}
	if name[0] != 'H' {
		return false
	}
	// Parse the number after H
	num := 0
	for i := 1; i < len(name); i++ {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
		num = num*10 + int(name[i]-'0')
	}
	return num >= 1 && num <= 12
}

// ── Uranian Report ──────────────────────────────────────────────────────

// UranianReport is the full Uranian/Hamburg School analysis.
type UranianReport struct {
	Name             string            `json:"name"`
	DialPositions    []DialEntry       `json:"dial_positions"`
	MidpointPictures []MidpointPicture `json:"midpoint_pictures"`
	TightPictures    []MidpointPicture `json:"tight_pictures"`
	Activations      []Activation      `json:"activations"`
}

// ComputeUranianReport computes the full Uranian analysis.
func ComputeUranianReport(name string, planets map[string]float64, houses map[string]float64) UranianReport {
	// Combine planets and house cusps as factors
	factors := make(map[string]float64)
	for k, v := range planets {
		factors[k] = v
	}
	for k, v := range houses {
		factors[k] = v
	}

	// Target longitudes: all planet positions
	targets := make([]float64, 0, len(planets))
	for _, lon := range planets {
		targets = append(targets, lon)
	}

	pics := FindMidpointPictures(factors, 1.0, true)
	tight := FindTightPictures(factors, 0.5)
	acts := FindActivations(targets, factors, 1.0)

	if pics == nil {
		pics = []MidpointPicture{}
	}
	if tight == nil {
		tight = []MidpointPicture{}
	}
	if acts == nil {
		acts = []Activation{}
	}

	return UranianReport{
		Name:             name,
		DialPositions:    DialSort(factors),
		MidpointPictures: pics,
		TightPictures:    tight,
		Activations:      acts,
	}
}

// ── General Midpoint Analysis (Ebertin-style) ─────────────────────────────

// DirectMidpointHit represents a planet occupying a midpoint of two other objects.
type DirectMidpointHit struct {
	PairA  string  `json:"pair_a"`
	PairB  string  `json:"pair_b"`
	Planet string  `json:"planet"`
	Orb    float64 `json:"orb"`
}

// FindDirectMidpoints finds all direct midpoint hits in a set of objects.
// For every pair (A, B), checks if any third object C sits at the A/B midpoint
// within maxOrb degrees. Results are sorted by orb ascending.
// A planet is never a hit for a pair it's part of.
func FindDirectMidpoints(objects map[string]float64, maxOrb float64) []DirectMidpointHit {
	names := sortedKeys(objects)
	var hits []DirectMidpointHit

	for i, nameA := range names {
		for _, nameB := range names[i+1:] {
			mp := Midpoint(objects[nameA], objects[nameB])
			for _, nameC := range names {
				if nameC == nameA || nameC == nameB {
					continue
				}
				orb := angularDistance(mp, objects[nameC])
				if orb <= maxOrb {
					hits = append(hits, DirectMidpointHit{
						PairA:  nameA,
						PairB:  nameB,
						Planet: nameC,
						Orb:    math.Round(orb*1e4) / 1e4,
					})
				}
			}
		}
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].Orb < hits[j].Orb })
	return hits
}

// angularDistance returns the shortest angular distance between two longitudes.
func angularDistance(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// ── Midpoint Report (Ebertin + Uranian combined) ──────────────────────────

// MidpointReport combines general midpoint analysis with Uranian harmonic pictures.
type MidpointReport struct {
	Name             string              `json:"name"`
	DialPositions    []DialEntry         `json:"dial_positions"`
	DirectMidpoints  []DirectMidpointHit `json:"direct_midpoints"`
	MidpointPictures []MidpointPicture   `json:"midpoint_pictures"`
	TightPictures    []MidpointPicture   `json:"tight_pictures"`
	Activations      []Activation        `json:"activations"`
}

// ComputeMidpointReport computes the full midpoint analysis including angles.
// planets: planet name → ecliptic longitude
// houses: house cusp name (e.g., "H1") → ecliptic longitude
// angles: angle name (e.g., "ASC", "MC", "DSC", "IC") → ecliptic longitude
func ComputeMidpointReport(name string, planets, houses, angles map[string]float64) MidpointReport {
	// Combine all factors
	factors := make(map[string]float64)
	for k, v := range planets {
		factors[k] = v
	}
	for k, v := range houses {
		factors[k] = v
	}
	for k, v := range angles {
		factors[k] = v
	}

	// Target longitudes: all planet + angle positions
	targets := make([]float64, 0, len(planets)+len(angles))
	for _, lon := range planets {
		targets = append(targets, lon)
	}
	for _, lon := range angles {
		targets = append(targets, lon)
	}

	direct := FindDirectMidpoints(factors, 1.0)
	pics := FindMidpointPictures(factors, 1.0, true)
	tight := FindTightPictures(factors, 0.5)
	acts := FindActivations(targets, factors, 1.0)

	if direct == nil {
		direct = []DirectMidpointHit{}
	}
	if pics == nil {
		pics = []MidpointPicture{}
	}
	if tight == nil {
		tight = []MidpointPicture{}
	}
	if acts == nil {
		acts = []Activation{}
	}

	return MidpointReport{
		Name:             name,
		DialPositions:    DialSort(factors),
		DirectMidpoints:  direct,
		MidpointPictures: pics,
		TightPictures:    tight,
		Activations:      acts,
	}
}
