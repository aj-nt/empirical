package dignity

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// ── Geometric Pattern Detection ──────────────────────────────────────────

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
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// PatternReport holds all detected patterns for a chart.
type PatternReport struct {
	Name     string    `json:"name"`
	Patterns []Pattern `json:"patterns"`
}

// DetectPatterns finds all geometric configurations in a set of planet longitudes.
// planets: planet name → ecliptic longitude (0-360)
// orbDeg: max orb for aspect matching (default 5°)
func DetectPatterns(planets map[string]float64, orbDeg float64) *PatternReport {
	if orbDeg <= 0 {
		orbDeg = 5.0
	}

	// Build planet list
	var names []string
	for n := range planets {
		names = append(names, n)
	}
	sort.Strings(names)

	// Compute all pairwise aspects
	type edge struct {
		p1, p2 string
		aspect string
		orb    float64
		angle  float64 // actual angular separation
	}
	var edges []edge
	aspectDefs := []struct {
		angle float64
		name  string
	}{
		{0, "conjunction"}, {30, "semi-sextile"}, {45, "semi-square"},
		{60, "sextile"}, {90, "square"}, {120, "trine"},
		{135, "sesquiquadrate"}, {150, "quincunx"}, {180, "opposition"},
	}

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			a, b := names[i], names[j]
			dist := angleDist(planets[a], planets[b])
			for _, ad := range aspectDefs {
				orb := math.Abs(dist - ad.angle)
				if orb <= orbDeg {
					edges = append(edges, edge{a, b, ad.name, math.Round(orb*100) / 100, dist})
					break // closest aspect only
				}
			}
		}
	}

	// Build adjacency: planet → list of (other planet, aspect, orb)
	adj := make(map[string][]edge)
	for _, e := range edges {
		adj[e.p1] = append(adj[e.p1], e)
		adj[e.p2] = append(adj[e.p2], edge{e.p2, e.p1, e.aspect, e.orb, e.angle})
	}

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

	// ── Stellium: 3+ planets within 8° in same sign ──────────────────
	stelliumOrb := 8.0
	// Group planets by sign
	signGroups := make(map[string][]string)
	for _, n := range names {
		sign := SignForLongitude(planets[n])
		signGroups[sign] = append(signGroups[sign], n)
	}
	for sign, group := range signGroups {
		if len(group) < 3 {
			continue
		}
		// Check if they're within stelliumOrb of each other
		// Sort by longitude within sign
		sort.Slice(group, func(i, j int) bool {
			return math.Mod(planets[group[i]], 30) < math.Mod(planets[group[j]], 30)
		})
		// Find maximal connected clusters
		visited := make(map[string]bool)
		for _, seed := range group {
			if visited[seed] {
				continue
			}
			cluster := []string{seed}
			visited[seed] = true
			for _, other := range group {
				if visited[other] {
					continue
				}
				// Check if other is within 8° of any planet in cluster
				for _, c := range cluster {
					if angleDist(planets[c], planets[other]) <= stelliumOrb {
						cluster = append(cluster, other)
						visited[other] = true
						break
					}
				}
			}
			if len(cluster) >= 3 {
				// Build aspect list for the stellium
				var paspects []PatternAspect
				for i := 0; i < len(cluster); i++ {
					for j := i + 1; j < len(cluster); j++ {
						ok, orb := hasAspect(cluster[i], cluster[j], "conjunction")
						if ok {
							paspects = append(paspects, PatternAspect{cluster[i], cluster[j], "conjunction", orb})
						} else {
							// Within orb but not classified as conjunction by our aspect defs
							dist := angleDist(planets[cluster[i]], planets[cluster[j]])
							paspects = append(paspects, PatternAspect{cluster[i], cluster[j], "conjunction", math.Round(dist*100) / 100})
						}
					}
				}
				patterns = append(patterns, Pattern{
					Kind:        Stellium,
					Name:        fmt.Sprintf("Stellium in %s", sign),
					Planets:     cluster,
					Description: fmt.Sprintf("%d planets clustered within %.0f° in %s — concentrated energy in one sign.", len(cluster), stelliumOrb, sign),
					Aspects:     paspects,
				})
			}
		}
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
						Description: fmt.Sprintf("%s, %s, %s form a Grand Trine — each pair in flowing 120° aspect. Self-reinforcing harmony in one element.",
							trio[0], trio[1], trio[2]),
						Aspects: []PatternAspect{
							{names[i], names[j], "trine", mustOrb(hasAspect(names[i], names[j], "trine"))},
							{names[i], names[k], "trine", mustOrb(hasAspect(names[i], names[k], "trine"))},
							{names[j], names[k], "trine", mustOrb(hasAspect(names[j], names[k], "trine"))},
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
						Description: fmt.Sprintf("Grand Trine (%s, %s, %s) with %s at the release point — opposite %s and sextile the other two. The trine's energy discharges through the opposition axis.",
							apex, base1, base2, fourth, apex),
						Aspects: []PatternAspect{
							{apex, base1, "trine", mustOrb(hasAspect(apex, base1, "trine"))},
							{apex, base2, "trine", mustOrb(hasAspect(apex, base2, "trine"))},
							{base1, base2, "trine", mustOrb(hasAspect(base1, base2, "trine"))},
							{fourth, apex, "opposition", mustOrb(hasAspect(fourth, apex, "opposition"))},
							{fourth, base1, "sextile", mustOrb(hasAspect(fourth, base1, "sextile"))},
							{fourth, base2, "sextile", mustOrb(hasAspect(fourth, base2, "sextile"))},
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
							{names[i], names[j], "opposition", mustOrb(hasAspect(names[i], names[j], "opposition"))},
							{names[k], names[i], "square", mustOrb(hasAspect(names[k], names[i], "square"))},
							{names[k], names[j], "square", mustOrb(hasAspect(names[k], names[j], "square"))},
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
								{names[i], names[j], "opposition", mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{names[k], names[l], "opposition", mustOrb(hasAspect(names[k], names[l], "opposition"))},
								{names[i], names[k], "square", mustOrb(hasAspect(names[i], names[k], "square"))},
								{names[i], names[l], "square", mustOrb(hasAspect(names[i], names[l], "square"))},
								{names[j], names[k], "square", mustOrb(hasAspect(names[j], names[k], "square"))},
								{names[j], names[l], "square", mustOrb(hasAspect(names[j], names[l], "square"))},
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
						Description: fmt.Sprintf("%s and %s sextile each other, both quincunx %s at the apex. A fated, uncomfortable pointing — the apex planet carries a special assignment.",
							names[i], names[j], names[k]),
						Aspects: []PatternAspect{
							{names[i], names[j], "sextile", mustOrb(hasAspect(names[i], names[j], "sextile"))},
							{names[k], names[i], "quincunx", mustOrb(hasAspect(names[k], names[i], "quincunx"))},
							{names[k], names[j], "quincunx", mustOrb(hasAspect(names[k], names[j], "quincunx"))},
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
						paspects = append(paspects, PatternAspect{names[i], names[j], "opposition", mustOrb(hasAspect(names[i], names[j], "opposition"))})
						paspects = append(paspects, PatternAspect{names[k], names[l], "opposition", mustOrb(hasAspect(names[k], names[l], "opposition"))})
						if config1 {
							paspects = append(paspects, PatternAspect{names[i], names[k], "sextile", mustOrb(hasAspect(names[i], names[k], "sextile"))})
							paspects = append(paspects, PatternAspect{names[j], names[l], "sextile", mustOrb(hasAspect(names[j], names[l], "sextile"))})
							paspects = append(paspects, PatternAspect{names[i], names[l], "trine", mustOrb(hasAspect(names[i], names[l], "trine"))})
							paspects = append(paspects, PatternAspect{names[j], names[k], "trine", mustOrb(hasAspect(names[j], names[k], "trine"))})
						} else {
							paspects = append(paspects, PatternAspect{names[i], names[l], "sextile", mustOrb(hasAspect(names[i], names[l], "sextile"))})
							paspects = append(paspects, PatternAspect{names[j], names[k], "sextile", mustOrb(hasAspect(names[j], names[k], "sextile"))})
							paspects = append(paspects, PatternAspect{names[i], names[k], "trine", mustOrb(hasAspect(names[i], names[k], "trine"))})
							paspects = append(paspects, PatternAspect{names[j], names[l], "trine", mustOrb(hasAspect(names[j], names[l], "trine"))})
						}
						patterns = append(patterns, Pattern{
							Kind:    MysticRectangle,
							Name:    "Mystic Rectangle",
							Planets: all,
							Description: fmt.Sprintf("Two oppositions (%s-%s, %s-%s) woven together by sextiles and trines. Balanced creative tension — a rare harmonizing structure.",
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
								{names[i], names[j], "opposition", mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{names[i], names[k], "sextile", mustOrb(hasAspect(names[i], names[k], "sextile"))},
								{names[k], names[j], "trine", mustOrb(hasAspect(names[k], names[j], "trine"))},
								{names[j], names[l], "sextile", mustOrb(hasAspect(names[j], names[l], "sextile"))},
								{names[l], names[i], "trine", mustOrb(hasAspect(names[l], names[i], "trine"))},
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
								{names[i], names[j], "opposition", mustOrb(hasAspect(names[i], names[j], "opposition"))},
								{names[i], names[l], "sextile", mustOrb(hasAspect(names[i], names[l], "sextile"))},
								{names[l], names[j], "trine", mustOrb(hasAspect(names[l], names[j], "trine"))},
								{names[j], names[k], "sextile", mustOrb(hasAspect(names[j], names[k], "sextile"))},
								{names[k], names[i], "trine", mustOrb(hasAspect(names[k], names[i], "trine"))},
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
							Description: fmt.Sprintf("%s, %s, %s form a right triangle: sextile + trine + square. Productive friction — the square provides drive, the sextile/trine provide flow.",
								trio[0], trio[1], trio[2]),
							Aspects: []PatternAspect{
								{trio[perm[0][0]], trio[perm[0][1]], "sextile", mustOrb(hasAspect(trio[perm[0][0]], trio[perm[0][1]], "sextile"))},
								{trio[perm[1][0]], trio[perm[1][1]], "trine", mustOrb(hasAspect(trio[perm[1][0]], trio[perm[1][1]], "trine"))},
								{trio[perm[2][0]], trio[perm[2][1]], "square", mustOrb(hasAspect(trio[perm[2][0]], trio[perm[2][1]], "square"))},
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

	return &PatternReport{
		Name:     "",
		Patterns: patterns,
	}
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
	b.WriteString(fmt.Sprintf("Geometric Pattern Report — %s\n\n", pr.Name))

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
			b.WriteString(fmt.Sprintf("    %s — %s %s (orb %.1f°)\n", a.Planet1, a.Planet2, a.Aspect, a.Orb))
		}
		b.WriteString("\n")
	}

	return b.String()
}
