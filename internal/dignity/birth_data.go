// Package dignity provides astrological computation primitives.
package dignity

// BirthData holds the input data for computing a natal chart.
// It is the canonical representation of birth information used across
// all function signatures, replacing the previous pattern of passing
// name, year, month, day, hour, minute, tzOffset, lat, lng as individual parameters.
type BirthData struct {
	Name     string
	Year     int
	Month    int
	Day      int
	Hour     int
	Minute   int
	Second   int
	TZOffset float64
	Lat      float64
	Lng      float64
}
