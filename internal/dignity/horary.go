package dignity

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// Horary Astrology — Question-based chart judgment
// ═══════════════════════════════════════════════════════════════════════════
//
// Horary answers a specific question using the chart of the moment the
// question is asked. Key principles:
//   - The querent is signified by the 1st house ruler
//   - The quesited is signified by the relevant house ruler
//   - Aspect between significators = yes
//   - No aspect = no
//   - Applying aspect = stronger yes
//   - Separating aspect = weaker yes / already happened
//   - Moon's next aspect is the key indicator
//   - Void of Course Moon = nothing will come of it
//   - Reception between significators strengthens the answer
//   - Prohibition/frustration can block a yes

// HoraryQuestion represents a horary question.
type HoraryQuestion struct {
	Question string `json:"question"`
	Category string `json:"category"` // auto-detected or user-specified
}

// HoraryJudgment is the complete horary answer.
type HoraryJudgment struct {
	Question       string              `json:"question"`
	Category       string              `json:"category"`
	QuerentHouse   int                 `json:"querent_house"`
	QuesitedHouse  int                 `json:"quesited_house"`
	QuerentSig     string              `json:"querent_significator"`
	QuesitedSig    string              `json:"quesited_significator"`
	MoonNextAspect *HoraryAspect       `json:"moon_next_aspect"`
	SigAspect      *HoraryAspect       `json:"significator_aspect"`
	Verdict        string              `json:"verdict"`        // yes / no / maybe
	Confidence     string              `json:"confidence"`     // strong / moderate / weak
	Reasoning      []string            `json:"reasoning"`
	Considerations []string            `json:"considerations"` // things to watch
	ChartMoment    string              `json:"chart_moment"`
}

// HoraryAspect describes an aspect between two planets.
type HoraryAspect struct {
	FromPlanet string  `json:"from_planet"`
	ToPlanet   string  `json:"to_planet"`
	Aspect     string  `json:"aspect"`
	Orb        float64 `json:"orb"`
	Applying   bool    `json:"applying"`
	Reception  string  `json:"reception"` // mutual / one-sided / none
}

// ── Horary House Categories ────────────────────────────────────────────────

// horaryCategories maps keywords to the house of the quesited.
var horaryCategories = map[string]struct {
	House int
	Keywords []string
}{
	"Career / Job":          {10, []string{"job", "career", "work", "promotion", "boss", "profession", "employment", "raise", "salary"}},
	"Money / Finances":      {2,  []string{"money", "cash", "income", "finance", "debt", "loan", "savings", "bank", "wealth"}},
	"Relationships / Love":  {7,  []string{"love", "relationship", "marriage", "partner", "girlfriend", "boyfriend", "spouse", "dating", "romance"}},
	"Home / Property":       {4,  []string{"home", "house", "property", "real estate", "move", "relocate", "apartment", "rent", "mortgage"}},
	"Health":                {6,  []string{"health", "illness", "sick", "disease", "surgery", "doctor", "hospital", "healing", "symptom"}},
	"Travel":                {9,  []string{"travel", "trip", "journey", "flight", "vacation", "abroad", "visa", "passport"}},
	"Education / Learning":  {9,  []string{"study", "learn", "school", "university", "college", "degree", "exam", "test", "course", "education"}},
	"Legal / Court":         {9,  []string{"legal", "court", "lawsuit", "lawyer", "attorney", "judge", "case", "trial"}},
	"Family / Children":     {5,  []string{"child", "children", "baby", "pregnancy", "family", "parent", "mother", "father", "son", "daughter"}},
	"Communication":         {3,  []string{"message", "email", "call", "phone", "text", "letter", "news", "communication", "hear from"}},
	"Lost Items":            {2,  []string{"lost", "missing", "find", "where is", "locate", "stolen"}},
	"Spirituality":          {12, []string{"spiritual", "meditation", "god", "prayer", "meaning", "purpose", "soul"}},
	"Friends / Community":   {11, []string{"friend", "social", "community", "group", "network", "club"}},
}

// detectHoraryCategory identifies the horary house from the question text.
func detectHoraryCategory(question string) (string, int) {
	lower := strings.ToLower(question)
	for cat, info := range horaryCategories {
		for _, kw := range info.Keywords {
			if strings.Contains(lower, kw) {
				return cat, info.House
			}
		}
	}
	// Default: 1st house (about the querent themselves)
	return "Personal / Self", 1
}

// ── ComputeHoraryJudgment ───────────────────────────────────────────────────

