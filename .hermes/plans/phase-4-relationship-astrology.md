# Phase 4: Relationship Astrology — Implementation Plan

**Goal:** Build a dedicated Synastry tab that surfaces all existing backend relationship endpoints in readable, interactive views.

**Backend already has:** synastry, composite, draconic-synastry, draconic-synastry-full, draconic, draconic-transits-cross.

**Frontend gap:** Types and API methods exist. No UI components. The "Synastry" tab in App.tsx exists but has no rendering block.

## Tasks

### 4.1: Partner data entry
- Partner chart form (name, birth data) — stored in IndexedDB alongside main chart
- Partner selector dropdown in Synastry tab
- Quick-select from existing saved charts

### 4.2: Synastry view
- Fetch `/api/synastry` with main + partner data
- Display: inter-aspects grouped by planet pair, sorted by orb
- House overlays (partner's planets in main's houses, and vice versa)
- Aspect summary (counts by type)

### 4.3: Composite view
- Fetch `/api/composite` with main + partner data
- Display: composite chart planets, houses, aspects
- Composite chart wheel (reuse ChartWheel with composite data)

### 4.4: Draconic synastry view
- Fetch `/api/draconic-synastry-full`
- Display: three-layer comparison (tropical-tropical, draconic-draconic, tropical-draconic)
- Highlight bridges (aspects that appear in all three layers)

### 4.5: Relationship synthesis
- Summary card: total aspects, strongest connections, house emphasis
- Element balance comparison between the two charts
- Key themes (e.g., "Saturn-heavy — karmic/structural bond")

### 4.6: End-to-end verification
- TypeScript, build, tests, browser verify
