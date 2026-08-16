# Phase 5: Professional Features — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Transform Empirical from a feature-complete demo into a daily-driver professional astrology tool with chart management, export/report generation, customization, and workspace polish.

**Architecture:** Phase 5 is primarily frontend work — the backend already has all computation endpoints. New features layer on top of existing IndexedDB (chartDB), localStorage (preferences, custom texts), and the `/api/report` endpoint. PDF generation uses client-side print-to-PDF via `@media print` CSS. No new backend endpoints needed.

**Tech Stack:** React, TypeScript, IndexedDB (idb), localStorage, CSS `@media print`, existing `api.report()` endpoint

---

## Survey: What Already Exists

| Feature | Status | Notes |
|---------|--------|-------|
| Reports tab (Natal, Transit, Page Designer) | ✅ Built | HTML export via PageDesigner, iframe preview for Natal/Transit |
| TextManager (customize interpretation texts) | ✅ Built | localStorage persistence, per-system, edit/revert |
| IndexedDB (chartDB) | ✅ Built | CRUD, search, tags index, exportAll/importAll (JSON) |
| Sidebar with search | ✅ Built | Text search, chart list, delete |
| Export SVG/PNG buttons | ✅ Built | On chart wheels |
| Custom aspect sets | ❌ Missing | No UI for customizing orbs or aspect types |
| PDF export | ❌ Missing | No PDF generation anywhere |
| Print layout | ❌ Missing | No `@media print` CSS, no print button |
| Tag management UI | ❌ Missing | Tags exist in DB schema but no UI to add/edit/remove |
| Bulk chart operations | ❌ Missing | No duplicate, merge, or multi-select delete |
| Backup/restore UI | ❌ Missing | exportAll/importAll exist in DB but no UI buttons |
| Settings/preferences | ❌ Missing | No default house system, ayanamsa, or orb preference |
| Keyboard shortcuts | ❌ Missing | None implemented |
| Synastry/Composite reports | ❌ Missing | Data exists but no report generation for relationships |

---

## Tasks

### Task 5.1: Settings Panel with Persistent Preferences

**Objective:** Add a Settings modal/popover that lets users set defaults for house system, ayanamsa, default orb, and theme — persisted to localStorage and applied globally.

**Files:**
- Create: `web/src/components/settings/SettingsPanel.tsx`
- Modify: `web/src/App.tsx` — add settings button + state
- Modify: `web/src/lib/types.ts` — add `UserPreferences` interface

**Step 1: Add `UserPreferences` type**

```typescript
// web/src/lib/types.ts — add after existing interfaces
export interface UserPreferences {
  defaultHouseSystem: string;    // 'placidus'
  defaultAyanamsa: string;       // 'tropical'
  defaultOrb: number;            // 3
  theme: 'light' | 'dark';       // 'dark'
}
```

**Step 2: Create `SettingsPanel.tsx`**

A modal with form fields for each preference. Load from localStorage on mount, save on change. Dispatch a custom event `'preferences-changed'` so other components can react.

**Step 3: Wire into App.tsx**

Add a gear icon button in the top bar. Pass preferences down via context or props to ChartForm (for default house system/ayanamsa), SynastryView (for default orb), and other components.

**Step 4: Apply preferences**

- ChartForm: pre-select default house system and ayanamsa
- SynastryView: use default orb for aspect grids
- ThemeSwitcher: sync with preferences

**Verification:** Open Settings, change default house system to "Whole Sign", create a new chart — form should pre-select "Whole Sign". Change default orb to 5, open Synastry — aspects should use 5° orb.

---

### Task 5.2: Tag Management UI

**Objective:** Add tag editing to the sidebar — inline add/remove tags on each chart entry, with a tag filter dropdown.

**Files:**
- Modify: `web/src/components/layout/Sidebar.tsx`
- Modify: `web/src/lib/db.ts` — add `updateTags()` method

**Step 1: Add `chartDB.updateTags()`**

```typescript
async updateTags(id: number, tags: string[]): Promise<void> {
  const db = await getDB();
  const chart = await db.get('charts', id);
  if (!chart) throw new Error(`Chart ${id} not found`);
  chart.tags = tags;
  chart.updatedAt = new Date().toISOString();
  await db.put('charts', chart);
}
```

