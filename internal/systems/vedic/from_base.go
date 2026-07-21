package vedic

import "github.com/aj-nt/empirical/internal/dignity"

// FromBase produces a full Vedic dignity convergence report from a BaseChart.
// It extracts tropical longitudes, applies the Lahiri ayanamsa from the chart,
// and computes Western/Vedic dignity convergence for all seven classical planets.
func FromBase(bc *dignity.BaseChart) *dignity.DignityConvergence {
	planetLons := dignity.TropicalToLonMap(bc.Tropical)
	return dignity.ComputeDignityConvergence(planetLons, bc.Ayanamsa, bc.Name)
}
