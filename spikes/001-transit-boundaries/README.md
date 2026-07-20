# 001: Transit Boundary Detection via Swiss Ephemeris Crossing Search

## Question

Can `swe_helio_cross_ut` find the exact ingress/egress dates when a transiting
planet enters and leaves orb of a natal aspect?

## Verdict: INVALIDATED (for helio_cross) → VALIDATED (for binary search)

### What worked

- **Binary search with `swe_calc_ut`** finds exact ingress/egress to sub-second
  precision in ~30 iterations. For Saturn square Cait's Moon:
  - Ingress: 2026-05-25 11:29 UT
  - Peak: 2026-08-05 13:41 UT (exact square, 0.0000° orb)
  - Egress: 2026-09-29 20:14 UT
  - Duration: 127.4 days
  - The scan window (July 10-12) was a 3-day slice of a 127-day transit.

- **Golden section search** finds the peak orb (exact aspect) between ingress
  and egress with the same precision.

### What didn't

- **`swe_helio_cross_ut` is heliocentric.** It finds when a planet crosses a
  longitude in its *heliocentric* orbit, not its geocentric apparent position.
  SWE has no built-in geocentric planet crossing function. The `dir` parameter
  controls planet motion direction (direct/retrograde), not search direction.

- **Day-by-day scan misses short transits.** Mercury square Moon produced 4
  ingress/egress pairs with the helio_cross approach but only 1 with 1-day
  scanning. Fast planets need sub-day resolution.

### Surprises

- Saturn's ingress was 46 days *before* the scan window started. The current
  `start_date`/`end_date` output is completely misleading — it reports scan
  window boundaries, not transit boundaries.

- The binary search approach is simpler than expected. No new CGo wrappers
  needed beyond what already exists. The `ComputeFunc` injection pattern
  already supports this.

### Recommendation for the real build

1. **Add `FindTransitBoundaries` to `transit.go`** — takes a `ComputeFunc`,
   natal longitude, transit planet ID, aspect angle, orb, and a search window.
   Returns `[]BoundaryPair` (ingress JD, egress JD, peak JD).

2. **Change the data model** — `TransitHit` currently has a single `Date`.
   Replace with `[]TransitPeriod` where each period has `Ingress`, `Egress`,
   `PeakOrb`, `PeakDate`. This handles retrograde re-entries.

3. **Algorithm**: For each transit planet × natal planet × aspect:
   - Coarse scan at 1-day resolution to find in-orb periods
   - Binary search to refine each ingress/egress to sub-second precision
   - Golden section search for peak orb within each period

4. **Performance**: ~30 `swe_calc_ut` calls per boundary × 2 boundaries ×
   N transits. For a 3-day window with 24 transit planets × 10 natal planets
   × 5 aspects, that's ~7,200 SWE calls. At ~0.1ms each, ~0.7 seconds total.
   Acceptable.

5. **Keep `HelioCrossUT` wrapper** — it's already added to `swe.go` and
   compiles. It's not useful for transits but may be useful for heliocentric
   work later. Don't remove it.
