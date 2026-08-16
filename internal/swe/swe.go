package swe

/*
#cgo pkg-config: swe
#include <swephexp.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// mu guards all Swiss Ephemeris C calls — the library is not thread-safe.
var mu sync.Mutex

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
	TRUE_NODE = C.SE_TRUE_NODE
	CHIRON    = C.SE_CHIRON
	CERES     = C.SE_CERES
	PALLAS    = C.SE_PALLAS
	JUNO      = C.SE_JUNO
	VESTA     = C.SE_VESTA
	MEAN_APOG = C.SE_MEAN_APOG

	// Dwarf planets and distant objects (asteroid numbers + SE_AST_OFFSET)
	ERIS     = C.SE_AST_OFFSET + 136199
	MAKEMAKE = C.SE_AST_OFFSET + 136472
	GONGGONG = C.SE_AST_OFFSET + 225088

	// Major asteroids (0-999 range, covered by seas_00.se1)
	// Numbers 1-4 (Ceres, Pallas, Juno, Vesta) use special SWE constants above.
	ASTRAEA    = C.SE_AST_OFFSET + 5
	HEBE       = C.SE_AST_OFFSET + 6
	IRIS       = C.SE_AST_OFFSET + 7
	FLORA      = C.SE_AST_OFFSET + 8
	METIS      = C.SE_AST_OFFSET + 9
	HYGIEA     = C.SE_AST_OFFSET + 10
	PSYCHE     = C.SE_AST_OFFSET + 16
	FORTUNA    = C.SE_AST_OFFSET + 19
	PROSERPINA = C.SE_AST_OFFSET + 26
	AMPHITRITE = C.SE_AST_OFFSET + 29
	PANDORA    = C.SE_AST_OFFSET + 55
	MNEMOSYNE  = C.SE_AST_OFFSET + 57
	CYBELE     = C.SE_AST_OFFSET + 65
	DIANA      = C.SE_AST_OFFSET + 78
	SAPPHO     = C.SE_AST_OFFSET + 80
	EROS       = C.SE_AST_OFFSET + 433
	// Distant objects (covered by seas_90.se1, seas_136.se1)
	ORCUS  = C.SE_AST_OFFSET + 90482
	SEDNA  = C.SE_AST_OFFSET + 90377
	HAUMEA = C.SE_AST_OFFSET + 136108

	// Uranian (Hamburg School) hypothetical planets
	CUPIDO   = C.SE_CUPIDO
	HADES    = C.SE_HADES
	ZEUS     = C.SE_ZEUS
	KRONOS   = C.SE_KRONOS
	APOLLON  = C.SE_APOLLON
	ADMETOS  = C.SE_ADMETOS
	POSEIDON = C.SE_POSEIDON
	VULKANUS = C.SE_VULKANUS
)

// Flag: use high-precision Swiss Ephemeris files.
const SEFLG_SWIEPH = C.SEFLG_SWIEPH

// Flag: compute speed (deg/day). Must be OR'd with SEFLG_SWIEPH.
const SEFLG_SPEED = C.SEFLG_SPEED

// Julday computes the Julian Day number for a given calendar date and time.
// hour is in UT. If gregflag is true, Gregorian calendar is used.
func Julday(year, month, day int, hour float64, gregflag bool) float64 {
	mu.Lock()
	defer mu.Unlock()
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
// Panics if the Swiss Ephemeris returns an error (missing file, out of range, etc.).
// For callers that can propagate errors, use CalcUTErr instead.
// planet must be one of the SE_* constants.
func CalcUT(jd float64, planet int) (lon, lat, dist, speed float64) {
	mu.Lock()
	defer mu.Unlock()
	var xx [6]C.double
	var serr [256]C.char
	ret := C.swe_calc_ut(
		C.double(jd),
		C.int(planet),
		C.int(SEFLG_SWIEPH|SEFLG_SPEED),
		(*C.double)(unsafe.Pointer(&xx[0])),
		(*C.char)(unsafe.Pointer(&serr[0])),
	)
	if int(ret) < 0 {
		panic(fmt.Sprintf("swe_calc_ut failed for planet %d: %s", planet, C.GoString(&serr[0])))
	}
	return float64(xx[0]), float64(xx[1]), float64(xx[2]), float64(xx[3])
}

// CalcUTErr is like CalcUT but returns an error instead of panicking.
// Use this when the caller can propagate errors to the user.
func CalcUTErr(jd float64, planet int) (lon, lat, dist, speed float64, err error) {
	mu.Lock()
	defer mu.Unlock()
	var xx [6]C.double
	var serr [256]C.char
	ret := C.swe_calc_ut(
		C.double(jd),
		C.int(planet),
		C.int(SEFLG_SWIEPH|SEFLG_SPEED),
		(*C.double)(unsafe.Pointer(&xx[0])),
		(*C.char)(unsafe.Pointer(&serr[0])),
	)
	if int(ret) < 0 {
		return 0, 0, 0, 0, fmt.Errorf("swe_calc_ut failed for planet %d: %s", planet, C.GoString(&serr[0]))
	}
	return float64(xx[0]), float64(xx[1]), float64(xx[2]), float64(xx[3]), nil
}

// SetEphePath points Swiss Ephemeris to its data files directory.
// Must be called before any calculation.
func SetEphePath(path string) {
	mu.Lock()
	defer mu.Unlock()
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.swe_set_ephe_path(cpath)
}

// SIDM_LAHIRI is the Lahiri ayanamsa.
const SIDM_LAHIRI = C.SE_SIDM_LAHIRI

// Other ayanamsas for sensitivity testing.
const (
	SIDM_FAGAN_BRADLEY = C.SE_SIDM_FAGAN_BRADLEY
	SIDM_RAMAN         = C.SE_SIDM_RAMAN
	SIDM_KRISHNAMURTI  = C.SE_SIDM_KRISHNAMURTI
)

// SetSidMode sets the sidereal mode (ayanamsa) for sidereal calculations.
// sidMode: e.g. SIDM_LAHIRI. t0 and ayanT0 are typically 0, 0.
func SetSidMode(sidMode int32, t0, ayanT0 float64) {
	mu.Lock()
	defer mu.Unlock()
	C.swe_set_sid_mode(C.int32_t(sidMode), C.double(t0), C.double(ayanT0))
}

// SetAyanamsaMode sets the sidereal mode by name.
// Supported: "lahiri", "fagan_bradley", "raman", "krishnamurti".
// Defaults to Lahiri for empty or unrecognized names.
func SetAyanamsaMode(name string) {
	switch name {
	case "fagan_bradley":
		SetSidMode(SIDM_FAGAN_BRADLEY, 0, 0)
	case "raman":
		SetSidMode(SIDM_RAMAN, 0, 0)
	case "krishnamurti":
		SetSidMode(SIDM_KRISHNAMURTI, 0, 0)
	default:
		SetSidMode(SIDM_LAHIRI, 0, 0)
	}
}

// GetAyanamsaUT returns the ayanamsa value for a given Julian Day.
func GetAyanamsaUT(jd float64) float64 {
	mu.Lock()
	defer mu.Unlock()
	return float64(C.swe_get_ayanamsa_ut(C.double(jd)))
}

// Houses computes house cusps and angles for a given time and location.
// hsys: 'P' = Placidus, 'W' = Whole Sign, etc.
// Returns 13 cusps (1-12) and 10 ascmc values:
//   ascmc[0] = ASC, ascmc[1] = MC, ascmc[2] = ARMC,
//   ascmc[3] = Vertex, ascmc[4] = Equatorial ASC, ...
func Houses(jd, lat, lon float64, hsys byte) (cusps [13]float64, ascmc [10]float64) {
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
	C.swe_close()
}

// Revjul converts a Julian Day back to calendar date and UT hour.
// Returns year, month, day, hour_ut (fractional).
func Revjul(jd float64) (year, month, day int, hour float64) {
	mu.Lock()
	defer mu.Unlock()
	var y, m, d C.int
	var h C.double
	C.swe_revjul(C.double(jd), C.int(1), &y, &m, &d, &h)
	return int(y), int(m), int(d), float64(h)
}

// FixstarMag returns the visual magnitude of a fixed star.
// starName must match a name in sefstars.txt.
// Returns 99 if the star is not found.
func FixstarMag(starName string) float64 {
	mu.Lock()
	defer mu.Unlock()
	cname := C.CString(starName)
	defer C.free(unsafe.Pointer(cname))
	var mag C.double
	var serr [256]C.char
	ret := C.swe_fixstar_mag(
		cname,
		&mag,
		(*C.char)(unsafe.Pointer(&serr[0])),
	)
	if int(ret) < 0 {
		return 99
	}
	return float64(mag)
}
// starName must match a name in sefstars.txt (e.g., "Aldebaran", "Sirius").
// Returns ecliptic longitude, latitude, distance (AU), and speed (deg/day).
// Returns 0,0,0,0 if the star is not found.
func Fixstar(starName string, jd float64) (lon, lat, dist, speed float64) {
	mu.Lock()
	defer mu.Unlock()
	cname := C.CString(starName)
	defer C.free(unsafe.Pointer(cname))
	var xx [6]C.double
	var serr [256]C.char
	ret := C.swe_fixstar(
		cname,
		C.double(jd),
		C.int(SEFLG_SWIEPH),
		(*C.double)(unsafe.Pointer(&xx[0])),
		(*C.char)(unsafe.Pointer(&serr[0])),
	)
	if int(ret) < 0 {
		return 0, 0, 0, 0
	}
	return float64(xx[0]), float64(xx[1]), float64(xx[2]), float64(xx[3])
}
