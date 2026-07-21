package bazi

import "github.com/aj-nt/empirical/internal/dignity"

// FromBase produces the Four Pillars and Day Master from a BaseChart.
// It delegates to ComputeBaZiPillars using the chart's birth year, month, day, and hour.
func FromBase(bc *dignity.BaseChart) dignity.BaZiFourPillars {
	return dignity.ComputeBaZiPillars(bc.Year, bc.Month, bc.Day, bc.Hour)
}
