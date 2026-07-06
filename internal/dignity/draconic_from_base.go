package dignity

// DraconicFromBase computes the draconic chart from a BaseChart.
// It extracts tropical longitudes and rotates by the chart's North Node offset.
func DraconicFromBase(bc *BaseChart) *DraconicChart {
	planetLons := TropicalToLonMap(bc.Tropical)
	return ComputeDraconic(planetLons, bc.NorthNode)
}
