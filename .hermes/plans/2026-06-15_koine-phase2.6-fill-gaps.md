# Koiné Astrology — Phase 2.6: Fill Content Gaps

**Goal:** Complete the interpretive content before family validation. Five gaps to fill.

**Tech Stack:** Go 1.21, `internal/dignity/`, no new dependencies.

---

### Gap 1: Outer planet interpretations (72 entries)

Uranus, Neptune, Pluto currently fall back to template assembly. Need:
- 36 planet-in-sign (3 planets × 12 signs)
- 36 planet-in-house (3 planets × 12 houses)

Files: `planet_in_sign.go`, `planet_in_house.go`

### Gap 2: Fixed star integration in synthesis

The synthesis engine ignores fixed stars. Add a body section that mentions the most significant star conjunctions (top 3 by orb).

File: `synthesis.go` — new `buildStarSection` function, new `StarConjunction` param

### Gap 3: Fortune integration in synthesis

Fortune (Pars Fortunae) is computed but not mentioned in synthesis. Add a brief mention in the opening or a body section.

File: `synthesis.go` — integrate Fortune sign+house into opening or body

### Gap 4: Synastry synthesis endpoint

New `/api/synastry-interpretation` endpoint that produces a synthesis reading for two charts. Uses the same engine but adapted for relationship reading: mutual aspects, house overlays, composite themes.

Files: `server.go` (new handler), `cmd/recover/main.go` (new compute function), `synthesis.go` (new `SynthesizeSynastry` function)

### Gap 5: Transit synthesis endpoint

New `/api/transit-interpretation` endpoint that produces a synthesis reading of current transits to the natal chart. Focuses on the most significant transits and their narrative meaning.

Files: `server.go` (new handler), `cmd/recover/main.go` (new compute function), `synthesis.go` (new `SynthesizeTransits` function)

---

**Order:** Gap 1 → Gap 2 → Gap 3 → Gap 4 → Gap 5
**Verification:** Build + tests after each gap.
