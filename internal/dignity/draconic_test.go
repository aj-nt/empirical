package dignity

import (
	"math"
	"testing"
)

func TestDraconic_Rotation(t *testing.T) {
	// NN at 10°, offset = 10°
	// Sun at 25° → draconic 15° (same sign Aries)
	// Mercury at 2° → draconic 352° (sign shift Aries→Pisces)
	// Mars at 350° → draconic 340° (sign shift Pisces→Pisces, no change)
	tropical := map[string]float64{
		"Sun":     25.0,
		"Moon":    120.0,
		"Mercury": 2.0,
		"Mars":    350.0,
	}
	nnLong := 10.0

	drac := ComputeDraconic(tropical, nnLong)

	// NN itself should be 0
	if math.Abs(drac.Planets["Node"]-0.0) > 0.001 {
		t.Errorf("draconic Node should be 0, got %.2f", drac.Planets["Node"])
	}

	// Sun: 25 - 10 = 15
	if math.Abs(drac.Planets["Sun"]-15.0) > 0.001 {
		t.Errorf("draconic Sun should be 15, got %.2f", drac.Planets["Sun"])
	}

	// Mercury: 2 - 10 = -8 → 352
	if math.Abs(drac.Planets["Mercury"]-352.0) > 0.001 {
		t.Errorf("draconic Mercury should be 352, got %.2f", drac.Planets["Mercury"])
	}

	// Mars: 350 - 10 = 340
	if math.Abs(drac.Planets["Mars"]-340.0) > 0.001 {
		t.Errorf("draconic Mars should be 340, got %.2f", drac.Planets["Mars"])
	}

	// Moon: 120 - 10 = 110
	if math.Abs(drac.Planets["Moon"]-110.0) > 0.001 {
		t.Errorf("draconic Moon should be 110, got %.2f", drac.Planets["Moon"])
	}

	// Offset should be 10
	if math.Abs(drac.Offset-10.0) > 0.001 {
		t.Errorf("offset should be 10, got %.2f", drac.Offset)
	}
}

func TestDraconic_SignShifts(t *testing.T) {
	tropical := map[string]float64{
		"Sun":     25.0,  // Aries
		"Mercury": 2.0,   // Aries
		"Venus":   350.0, // Pisces
	}
	nnLong := 10.0

	shifts := ComputeDraconicSignShifts(tropical, nnLong)

	// Sun: 25→15, both Aries — no shift
	if _, ok := shifts["Sun"]; ok {
		t.Error("Sun should not shift sign")
	}

	// Mercury: 2→352, Aries→Pisces — shift
	s, ok := shifts["Mercury"]
	if !ok {
		t.Fatal("Mercury should shift sign")
	}
	if s.TropSign != "Aries" || s.DracSign != "Pisces" {
		t.Errorf("Mercury shift: expected Aries→Pisces, got %s→%s", s.TropSign, s.DracSign)
	}

	// Venus: 350→340, Pisces→Pisces — no shift
	if _, ok := shifts["Venus"]; ok {
		t.Error("Venus should not shift sign")
	}
}

func TestDraconic_Bridges(t *testing.T) {
	// Tropical: Sun at 25°, Mars at 120°
	// Draconic (NN=10°): Sun at 15°, Mars at 110°
	// Bridge: draconic Sun (15°) conjunct tropical Mars (120°)? No — 105° apart
	// Bridge: draconic Mars (110°) trine tropical Sun (25°)? |110-25|=85, |85-120|=35 — no
	// Let's design a case that produces a bridge:
	// Tropical: Sun at 0°, Venus at 120°
	// NN at 120° → draconic Sun at 240°, draconic Venus at 0°
	// Bridge: draconic Venus (0°) conjunct tropical Sun (0°) — yes, 0° orb
	tropical := map[string]float64{
		"Sun":   0.0,
		"Venus": 120.0,
	}
	nnLong := 120.0

	bridges := ComputeDraconicBridges(tropical, nnLong, ClassicalPlanets, DefaultAspects(), 2.0)

	// Should find draconic Venus conjunct tropical Sun
	found := false
	for _, b := range bridges {
		if b.Planet1 == "Venus" && b.Planet2 == "Sun" && b.Aspect == "conjunction" {
			found = true
			if b.Orb > 0.01 {
				t.Errorf("Venus-Sun bridge orb should be ~0, got %.2f", b.Orb)
			}
		}
	}
	if !found {
		t.Error("missing draconic Venus conjunct tropical Sun bridge")
	}

	// Also: draconic Sun (240°) trine tropical Venus (120°) — |240-120|=120, exact
	found = false
	for _, b := range bridges {
		if b.Planet1 == "Sun" && b.Planet2 == "Venus" && b.Aspect == "trine" {
			found = true
		}
	}
	if !found {
		t.Error("missing draconic Sun trine tropical Venus bridge")
	}
}

