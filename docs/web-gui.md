# Web GUI

The Empirical web interface runs at `http://localhost:5000` after starting the server with `empirical serve 5000`.

## Layout

The interface has three main areas:

- **Sidebar** (left) — chart list, search, tag management, import/export
- **Tab bar** (top) — navigation between views
- **Main content** (center) — the active view

## Views

### Wheel

The natal chart wheel. Shows planets, house cusps, aspects, and sign glyphs. Controls at the top let you switch ayanamsa and house system on the fly. Export buttons save the chart as SVG or PNG.

### Bi-Wheel

Two concentric chart wheels — typically natal (inner) + transits (outer). Aspect lines connect planets between the two charts. Useful for transit analysis.

### Tri-Wheel

Three concentric wheels — natal + secondary + tertiary. Used for multi-layer transit analysis.

### Natal

Tabular view of all planet positions, signs, houses, and retrograde status. Includes asteroids (Chiron, Ceres, Pallas, Juno, Vesta), dwarf planets (Eris, Makemake, Gonggong), Lilith, and the lunar nodes.

### Transits

Transit analysis for a date range. Shows which transiting planets are aspecting your natal planets, with orb information and interpretation.

### Synastry

Relationship astrology. Enter a partner's birth data to see:
- **Aspects** — 201+ aspects between all bodies in both charts
- **Composite** — midpoint chart with 31 midpoints and 61+ aspects
- **Draconic** — three-layer draconic comparison with bridges
- **Synthesis** — relationship summary with house overlays
- **Report** — printable HTML report

### Maps

Astrocartography — planetary line maps showing where each planet's lines cross the globe. Uses Leaflet for interactive zoom and pan.

### Reports

Printable reports:
- **Natal Report** — full natal chart interpretation
- **Transit Report** — transit interpretation for a date range
- **Page Designer** — custom report builder with drag-and-drop blocks

### Ephemeris

Planetary positions table for a date range. Shows daily positions for all planets.

### Calendar

Transit calendar showing significant transits for a month.

### Research

Mundane astrology — transit analysis for national charts (67 countries).

### Interpretation

Full natal interpretation with:
- Planet-in-sign and planet-in-house descriptions
- Aspect interpretations
- Pattern detection (T-square, Grand Trine, Yod, Cradle, Stellium)
- Dignity scores and dispositor tree
- Element and modality balance
- Midpoints and star aspects
- Western/Koiné system toggle

### Predictive

Predictive techniques:
- **Solar Return** — chart for the moment the Sun returns to its natal position
- **Progressions** — secondary progressions (day-for-a-year)
- **Directions** — primary directions
- **Profections** — annual profections
- **Firdaria** — planetary periods
- **Zodiacal Releasing** — from the Lot of Fortune
- **Timing Convergence** — overlay of multiple predictive techniques

## Chart Management

### Saving Charts

Charts are saved to IndexedDB in your browser. Click **+ New Chart** in the sidebar, fill in birth data, and the chart is saved automatically.

### Tags

Add tags to charts for organization (e.g., "client", "family", "research"). Filter the sidebar by tag using the dropdown.

### Bulk Operations

Click **Select** in the sidebar to enter selection mode. Check multiple charts, then delete or tag them in bulk.

### Backup & Restore

Click **Export** to download all charts as a JSON file. Click **Import** to restore from a backup file or paste JSON.

## Settings

Click the ⚙ button in the tab bar to open settings:

- **General** — default house system, ayanamsa, orb, theme (dark/light)
- **Aspects** — custom aspect sets with per-aspect orb controls

Settings persist in localStorage.

## Keyboard Shortcuts

Press `?` to show the keyboard shortcut overlay.

| Shortcut | Action |
|----------|--------|
| `Ctrl+N` | New chart |
| `Ctrl+S` | Save chart |
| `Ctrl+P` | Print |
| `Ctrl+,` | Settings |
| `1`–`9` | Switch to tab 1–9 |
| `Escape` | Close modal |
| `?` | Show shortcuts |
