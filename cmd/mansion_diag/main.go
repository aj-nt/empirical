package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/aj-nt/empirical"
    "github.com/aj-nt/empirical/internal/dignity"
    "github.com/aj-nt/empirical/internal/swe"
)

func main() {
    cacheDir, err := empirical.EnsureEpheCache()
    if err != nil {
        fmt.Fprintf(os.Stderr, "ephe cache: %v\n", err)
        os.Exit(1)
    }
    swe.SetEphePath(cacheDir)

    // AJ's chart: Feb 15 1969, 23:10 PST (UTC-8), Olympia WA
    jd := swe.Julday(1969, 2, 15, 23.1667+8, true) // UT
    ayan := swe.GetAyanamsaUT(jd)

    planets := []struct{name string; id int}{
        {"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
        {"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
        {"Saturn", swe.SATURN},
    }

    tropical := make(map[string]float64)
    for _, p := range planets {
        lon, _, _, _ := swe.CalcUT(jd, p.id)
        for lon < 0 { lon += 360 }
        for lon >= 360 { lon -= 360 }
        tropical[p.name] = lon
    }

    result := dignity.ComputeMansionThreeWayConvergence("AJ", tropical, ayan)
    out, _ := json.MarshalIndent(result, "", "  ")
    fmt.Println(string(out))
}
