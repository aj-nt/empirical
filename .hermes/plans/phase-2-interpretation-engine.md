# Phase 2: Interpretation Engine — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a full natal chart interpretation panel that displays all sections of the existing `ChartInterpretation` backend data in a readable, navigable format, plus transit interpretation and text management.

**Architecture:** The backend already produces rich `ChartInterpretation` JSON via `WesternFromBase()` and `KoinéFromBase()`. The frontend currently only shows a tooltip on click. Phase 2 builds a dedicated interpretation view with collapsible sections, a transit interpretation view, and a text management UI for customizing interpretation templates.

**Tech Stack:** Go (existing backend), React/TypeScript (existing frontend), IndexedDB (existing local storage)

---

## Pre-Plan Verification

Before writing tasks, verify what already exists:

- [x] `planet_in_house.go` — 144 rich interpretations (12 planets × 12 houses) ✅
- [x] `western_interpretation.go` — planet-in-sign, aspect, pattern, pair-dynamics, star conjunction ✅
- [x] `koine_interpretation.go` — Hellenistic planet-in-sign, planet-in-house, aspect, pattern ✅
- [x] `WesternFromBase()` — full chart interpretation with all sections ✅
- [x] `KoinéFromBase()` — full Koiné chart interpretation ✅
- [x] `/api/interpretation` endpoint — returns `ChartInterpretation` JSON ✅
- [x] `WheelTooltip.tsx` — preloads interpretation data, shows tooltip on click ✅
- [ ] Full interpretation panel — **MISSING**
- [ ] Transit interpretation view — **MISSING**
- [ ] Interpretation text management UI — **MISSING**

---

### Task 1: Add interpretation tab to App.tsx

**Objective:** Add a new "Interpretation" tab that loads and displays the full `ChartInterpretation` data.

**Files:**
- Modify: `web/src/App.tsx`
- Create: `web/src/components/interpretation/InterpretationPanel.tsx`

**Step 1: Add interpretation tab to the tab bar**

In `App.tsx`, add `'interpretation'` to the `view` state union type and add a tab button in the view selector bar (alongside Wheel, Table, Transits, etc.).

**Step 2: Create InterpretationPanel component**

```tsx
// web/src/components/interpretation/InterpretationPanel.tsx
import { useState, useEffect } from 'react';
import type { BirthData, ChartInterpretation } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
  system?: 'western' | 'koiné';
}

export function InterpretationPanel({ data, system = 'western' }: Props) {
  const [interp, setInterp] = useState<ChartInterpretation | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.interpretation(data, system, 3)
      .then(setInterp)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng, system]);

  if (loading) return <p className="text-yellow text-sm">Loading interpretation...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!interp) return <p className="text-muted text-sm">No interpretation available.</p>;

  return (
    <div className="space-y-6 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 200px)' }}>
      {/* Planet in Sign */}
      <Section title="Planets in Signs">
        {interp.planet_signs?.map((text, i) => (
          <p key={i} className="text-sm mb-2">{text}</p>
        ))}
      </Section>

      {/* Planet in House */}
      <Section title="Planets in Houses">
        {interp.planet_houses?.map((text, i) => (
          <p key={i} className="text-sm mb-2">{text}</p>
        ))}
      </Section>

      {/* Aspects */}
      <Section title="Aspects">
        {interp.aspects?.map((text, i) => (
          <p key={i} className="text-sm mb-2">{text}</p>
        ))}
      </Section>

      {/* Patterns */}
      {interp.patterns && interp.patterns.length > 0 && (
        <Section title="Patterns">
          {interp.patterns.map((text, i) => (
            <p key={i} className="text-sm mb-2">{text}</p>
          ))}
        </Section>
      )}

      {/* Element Balance */}
      {interp.element_balance && (
        <Section title="Element Balance">
          <ElementBar balance={interp.element_balance} />
        </Section>
      )}

      {/* Modality Balance */}
      {interp.modality_balance && (
        <Section title="Modality Balance">
          <ModalityBar balance={interp.modality_balance} />
        </Section>
      )}

      {/* Rulership Chains */}
      {interp.rulership_chains && (
        <Section title="Rulership Chains">
          {Object.entries(interp.rulership_chains).map(([house, chain]) => (
            <p key={house} className="text-sm mb-1">
              <span className="text-muted">House {house}:</span> {chain.join(' → ')}
            </p>
          ))}
        </Section>
      )}

      {/* Midpoints */}
      {interp.midpoints && interp.midpoints.length > 0 && (
        <Section title="Midpoints">
          {interp.midpoints.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Stars */}
      {interp.stars && interp.stars.length > 0 && (
        <Section title="Fixed Star Contacts">
          {interp.stars.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(true);
  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-4 py-2 bg-surface text-sm font-semibold hover:bg-bg-hover"
      >
        {title}
        <span className="text-muted">{open ? '▾' : '▸'}</span>
      </button>
      {open && <div className="px-4 py-3">{children}</div>}
    </div>
  );
}
```

