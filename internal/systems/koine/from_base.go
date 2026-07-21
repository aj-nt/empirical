package koine

import "github.com/aj-nt/empirical/internal/dignity"

// FromBase produces a full Koiné chart interpretation from a BaseChart.
// It extracts tropical positions, computes whole-sign houses from the ASC,
// finds natal aspects and patterns, and runs the Hellenistic interpretation engine.
// orbDeg controls the aspect orb tolerance (default 5.0).
func FromBase(bc *dignity.BaseChart, orbDeg float64) *dignity.ChartInterpretation {
	return dignity.KoinéFromBase(bc, orbDeg)
}
