package dignity

// BaZiFromBase produces the Four Pillars and Day Master from a BaseChart.
// It delegates to ComputeBaZiPillars using the chart's birth year, month, day, and hour.
func BaZiFromBase(bc *BaseChart) BaZiFourPillars {
	return ComputeBaZiPillars(bc.Year, bc.Month, bc.Day, bc.Hour)
}
