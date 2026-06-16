package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/aj-nt/empirical"
    "github.com/aj-nt/empirical/internal/swe"
)

func main() {
    cacheDir, err := empirical.EnsureEpheCache()
    if err != nil {
        fmt.Fprintf(os.Stderr, "ephe cache: %v\n", err)
        os.Exit(1)
    }
    swe.SetEphePath(cacheDir)

    // J2000.0
    jd := 2451545.0

    stars := []string{
        "41 Arietis",
        "Alhena",
        "Alterf",
        "Zavijava",
        "iota Virginis",
        "Shaula",
        "Albaldah",
        "Sadachbia",
        "Alpheratz",
        "Mirach",
        "Praesepe",
        "Zubeneschamali",
        "Algedi",
        "Hamal",
        "Aldebaran",
        "Regulus",
        "Denebola",
        "Castor",
        "Pollux",
        "Alcyone",
        "Electra",
        "Ain",
    }

    results := make(map[string]map[string]interface{})
    for _, name := range stars {
        lon, lat, _, _ := swe.Fixstar(name, jd)
        mag := swe.FixstarMag(name)
        results[name] = map[string]interface{}{
            "lon": lon,
            "lat": lat,
            "mag": mag,
        }
    }

    out, _ := json.MarshalIndent(results, "", "  ")
    fmt.Println(string(out))
}
