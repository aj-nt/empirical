package dignity

// VedicFromBase produces a full Vedic dignity convergence report from a BaseChart.
// It extracts tropical longitudes, applies the Lahiri ayanamsa from the chart,
// and computes Western/Vedic dignity convergence for all seven classical planets.
func VedicFromBase(bc *BaseChart) *DignityConvergence {
	planetLons := TropicalToLonMap(bc.Tropical)
	return ComputeDignityConvergence(planetLons, bc.Ayanamsa, bc.Name)
}