func TestDraconic_Bridges_NoSelfPairs(t *testing.T) {
	// Same-name pairs (draconic Sun to tropical Sun) should be excluded.
	// They're always ~offset degrees apart and are tautological.
	tropical := map[string]float64{
		"Sun": 0.0,
	}
	nnLong := 0.0 // offset=0, draconic Sun = 0 — exact conjunction with itself

	bridges := ComputeDraconicBridges(tropical, nnLong, []string{"Sun"}, DefaultAspects(), 2.0)

	for _, b := range bridges {
		if b.Planet1 == b.Planet2 {
			t.Errorf("same-name bridge should be excluded: %s-%s", b.Planet1, b.Planet2)
		}
	}
}

func TestDraconic_Integration_AJ(t *testing.T) {
	// Cross-validate against known AJ draconic data.
	// NN at 2.16° offset. Mercury shifts Aquarius→Capricorn.
	// ASC shifts Scorpio→Libra.
	// Bridges: draconic Jupiter conjunct tropical Uranus (0.56°),
	//          draconic Jupiter conjunct tropical SouthNode (0.62°),
	//          draconic Uranus conjunct tropical SouthNode (0.98°),
	//          draconic Neptune conjunct tropical Mars (0.70°).
	ajTropical := map[string]float64{
		"Sun":     327.22, // Aquarius 27.22
		"Moon":    322.30, // Aquarius 22.30
		"Mercury": 302.10, // Aquarius 2.10
		"Venus":   12.60,  // Aries 12.60
		"Mars":    235.80, // Scorpio 25.80
		"Jupiter": 184.90, // Libra 4.90
		"Saturn":  21.50,  // Aries 21.50
		"Uranus":  185.00, // Libra 5.00
		"Neptune": 240.53, // Sagittarius 0.53
		"Pluto":   176.00, // Virgo 26.00
	}
	nnLong := 2.16 // Aries 2.16

	drac := ComputeDraconic(ajTropical, nnLong)

	// Mercury: 302.10 - 2.16 = 299.94 → Capricorn 29.94 (shift from Aquarius)
	mercDrac := drac.Planets["Mercury"]
	if math.Abs(mercDrac-299.94) > 0.1 {
		t.Errorf("draconic Mercury: expected ~299.94, got %.2f", mercDrac)
	}
	if SignForLongitude(mercDrac) != "Capricorn" {
		t.Errorf("draconic Mercury sign: expected Capricorn, got %s", SignForLongitude(mercDrac))
	}

	// Jupiter: 184.90 - 2.16 = 182.74 → Libra 2.74
	jupDrac := drac.Planets["Jupiter"]
	if math.Abs(jupDrac-182.74) > 0.1 {
		t.Errorf("draconic Jupiter: expected ~182.74, got %.2f", jupDrac)
	}

	// Bridges at 3° orb (test positions differ slightly from pyswisseph reference)
	// Use all 10 classical+outer planets since bridges involve Uranus and Neptune
	allPlanets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"}
	bridges := ComputeDraconicBridges(ajTropical, nnLong, allPlanets, DefaultAspects(), 3.0)

	find := func(p1, p2, asp string) *SynastryHit {
		for i := range bridges {
			b := &bridges[i]
			if b.Planet1 == p1 && b.Planet2 == p2 && b.Aspect == asp {
				return b
			}
		}
		return nil
	}

	// drac Jupiter conjunct tropical Uranus — the broadcast tower
	h := find("Jupiter", "Uranus", "conjunction")
	if h == nil {
		t.Error("missing draconic Jupiter conjunct tropical Uranus bridge")
	} else if h.Orb > 3.0 {
		t.Errorf("Jupiter-Uranus bridge orb too wide: %.2f", h.Orb)
	}

	// drac Neptune conjunct tropical Mars — soul intuition wired to action
	h = find("Neptune", "Mars", "conjunction")
	if h == nil {
		t.Error("missing draconic Neptune conjunct tropical Mars bridge")
	} else if h.Orb > 3.0 {
		t.Errorf("Neptune-Mars bridge orb too wide: %.2f", h.Orb)
	}

	// drac Jupiter trine tropical Mercury — soul wisdom to personality mind
	h = find("Jupiter", "Mercury", "trine")
	if h == nil {
		t.Error("missing draconic Jupiter trine tropical Mercury bridge")
	}

	// drac Uranus trine tropical Mercury — soul breakthrough to personality mind
	h = find("Uranus", "Mercury", "trine")
	if h == nil {
		t.Error("missing draconic Uranus trine tropical Mercury bridge")
	}

	// Sign shifts: only Mercury should shift (Aquarius→Capricorn)
	shifts := ComputeDraconicSignShifts(ajTropical, nnLong)
	if s, ok := shifts["Mercury"]; !ok {
		t.Error("Mercury should shift sign")
	} else if s.TropSign != "Aquarius" || s.DracSign != "Capricorn" {
		t.Errorf("Mercury shift: expected Aquarius→Capricorn, got %s→%s", s.TropSign, s.DracSign)
	}

	// Sun, Moon, Venus, Mars, Jupiter, Saturn should NOT shift
	noShift := []string{"Sun", "Moon", "Venus", "Mars", "Jupiter", "Saturn"}
	for _, p := range noShift {
		if _, ok := shifts[p]; ok {
			t.Errorf("%s should not shift sign", p)
		}
	}
}

