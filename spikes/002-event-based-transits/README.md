# Spike 002: Event-Based Transit Detection

## Question

Can we find transit periods by detecting discrete events (orb-threshold crossings) instead of scanning every day?

## Verdict: VALIDATED — IMPLEMENTED

**Built into `internal/dignity/transit.go` as `FindTransitPeriods` (2026-07-10).**

### What worked

1. **Event-based detection correctly finds transit boundaries.** Saturn ☌ Saturn shows two distinct periods in 2026-2027: a direct pass (52.3 days) and a retrograde pass (76.9 days). The retrograde flag correctly identifies which contacts are retrograde.

2. **Ephemeris caching is the key insight.** Precomputing all planet positions at 1-day resolution (7310 SWE calls, 225ms) eliminates SWE calls from the refinement loop. The scan itself runs in 5-10ms — a ~3000x speedup over calling SWE directly.

3. **Correctly reports zero when no transit exists.** Saturn □ Moon returns 0 periods in 2026-2027 because Saturn never comes within 3° of squaring natal Moon. Spike 001's "Saturn square Moon" was a false positive from scan-window boundary confusion.

4. **Retrograde multi-pass is handled.** The data model (`TransitPeriod` with `[]Contact`) naturally represents direct + retrograde passes as separate periods. Each contact carries a `Retrograde` flag.

5. **Performance scales linearly with window size.** 1 year: 5ms, 2 years: 10ms. Cache build dominates (225ms for 2 years) but is a one-time cost.

### What didn't

1. **1-day coarse step misses Moon transits.** The Moon moves ~13°/day, so a transit can enter and leave a 3° orb between daily samples. For daily horoscopes that include Moon transits, need 4×/day or 6×/day step. For weekly/monthly/yearly horoscopes, skipping the Moon is standard practice.

2. **Window-boundary egress is partial.** Period 2 of Saturn ☌ Saturn shows egress at 2028-01-01 with orb 0.39° — the real egress is after the window. This is correct behavior but should be flagged as "still in orb at window end."

3. **No station-aware period merging.** Saturn ☌ Saturn shows two separate periods (direct + retrograde) when they're really one continuous transit with a retrograde phase. The data model supports merging them into a single `TransitPeriod` with 5+ contacts, but the current algorithm doesn't do it.

### Surprises

1. **Spike 001's data was wrong.** The "Saturn square Moon" transit it reported (2026-05-25 to 2026-09-29) doesn't exist. Saturn's angular distance from natal Moon in 2026 ranges from 34° to 52° — nowhere near 90°. Spike 001 was reporting scan-window boundaries, not transit boundaries.

2. **The cache is the algorithm.** Once you have the ephemeris in memory, the "event detection" is just iterating over cached arrays and doing simple arithmetic. The refinement (binary search + golden section) uses the cache, not SWE. This is what makes it fast.

3. **1485 periods in 2 years is a lot.** That's 10 planets × 9 natal planets × 5 aspects × ~3.3 periods each. For horoscope generation, you'd filter this down significantly — only show periods active in the current week/month, only show significant aspects, etc.

### Recommendation for the real build

1. **Add `FindTransitPeriods` to `internal/dignity/transit.go`** using the cached approach. The function signature:
   ```go
   func FindTransitPeriods(
       natalLongs map[string]float64,
       natalPlanets []string,
       startDate, endDate string,
       aspects []AspectDef,
       orbDeg float64,
       coarseStepDays float64, // 1.0 for slow planets, 0.25 for fast
   ) ([]TransitPeriod, error)
   ```

2. **New types in `transit.go`:**
   ```go
   type ContactType int  // Ingress, Peak, Egress
   type Contact struct {
       Type       ContactType
       JD         float64
       Orb        float64
       Retrograde bool
   }
   type TransitPeriod struct {
       TransitPlanet string
       NatalPlanet   string
       Aspect        string
       Contacts      []Contact
   }
   ```

3. **Keep `ScanTransits` for backward compatibility** but deprecate it. It can call `FindTransitPeriods` internally and flatten to `[]TransitHit`.

4. **Add `CompactTransitPeriods`** that merges adjacent periods where the transit planet is retrograde between them (station-aware merging).

5. **For Moon transits**, use `coarseStepDays=0.25` (6-hour step). This 4×'s the cache size but is only needed for daily horoscopes.

6. **Cache should be internal to the function**, not exposed. Build it once per call, use it for all planet pairs, discard it.

## Data model comparison

| | Old (`TransitHit`) | New (`TransitPeriod`) |
|---|---|---|
| Single date | ✅ | ❌ (replaced by Contacts) |
| Ingress/egress/peak | ❌ | ✅ |
| Retrograde flag | ❌ | ✅ |
| Multi-pass support | ❌ | ✅ |
| Duration | ❌ (derivable from Compact) | ✅ (Contacts[last].JD - Contacts[0].JD) |
| JSON-serializable | ✅ | ✅ |
