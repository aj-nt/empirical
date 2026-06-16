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

    // Run three-way null model
    cfg := dignity.NullModelConfig{
        Name:            "Three-Way Brightness-Weighted",
        Iterations:      10000,
        NakshatraDraws:  27,
        XiuDraws:        28,
        FaintThreshold:  2.5,
        Seed:            42,
        UseEclipticWeight: false,
    }
    result := dignity.RunThreeWayNullBrightness(cfg)
    out, _ := json.MarshalIndent(result, "", "  ")
    fmt.Println(string(out))
}
