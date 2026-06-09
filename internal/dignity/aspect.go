package dignity

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ── Phase 2: Aspect Geometry Convergence ─────────────────────────────────

// EarthlyBranches is the ordered list of Chinese Earthly Branches in zodiac order.
var EarthlyBranches = []string{
	"Zi", "Chou", "Yin", "Mao", "Chen", "Si",
	"Wu", "Wei", "Shen", "You", "Xu", "Hai",
}

// BranchAngle returns the angular separation (0-180 deg) between two
// Earthly Branches. Each branch spans 30 degrees.
func BranchAngle(b1, b2 string) int {
	idx1 := indexOf(EarthlyBranches, b1)
	idx2 := indexOf(EarthlyBranches, b2)
	diff := idx1 - idx2
	if diff < 0 {
		diff = -diff
	}
	if diff > 6 {
		diff = 12 - diff
	}
	return diff * 30
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// AspectEntry describes one aspect angle and its universality across traditions.
type AspectEntry struct {
	Name         string
	AngleDegrees int
	Universality string // "universal", "partial", "single_source"
}

// AspectCatalog returns the catalog of aspect angle convergence across Western,
// Vedic, and Chinese traditions.
func AspectCatalog() []AspectEntry {
	allAngles := []int{0, 30, 60, 90, 120, 150, 180}
	catalog := make([]AspectEntry, len(allAngles))

	nameMap := map[int]string{
		0: "conjunction", 30: "semi-sextile", 60: "sextile",
		90: "square", 120: "trine", 150: "quincunx", 180: "opposition",
	}

	for i, angle := range allAngles {
		// Chinese classification
		var chinese string
		switch angle {
		case 0:
			chinese = "explicit" // 三会 (same-element trio) + same branch
		case 120:
			chinese = "explicit" // 三合 (three harmonies)
		case 180:
			chinese = "explicit" // 六冲 (six clashes)
		case 90:
			chinese = "implicit" // 刑 (punishment)
		case 30, 150:
			chinese = "partial" // 六合 subset
		default:
			chinese = "absent"
		}

		// Western and Vedic recognize all seven standard angles
		var universality string
		switch {
		case chinese == "explicit":
			universality = "universal"
		case chinese == "implicit":
			universality = "partial"
		case chinese == "partial":
			universality = "partial"
		default:
			universality = "partial"
		}

		catalog[i] = AspectEntry{
			Name:         nameMap[angle],
			AngleDegrees: angle,
			Universality: universality,
		}
	}

	return catalog
}

// FormatAspectCatalog formats the aspect convergence catalog as human-readable text.
func FormatAspectCatalog() string {
	catalog := AspectCatalog()
	sort.Slice(catalog, func(i, j int) bool {
		return catalog[i].AngleDegrees < catalog[j].AngleDegrees
	})

	var b strings.Builder
	b.WriteString("Aspect Angle Convergence Catalog\n\n")
	b.WriteString(fmt.Sprintf("%-8s %-16s %-10s %-10s %-12s %s\n",
		"Angle", "Name", "Western", "Vedic", "Chinese", "Universality"))
	b.WriteString(strings.Repeat("-", 68) + "\n")

	chMap := map[int]string{
		0: "三会/同辰", 30: "partial (六合)", 60: "absent", 90: "implicit (刑)",
		120: "三合", 150: "partial (六合)", 180: "六冲",
	}

	for _, a := range catalog {
		w := "yes"
		v := "yes"
		ch := chMap[a.AngleDegrees]
		univ := fmt.Sprintf("SIGNAL (%s)", a.Universality)
		b.WriteString(fmt.Sprintf("%-8d %-16s %-10s %-10s %-12s %s\n",
			a.AngleDegrees, a.Name, w, v, ch, univ))
	}

	b.WriteString("\n")

	// Summary
	var universal, partial []AspectEntry
	for _, a := range catalog {
		if a.Universality == "universal" {
			universal = append(universal, a)
		} else {
			partial = append(partial, a)
		}
	}

	var uniParts, partParts []string
	for _, a := range universal {
		uniParts = append(uniParts, fmt.Sprintf("%s (%d deg)", a.Name, a.AngleDegrees))
	}
	for _, a := range partial {
		partParts = append(partParts, fmt.Sprintf("%s (%d deg)", a.Name, a.AngleDegrees))
	}

	b.WriteString(fmt.Sprintf("UNIVERSAL (%d): %s\n", len(universal), strings.Join(uniParts, ", ")))
	b.WriteString(fmt.Sprintf("PARTIAL  (%d): %s\n\n", len(partial), strings.Join(partParts, ", ")))

	b.WriteString("Vedic drishti note: Graha drishti (Mars 4/7/8, Jupiter 5/7/9, " +
		"Saturn 3/7/10, all 7) uses whole-sign relationships — a " +
		"parallel aspect system possibly preserving an older layer.\n\n")

	b.WriteString("RECOVERY IMPLICATION: Trine (120), opposition (180), and " +
		"conjunction (0) are the deepest invariants. Square (90) is " +
		"nearly universal if Chinese 刑 is counted. Sextile (60) is " +
		"Western/Vedic only — a candidate for later accretion or " +
		"lost in the Chinese fragment.\n")

	return b.String()
}

// FormatAspectJSON returns the aspect catalog as indented JSON.
func FormatAspectJSON(catalog []AspectEntry) ([]byte, error) {
	type out struct {
		Angles       []AspectEntry `json:"angles"`
		Universal    int           `json:"universal_count"`
		Partial      int           `json:"partial_count"`
		Summary      string        `json:"summary"`
	}
	var universal, partial int
	for _, a := range catalog {
		if a.Universality == "universal" {
			universal++
		} else {
			partial++
		}
	}
	o := out{
		Angles:    catalog,
		Universal: universal,
		Partial:   partial,
		Summary: "Conjunction (0), trine (120), and opposition (180) are the " +
			"deepest invariants. Square (90) is nearly universal. " +
			"Sextile (60) is Western/Vedic only — candidate for later accretion.",
	}
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return nil, err
	}
	return b, nil
}