**Step 2: Add tag UI to Sidebar**

Each chart entry shows existing tags as small badges. Click a badge to remove it. An "+" button opens an inline input to add a new tag. A tag filter dropdown at the top filters the chart list by tag.

**Step 3: Add tag suggestions**

When typing a new tag, show autocomplete suggestions from all existing tags across all charts (gathered from `chartDB.getAll()`).

**Verification:** Add tags "client" and "family" to a chart. Filter by "family" — only that chart shows. Remove "family" tag — filter clears. Tags persist across page reloads.

---

### Task 5.3: Bulk Chart Operations

**Objective:** Add multi-select mode to sidebar with bulk delete, bulk tag, and duplicate operations.

**Files:**
- Modify: `web/src/components/layout/Sidebar.tsx`
- Modify: `web/src/lib/db.ts` — add `duplicate()`, `bulkDelete()`, `bulkTag()`

**Step 1: Add DB methods**

```typescript
async duplicate(id: number): Promise<number> {
  const chart = await this.getById(id);
  if (!chart) throw new Error(`Chart ${id} not found`);
  const { id: _, ...rest } = chart;
  return this.add({
    ...rest,
    name: `${rest.name} (copy)`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  });
}

async bulkDelete(ids: number[]): Promise<void> {
  const db = await getDB();
  const tx = db.transaction('charts', 'readwrite');
  for (const id of ids) await tx.store.delete(id);
  await tx.done;
}

async bulkTag(ids: number[], tag: string): Promise<void> {
  const db = await getDB();
  const tx = db.transaction('charts', 'readwrite');
  for (const id of ids) {
    const chart = await tx.store.get(id);
    if (chart && !chart.tags.includes(tag)) {
      chart.tags.push(tag);
      chart.updatedAt = new Date().toISOString();
      await tx.store.put(chart);
    }
  }
  await tx.done;
}
```

**Step 2: Add multi-select mode to Sidebar**

A "Select" button toggles multi-select mode. Each chart entry gets a checkbox. Selected count shown in a toolbar with "Delete Selected", "Tag Selected", "Cancel" buttons. Each chart entry also gets a "Duplicate" icon button (always visible, not just in select mode).

**Verification:** Select 3 charts, add tag "backup", verify all 3 have the tag. Duplicate a chart — copy appears with "(copy)" suffix. Delete 2 charts — they're gone from the list and DB.

---

### Task 5.4: Backup & Restore UI

**Objective:** Add visible Import/Export buttons to the sidebar for JSON backup/restore of the entire chart database.

**Files:**
- Modify: `web/src/components/layout/Sidebar.tsx`

**Step 1: Add Export button**

Button in sidebar header: "Export All". Calls `chartDB.exportAll()`, creates a Blob, triggers download as `empirical-charts-YYYY-MM-DD.json`.

**Step 2: Add Import button**

Button in sidebar header: "Import". Opens a file picker (`<input type="file" accept=".json">`). On file select, reads the file, calls `chartDB.importAll(json)`, shows a toast with count of imported charts, refreshes the chart list.

**Step 3: Add confirmation for import**

Before importing, show a confirmation dialog: "This will add N charts to your database. Existing charts will not be affected. Continue?" with count from parsing the JSON.

**Verification:** Export all charts — downloads a JSON file. Delete all charts. Import the JSON file — all charts restored with new IDs. Import again — charts are duplicated (expected behavior, noted in confirmation).

---

### Task 5.5: Print Layout & PDF Export

**Objective:** Add `@media print` CSS and a "Print" button to chart views, natal dashboard, and reports. PDF is generated via browser print-to-PDF.

**Files:**
- Create: `web/src/lib/print.css`
- Modify: `web/src/index.css` — import print.css
- Modify: `web/src/components/chart/ChartWheel.tsx` — add print button
- Modify: `web/src/components/natal/Dashboard.tsx` — add print button
- Modify: `web/src/App.tsx` — add print button to reports view

**Step 1: Create `print.css`**