// ComputeHoraryJudgment judges a horary question using the chart of the moment.
// bc: the base chart for the moment the question is asked
// question: the question text
func ComputeHoraryJudgment(bc *BaseChart, question string) (*HoraryJudgment, error) {
	category, quesitedHouse := detectHoraryCategory(question)

	judgment := &HoraryJudgment{
		Question:      question,
		Category:      category,
		QuerentHouse:  1,
		QuesitedHouse: quesitedHouse,
		ChartMoment:   time.Now().Format(time.RFC3339),
		Reasoning:     []string{},
		Considerations: []string{},
	}

	// ── Significators ──────────────────────────────────────────────────
	// Querent = ruler of 1st house
	// Quesited = ruler of the relevant house
	ascSignIdx := int(bc.ASC/30) % 12
	ascSign := Signs[ascSignIdx]
	judgment.QuerentSig = signRuler(ascSign)

	quesitedSignIdx := (ascSignIdx + quesitedHouse - 1) % 12
	quesitedSign := Signs[quesitedSignIdx]
	judgment.QuesitedSig = signRuler(quesitedSign)

	judgment.Reasoning = append(judgment.Reasoning,
		fmt.Sprintf("Querent (1H) = %s, ruler = %s", ascSign, judgment.QuerentSig))
	judgment.Reasoning = append(judgment.Reasoning,
		fmt.Sprintf("Quesited (%dH) = %s, ruler = %s", quesitedHouse, quesitedSign, judgment.QuesitedSig))

	// ── Planet positions ───────────────────────────────────────────────
	planetLons := TropicalToLonMap(bc.Tropical)
	planetSpeeds := make(map[string]float64)
	for name, pos := range bc.Tropical {
		planetSpeeds[name] = pos.Speed
	}

	// ── Moon's next aspect ─────────────────────────────────────────────
	moonLon, moonOK := planetLons["Moon"]
	if moonOK {
		moonNext := findNextAspect("Moon", moonLon, planetSpeeds["Moon"], planetLons, planetSpeeds)
		judgment.MoonNextAspect = moonNext
		if moonNext != nil {
			judgment.Reasoning = append(judgment.Reasoning,
				fmt.Sprintf("Moon's next aspect: %s %s %s (%.1f°)",
					moonNext.FromPlanet, moonNext.Aspect, moonNext.ToPlanet, moonNext.Orb))
		} else {
			judgment.Reasoning = append(judgment.Reasoning, "Moon is Void of Course — nothing will come of it")
			judgment.Considerations = append(judgment.Considerations, "VoC Moon: the matter may stall or fizzle out")
		}
	}

	// ── Significator aspect ────────────────────────────────────────────
	querentLon, qOK := planetLons[judgment.QuerentSig]
	quesitedLon, qsOK := planetLons[judgment.QuesitedSig]

	if qOK && qsOK {
		sigAspect := findAspectBetween(judgment.QuerentSig, querentLon, planetSpeeds[judgment.QuerentSig],
			judgment.QuesitedSig, quesitedLon, planetSpeeds[judgment.QuesitedSig])
		judgment.SigAspect = sigAspect

		if sigAspect != nil {
			judgment.Reasoning = append(judgment.Reasoning,
				fmt.Sprintf("Significator aspect: %s %s %s (%.1f°, %s)",
					sigAspect.FromPlanet, sigAspect.Aspect, sigAspect.ToPlanet,
					sigAspect.Orb, map[bool]string{true: "applying", false: "separating"}[sigAspect.Applying]))
		} else {
			judgment.Reasoning = append(judgment.Reasoning,
				fmt.Sprintf("No aspect between significators %s and %s", judgment.QuerentSig, judgment.QuesitedSig))
		}
	}

	// ── Verdict ─────────────────────────────────────────────────────────
	judgment.Verdict, judgment.Confidence = determineVerdict(judgment)

	return judgment, nil
}

// ── Aspect Finding ─────────────────────────────────────────────────────────

// findNextAspect finds the next applying aspect the Moon makes.
func findNextAspect(planet string, lon, speed float64, allLons, allSpeeds map[string]float64) *HoraryAspect {
	order := []string{"Sun", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}
	aspectTypes := []struct {
		name string
		deg  float64
	}{
		{"conjunction", 0}, {"sextile", 60}, {"square", 90},
		{"trine", 120}, {"opposition", 180},
	}

	var best *HoraryAspect
	bestDist := 360.0

	for _, target := range order {
		targetLon, ok := allLons[target]
		if !ok || target == planet {
			continue
		}
		targetSpeed, _ := allSpeeds[target]

		for _, asp := range aspectTypes {
			// Find the applying position
			aspLon := math.Mod(targetLon+asp.deg, 360)
			dist := aspLon - lon
			if dist < 0 {
				dist += 360
			}

			// Check if applying (planet moving toward aspect)
			applying := isApplying(lon, speed, targetLon, targetSpeed, asp.deg)

			if dist < bestDist && applying {
				bestDist = dist
				best = &HoraryAspect{
					FromPlanet: planet,
					ToPlanet:   target,
					Aspect:     asp.name,
					Orb:        math.Round(dist*100) / 100,
					Applying:   true,
					Reception:  checkReception(planet, target, allLons),
				}
			}
		}
	}

	return best
}

