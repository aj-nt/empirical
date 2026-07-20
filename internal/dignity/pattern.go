package dignity

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// ── Geometric Pattern Detection ──────────────────────────────────────────

// DefaultPatternOrb is the default maximum orb (in degrees) for pattern detection.
const DefaultPatternOrb = 3.0

// PatternKind names a known geometric configuration.
type PatternKind string

const (
	Stellium        PatternKind = "stellium"
	GrandTrine      PatternKind = "grand_trine"
	Kite            PatternKind = "kite"
	TSquare         PatternKind = "t_square"
	GrandCross      PatternKind = "grand_cross"
	Yod             PatternKind = "yod"
	MysticRectangle PatternKind = "mystic_rectangle"
	Cradle          PatternKind = "cradle"
	Wedge           PatternKind = "wedge"
)

// Pattern describes a detected geometric configuration.
type Pattern struct {
	Kind        PatternKind  `json:"kind"`
	Name        string       `json:"name"`
	Planets     []string     `json:"planets"`
	Description string       `json:"description"`
	Aspects     []PatternAspect `json:"aspects"`
}

// PatternAspect is one edge in the pattern.
type PatternAspect struct {
	Planet1   string  `json:"planet1"`
	Planet2   string  `json:"planet2"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
	StartDate string  `json:"start_date,omitempty"`
	EndDate   string  `json:"end_date,omitempty"`
	PeakOrb   float64 `json:"peak_orb,omitempty"`
}

// PatternReport holds all detected patterns for a chart.
type PatternReport struct {
	Name     string    `json:"name"`
	Patterns []Pattern `json:"patterns"`
}

// DetectPatterns finds all geometric configurations in a set of planet longitudes.
// planets: planet name → ecliptic longitude (0-360)
// orbDeg: max orb for aspect matching (default 3°)
func DetectPatterns(planets map[string]float64, orbDeg float64) *PatternReport {
	if orbDeg <= 0 {
		orbDeg = DefaultPatternOrb
	}

	names, edges, adj := buildEdges(planets, nil, orbDeg, false)
	patterns := detectPatternsFromEdges(names, edges, adj, planets)

	return &PatternReport{
		Name:     "",
		Patterns: patterns,
	}
}

// DetectPatternsWithStars finds geometric configurations that may include
// fixed stars as additional nodes. Stars aspect planets via conjunction,
// opposition, square, and trine only (matching FindStarConjunctions).
// Stars do not aspect each other. Planet-planet aspects use the full
// 6-aspect set (conjunction, sextile, square, trine, quincunx, opposition).
func DetectPatternsWithStars(planets, stars map[string]float64, orbDeg float64) *PatternReport {
	if orbDeg <= 0 {
		orbDeg = DefaultPatternOrb
	}

	names, edges, adj := buildEdges(planets, stars, orbDeg, true)
	// Build combined position map for stellium detection
	combined := make(map[string]float64)
	for k, v := range planets {
		combined[k] = v
	}
	for k, v := range stars {
		combined[k] = v
	}
	patterns := detectPatternsFromEdges(names, edges, adj, combined)

	return &PatternReport{
		Name:     "",
		Patterns: patterns,
	}
}

type edge struct {
	p1, p2 string
	aspect string
	orb    float64
	angle  float64 // actual angular separation
}

// buildEdges computes all pairwise aspects and builds the adjacency map.
// When includeStars is true, star positions are added as nodes that aspect
// planets via the 4 universal aspects (conjunction, opposition, square, trine)
// and do not aspect each other.
func buildEdges(planets, stars map[string]float64, orbDeg float64, includeStars bool) (names []string, edges []edge, adj map[string][]edge) {
	for n := range planets {
		names = append(names, n)
	}
	if includeStars {
		for n := range stars {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	// Build a lookup: is this name a star?
	isStar := make(map[string]bool)
	if includeStars {
		for n := range stars {
			isStar[n] = true
		}
	}

	// Full 6-aspect set for planet-planet
	allAspectDefs := []struct {
		angle float64
		name  string
	}{
		{0, "conjunction"}, {60, "sextile"}, {90, "square"},
		{120, "trine"}, {150, "quincunx"}, {180, "opposition"},
	}

	// 4-aspect set for star-planet (matches FindStarConjunctions)
	starAspectDefs := []struct {
		angle float64
		name  string
	}{
		{0, "conjunction"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
	}

	// Build combined position map
	allPos := make(map[string]float64)
	for k, v := range planets {
		allPos[k] = v
	}
	for k, v := range stars {
		allPos[k] = v
	}

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			a, b := names[i], names[j]

			// Stars don't aspect each other
			if isStar[a] && isStar[b] {
				continue
			}

			dist := angleDist(allPos[a], allPos[b])

			// Choose aspect set: star-involved pairs use 4 aspects
			aspectDefs := allAspectDefs
			if isStar[a] || isStar[b] {
				aspectDefs = starAspectDefs
			}

			for _, ad := range aspectDefs {
				orb := math.Abs(dist - ad.angle)
				if orb <= orbDeg {
					edges = append(edges, edge{a, b, ad.name, math.Round(orb*100) / 100, dist})
					break // closest aspect only
				}
			}
		}
	}

	// Build adjacency
	adj = make(map[string][]edge)
	for _, e := range edges {
		adj[e.p1] = append(adj[e.p1], e)
		adj[e.p2] = append(adj[e.p2], edge{e.p2, e.p1, e.aspect, e.orb, e.angle})
	}

	return
}

// detectPatternsFromEdges runs pattern detection given pre-computed edges and adjacency.
// allPos is the full position map (needed for stellium sign grouping).
func detectPatternsFromEdges(names []string, edges []edge, adj map[string][]edge, allPos map[string]float64) []Pattern {

	// Helper: does edge exist between two planets with a specific aspect?
	hasAspect := func(p1, p2, aspect string) (bool, float64) {
		for _, e := range adj[p1] {
			if e.p2 == p2 && e.aspect == aspect {
				return true, e.orb
			}
		}
		return false, 0
	}

	var patterns []Pattern

	// ── Stellium: 3+ bodies in same sign ──────────────────────────────
	// Group everything except stars by sign (stars don't count toward stelliums)
	starSet := make(map[string]bool, len(StarNames))
	for _, s := range StarNames {
		starSet[s] = true
	}
	signGroups := make(map[string][]string)
	for _, n := range names {
		if starSet[n] {
			continue
		}
		sign := SignForLongitude(allPos[n])
		signGroups[sign] = append(signGroups[sign], n)
	}
	for sign, group := range signGroups {
		if len(group) < 3 {
			continue
		}
		// Sort by longitude within sign for consistent ordering
		sort.Slice(group, func(i, j int) bool {
			return math.Mod(allPos[group[i]], 30) < math.Mod(allPos[group[j]], 30)
		})
		// Build aspect list for the stellium
		var paspects []PatternAspect
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				dist := angleDist(allPos[group[i]], allPos[group[j]])
				paspects = append(paspects, PatternAspect{Planet1: group[i], Planet2: group[j], Aspect: "conjunction", Orb: math.Round(dist*100) / 100})
			}
		}
		patterns = append(patterns, Pattern{
			Kind:        Stellium,
			Name:        fmt.Sprintf("Stellium in %s", sign),
			Planets:     group,
			Description: fmt.Sprintf("%d planets in %s. Concentrated energy in one sign.", len(group), sign),
			Aspects:     paspects,
		})
	}

	// ── Grand Trine: 3 planets, each pair in trine ───────────────────
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			ok1, _ := hasAspect(names[i], names[j], "trine")
			if !ok1 {
				continue
			}
			for k := j + 1; k < len(names); k++ {
				ok2, _ := hasAspect(names[i], names[k], "trine")
				ok3, _ := hasAspect(names[j], names[k], "trine")
				if ok2 && ok3 {
					trio := []string{names[i], names[j], names[k]}
					patterns = append(patterns, Pattern{
						Kind:    GrandTrine,
						Name:    "Grand Trine",
						Planets: trio,
						Description: fmt.Sprintf("%s, %s, %s form a Grand Trine. Each pair in flowing 120° aspect. Self-reinforcing harmony in one element.",
							trio[0], trio[1], trio[2]),
						Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "trine", Orb: mustOrb(hasAspect(names[i], names[j], "trine"))},
							{Planet1: names[i], Planet2: names[k], Aspect: "trine", Orb: mustOrb(hasAspect(names[i], names[k], "trine"))},
							{Planet1: names[j], Planet2: names[k], Aspect: "trine", Orb: mustOrb(hasAspect(names[j], names[k], "trine"))},
						},
					})
				}
			}
		}
	}

	// ── Kite: Grand Trine + 4th planet opposite one, sextile other two ─
	for _, p := range patterns {
		if p.Kind != GrandTrine {
			continue
		}
		trio := p.Planets
		for _, fourth := range names {
			if fourth == trio[0] || fourth == trio[1] || fourth == trio[2] {
				continue
			}
			// Try each trio member as the opposition point
			for apexIdx := 0; apexIdx < 3; apexIdx++ {
				apex := trio[apexIdx]
				base1 := trio[(apexIdx+1)%3]
				base2 := trio[(apexIdx+2)%3]

				oppOk, _ := hasAspect(fourth, apex, "opposition")
				sex1Ok, _ := hasAspect(fourth, base1, "sextile")
				sex2Ok, _ := hasAspect(fourth, base2, "sextile")

				if oppOk && sex1Ok && sex2Ok {
					all := []string{apex, base1, base2, fourth}
					patterns = append(patterns, Pattern{
						Kind:    Kite,
						Name:    "Kite",
						Planets: all,
						Description: fmt.Sprintf("Grand Trine (%s, %s, %s) with %s at the release point. Opposite %s and sextile the other two. The trine's energy discharges through the opposition axis.",
							apex, base1, base2, fourth, apex),
						Aspects: []PatternAspect{
							{Planet1: apex, Planet2: base1, Aspect: "trine", Orb: mustOrb(hasAspect(apex, base1, "trine"))},
							{Planet1: apex, Planet2: base2, Aspect: "trine", Orb: mustOrb(hasAspect(apex, base2, "trine"))},
							{Planet1: base1, Planet2: base2, Aspect: "trine", Orb: mustOrb(hasAspect(base1, base2, "trine"))},
							{Planet1: fourth, Planet2: apex, Aspect: "opposition", Orb: mustOrb(hasAspect(fourth, apex, "opposition"))},
							{Planet1: fourth, Planet2: base1, Aspect: "sextile", Orb: mustOrb(hasAspect(fourth, base1, "sextile"))},
							{Planet1: fourth, Planet2: base2, Aspect: "sextile", Orb: mustOrb(hasAspect(fourth, base2, "sextile"))},
						},
					})
				}
			}
		}
	}

	// ── T-Square: 2 planets in opposition, both square a 3rd ─────────
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			oppOk, _ := hasAspect(names[i], names[j], "opposition")
			if !oppOk {
				continue
			}
			for k := 0; k < len(names); k++ {
				if k == i || k == j {
					continue
				}
				sq1Ok, _ := hasAspect(names[k], names[i], "square")
				sq2Ok, _ := hasAspect(names[k], names[j], "square")
				if sq1Ok && sq2Ok {
					trio := []string{names[i], names[j], names[k]}
					patterns = append(patterns, Pattern{
						Kind:    TSquare,
						Name:    "T-Square",
						Planets: trio,
						Description: fmt.Sprintf("%s opposes %s, both square %s at the apex. Dynamic tension demanding resolution through the apex planet.",
							names[i], names[j], names[k]),
						Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "opposition", Orb: mustOrb(hasAspect(names[i], names[j], "opposition"))},
							{Planet1: names[k], Planet2: names[i], Aspect: "square", Orb: mustOrb(hasAspect(names[k], names[i], "square"))},
							{Planet1: names[k], Planet2: names[j], Aspect: "square", Orb: mustOrb(hasAspect(names[k], names[j], "square"))},
						},
					})
				}
			}
		}
	}

	// ── Grand Cross: 4 planets, two oppositions, all mutual squares ──
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			opp1Ok, _ := hasAspect(names[i], names[j], "opposition")
			if !opp1Ok {
				continue
			}
			for k := 0; k < len(names); k++ {
				if k == i || k == j {
					continue
				}
				for l := k + 1; l < len(names); l++ {
					if l == i || l == j {
						continue
					}
					opp2Ok, _ := hasAspect(names[k], names[l], "opposition")
					if !opp2Ok {
						continue
					}
					// Check all four cross-squares
					sqIK, _ := hasAspect(names[i], names[k], "square")
					sqIL, _ := hasAspect(names[i], names[l], "square")
					sqJK, _ := hasAspect(names[j], names[k], "square")
					sqJL, _ := hasAspect(names[j], names[l], "square")
					if sqIK && sqIL && sqJK && sqJL {
						all := []string{names[i], names[j], names[k], names[l]}
						patterns = append(patterns, Pattern{
							Kind:    GrandCross,
							Name:    "Grand Cross",
							Planets: all,
							Description: fmt.Sprintf("%s-%s and %s-%s form two oppositions, all four in mutual square. Intense dynamic pressure across four points.",
								names[i], names[j], names[k], names[l]),
							Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "opposition", Orb: mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{Planet1: names[k], Planet2: names[l], Aspect: "opposition", Orb: mustOrb(hasAspect(names[k], names[l], "opposition"))},
								{Planet1: names[i], Planet2: names[k], Aspect: "square", Orb: mustOrb(hasAspect(names[i], names[k], "square"))},
								{Planet1: names[i], Planet2: names[l], Aspect: "square", Orb: mustOrb(hasAspect(names[i], names[l], "square"))},
								{Planet1: names[j], Planet2: names[k], Aspect: "square", Orb: mustOrb(hasAspect(names[j], names[k], "square"))},
								{Planet1: names[j], Planet2: names[l], Aspect: "square", Orb: mustOrb(hasAspect(names[j], names[l], "square"))},
							},
						})
					}
				}
			}
		}
	}

	// ── Yod: 2 planets sextile, both quincunx a 3rd ──────────────────
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			sexOk, _ := hasAspect(names[i], names[j], "sextile")
			if !sexOk {
				continue
			}
			for k := 0; k < len(names); k++ {
				if k == i || k == j {
					continue
				}
				q1Ok, _ := hasAspect(names[k], names[i], "quincunx")
				q2Ok, _ := hasAspect(names[k], names[j], "quincunx")
				if q1Ok && q2Ok {
					trio := []string{names[i], names[j], names[k]}
					patterns = append(patterns, Pattern{
						Kind:    Yod,
						Name:    "Yod (Finger of God)",
						Planets: trio,
						Description: fmt.Sprintf("%s and %s sextile each other, both quincunx %s at the apex. A fated, uncomfortable pointing. The apex planet carries a special assignment.",
							names[i], names[j], names[k]),
						Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "sextile", Orb: mustOrb(hasAspect(names[i], names[j], "sextile"))},
							{Planet1: names[k], Planet2: names[i], Aspect: "quincunx", Orb: mustOrb(hasAspect(names[k], names[i], "quincunx"))},
							{Planet1: names[k], Planet2: names[j], Aspect: "quincunx", Orb: mustOrb(hasAspect(names[k], names[j], "quincunx"))},
						},
					})
				}
			}
		}
	}

	// ── Mystic Rectangle: 2 oppositions connected by sextiles + trines ─
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			opp1Ok, _ := hasAspect(names[i], names[j], "opposition")
			if !opp1Ok {
				continue
			}
			for k := 0; k < len(names); k++ {
				if k == i || k == j {
					continue
				}
				for l := k + 1; l < len(names); l++ {
					if l == i || l == j {
						continue
					}
					opp2Ok, _ := hasAspect(names[k], names[l], "opposition")
					if !opp2Ok {
						continue
					}
					// Check sextile connections between the two oppositions
					// Two possible configurations: i-k sextile + j-l sextile, or i-l sextile + j-k sextile
					config1 := false
					sexIK, _ := hasAspect(names[i], names[k], "sextile")
					sexJL, _ := hasAspect(names[j], names[l], "sextile")
					if sexIK && sexJL {
						// Check trines for the other cross-pairs
						trIL, _ := hasAspect(names[i], names[l], "trine")
						trJK, _ := hasAspect(names[j], names[k], "trine")
						if trIL && trJK {
							config1 = true
						}
					}
					config2 := false
					sexIL, _ := hasAspect(names[i], names[l], "sextile")
					sexJK, _ := hasAspect(names[j], names[k], "sextile")
					if sexIL && sexJK {
						trIK, _ := hasAspect(names[i], names[k], "trine")
						trJL, _ := hasAspect(names[j], names[l], "trine")
						if trIK && trJL {
							config2 = true
						}
					}
					if config1 || config2 {
						all := []string{names[i], names[j], names[k], names[l]}
						var paspects []PatternAspect
						paspects = append(paspects, PatternAspect{Planet1: names[i], Planet2: names[j], Aspect: "opposition", Orb: mustOrb(hasAspect(names[i], names[j], "opposition"))})
						paspects = append(paspects, PatternAspect{Planet1: names[k], Planet2: names[l], Aspect: "opposition", Orb: mustOrb(hasAspect(names[k], names[l], "opposition"))})
						if config1 {
							paspects = append(paspects, PatternAspect{Planet1: names[i], Planet2: names[k], Aspect: "sextile", Orb: mustOrb(hasAspect(names[i], names[k], "sextile"))})
							paspects = append(paspects, PatternAspect{Planet1: names[j], Planet2: names[l], Aspect: "sextile", Orb: mustOrb(hasAspect(names[j], names[l], "sextile"))})
							paspects = append(paspects, PatternAspect{Planet1: names[i], Planet2: names[l], Aspect: "trine", Orb: mustOrb(hasAspect(names[i], names[l], "trine"))})
							paspects = append(paspects, PatternAspect{Planet1: names[j], Planet2: names[k], Aspect: "trine", Orb: mustOrb(hasAspect(names[j], names[k], "trine"))})
						} else {
							paspects = append(paspects, PatternAspect{Planet1: names[i], Planet2: names[l], Aspect: "sextile", Orb: mustOrb(hasAspect(names[i], names[l], "sextile"))})
							paspects = append(paspects, PatternAspect{Planet1: names[j], Planet2: names[k], Aspect: "sextile", Orb: mustOrb(hasAspect(names[j], names[k], "sextile"))})
							paspects = append(paspects, PatternAspect{Planet1: names[i], Planet2: names[k], Aspect: "trine", Orb: mustOrb(hasAspect(names[i], names[k], "trine"))})
							paspects = append(paspects, PatternAspect{Planet1: names[j], Planet2: names[l], Aspect: "trine", Orb: mustOrb(hasAspect(names[j], names[l], "trine"))})
						}
						patterns = append(patterns, Pattern{
							Kind:    MysticRectangle,
							Name:    "Mystic Rectangle",
							Planets: all,
							Description: fmt.Sprintf("Two oppositions (%s-%s, %s-%s) woven together by sextiles and trines. Balanced creative tension. A rare harmonizing structure.",
								names[i], names[j], names[k], names[l]),
							Aspects: paspects,
						})
					}
				}
			}
		}
	}

	// ── Cradle: opposition + two sextiles + two trines (half mystic rectangle) ─
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			oppOk, _ := hasAspect(names[i], names[j], "opposition")
			if !oppOk {
				continue
			}
			for k := 0; k < len(names); k++ {
				if k == i || k == j {
					continue
				}
				for l := k + 1; l < len(names); l++ {
					if l == i || l == j {
						continue
					}
					// Check: i sextile k, k trine j, j sextile l, l trine i
					sexIK, _ := hasAspect(names[i], names[k], "sextile")
					trKJ, _ := hasAspect(names[k], names[j], "trine")
					sexJL, _ := hasAspect(names[j], names[l], "sextile")
					trLI, _ := hasAspect(names[l], names[i], "trine")
					if sexIK && trKJ && sexJL && trLI {
						all := []string{names[i], names[j], names[k], names[l]}
						patterns = append(patterns, Pattern{
							Kind:    Cradle,
							Name:    "Cradle",
							Planets: all,
							Description: fmt.Sprintf("%s opposes %s, with %s and %s forming sextile-trine bridges on each side. A supportive container for the opposition's tension.",
								names[i], names[j], names[k], names[l]),
							Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "opposition", Orb: mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{Planet1: names[i], Planet2: names[k], Aspect: "sextile", Orb: mustOrb(hasAspect(names[i], names[k], "sextile"))},
								{Planet1: names[k], Planet2: names[j], Aspect: "trine", Orb: mustOrb(hasAspect(names[k], names[j], "trine"))},
								{Planet1: names[j], Planet2: names[l], Aspect: "sextile", Orb: mustOrb(hasAspect(names[j], names[l], "sextile"))},
								{Planet1: names[l], Planet2: names[i], Aspect: "trine", Orb: mustOrb(hasAspect(names[l], names[i], "trine"))},
							},
						})
					}
					// Also check the mirror: i sextile l, l trine j, j sextile k, k trine i
					sexIL, _ := hasAspect(names[i], names[l], "sextile")
					trLJ, _ := hasAspect(names[l], names[j], "trine")
					sexJK, _ := hasAspect(names[j], names[k], "sextile")
					trKI, _ := hasAspect(names[k], names[i], "trine")
					if sexIL && trLJ && sexJK && trKI {
						all := []string{names[i], names[j], names[k], names[l]}
						patterns = append(patterns, Pattern{
							Kind:    Cradle,
							Name:    "Cradle",
							Planets: all,
							Description: fmt.Sprintf("%s opposes %s, with %s and %s forming sextile-trine bridges on each side. A supportive container for the opposition's tension.",
								names[i], names[j], names[l], names[k]),
							Aspects: []PatternAspect{
							{Planet1: names[i], Planet2: names[j], Aspect: "opposition", Orb: mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{Planet1: names[i], Planet2: names[l], Aspect: "sextile", Orb: mustOrb(hasAspect(names[i], names[l], "sextile"))},
								{Planet1: names[l], Planet2: names[j], Aspect: "trine", Orb: mustOrb(hasAspect(names[l], names[j], "trine"))},
								{Planet1: names[j], Planet2: names[k], Aspect: "sextile", Orb: mustOrb(hasAspect(names[j], names[k], "sextile"))},
								{Planet1: names[k], Planet2: names[i], Aspect: "trine", Orb: mustOrb(hasAspect(names[k], names[i], "trine"))},
							},
						})
					}
				}
			}
		}
	}

	// ── Wedge: sextile + trine + square (right triangle) ─────────────
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			for k := j + 1; k < len(names); k++ {
				// Check all 3 permutations of (sextile, trine, square)
				perms := [][3][2]int{
					{{0, 1}, {1, 2}, {0, 2}}, // i-j sextile, j-k trine, i-k square
					{{0, 1}, {0, 2}, {1, 2}}, // i-j sextile, i-k trine, j-k square
					{{1, 2}, {0, 2}, {0, 1}}, // j-k sextile, i-k trine, i-j square
				}
				trio := []string{names[i], names[j], names[k]}
				for _, perm := range perms {
					sexOk, _ := hasAspect(trio[perm[0][0]], trio[perm[0][1]], "sextile")
					trOk, _ := hasAspect(trio[perm[1][0]], trio[perm[1][1]], "trine")
					sqOk, _ := hasAspect(trio[perm[2][0]], trio[perm[2][1]], "square")
					if sexOk && trOk && sqOk {
						patterns = append(patterns, Pattern{
							Kind:    Wedge,
							Name:    "Wedge",
							Planets: trio,
							Description: fmt.Sprintf("%s, %s, %s form a right triangle: sextile + trine + square. Productive friction. The square provides drive, the sextile/trine provide flow.",
								trio[0], trio[1], trio[2]),
							Aspects: []PatternAspect{
							{Planet1: trio[perm[0][0]], Planet2: trio[perm[0][1]], Aspect: "sextile", Orb: mustOrb(hasAspect(trio[perm[0][0]], trio[perm[0][1]], "sextile"))},
								{Planet1: trio[perm[1][0]], Planet2: trio[perm[1][1]], Aspect: "trine", Orb: mustOrb(hasAspect(trio[perm[1][0]], trio[perm[1][1]], "trine"))},
								{Planet1: trio[perm[2][0]], Planet2: trio[perm[2][1]], Aspect: "square", Orb: mustOrb(hasAspect(trio[perm[2][0]], trio[perm[2][1]], "square"))},
							},
						})
						break // one permutation is enough
					}
				}
			}
		}
	}

	// Deduplicate: remove patterns that are subsets of larger patterns of the same kind
	patterns = deduplicatePatterns(patterns)

	// Sort by kind then by planet count
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].Kind != patterns[j].Kind {
			return patternKindOrder(patterns[i].Kind) < patternKindOrder(patterns[j].Kind)
		}
		return len(patterns[i].Planets) > len(patterns[j].Planets)
	})

	return patterns
}

// mustOrb extracts the orb from hasAspect result, or returns 0.
func mustOrb(ok bool, orb float64) float64 {
	if !ok {
		return 0
	}
	return orb
}

// patternKindOrder returns a sort order for pattern kinds.
func patternKindOrder(k PatternKind) int {
	switch k {
	case Stellium:
		return 0
	case GrandTrine:
		return 1
	case Kite:
		return 2
	case TSquare:
		return 3
	case GrandCross:
		return 4
	case Yod:
		return 5
	case MysticRectangle:
		return 6
	case Cradle:
		return 7
	case Wedge:
		return 8
	default:
		return 9
	}
}

// deduplicatePatterns removes patterns that are subsets of larger patterns of the same kind.
func deduplicatePatterns(patterns []Pattern) []Pattern {
	var out []Pattern
	for i, p := range patterns {
		subset := false
		for j, q := range patterns {
			if i == j {
				continue
			}
			if p.Kind != q.Kind {
				continue
			}
			if isSubset(p.Planets, q.Planets) {
				subset = true
				break
			}
		}
		if !subset {
			out = append(out, p)
		}
	}
	return out
}

// isSubset returns true if a's planets are all contained in b's planets.
func isSubset(a, b []string) bool {
	bSet := make(map[string]bool)
	for _, p := range b {
		bSet[p] = true
	}
	for _, p := range a {
		if !bSet[p] {
			return false
		}
	}
	return len(a) < len(b)
}

// PatternReportJSON serializes a pattern report to JSON.
func (pr *PatternReport) PatternReportJSON() ([]byte, error) {
	return json.MarshalIndent(pr, "", "  ")
}

// FormatPatternReport returns a human-readable pattern report.
func FormatPatternReport(pr *PatternReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Geometric Pattern Report .  %s\n\n", pr.Name))

	if len(pr.Patterns) == 0 {
		b.WriteString("No geometric patterns detected.\n")
		return b.String()
	}

	for _, p := range pr.Patterns {
		b.WriteString(fmt.Sprintf("▸ %s (%s)\n", p.Name, p.Kind))
		b.WriteString(fmt.Sprintf("  Planets: %s\n", strings.Join(p.Planets, ", ")))
		b.WriteString(fmt.Sprintf("  %s\n", p.Description))
		b.WriteString("  Aspects:\n")
		for _, a := range p.Aspects {
			b.WriteString(fmt.Sprintf("    %s .  %s %s (orb %.1f°)\n", a.Planet1, a.Planet2, a.Aspect, a.Orb))
		}
		b.WriteString("\n")
	}

	return b.String()
}