```css
@media print {
  /* Hide UI chrome */
  nav, .sidebar, button, .tab-bar, .no-print {
    display: none !important;
  }
  
  /* Full-width content */
  body { margin: 0; padding: 0; }
  .print-area { width: 100%; max-width: none; }
  
  /* Chart wheels at reasonable size */
  .chart-wheel svg { max-width: 400px; max-height: 400px; }
  
  /* Page breaks */
  .page-break { page-break-before: always; }
  
  /* Dark theme → light for print */
  body { background: white !important; color: black !important; }
  .text-muted { color: #666 !important; }
  .bg-surface, .bg-bg { background: white !important; border-color: #ccc !important; }
}
```

**Step 2: Add Print buttons**

Add a printer icon button to:
- Chart wheel view (prints the wheel + planet table)
- Natal dashboard (prints the full dashboard)
- Reports view (prints the current report)
- Synastry view (prints both wheels + aspect grid)

Each button calls `window.print()`.

**Step 3: Add print-specific class names**

Wrap printable content in `<div className="print-area">` and hide UI elements with `className="no-print"`.

**Verification:** Click Print on a chart wheel. Print preview shows the wheel and planet table without sidebar, buttons, or tab bar. Background is white, text is black. Same for natal dashboard and reports.

---

### Task 5.6: Custom Aspect Sets

**Objective:** Add a UI to customize which aspects are used and their orbs. Persisted to localStorage, applied to all aspect grids and API calls.

**Files:**
- Create: `web/src/components/settings/AspectSetEditor.tsx`
- Modify: `web/src/components/settings/SettingsPanel.tsx` — add Aspect Sets tab
- Modify: `web/src/lib/types.ts` — add `AspectSet` interface

**Step 1: Add types**

```typescript
export interface AspectDef {
  name: string;           // 'conjunction'
  angle: number;          // 0
  orb: number;            // 8
  enabled: boolean;       // true
  glyph: string;          // '☌'
}

export interface AspectSet {
  name: string;           // 'Modern (tight orbs)'
  aspects: AspectDef[];
}
```

**Step 2: Create `AspectSetEditor.tsx`**

A panel showing all aspect types with:
- Checkbox to enable/disable
- Number input for orb
- Preset dropdown: "Modern (tight)", "Traditional (wide)", "Custom"
- Save/load from localStorage key `empirical-aspect-set`

Default preset (matching current backend defaults):
```
Conjunction 0°  orb 8°
Opposition  180° orb 8°
Trine       120° orb 8°
Square      90°  orb 7°
Sextile     60°  orb 6°
Quincunx    150° orb 3°
```

**Step 3: Pass aspect set to API calls**

Modify `api.synastry()`, `api.composite()`, `api.draconicSynastryFull()` to accept an optional `aspectSet: AspectSet` parameter. When provided, pass custom orbs to the backend. (Backend already accepts `orb` parameter — for now, use the widest enabled orb as the `orb` parameter and filter client-side.)

**Step 4: Apply to AspectGrid**

Modify `AspectGrid` (both natal and synastry) to accept an optional `aspectSet` prop. When provided, only show aspects whose types are enabled and whose orbs are within the custom range.

**Verification:** Open Aspect Set Editor, disable quincunx, set conjunction orb to 3°. Open a synastry — quincunx aspects are gone, only conjunctions within 3° show. Reset to defaults — all aspects return.

---

### Task 5.7: Keyboard Shortcuts

**Objective:** Add global keyboard shortcuts for common operations.

**Files:**
- Create: `web/src/lib/shortcuts.ts`
- Modify: `web/src/main.tsx` — register global listener

**Step 1: Create `shortcuts.ts`**

```typescript
const shortcuts: Record<string, { key: string; ctrl?: boolean; description: string; action: () => void }> = {};

export function registerShortcut(id: string, key: string, ctrl: boolean, description: string, action: () => void) {
  shortcuts[id] = { key, ctrl, description, action };
}

export function initShortcuts() {
  window.addEventListener('keydown', (e) => {
    // Don't fire when typing in inputs
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
    
    for (const [id, s] of Object.entries(shortcuts)) {
      const ctrlMatch = s.ctrl ? (e.ctrlKey || e.metaKey) : !e.ctrlKey && !e.metaKey;
      if (e.key.toLowerCase() === s.key.toLowerCase() && ctrlMatch) {
        e.preventDefault();
        s.action();
        return;
      }
    }
  });
}
```

**Step 2: Register shortcuts in App.tsx**

