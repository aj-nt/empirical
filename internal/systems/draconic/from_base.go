package draconic

import "github.com/aj-nt/empirical/internal/dignity"

// FromBase computes the draconic chart from a BaseChart.
// It extracts tropical longitudes and rotates by the chart's North Node offset.
func FromBase(bc *dignity.BaseChart) *dignity.DraconicChart {
	planetLons := dignity.TropicalToLonMap(bc.Tropical)
	return dignity.ComputeDraconic(planetLons, bc.NorthNode)
}