func TestDraconicSynastry(t *testing.T) {
	// Two draconic charts: compute synastry between them.
	// Person A: NN=10°, Sun=25° → drac Sun=15°
	// Person B: NN=120°, Sun=15° → drac Sun=255°
	// Draconic synastry: dracA.Sun(15°) vs dracB.Sun(255°) — 120° apart = trine
	tropicalA := map[string]float64{"Sun": 25.0}
	tropicalB := map[string]float64{"Sun": 15.0}
	nnA := 10.0
	nnB := 120.0

	hits := ComputeDraconicSynastry(tropicalA, nnA, tropicalB, nnB, []string{"Sun"}, DefaultAspects(), 3.0)

	if len(hits) != 1 {
		t.Fatalf("expected 1 draconic synastry hit, got %d", len(hits))
	}
	if hits[0].Aspect != "trine" {
		t.Errorf("expected trine, got %s", hits[0].Aspect)
	}
	// |255-15| = 240, min(240, 120) = 120, |120-120| = 0
	if hits[0].Orb > 0.01 {
		t.Errorf("expected orb ~0, got %.2f", hits[0].Orb)
	}
}

func TestDraconic_Offset(t *testing.T) {
	// Offset is just the NN longitude normalized to 0-360
	// NN at 362° (2° past circle) → offset 2°
	nn := 362.0
	offset := normalizeLon(nn)
	if math.Abs(offset-2.0) > 0.001 {
		t.Errorf("offset for NN=362 should be 2, got %.2f", offset)
	}
}

func TestDraconicToTropical(t *testing.T) {
	// Roundtrip: tropical → draconic → tropical should be identity
	tropical := 150.0
	node := 45.0
	drac := normalizeLon(tropical - node)
	back := DraconicToTropical(drac, node)
	if math.Abs(back-tropical) > 0.001 {
		t.Errorf("roundtrip: expected %.2f, got %.2f", tropical, back)
	}

	// Wraparound: tropical 10°, node 350° → drac 20°, back 10°
	tropical = 10.0
	node = 350.0
	drac = normalizeLon(tropical - node)
	back = DraconicToTropical(drac, node)
	if math.Abs(back-tropical) > 0.001 {
		t.Errorf("wraparound roundtrip: expected %.2f, got %.2f", tropical, back)
	}
}