**Step 3: Wire into App.tsx**

In the view selector, add the interpretation tab. When `view === 'interpretation'`, render `<InterpretationPanel data={activeChart.birthData} />`.

**Step 4: Verify**

- Run `npx tsc --noEmit` — should be clean
- Run `npm run build` — should succeed
- Open browser, navigate to a chart, click "Interpretation" tab — should see collapsible sections

---

### Task 2: Add TypeScript types for ChartInterpretation

**Objective:** Ensure `web/src/lib/types.ts` has complete types matching the backend `ChartInterpretation` struct.

**Files:**
- Modify: `web/src/lib/types.ts`

**Step 1: Add missing types**

```typescript
export interface ChartInterpretation {
  name: string;
  planet_signs: string[];
  planet_houses: string[];
  aspects: string[];
  patterns: string[];
  stars?: string[];
  midpoints?: string[];
  declinations?: string[];
  contraparallels?: string[];
  element_balance?: Record<string, number>;
  modality_balance?: Record<string, number>;
  hemisphere?: HemisphereEmphasis;
  rulership_chains?: Record<number, string[]>;
  dispositor_trees?: Record<string, string[]>;
  is_day: boolean;
  lunar_phase?: string;
  lunar_phase_angle?: number;
  retrogrades?: string[];
  antiscia?: string[];
  antiscia_contacts?: string[];
  mutual_receptions?: string[];
  decans?: string[];
  terms?: string[];
  voc_moon?: string;
  sect?: string;
  chart_ruler?: string;
  chart_ruler_traditional?: string;
  chart_ruler_house?: number;
  chart_ruler_sign?: string;
  chart_ruler_dignity?: string;
  final_dispositor?: string;
  final_dispositor_traditional?: string;
  weighted_aspects?: WeightedAspect[];
  key_midpoints?: string[];
  key_star_aspects?: string[];
  angular_planets?: string[];
}

export interface HemisphereEmphasis {
  upper: number;
  lower: number;
  east: number;
  west: number;
  description: string;
}

export interface WeightedAspect {
  planet1: string;
  planet2: string;
  aspect: string;
  orb: number;
  weight: number;
}
```

**Step 2: Update api.ts interpretation method**

Ensure `api.interpretation()` returns `ChartInterpretation`:

```typescript
interpretation: (data: BirthData, system = 'western', orb = 3) =>
  post<ChartInterpretation>('/api/interpretation', { ...bd(data), system, orb }),
```

**Step 3: Verify**

- Run `npx tsc --noEmit` — should be clean

---

### Task 3: Add element/modality bar charts

**Objective:** Visualize element and modality balance as horizontal bar charts.

**Files:**
- Create: `web/src/components/interpretation/ElementBar.tsx`
- Create: `web/src/components/interpretation/ModalityBar.tsx`

**Step 1: Create ElementBar**

