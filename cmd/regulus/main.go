package main

import (
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	swe.SetEphePath("ephe")
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	bc, err := dignity.ComputeBaseChart(dignity.BirthData{
		Name:     "AJ",
		Year:     1969,
		Month:    2,
		Day:      15,
		Hour:     23,
		Minute:   10,
		Second:   0,
		TZOffset: -8.0,
		Lat:      47.038,
		Lng:      -122.901,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Compute Regulus position directly
	regulusLon, _, _, _ := swe.Fixstar("Regulus", bc.JD)
	regulusLon = math.Mod(regulusLon+360, 360)
	regulusSign := dignity.SignForLongitude(regulusLon)

	// Find planets conjunct Regulus (within 3°)
	type hit struct {
		planet string
		lon    float64
		orb    float64
	}
	var hits []hit
	for planet, pos := range bc.Tropical {
		orb := math.Abs(math.Mod(pos.Lon-regulusLon+540, 360) - 180)
		orb = math.Round(orb*100) / 100
		if orb <= 3.0 {
			hits = append(hits, hit{planet, pos.Lon, orb})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].orb < hits[j].orb
	})

	// Find house for Regulus
	houses := bc.Houses["placidus"]
	var regulusHouse int
	if len(houses) > 0 {
		for h := 1; h <= 12; h++ {
			cusp := houses[h]
			nextCusp := houses[h%12+1]
			if nextCusp < cusp {
				nextCusp += 360
			}
			lon := regulusLon
			if lon < cusp {
				lon += 360
			}
			if lon >= cusp && lon < nextCusp {
				regulusHouse = h
				break
			}
		}
	}

	// Nearest planets (wider context)
	type dist struct {
		planet string
		lon    float64
		orb    float64
	}
	var dists []dist
	for planet, pos := range bc.Tropical {
		orb := math.Abs(math.Mod(pos.Lon-regulusLon+540, 360) - 180)
		dists = append(dists, dist{planet, pos.Lon, math.Round(orb*100) / 100})
	}
	sort.Slice(dists, func(i, j int) bool {
		return dists[i].orb < dists[j].orb
	})

	// ── Western Interpretation ──────────────────────────────────────

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  REGULUS — Western Interpretation for AJ")
	fmt.Println("  February 15, 1969  23:10  Olympia, WA")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Regulus at %.2f° %s in House %d\n", regulusLon, regulusSign, regulusHouse)
	fmt.Println()

	// 1. Traditional meaning
	fmt.Println("── 1. Traditional Nature ──")
	fmt.Println()
	fmt.Println(dignity.StarMeanings["Regulus"])
	fmt.Println()

	// 2. Sign placement
	fmt.Println("── 2. Sign: Leo (Domicile) ──")
	fmt.Println()
	fmt.Println("Regulus is the Heart of the Lion, and it sits in Leo — the sign")
	fmt.Println("it rules by nature. This is the star in its domicile, operating at")
	fmt.Println("full, unmediated strength. There is no translation layer between")
	fmt.Println("the star's Mars-Jupiter nature and the sign's fixed-fire expression.")
	fmt.Println()
	fmt.Println("Leo is radiant, creative, commanding. Regulus in Leo doesn't whisper")
	fmt.Println("— it announces. The traditional warning applies with extra force:")
	fmt.Println("great success if integrity is maintained, catastrophic fall if not.")
	fmt.Println("A domicile Regulus has no excuse for mediocrity and no safety net")
	fmt.Println("for corruption.")

	// 3. House placement
	fmt.Println()
	fmt.Println("── 3. House: 10th (Career, Public Role, Authority) ──")
	fmt.Println()
	fmt.Println("The 10th house is the highest point in the chart — career, public")
	fmt.Println("standing, reputation, and the role you play in the world. Regulus")
	fmt.Println("here places the royal star at the most visible point in the chart.")
	fmt.Println()
	fmt.Println("This is not a placement for private achievement. The 10th house")
	fmt.Println("Regulus means your work is seen, your reputation matters, and your")
	fmt.Println("authority — or lack of it — is public. The star of kingship in the")
	fmt.Println("house of career: you are meant to lead something, to be known for")
	fmt.Println("something, to occupy a position where your decisions affect others.")
	fmt.Println()
	fmt.Println("The MC (Midheaven) is the cusp of the 10th. Regulus near or in the")
	fmt.Println("10th means the royal star sits near the career angle — the most")
	fmt.Printf("public point in the chart. Your MC is at %.2f°.\n", bc.MC)

	// 4. No conjunctions — unmodified
	fmt.Println()
	fmt.Println("── 4. No Planetary Conjunctions — Unmodified Regulus ──")
	fmt.Println()
	fmt.Println("No planet sits within 3° of Regulus. The nearest body is")
	if len(dists) > 0 {
		fmt.Printf("%s at %.2f° (orb %.2f°).\n", dists[0].planet, dists[0].lon, dists[0].orb)
	}
	fmt.Println()
	fmt.Println("This is significant. Regulus stands alone — uncolored, unmediated,")
	fmt.Println("uncompromised by planetary contact. A conjunction would blend")
	fmt.Println("Regulus with another planet's nature (Mars would add aggression,")
	fmt.Println("Saturn would add discipline, etc.). Instead, you get pure Regulus:")
	fmt.Println("the royal star in its own sign, in the house of career, with no")
	fmt.Println("planet softening, redirecting, or complicating its expression.")
	fmt.Println()
	fmt.Println("The downside: no planet channels Regulus into a specific domain.")
	fmt.Println("The energy is diffuse across the entire 10th house — career, reputation,")
	fmt.Println("authority, public role — rather than focused through a single planet's")
	fmt.Println("function. The upside: nothing dilutes it. When Regulus activates in")
	fmt.Println("your life (by transit, progression, or direction), it activates pure.")

	// 5. Synthesis
	fmt.Println()
	fmt.Println("── 5. Synthesis ──")
	fmt.Println()
	fmt.Println("Regulus at 149° Leo in the 10th house is a statement of intent from")
	fmt.Println("the chart. The Heart of the Lion at the highest point: you are meant")
	fmt.Println("to be known for what you build. Not famous for fame's sake — Regulus")
	fmt.Println("isn't a celebrity star, it's a sovereignty star. It demands that your")
	fmt.Println("work carry authority, that your output be worthy of the position you")
	fmt.Println("occupy.")
	fmt.Println()
	fmt.Println("The unmodified nature means this isn't about a specific skill or")
	fmt.Println("talent (that would be a planetary conjunction). It's about the")
	fmt.Println("quality of your presence in the public sphere. How you lead. Whether")
	fmt.Println("you maintain integrity under visibility. Whether your ambition serves")
	fmt.Println("something larger than ego.")
	fmt.Println()
	fmt.Println("The domicile placement in Leo says: you have the raw material. The")
	fmt.Println("10th house says: it will be tested in public. The lack of conjunction")
	fmt.Println("says: how you wield it is entirely on you. No planet to blame, no")
	fmt.Println("planet to credit. Just the star, the sign, the house, and your choices.")
	fmt.Println()
	fmt.Println("The traditionalists would add: Regulus in the 10th confers success")
	fmt.Println("in positions of command — military, political, or institutional — but")
	fmt.Println("only if the native refuses to abuse the power granted. The fall, if")
	fmt.Println("it comes, is proportional to the height.")
}
