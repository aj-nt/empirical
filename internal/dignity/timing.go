package dignity

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// Phase 4: Timing layer convergence
// ═══════════════════════════════════════════════════════════════════════════
//
// Every surviving tradition has a timing mechanic. They have fundamentally
// different architectures — different cycle lengths, different triggers,
// different abstractions. They cannot agree on exact dates.
//
// The convergence question: do they agree on themes? When Vedic says
// "Mercury period," does Hellenistic also activate Mercury's house?
// When Ba Zi says "Metal luck pillar," does that planet appear elsewhere?
//
// System              Mechanism               Cycle basis       Output
// ─────────────────── ─────────────────────── ────────────────  ─────────
// Vedic (Vimshottari) Moon nakshatra → planet  120-year cycle   Planet name
// Ba Zi (Luck Pillars) Day stem → element pair  10-year pillar  Stem+Branch
// Hellenistic (Prof)   Age → annual house        1-year cycle   House+Lord
//
// Convergence metric: map each system's current period to a set of
// "activated planets", then check for overlap. If >= 2 systems activate
// the same planet, that's a timing convergence.

// ── Types ───────────────────────────────────────────────────────────────

// TimingPeriod is a single timing period from one system.
type TimingPeriod struct {
	System   string   `json:"system"`
	Name     string   `json:"name"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Years    float64  `json:"years"`
	Planets  []string `json:"planets"`
	Elements []string `json:"elements"`
}

// TimingConvergence holds timing convergence across systems for a given date.
type TimingConvergence struct {
	Name              string          `json:"name"`
	TargetDate        string          `json:"target_date"`
	Periods           []TimingPeriod  `json:"periods"`
	PlanetConvergences []string       `json:"planet_convergences"`
	ConvergenceCount  int             `json:"convergence_count"`
	HasConvergence    bool            `json:"has_convergence"`
	AllSystemsAgree   bool            `json:"all_systems_agree"`
}

// PlanetConvergences returns planets that appear in 2+ systems.
func (tc *TimingConvergence) computePlanetConvergences() []string {
	planetSystems := make(map[string]map[string]bool)
	for _, p := range tc.Periods {
		seen := make(map[string]bool)
		for _, planet := range p.Planets {
			if planet == "" {
				continue
			}
			seen[planet] = true
		}
		for planet := range seen {
			if planetSystems[planet] == nil {
				planetSystems[planet] = make(map[string]bool)
			}
			planetSystems[planet][p.System] = true
		}
	}
	var result []string
	for planet, systems := range planetSystems {
		if len(systems) >= 2 {
			result = append(result, planet)
		}
	}
	return result
}

// ConvergenceCount returns the number of converging planets.
func (tc *TimingConvergence) computeConvergenceCount() int {
	return len(tc.computePlanetConvergences())
}

// HasConvergence checks if any planet appears in 2+ systems.
func (tc *TimingConvergence) computeHasConvergence() bool {
	return tc.computeConvergenceCount() > 0
}

// AllSystemsAgree checks if at least one planet appears in ALL systems.
func (tc *TimingConvergence) computeAllSystemsAgree() bool {
	if len(tc.Periods) < 2 {
		return false
	}
	nSystems := make(map[string]bool)
	for _, p := range tc.Periods {
		nSystems[p.System] = true
	}
	total := len(nSystems)

	planetSystems := make(map[string]map[string]bool)
	for _, p := range tc.Periods {
		seen := make(map[string]bool)
		for _, planet := range p.Planets {
			if planet == "" {
				continue
			}
			seen[planet] = true
		}
		for planet := range seen {
			if planetSystems[planet] == nil {
				planetSystems[planet] = make(map[string]bool)
			}
			planetSystems[planet][p.System] = true
		}
	}
	for _, systems := range planetSystems {
		if len(systems) == total {
			return true
		}
	}
	return false
}

// ── Ba Zi: Heavenly Stems & Earthly Branches ─────────────────────────────

var baZiHeavenlyStems = []string{
	"Jia", "Yi", "Bing", "Ding", "Wu", "Ji", "Geng", "Xin", "Ren", "Gui",
}

var baZiStemElements = []string{
	"Wood", "Wood", "Fire", "Fire", "Earth", "Earth", "Metal", "Metal", "Water", "Water",
}

var baZiStemYinYang = []string{
	"Yang", "Yin", "Yang", "Yin", "Yang", "Yin", "Yang", "Yin", "Yang", "Yin",
}

var baZiEarthlyBranches = []string{
	"Zi", "Chou", "Yin", "Mao", "Chen", "Si", "Wu", "Wei", "Shen", "You", "Xu", "Hai",
}

var baZiBranchElements = []string{
	"Water", "Earth", "Wood", "Wood", "Earth", "Fire", "Fire", "Earth", "Metal", "Metal", "Earth", "Water",
}

// Solar term boundary dates (month, day) for month branch determination.
// Index → branch: 0=Yin, 1=Mao, ..., 8=Shen, 9=You, 10=Xu, 11=Hai, 10=Zi, 11=Chou.
var baZiSolarMonthStarts = [][2]int{
	{2, 4},  // Li Chun — Yin (Tiger)
	{3, 6},  // Jing Zhe — Mao (Rabbit)
	{4, 5},  // Qing Ming — Chen (Dragon)
	{5, 6},  // Li Xia — Si (Snake)
	{6, 6},  // Mang Zhong — Wu (Horse)
	{7, 7},  // Xiao Shu — Wei (Goat)
	{8, 8},  // Li Qiu — Shen (Monkey)
	{9, 8},  // Bai Lu — You (Rooster)
	{10, 8}, // Han Lu — Xu (Dog)
	{11, 7}, // Li Dong — Hai (Pig)
	{12, 7}, // Da Xue — Zi (Rat)
	{1, 6},  // Xiao Han — Chou (Ox)
}

var baZiSolarMonthBranches = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0, 1}

// Year stem → first month stem: Jia/Ji→Bing(2), Yi/Geng→Wu(4), Bing/Xin→Geng(6), Ding/Ren→Ren(8), Wu/Gui→Jia(0).
var baZiYearToMonthStem = []int{2, 4, 6, 8, 0}

// ── Ba Zi calculation functions ─────────────────────────────────────────

func baZiYearStemBranch(year int) (stemIdx, branchIdx int) {
	return ((year-4)%10 + 10) % 10, ((year-4)%12 + 12) % 12
}

func baZiMonthStemBranch(year, month, day int) (stemIdx, branchIdx int) {
	yStem, _ := baZiYearStemBranch(year)

	// Solar month: iterate forward through terms, find the one we're BEFORE.
	// The correct solar month is the PREVIOUS term. If we never find one,
	// we're after the last term → solar month 11.
	solarMonth := 11 // after Xiao Han, before Li Chun
	for i, sm := range baZiSolarMonthStarts {
		if month < sm[0] || (month == sm[0] && day < sm[1]) {
			if i == 0 {
				solarMonth = 11
			} else {
				solarMonth = (i - 1) % 12
			}
			break
		}
	}

	branchIdx = (solarMonth + 2) % 12 // solar month 0 → Yin(2)
	firstStemTable := []int{2, 4, 6, 8, 0, 2, 4, 6, 8, 0}
	firstStem := firstStemTable[yStem]
	stemIdx = (firstStem + solarMonth) % 10
	return
}

func baZiDayStemBranch(year, month, day int) (stemIdx, branchIdx int) {
	// Use date ordinal + offset (exact match to Python's _julian_day / toordinal)
	// Jan 1, 2000 = ordinal 730120, JD = 2451544 (midnight)
	// ordinal of date(1,1,1) = 1, JD = 1721424.5 → offset = 1721424
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	// Go doesn't have toordinal(), compute manually from Unix epoch
	// Jan 1, 1970 = ordinal 719528 = JD 2440587 (noon-relative: 2440587.5)
	// Use the standard formula: days since Jan 1, 1970 + offset
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	days := int(t.Sub(epoch).Hours() / 24)
	jd := days + 2440587 // 2440587 = JD at midnight Jan 1, 1970

	stemIdx = ((jd % 10) + 10) % 10
	branchIdx = ((jd+2)%12 + 12) % 12
	return
}

func baZiHourBranch(hour int) int {
	return ((hour + 1) / 2) % 12
}

// Hour stem: base_stems[dayStemIdx] + hourBranchIdx, as per Python set_hour_stem.
// base_stems = [0, 2, 4, 6, 8, 0, 2, 4, 6, 8]
func baZiHourStem(hBranch, dStem int) int {
	base := []int{0, 2, 4, 6, 8, 0, 2, 4, 6, 8}[dStem]
	return (base + hBranch) % 10
}

// BaZiPillar holds a single pillar.
type BaZiPillar struct {
	Stem    string
	Branch  string
	Element string
}

// BaZiFourPillars holds the complete chart.
type BaZiFourPillars struct {
	Year, Month, Day, Hour BaZiPillar
	DayMaster              BaZiDayMaster
}

// BaZiDayMaster holds the day-stem yin-yang and element.
type BaZiDayMaster struct {
	YinYang string
	Element string
}

// ComputeBaZiPillars returns the Four Pillars + Day Master.
func ComputeBaZiPillars(year, month, day, hour int) BaZiFourPillars {
	yStem, yBranch := baZiYearStemBranch(year)
	mStem, mBranch := baZiMonthStemBranch(year, month, day)
	dStem, dBranch := baZiDayStemBranch(year, month, day)
	hBranch := baZiHourBranch(hour)
	hStem := baZiHourStem(hBranch, dStem)

	return BaZiFourPillars{
		Year:  BaZiPillar{baZiHeavenlyStems[yStem], baZiEarthlyBranches[yBranch], baZiStemElements[yStem]},
		Month: BaZiPillar{baZiHeavenlyStems[mStem], baZiEarthlyBranches[mBranch], baZiStemElements[mStem]},
		Day:   BaZiPillar{baZiHeavenlyStems[dStem], baZiEarthlyBranches[dBranch], baZiStemElements[dStem]},
		Hour:  BaZiPillar{baZiHeavenlyStems[hStem], baZiEarthlyBranches[hBranch], baZiStemElements[hStem]},
		DayMaster: BaZiDayMaster{
			YinYang: baZiStemYinYang[dStem],
			Element: baZiStemElements[dStem],
		},
	}
}

// baZiStemIndex returns the index of a heavenly stem name, or -1.
func baZiStemIndex(name string) int {
	for i, s := range baZiHeavenlyStems {
		if s == name {
			return i
		}
	}
	return -1
}

// baZiBranchIndex returns the index of an earthly branch name, or -1.
func baZiBranchIndex(name string) int {
	for i, b := range baZiEarthlyBranches {
		if b == name {
			return i
		}
	}
	return -1
}

// intMod returns positive modulo: (a % m + m) % m.
func intMod(a, m int) int {
	return ((a % m) + m) % m
}

// ── Vedic: Nakshatra & Vimshottari Dasha ─────────────────────────────────

var vedicNakshatras = []string{
	"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashirsha", "Ardra",
	"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
	"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
	"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
	"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
}

const vedicNakshatraSpan = 360.0 / 27.0

var vedicVimshottariOrder = []string{
	"Ketu", "Venus", "Sun", "Moon", "Mars",
	"Rahu", "Jupiter", "Saturn", "Mercury",
}

var vedicVimshottariPeriods = map[string]float64{
	"Ketu": 7, "Venus": 20, "Sun": 6, "Moon": 10, "Mars": 7,
	"Rahu": 18, "Jupiter": 16, "Saturn": 19, "Mercury": 17,
}

// Nakshatra rulers: 9-planet cycle repeats 3x across 27 nakshatras.
var vedicNakshatraRulers = func() []string {
	r := make([]string, 27)
	for i := range r {
		r[i] = vedicVimshottariOrder[i%9]
	}
	return r
}()

// NakshatraPosition holds nakshatra details for a sidereal longitude.
type NakshatraPosition struct {
	Nakshatra         string
	Pada              int
	DegreeInNakshatra float64
	Ruler             string
}

// GetNakshatra computes nakshatra from sidereal longitude.
func GetNakshatra(siderealLon float64) NakshatraPosition {
	sid := math.Mod(siderealLon, 360)
	if sid < 0 {
		sid += 360
	}
	idx := int(sid/vedicNakshatraSpan) % 27
	degInNak := math.Mod(sid, vedicNakshatraSpan)
	pada := int(degInNak/(vedicNakshatraSpan/4)) + 1
	if pada > 4 {
		pada = 4
	}
	return NakshatraPosition{
		Nakshatra:         vedicNakshatras[idx],
		Pada:              pada,
		DegreeInNakshatra: degInNak,
		Ruler:             vedicNakshatraRulers[idx],
	}
}

// VimshottariDashaEntry is one mahadasha period.
type VimshottariDashaEntry struct {
	Planet string
	Start  string
	End    string
	Years  float64
}

// ComputeVimshottariDasha computes the full dasha sequence from birth.
func ComputeVimshottariDasha(nak NakshatraPosition, birthYear, birthMonth, birthDay int) []VimshottariDashaEntry {
	nakshatraIdx := 0
	for i, n := range vedicNakshatras {
		if n == nak.Nakshatra {
			nakshatraIdx = i
			break
		}
	}
	ruler := vedicNakshatraRulers[nakshatraIdx]
	proportion := nak.DegreeInNakshatra / vedicNakshatraSpan

	birth := time.Date(birthYear, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)
	startIdx := 0
	for i, p := range vedicVimshottariOrder {
		if p == ruler {
			startIdx = i
			break
		}
	}

	var seq []VimshottariDashaEntry
	current := birth

	// First period (partial — remaining from birth)
	firstYears := vedicVimshottariPeriods[ruler] * (1 - proportion)
	firstEnd := current.Add(time.Duration(firstYears*365.25*24) * time.Hour)
	seq = append(seq, VimshottariDashaEntry{
		Planet: ruler,
		Start:  current.Format("2006-01-02"),
		End:    firstEnd.Format("2006-01-02"),
		Years:  math.Round(firstYears*100) / 100,
	})
	current = firstEnd

	// Full periods
	for i := 1; i < 9; i++ {
		planet := vedicVimshottariOrder[(startIdx+i)%9]
		years := vedicVimshottariPeriods[planet]
		end := current.Add(time.Duration(years*365.25*24) * time.Hour)
		seq = append(seq, VimshottariDashaEntry{
			Planet: planet,
			Start:  current.Format("2006-01-02"),
			End:    end.Format("2006-01-02"),
			Years:  years,
		})
		current = end
	}

	return seq
}

// CurrentDasha returns the active dasha entry for a target date, or nil.
func CurrentDasha(seq []VimshottariDashaEntry, targetDateStr string) *VimshottariDashaEntry {
	target, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		return nil
	}
	for i := range seq {
		start, _ := time.Parse("2006-01-02", seq[i].Start)
		end, _ := time.Parse("2006-01-02", seq[i].End)
		if !target.Before(start) && target.Before(end) {
			return &seq[i]
		}
	}
	return nil
}

// ── Hellenistic: Annual Profections ─────────────────────────────────────

var hellSignRulerships = map[string]string{
	"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury", "Cancer": "Moon",
	"Leo": "Sun", "Virgo": "Mercury", "Libra": "Venus", "Scorpio": "Mars",
	"Sagittarius": "Jupiter", "Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
}

var hellSignElements = map[string]string{
	"Aries": "Fire", "Taurus": "Earth", "Gemini": "Air", "Cancer": "Water",
	"Leo": "Fire", "Virgo": "Earth", "Libra": "Air", "Scorpio": "Water",
	"Sagittarius": "Fire", "Capricorn": "Earth", "Aquarius": "Air", "Pisces": "Water",
}

var hellZodiacSigns = []string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// ProfectionInfo holds the annual profection for a target date.
type ProfectionInfo struct {
	Age             int
	ProfectedHouse  int
	ProfectedSign   string
	LordOfYear      string
	ProfectionStart string
	ProfectionEnd   string
	Element         string
}

// ComputeProfection computes the annual profection for a target date.
// ascLon is the tropical ascendant longitude in degrees.
func ComputeProfection(birthYear, birthMonth, birthDay int, targetDate time.Time, ascLon float64) ProfectionInfo {
	birthdayThisYear := time.Date(targetDate.Year(), time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)

	var profStart time.Time
	var age int
	if targetDate.Before(birthdayThisYear) {
		profStart = time.Date(targetDate.Year()-1, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)
		age = targetDate.Year() - birthYear - 1
	} else {
		profStart = birthdayThisYear
		age = targetDate.Year() - birthYear
	}

	profEnd := time.Date(profStart.Year()+1, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)
	house := (age % 12) + 1

	// Sign on profected house cusp (whole-sign from ASC)
	ascSignIdx := int(math.Mod(ascLon, 360)/30) % 12
	for ascSignIdx < 0 {
		ascSignIdx += 12
	}
	profSignIdx := (ascSignIdx + house - 1) % 12
	for profSignIdx < 0 {
		profSignIdx += 12
	}
	profSign := hellZodiacSigns[profSignIdx]
	lord := hellSignRulerships[profSign]

	return ProfectionInfo{
		Age:             age,
		ProfectedHouse:  house,
		ProfectedSign:   profSign,
		LordOfYear:      lord,
		ProfectionStart: profStart.Format("2006-01-02"),
		ProfectionEnd:   profEnd.Format("2006-01-02"),
		Element:         hellSignElements[profSign],
	}
}

// ── Compute Timing Convergence ──────────────────────────────────────────

// Vedic planet → planet/element mapping.
var vedicPlanetMap = map[string]struct {
	Planets  []string
	Elements []string
}{
	"Mercury": {[]string{"Mercury"}, []string{"Earth"}},
	"Venus":   {[]string{"Venus"}, []string{"Water"}},
	"Mars":    {[]string{"Mars"}, []string{"Fire"}},
	"Jupiter": {[]string{"Jupiter"}, []string{"Ether"}},
	"Saturn":  {[]string{"Saturn"}, []string{"Air"}},
	"Sun":     {[]string{"Sun"}, []string{"Fire"}},
	"Moon":    {[]string{"Moon"}, []string{"Water"}},
	"Rahu":    {[]string{"Rahu", "Uranus"}, nil},
	"Ketu":    {[]string{"Ketu", "Neptune"}, nil},
}

// Ba Zi element → planet mapping.
var elementToPlanets = map[string][]string{
	"Metal": {"Saturn", "Venus"},
	"Wood":  {"Jupiter"},
	"Water": {"Moon", "Mercury"},
	"Fire":  {"Mars", "Sun"},
	"Earth": {"Saturn"},
}

// ComputeTimingConvergence computes timing convergence for a target date.
//
// Caller provides:
//   - dashaEntry: the active Vimshottari dasha on the target date (may be nil)
//   - baZiPillars: pre-computed Four Pillars
//   - profection: pre-computed profection info
func ComputeTimingConvergence(
	targetDateStr string,
	name string,
	dashaEntry *VimshottariDashaEntry,
	baZiPillars BaZiFourPillars,
	birthYear, birthMonth, birthDay int,
	profection ProfectionInfo,
) *TimingConvergence {
	tc := &TimingConvergence{
		Name:       name,
		TargetDate: targetDateStr,
	}

	// ── Vedic Dasha ──
	if dashaEntry != nil {
		mapping, ok := vedicPlanetMap[dashaEntry.Planet]
		if !ok {
			mapping = struct {
				Planets  []string
				Elements []string
			}{[]string{dashaEntry.Planet}, nil}
		}
		tc.Periods = append(tc.Periods, TimingPeriod{
			System:   "vedic_dasha",
			Name:     dashaEntry.Planet + " mahadasha",
			Start:    dashaEntry.Start,
			End:      dashaEntry.End,
			Years:    dashaEntry.Years,
			Planets:  mapping.Planets,
			Elements: mapping.Elements,
		})
	}

	// ── Ba Zi Luck Pillars ──
	dm := baZiPillars.DayMaster
	direction := 1
	if dm.YinYang == "Yin" {
		direction = -1
	}
	msIdx := baZiStemIndex(baZiPillars.Month.Stem)
	mbIdx := baZiBranchIndex(baZiPillars.Month.Branch)

	target, _ := time.Parse("2006-01-02", targetDateStr)

	found := false
	for i := 0; i < 10 && !found; i++ {
		lpStem := baZiHeavenlyStems[intMod(msIdx+direction*(i+1), 10)]
		lpBranch := baZiEarthlyBranches[intMod(mbIdx+direction*(i+1), 12)]
		lpStemIdx := baZiStemIndex(lpStem)
		lpBranchIdx := baZiBranchIndex(lpBranch)
		lpElement := baZiStemElements[lpStemIdx]
		lpBrElem := baZiBranchElements[lpBranchIdx]

		lpStart := time.Date(birthYear+i*10, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)
		lpEnd := time.Date(birthYear+(i+1)*10, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)

		if !target.Before(lpStart) && target.Before(lpEnd) {
			// Dedupe planets from stem + branch elements
			planetSet := make(map[string]bool)
			for _, p := range elementToPlanets[lpElement] {
				planetSet[p] = true
			}
			for _, p := range elementToPlanets[lpBrElem] {
				planetSet[p] = true
			}
			var allPlanets []string
			for p := range planetSet {
				allPlanets = append(allPlanets, p)
			}

			tc.Periods = append(tc.Periods, TimingPeriod{
				System:   "bazi_luck",
				Name:     fmt.Sprintf("%s%s (%s+%s)", lpStem, lpBranch, lpElement, lpBrElem),
				Start:    lpStart.Format("2006-01-02"),
				End:      lpEnd.Format("2006-01-02"),
				Years:    10,
				Planets:  allPlanets,
				Elements: []string{lpElement, lpBrElem},
			})
			found = true
		}
	}

	if !found {
		// No luck pillar found — shouldn't happen with 100-year span
		// but add a placeholder to keep period count consistent
		lpStart := time.Date(birthYear, time.Month(birthMonth), birthDay, 0, 0, 0, 0, time.UTC)
		lpEnd := lpStart.AddDate(10, 0, 0)
		tc.Periods = append(tc.Periods, TimingPeriod{
			System:   "bazi_luck",
			Name:     "unknown",
			Start:    lpStart.Format("2006-01-02"),
			End:      lpEnd.Format("2006-01-02"),
			Years:    10,
		})
	}

	// ── Hellenistic Profection ──
	var profPlanets []string
	if profection.LordOfYear != "" {
		profPlanets = append(profPlanets, profection.LordOfYear)
		// Also add the sign ruler (same as lord of year for simple profections)
		if profection.ProfectedSign != "" {
			if ruler, ok := hellSignRulerships[profection.ProfectedSign]; ok && ruler != profection.LordOfYear {
				profPlanets = append(profPlanets, ruler)
			}
		}
	}

	tc.Periods = append(tc.Periods, TimingPeriod{
		System:   "profection",
		Name:     fmt.Sprintf("H%d %s (lord: %s)", profection.ProfectedHouse, profection.ProfectedSign, profection.LordOfYear),
		Start:    profection.ProfectionStart,
		End:      profection.ProfectionEnd,
		Years:    1,
		Planets:  profPlanets,
		Elements: []string{profection.Element},
	})

	// Populate computed fields
	tc.finalize()

	return tc
}

// finalize populates computed convergence fields.
func (tc *TimingConvergence) finalize() {
	convs := tc.computePlanetConvergences()
	if convs == nil {
		convs = []string{}
	}
	tc.PlanetConvergences = convs
	tc.ConvergenceCount = len(convs)
	tc.HasConvergence = tc.ConvergenceCount > 0
	tc.AllSystemsAgree = tc.computeAllSystemsAgree()
}

// ── Formatting ──────────────────────────────────────────────────────────

// FormatTimingConvergence formats a report as human-readable text.
func FormatTimingConvergence(tc *TimingConvergence) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Timing Convergence Report — %s\n", tc.Name))
	b.WriteString(fmt.Sprintf("Target date: %s\n\n", tc.TargetDate))

	for _, p := range tc.Periods {
		planetStr := "none"
		if len(p.Planets) > 0 {
			planetStr = strings.Join(p.Planets, ", ")
		}
		b.WriteString(fmt.Sprintf("  %-14s %s\n", p.System, p.Name))
		b.WriteString(fmt.Sprintf("                  %s to %s (%.0fy)\n", p.Start, p.End, p.Years))
		b.WriteString(fmt.Sprintf("                  planets: %s\n", planetStr))
		b.WriteString("\n")
	}

	if tc.HasConvergence {
		b.WriteString(fmt.Sprintf("CONVERGENCE: %s appear in 2+ systems\n", strings.Join(tc.PlanetConvergences, ", ")))
	} else {
		b.WriteString("NO CONVERGENCE: no planet appears in 2+ systems\n")
	}

	if tc.AllSystemsAgree {
		b.WriteString("STRONG: at least one planet appears in ALL systems\n")
	} else if tc.HasConvergence {
		b.WriteString("PARTIAL: planet overlap but no single planet in all systems\n")
	}

	b.WriteString("\n")
	b.WriteString("RECOVERY IMPLICATION: Timing layers use fundamentally different architectures (planet cycles vs element cycles vs annual houses). Direct date convergence is impossible. Planet-level thematic convergence suggests the original system had a unified timing grammar — different traditions preserved different facets. When two or more systems activate the same planet, that period's theme is likely signal, not noise.\n")

	return b.String()
}