```tsx
// web/src/components/interpretation/ElementBar.tsx
const ELEMENT_COLORS: Record<string, string> = {
  Fire: '#f85149',
  Earth: '#3fb950',
  Air: '#f0c040',
  Water: '#58a6ff',
};

export function ElementBar({ balance }: { balance: Record<string, number> }) {
  const total = Object.values(balance).reduce((a, b) => a + b, 0) || 1;
  return (
    <div className="space-y-2">
      {Object.entries(balance).map(([element, count]) => (
        <div key={element} className="flex items-center gap-2">
          <span className="text-xs w-12 text-muted">{element}</span>
          <div className="flex-1 h-4 bg-surface rounded overflow-hidden">
            <div
              className="h-full rounded transition-all"
              style={{
                width: `${(count / total) * 100}%`,
                backgroundColor: ELEMENT_COLORS[element] || '#888',
              }}
            />
          </div>
          <span className="text-xs text-muted w-6 text-right">{count}</span>
        </div>
      ))}
    </div>
  );
}
```

**Step 2: Create ModalityBar (same pattern, different colors)**

**Step 3: Verify**

- Run `npx tsc --noEmit` — should be clean

---

### Task 4: Add transit interpretation view

**Objective:** Display transit-to-natal aspects and house overlays in a readable format.

**Files:**
- Create: `web/src/components/interpretation/TransitInterpretation.tsx`
- Modify: `web/src/App.tsx` (TransitsView)

**Step 1: Create TransitInterpretation component**

The backend already has `TransitReport` with aspects and house overlays. The `/api/transits` endpoint returns `TransitResponse` which includes `TransitHit[]`. Build a component that:

1. Fetches transit data for a date range
2. Groups transit aspects by transiting planet
3. Shows house overlays (transiting planet in natal house)
4. Renders in collapsible sections

**Step 2: Wire into TransitsView**

Add a sub-tab "Interpretation" alongside "Table" and "Animation" in the TransitsView.

**Step 3: Verify**

- Run `npx tsc --noEmit` — should be clean
- Run `npm run build` — should succeed

---

### Task 5: Add interpretation text management UI

**Objective:** Allow users to view and customize interpretation template texts stored in IndexedDB.

**Files:**
- Create: `web/src/components/interpretation/TextManager.tsx`
- Modify: `web/src/App.tsx` (add settings/management tab)

**Step 1: Create TextManager component**

A simple two-panel view:
- Left: list of interpretation categories (Planet in Sign, Planet in House, Aspects, Patterns)
- Right: editable text area for the selected template
- Save to IndexedDB, load on mount
- "Reset to defaults" button

**Step 2: Add to App.tsx**

Add a "Manage Texts" button in the interpretation panel header, or a settings tab.

**Step 3: Verify**

- Run `npx tsc --noEmit` — should be clean
- Run `npm run build` — should succeed

---

### Task 6: Add system toggle (Western / Koiné)

**Objective:** Let users switch between Western and Koiné interpretation systems.

**Files:**
- Modify: `web/src/components/interpretation/InterpretationPanel.tsx`

**Step 1: Add system selector**

Add a toggle/dropdown at the top of the InterpretationPanel to switch between 'western' and 'koiné'. When changed, re-fetch interpretation data.

**Step 2: Verify**

- Run `npx tsc --noEmit` — should be clean
- Run `npm run build` — should succeed

---

### Task 7: End-to-end verification

**Objective:** Verify the full interpretation flow works end-to-end.

**Steps:**
1. Rebuild Go binary: `go build -buildvcs=false -o /tmp/empirical ./cmd/recover/`
2. Start server: `/tmp/empirical serve 5000`
3. Build frontend: `cd web && npm run build`
4. Open browser, create a chart, click "Interpretation" tab
5. Verify all sections render with data
6. Toggle between Western and Koiné
7. Verify transit interpretation shows aspect data
8. Run all tests: `cd web && npx vitest run`

---

## Principles

- **DRY:** Reuse existing `ChartInterpretation` types and API — don't create new endpoints
- **YAGNI:** Don't build an interpretation editor that supports custom templates yet — just display what the backend produces
- **TDD:** Each component gets a basic render test
- **Frequent commits:** Commit after each task
