package swe

/*
#cgo darwin,amd64 LDFLAGS: /Users/aj/.local/lib/libswe.a -lm
#cgo darwin,arm64 LDFLAGS: /Users/aj/.local/lib/libswe.a -lm
#cgo linux LDFLAGS: -lswe -lm
#cgo CFLAGS: -I/Users/aj/.local/include
#include <swephexp.h>
*/
import "C"
import "unsafe"

// Swiss Ephemeris planet indices.
const (
	SUN       = C.SE_SUN
	MOON      = C.SE_MOON
	MERCURY   = C.SE_MERCURY
	VENUS     = C.SE_VENUS
	MARS      = C.SE_MARS
	JUPITER   = C.SE_JUPITER
	SATURN    = C.SE_SATURN
	URANUS    = C.SE_URANUS
	NEPTUNE   = C.SE_NEPTUNE
	PLUTO     = C.SE_PLUTO
	MEAN_NODE = C.SE_MEAN_NODE
	CHIRON    = C.SE_CHIRON
	CERES     = C.SE_CERES
	PALLAS    = C.SE_PALLAS
	JUNO      = C.SE_JUNO
	VESTA     = C.SE_VESTA
	MEAN_APOG = C.SE_MEAN_APOG
)

// Flag: use high-precision Swiss Ephemeris files.
const SEFLG_SWIEPH = C.SEFLG_SWIEPH

// Julday computes the Julian Day number for a given calendar date and time.
// hour is in UT. If gregflag is true, Gregorian calendar is used.
func Julday(year, month, day int, hour float64, gregflag bool) float64 {
	var gflag C.int
	if gregflag {
		gflag = 1
	}
	return float64(C.swe_julday(
		C.int(year), C.int(month), C.int(day),
		C.double(hour), gflag,
	))
}

// CalcUT computes a planet's ecliptic position for the given Julian Day.
// Returns longitude, latitude, distance (AU), and speed in longitude (deg/day).
// planet must be one of the SE_* constants.
func CalcUT(jd float64, planet int) (lon, lat, dist, speed float64) {
	var xx [6]C.double
	var serr [256]C.char
	ret := C.swe_calc_ut(
		C.double(jd),
		C.int(planet),
		C.int(SEFLG_SWIEPH),
		(*C.double)(unsafe.Pointer(&xx[0])),
		(*C.char)(unsafe.Pointer(&serr[0])),
	)
	if int(ret) < 0 {
		return 0, 0, 0, 0
	}
	return float64(xx[0]), float64(xx[1]), float64(xx[2]), float64(xx[3])
}

// SetEphePath points Swiss Ephemeris to its data files directory.
// Must be called before any calculation.
func SetEphePath(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.swe_set_ephe_path(cpath)
}

// SIDM_LAHIRI is the Lahiri ayanamsa.
const SIDM_LAHIRI = C.SE_SIDM_LAHIRI

// SetSidMode sets the sidereal mode (ayanamsa) for sidereal calculations.
// sidMode: e.g. SIDM_LAHIRI. t0 and ayanT0 are typically 0, 0.
func SetSidMode(sidMode int32, t0, ayanT0 float64) {
	C.swe_set_sid_mode(C.int32_t(sidMode), C.double(t0), C.double(ayanT0))
}

// GetAyanamsaUT returns the ayanamsa value for a given Julian Day.
func GetAyanamsaUT(jd float64) float64 {
	return float64(C.swe_get_ayanamsa_ut(C.double(jd)))
}

// Houses computes house cusps and angles for a given time and location.
// hsys: 'P' = Placidus, 'W' = Whole Sign, etc.
// Returns 13 cusps (1-12) and 10 ascmc values:
//   ascmc[0] = ASC, ascmc[1] = MC, ascmc[2] = ARMC,
//   ascmc[3] = Vertex, ascmc[4] = Equatorial ASC, ...
func Houses(jd, lat, lon float64, hsys byte) (cusps [13]float64, ascmc [10]float64) {
	var ccusps [13]C.double
	var cascmc [10]C.double
	C.swe_houses(
		C.double(jd),
		C.double(lat),
		C.double(lon),
		C.int(hsys),
		(*C.double)(unsafe.Pointer(&ccusps[0])),
		(*C.double)(unsafe.Pointer(&cascmc[0])),
	)
	for i := range ccusps {
		cusps[i] = float64(ccusps[i])
	}
	for i := range cascmc {
		ascmc[i] = float64(cascmc[i])
	}
	return
}

// Close cleans up Swiss Ephemeris resources.
func Close() {
	C.swe_close()
}
