package dignity

// ── Hemisphere Emphasis ────────────────────────────────────────────────────
//
// Pure math — no SWE calls. Counts planets in each hemisphere quadrant
// relative to the Ascendant-Descendant and MC-IC axes.

// HemisphereEmphasis holds planet counts in each hemisphere.
type HemisphereEmphasis struct {
	Above int `json:"above"` // houses 7-12 (above horizon)
	Below int `json:"below"` // houses 1-6 (below horizon)
	East  int `json:"east"`  // houses 10,11,12,1,2,3 (rising side)
	West  int `json:"west"`  // houses 4,5,6,7,8,9 (setting side)
}

// ComputeHemisphereEmphasis counts planets in each hemisphere.
// asc is the tropical Ascendant longitude.
func ComputeHemisphereEmphasis(planets map[string]float64, asc float64) *HemisphereEmphasis {
	he := &HemisphereEmphasis{}
	for _, lon := range planets {
		// Whole-sign house from ASC
		house := ((int(lon/30) - int(asc/30) + 12) % 12) + 1

		// Above/below horizon: houses 7-12 above, 1-6 below
		if house >= 7 {
			he.Above++
		} else {
			he.Below++
		}

		// East/west: houses 10,11,12,1,2,3 east; 4,5,6,7,8,9 west
		if house >= 10 || house <= 3 {
			he.East++
		} else {
			he.West++
		}
	}
	return he
}