func TestDraconicSynastryFull(t *testing.T) {
	// Person A: NN=10°, Sun=25° → drac Sun=15°
	// Person B: NN=120°, Sun=15° → drac Sun=255°
	// drac-to-drac: dracA.Sun(15°) vs dracB.Sun(255°) — 120° apart = trine
	// tropA-to-dracB: tropA.Sun(25°) vs dracB.Sun(255°) — |255-25|=230, min(230,130)=130, |130-120|=10 — no hit at 3° orb
	// tropB-to-dracA: tropB.Sun(15°) vs dracA.Sun(15°) — 0° apart = conjunction
	tropicalA := map[string]float64{"Sun": 25.0}
	tropicalB := map[string]float64{"Sun": 15.0}
	nnA := 10.0
	nnB := 120.0

	result := ComputeDraconicSynastryFull(tropicalA, nnA, tropicalB, nnB, []string{"Sun"}, DefaultAspects(), 3.0)

	// drac-to-drac: 1 hit (trine)
	if len(result.DracToDrac) != 1 {
		t.Fatalf("drac-to-drac: expected 1 hit, got %d", len(result.DracToDrac))
	}
	if result.DracToDrac[0].Aspect != "trine" {
		t.Errorf("drac-to-drac: expected trine, got %s", result.DracToDrac[0].Aspect)
	}

	// tropA-to-dracB: 0 hits (25° vs 255° — 130° gap, no aspect at 3° orb)
	if len(result.TropAToDracB) != 0 {
		t.Errorf("tropA-to-dracB: expected 0 hits, got %d", len(result.TropAToDracB))
	}

	// tropB-to-dracA: 1 hit (conjunction at 0°)
	if len(result.TropBToDracA) != 1 {
		t.Fatalf("tropB-to-dracA: expected 1 hit, got %d", len(result.TropBToDracA))
	}
	if result.TropBToDracA[0].Aspect != "conjunction" {
		t.Errorf("tropB-to-dracA: expected conjunction, got %s", result.TropBToDracA[0].Aspect)
	}
	if result.TropBToDracA[0].Orb > 0.01 {
		t.Errorf("tropB-to-dracA: expected orb ~0, got %.2f", result.TropBToDracA[0].Orb)
	}
}

func TestDraconicSynastryFull_NoSelfPairs(t *testing.T) {
	// Unlike bridges (same-person tropical-to-draconic), synastry between
	// two people does NOT filter same-name pairs. Sun-to-Sun between two
	// people is a real soul contact. This test verifies that same-name
	// pairs ARE present in the output.
	tropicalA := map[string]float64{"Sun": 0.0}
	tropicalB := map[string]float64{"Sun": 0.0}
	nnA := 0.0
	nnB := 0.0

	result := ComputeDraconicSynastryFull(tropicalA, nnA, tropicalB, nnB, []string{"Sun"}, DefaultAspects(), 2.0)

	// drac-to-drac: Sun-Sun at 0° — should be present (real soul contact)
	found := false
	for _, h := range result.DracToDrac {
		if h.Planet1 == "Sun" && h.Planet2 == "Sun" {
			found = true
			if h.Aspect != "conjunction" {
				t.Errorf("drac-to-drac Sun-Sun: expected conjunction, got %s", h.Aspect)
			}
		}
	}
	if !found {
		t.Error("drac-to-drac: missing Sun-Sun conjunction (real soul contact)")
	}

	// tropA-to-dracB: tropA.Sun(0°) vs dracB.Sun(0°) — conjunction
	found = false
	for _, h := range result.TropAToDracB {
		if h.Planet1 == "Sun" && h.Planet2 == "Sun" {
			found = true
		}
	}
	if !found {
		t.Error("tropA-to-dracB: missing Sun-Sun conjunction")
	}

	// tropB-to-dracA: tropB.Sun(0°) vs dracA.Sun(0°) — conjunction
	found = false
	for _, h := range result.TropBToDracA {
		if h.Planet1 == "Sun" && h.Planet2 == "Sun" {
			found = true
		}
	}
	if !found {
		t.Error("tropB-to-dracA: missing Sun-Sun conjunction")
	}
}