// findAspectBetween checks if two planets are in aspect.
func findAspectBetween(p1 string, lon1, speed1 float64, p2 string, lon2, speed2 float64) *HoraryAspect {
	dist := math.Abs(lon1 - lon2)
	if dist > 180 {
		dist = 360 - dist
	}

	aspectTypes := map[float64]string{
		0: "conjunction", 60: "sextile", 90: "square", 120: "trine", 180: "opposition",
	}

	// Check within 8° orb
	for aspDeg, aspName := range aspectTypes {
		orb := math.Abs(dist - aspDeg)
		if orb <= 8.0 {
			applying := isApplying(lon1, speed1, lon2, speed2, aspDeg)
			return &HoraryAspect{
				FromPlanet: p1,
				ToPlanet:   p2,
				Aspect:     aspName,
				Orb:        math.Round(orb*100) / 100,
				Applying:   applying,
				Reception:  checkReception(p1, p2, map[string]float64{p1: lon1, p2: lon2}),
			}
		}
	}

	return nil
}

// isApplying determines if planet1 is moving toward an aspect with planet2.
func isApplying(lon1, speed1, lon2, speed2, aspDeg float64) bool {
	// Simplified: if the faster planet is behind the aspect point, it's applying
	aspPoint := math.Mod(lon2+aspDeg, 360)
	dist := aspPoint - lon1
	if dist < 0 {
		dist += 360
	}
	// If the planet is moving toward the aspect point (dist decreasing), it's applying
	// For direct motion: planet is behind the aspect point
	// For retrograde: planet is ahead of the aspect point
	if speed1 > 0 {
		return dist < 180 // planet behind aspect point
	}
	return dist > 180 // planet ahead of aspect point (retrograde, moving backward toward it)
}

// checkReception checks for mutual reception between two planets.
func checkReception(p1, p2 string, lons map[string]float64) string {
	lon1, ok1 := lons[p1]
	lon2, ok2 := lons[p2]
	if !ok1 || !ok2 {
		return "none"
	}

	sign1 := Signs[int(lon1/30)%12]
	sign2 := Signs[int(lon2/30)%12]

	ruler1 := signRuler(sign1)
	ruler2 := signRuler(sign2)

	if ruler1 == p2 && ruler2 == p1 {
		return "mutual"
	}
	if ruler1 == p2 || ruler2 == p1 {
		return "one-sided"
	}
	return "none"
}

// ── Verdict ────────────────────────────────────────────────────────────────

func determineVerdict(j *HoraryJudgment) (string, string) {
	yesSignals := 0
	noSignals := 0

	// Moon VoC = strong no
	if j.MoonNextAspect == nil {
		noSignals += 3
	}

	// Significator aspect
	if j.SigAspect != nil {
		if j.SigAspect.Applying {
			yesSignals += 3
		} else {
			yesSignals += 1 // separating = weaker yes
		}
		if j.SigAspect.Reception == "mutual" {
			yesSignals += 2
		} else if j.SigAspect.Reception == "one-sided" {
			yesSignals += 1
		}
	} else {
		noSignals += 2
	}

	// Moon's next aspect to a significator = strong yes
	if j.MoonNextAspect != nil {
		if j.MoonNextAspect.ToPlanet == j.QuerentSig || j.MoonNextAspect.ToPlanet == j.QuesitedSig {
			yesSignals += 2
		}
		// Benefic aspect (trine/sextile) = positive
		if j.MoonNextAspect.Aspect == "trine" || j.MoonNextAspect.Aspect == "sextile" {
			yesSignals += 1
		}
		// Malefic aspect (square/opposition) = obstacles
		if j.MoonNextAspect.Aspect == "square" || j.MoonNextAspect.Aspect == "opposition" {
			noSignals += 1
		}
	}

	// Determine verdict
	net := yesSignals - noSignals
	switch {
	case net >= 4:
		return "yes", "strong"
	case net >= 2:
		return "yes", "moderate"
	case net >= 0:
		return "maybe", "weak"
	case net >= -2:
		return "no", "moderate"
	default:
		return "no", "strong"
	}
}

// ── JSON ───────────────────────────────────────────────────────────────────

// JSON serializes the horary judgment to JSON.
func (j *HoraryJudgment) JSON() ([]byte, error) {
	return json.Marshal(j)
}
