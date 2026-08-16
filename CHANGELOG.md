# Changelog

## v1.0.0 (2026-07-30)

Initial release. Cross-system empirical astrology engine.

### Core
- Swiss Ephemeris integration (JPL DE ephemeris data)
- Tropical and sidereal zodiac (Lahiri, Fagan-Bradley, Raman, Krishnamurti)
- 8 house systems (Placidus, Whole Sign, Equal, Koch, Porphyry, Regiomontanus, Alcabitius, Campanus)
- Classical planets + Uranus, Neptune, Pluto
- Asteroids: Chiron, Ceres, Pallas, Juno, Vesta
- Dwarf planets: Eris, Makemake, Gonggong
- Points: Lilith, North Node, Part of Fortune
- Trans-Neptunian Points (8 TNPs)
- 201+ aspects across all bodies

### Interpretation
- Western (tropical) interpretation engine
- Vedic (Jyotish) interpretation with nakshatras and shadbala
- Koiné (Hellenistic) synthesis system
- Ba Zi (Four Pillars) Chinese astrology
- Planet-in-sign, planet-in-house, aspect interpretations
- Pattern detection (T-square, Grand Trine, Yod, Cradle, Stellium, Grand Cross, Kite, Mystic Rectangle)
- Dignity scoring and dispositor trees
- Element and modality balance
- Midpoints and star aspects
- Declination parallels and antiscia

### Predictive Tools
- Solar returns
- Secondary progressions
- Primary directions
- Annual profections
- Firdaria periods
- Zodiacal releasing
- Transit analysis with interpretation
- Timing convergence (multi-technique overlay)

### Relationship Astrology
- Synastry (201+ aspects, house overlays)
- Composite (midpoint) charts
- Draconic synastry (three-layer comparison with bridges)
- Relationship synthesis

### Research
- Recover protocol (cross-system convergence report)
- Ephemeris table
- Transit calendar
- Mundane astrology (67 national charts)
- Astrocartography (planetary line maps)
- Chart database (15 verified public figure charts)

### Web GUI
- Interactive chart wheel (SVG, bi-wheel, tri-wheel)
- Full interpretation panel with system toggle
- Predictive tools tab (7 sub-views)
- Relationship astrology tab (4 sub-views)
- Astrocartography map (Leaflet)
- Printable reports (natal, transit, synastry)
- Page designer (custom report builder)
- Dark/light theme
- Keyboard shortcuts
- Tag management with filter
- Bulk operations (select, delete, tag)
- Backup/restore (JSON export/import)
- Custom aspect sets with per-aspect orbs
- Settings panel (house system, ayanamsa, orb, theme)
- PWA support (offline-capable, installable)

### CLI
- `recover` — cross-system convergence report
- `transit` — transit analysis
- `serve` — HTTP server + web UI
- `solar-return`, `progressions`, `directions`, `profections`, `firdaria`, `zodiacal-releasing`
- `synastry`, `composite`, `draconic`
- `ephemeris`, `astrocartography`, `mundane`
- `chartdb` — embedded chart database (list, show, recover, search)
- `interpretation` — full natal interpretation
- `lunar-mansion` — lunar mansion research
- `vedic` — Vedic natal report
- `--json` flag on all subcommands

### Distribution
- Single Go binary (zero runtime dependencies)
- GitHub Actions CI (Go + web build, test, lint)
- GoReleaser for tagged releases
- Homebrew formula
- PWA (installable web app)
- Documentation site (getting started, API ref, CLI ref, web GUI, astrology systems)
- Community infrastructure (CONTRIBUTING.md, issue templates, PR template, code of conduct)
- Plugin system foundation (Interpreter interface, registry, .so loader)