```
Ctrl+N      — New chart
Ctrl+S      — Save current chart
Ctrl+P      — Print
Ctrl+F      — Focus search
1-9         — Switch to tab 1-9
Escape      — Close modal / cancel
?           — Show shortcut help overlay
```

**Step 3: Add shortcut help overlay**

A "?" button in the top bar opens a modal listing all shortcuts. Also shown on first visit (localStorage flag).

**Verification:** Press Ctrl+N — new chart form opens. Press 3 — switches to Tri-Wheel tab. Press ? — shortcut help overlay appears. Type in a text input — shortcuts don't fire.

---

### Task 5.8: Synastry & Composite Reports

**Objective:** Add report generation for relationship charts — printable HTML reports for synastry, composite, and draconic synastry.

**Files:**
- Create: `web/src/components/reports/SynastryReport.tsx`
- Modify: `web/src/App.tsx` — add "Synastry Report" option to Reports tab
- Modify: `web/src/components/synastry/SynastryView.tsx` — add "Generate Report" button

**Step 1: Create `SynastryReport.tsx`**

Similar to `NatalReport` — fetches synastry + composite + draconic data, renders an HTML report with:
- Header: both names, dates, locations
- Section 1: Aspect summary (strongest connections by orb)
- Section 2: House overlays
- Section 3: Composite chart planets + patterns
- Section 4: Draconic bridges
- Print button

Uses the existing `api.synastry()`, `api.composite()`, `api.draconicSynastryFull()` endpoints.

**Step 2: Wire into Reports tab**

Add "Synastry Report" as a 4th report type in the Reports view. When selected, prompt user to select a partner chart from a dropdown (or use the currently selected partner from SynastryView).

**Step 3: Add "Generate Report" button to SynastryView**

A button in the SynastryView header: "Generate Report" — switches to Reports tab with Synastry Report pre-selected and partner pre-filled.

**Verification:** Open Synastry with AJ + Cait. Click "Generate Report". Report shows both names, strongest aspects, house overlays, composite patterns, draconic bridges. Click Print — clean print layout.

---

### Task 5.9: End-to-End Verification

**Objective:** Verify all Phase 5 features work together with no regressions.

**Steps:**

1. **Settings persistence:** Change default house system to Whole Sign, close tab, reopen — Whole Sign is pre-selected in new chart form.
2. **Tag workflow:** Add tags to 3 charts, filter by tag, remove a tag, verify filter updates.
3. **Bulk operations:** Select 2 charts, add tag "test", verify both have it. Duplicate one, verify copy exists. Delete both, verify removed.
4. **Backup/restore:** Export all, delete all, import, verify all charts restored.
5. **Print:** Print chart wheel, natal dashboard, synastry report — all produce clean print output.
6. **Custom aspects:** Disable quincunx, set tight orbs, verify aspect grids reflect changes.
7. **Keyboard shortcuts:** Test all registered shortcuts.
8. **Synastry report:** Generate report for AJ + Cait, verify all sections populated.
9. **Regression:** Run `npx tsc --noEmit`, `npm run build`, `npx vitest run`, `go build` — all must pass.
10. **Browser test:** Load the app, click through all tabs, verify no console errors.

---

## Summary

| # | Task | Est. Session |
|---|------|-------------|
| 5.1 | Settings Panel | 1 session |
| 5.2 | Tag Management UI | 1 session |
| 5.3 | Bulk Chart Operations | 1 session |
| 5.4 | Backup & Restore UI | 0.5 session |
| 5.5 | Print Layout & PDF Export | 1 session |
| 5.6 | Custom Aspect Sets | 1.5 sessions |
| 5.7 | Keyboard Shortcuts | 0.5 session |
| 5.8 | Synastry & Composite Reports | 1 session |
| 5.9 | End-to-End Verification | 0.5 session |
| **Total** | | **~8 sessions** |

## Risks

- **Custom aspect sets** (5.6) is the riskiest task — it touches aspect grid rendering, API calls, and localStorage. The backend doesn't support per-aspect-type orbs, so client-side filtering is needed. This is acceptable for v1.
- **Print layout** (5.5) requires testing across browsers. Chrome and Safari handle `@media print` differently. Target Chrome first (the primary browser for this app).
- **Bulk operations** (5.3) need careful IndexedDB transaction handling to avoid partial failures.