func TestCrossSystemTransitComparison(t *testing.T) {
	// Simulate: natal draconic positions (same in both zodiacs)
	// Transiting positions in tropical vs sidereal (differ by ~24°)
	// Verify the comparison correctly identifies survivors vs zodiac-dependent

	natalDrac := map[string]float64{
		"Sun":     325.3,
		"Moon":    320.1,
		"Mercury": 299.9,
		"Venus":   22.0,
		"Mars":    233.6,
		"Jupiter": 182.8,
		"Saturn":  329.1,
	}

	// Transiting positions: tropical
	tropTransits := map[string]float64{
		"Sun":     325.3, // exact conjunction with natal D.Sun
		"Moon":    140.1, // 180° from natal D.Moon (opposition)
		"Mars":    233.6, // exact conjunction with natal D.Mars
		"Jupiter": 62.8,  // 120° from natal D.Jupiter (trine)
	}

	// Transiting positions: sidereal (shifted by ~24° Lahiri)
	// Sun: 325.3 - 24.0 = 301.3 — no longer conjunct D.Sun (325.3)
	// Moon: 140.1 - 24.0 = 116.1 — no longer opposite D.Moon (320.1)
	// Mars: 233.6 - 24.0 = 209.6 — no longer conjunct D.Mars (233.6)
	// Jupiter: 62.8 - 24.0 = 38.8 — still trine D.Jupiter? 182.8 - 38.8 = 144°, not 120°
	sidTransits := map[string]float64{
		"Sun":     301.3,
		"Moon":    116.1,
		"Mars":    209.6,
		"Jupiter": 38.8,
	}

	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemTransits(natalDrac, tropTransits, sidTransits, aspects, orb)

	// Survivors: aspects that appear in BOTH zodiacs
	// Zodiac-dependent: aspects in one but not the other

	if len(result.Survivors) != 0 {
		t.Errorf("expected 0 survivors, got %d: %v", len(result.Survivors), result.Survivors)
	}

	// Both tropical and sidereal should produce hits
	if len(result.TropicalOnly) == 0 {
		t.Error("expected some tropical-only hits")
	}
	if len(result.SiderealOnly) == 0 {
		t.Error("expected some sidereal-only hits")
	}

	// Verify the tropical-only entries have correct data for known hits
	foundSun := false
	foundMars := false
	for _, h := range result.TropicalOnly {
		if h.TransitPlanet == "Sun" && h.NatalPlanet == "Sun" {
			foundSun = true
			if h.Aspect != "conjunction" {
				t.Errorf("Sun-Sun aspect: expected conjunction, got %s", h.Aspect)
			}
		}
		if h.TransitPlanet == "Mars" && h.NatalPlanet == "Mars" {
			foundMars = true
		}
	}
	if !foundSun {
		t.Error("missing Sun-Sun tropical-only hit")
	}
	if !foundMars {
		t.Error("missing Mars-Mars tropical-only hit")
	}
}

func TestCrossSystemTransitComparison_NoSurvivor(t *testing.T) {
	// With 3° orb and ~24° ayanamsa shift, no aspect survives.
	natalDrac := map[string]float64{"Sun": 100.0}

	tropTransits := map[string]float64{"Sun": 100.5}  // 0.5° conjunction
	sidTransits := map[string]float64{"Sun": 76.5}    // shifted by 24°

	aspects := DefaultAspects()
	orb := 3.0

	result := CompareCrossSystemTransits(natalDrac, tropTransits, sidTransits, aspects, orb)

	if len(result.Survivors) != 0 {
		t.Errorf("with 3° orb and 24° shift, expected 0 survivors, got %d", len(result.Survivors))
	}
	if len(result.TropicalOnly) != 1 {
		t.Errorf("expected 1 tropical-only, got %d", len(result.TropicalOnly))
	}
}

func TestCrossSystemTransitComparison_WideOrbSurvivor(t *testing.T) {
	// With a wide enough orb, an aspect can survive the zodiac shift
	natalDrac := map[string]float64{"Sun": 100.0}

	// Tropical: transit at 98.0° — 2° from conjunction
	// Sidereal: transit at 98.0 - 24.0 = 74.0° — 26° from conjunction
	// With 30° orb, both are within orb → survivor
	tropTransits := map[string]float64{"Sun": 98.0}
	sidTransits := map[string]float64{"Sun": 74.0}

	aspects := DefaultAspects()
	orb := 30.0

	result := CompareCrossSystemTransits(natalDrac, tropTransits, sidTransits, aspects, orb)

	if len(result.Survivors) != 1 {
		t.Fatalf("expected 1 survivor with 30° orb, got %d", len(result.Survivors))
	}
	if result.Survivors[0].Aspect != "conjunction" {
		t.Errorf("expected conjunction, got %s", result.Survivors[0].Aspect)
	}
	if len(result.TropicalOnly) != 0 {
		t.Errorf("expected 0 tropical-only, got %d", len(result.TropicalOnly))
	}
}
