# Koiné Astrology — Phase 2.5: Rich Interpretive Content

**Goal:** Replace thin template-assembled planet-in-sign and planet-in-house descriptions with rich, authored interpretations. 7 planets × 12 signs = 84 planet-in-sign. 7 planets × 12 houses = 84 planet-in-house. Plus richer house meanings.

**Why:** The synthesis engine is solid but the content it synthesizes is one-liners. "Sun in Aries: core identity, cardinal fire — initiatory, direct, competitive. in domicile." That's a label, not a reading. The synthesis can't produce good output from thin input.

**Tech Stack:** Go 1.21, `internal/dignity/interpretation.go`, no new dependencies.

**Approach:** Replace the `planetDescriptions`, `signDescriptions`, and `houseDescriptions` maps with richer content. Add a new `planetInSign` map that combines planet+sign into a single rich interpretation (replacing the template assembly in `InterpretPlanetInSign`). Add a new `planetInHouse` map for planet+house combinations.

**Current state:**
- `planetDescriptions` — 7 entries, one phrase each (e.g., "core identity, vitality, conscious self")
- `signDescriptions` — 12 entries, one phrase each (e.g., "cardinal fire — initiatory, direct, competitive")
- `houseDescriptions` — 12 entries, one phrase each (e.g., "self, persona, physical body, how you show up")
- `InterpretPlanetInSign` assembles: `{planetDesc}, {signDesc}. {dignity}.` — a label
- `InterpretPlanetInHouse` assembles: `{planetDesc} in the {houseDesc}.` — a label

**Target state:**
- `planetInSign` map — 84 entries, each 2-4 sentences describing how that planet expresses in that sign
- `planetInHouse` map — 84 entries, each 2-4 sentences describing how that planet expresses in that house
- `houseDescriptions` — expanded to 2-3 sentences per house
- `InterpretPlanetInSign` uses `planetInSign[planet][sign]` + dignity note
- `InterpretPlanetInHouse` uses `planetInHouse[planet][house]`
- Synthesis engine unchanged — it already calls these functions, so richer input automatically produces richer output

**Planet order for authoring:** Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn
**Sign order:** Aries, Taurus, Gemini, Cancer, Leo, Virgo, Libra, Scorpio, Sagittarius, Capricorn, Aquarius, Pisces

---

### Task 1: Expand house descriptions (12 houses)

Replace the one-liner `houseDescriptions` with 2-3 sentence descriptions that give each house domain, quality, and life area.

File: `internal/dignity/interpretation.go:120-133`

### Task 2: Write planet-in-sign interpretations (84 entries)

Create `planetInSign` map: `map[string]map[string]string` — planet → sign → 2-4 sentence interpretation.

Each entry should cover:
- How the planet's nature expresses through the sign's element and modality
- The felt experience of this placement
- A distinctive quality or tension
- Dignity is handled separately (appended by InterpretPlanetInSign)

File: `internal/dignity/interpretation.go` — new map after signDescriptions

### Task 3: Write planet-in-house interpretations (84 entries)

Create `planetInHouse` map: `map[string]map[int]string` — planet → house → 2-4 sentence interpretation.

Each entry should cover:
- How the planet's domain manifests in this house's life area
- The native's experience of this placement
- What it drives or seeks

File: `internal/dignity/interpretation.go` — new map after planetInSign

### Task 4: Update InterpretPlanetInSign and InterpretPlanetInHouse

Replace template assembly with map lookup. Keep dignity note appended.

### Task 5: Run tests, verify build, commit

---

**Authoring notes:**
- These are Koiné interpretations — 2-state dignity, whole-sign houses, 3 universal aspects
- Tone: direct, observational, no fortune-telling. Describe the configuration, not the destiny.
- Avoid: "you will," "you are," "your destiny." Prefer: "this placement suggests," "the native experiences," "there is a tendency toward."
- Each entry should be distinct — not just remixing the same phrases.
- Quality over speed. If a batch needs review, pause and ask.
