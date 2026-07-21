package western

import "github.com/aj-nt/empirical/internal/dignity"

// FromBase produces a full modern Western chart interpretation from a BaseChart.
// When reading is true, additional reading-optimized fields are computed
// (chart ruler, final dispositor, weighted aspects, key midpoints, key star aspects,
// angular planets).
func FromBase(bc *dignity.BaseChart, orbDeg float64, reading bool) *dignity.ChartInterpretation {
	return dignity.WesternFromBase(bc, orbDeg, reading)
}
