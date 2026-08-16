# Phase 6: Ecosystem & Distribution — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Transform Empirical from a local-only development project into a shippable open-source product with CI/CD, distribution packaging, PWA support, documentation, and community infrastructure.

**Architecture:** Phase 6 is infrastructure and polish — no new astrological computation. The Go binary already embeds the web frontend and ephemeris files. We add GitHub Actions CI/CD, Homebrew distribution, PWA manifest + service worker, a documentation site, community contribution templates, and the foundation for a plugin system. All tasks are independent except 6.6 depends on 6.1 (CI must pass before plugin PRs land).

**Tech Stack:** GitHub Actions, GoReleaser, Homebrew formula, Vite PWA plugin, service worker, Markdown/MDX docs, Go plugin interface

---

## Survey: What Already Exists

| Feature | Status | Notes |
|---------|--------|-------|
| Single Go binary | ✅ | Embeds web/dist + ephemeris files via `//go:embed` |
| MSW API handlers | ✅ | `web/src/test/mocks/handlers.ts` — 204 lines, covers chart/bi-wheel/tri-wheel/interpretation/transits/recover |
| Playwright | ✅ | Installed in devDependencies, `test:e2e` and `test:visual` scripts exist |
| README.md | ✅ | 72 lines, recover-protocol focused, needs expansion |
| MANUAL.md | ✅ | 548 lines, comprehensive API/CLI reference |
| DESIGN.md | ✅ | 102 lines, Koiné design philosophy |
| Go module | ✅ | `github.com/aj-nt/empirical`, Go 1.25 |
| Web build | ✅ | Vite + React + Tailwind, `npm run build` → `web/dist/` |
| Template rendering | ✅ | Go `html/template` for natal reports |
| GitHub repo | ✅ | `github.com/aj-nt/empirical` |
| Playwright visual config | ❌ | Script exists in package.json but config file missing |
| `.github/` directory | ❌ | No CI, no issue templates, no PR templates |
| Makefile | ❌ | No build automation |
| GoReleaser config | ❌ | No release automation |
| Homebrew formula | ❌ | No distribution |
| PWA manifest | ❌ | No offline support |
| Service worker | ❌ | No caching |
| Documentation site | ❌ | Only README + MANUAL.md |
| CONTRIBUTING.md | ❌ | No contributor guide |
| Plugin system | ❌ | No extension mechanism |
| Chart database | ❌ | No community chart repo |

---

## Task 6.1: GitHub CI/CD Pipeline

**Objective:** Add GitHub Actions workflows for CI (build, test, lint) and release automation.

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yml`

**Step 1: Create CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go build -buildvcs=false ./...
      - run: go test ./...
      - run: go vet ./...

  web:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: web
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - run: npm ci
      - run: npx tsc --noEmit
      - run: npx vitest run
      - run: npm run build
```

**Step 2: Create release workflow with GoReleaser**

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: ['v*']

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - run: cd web && npm ci && npm run build
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 3: Create GoReleaser config**

```yaml
# .goreleaser.yml
builds:
  - main: ./cmd/recover/
    binary: empirical
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]
    env:
      - CGO_ENABLED=1
    flags:
      - -buildvcs=false

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: empirical
    repository:
      owner: aj-nt
      name: homebrew-tap
    homepage: "https://github.com/aj-nt/empirical"
    description: "Cross-system empirical astrology engine"
    license: "MIT"
```

**Verification:** Push a tag `v0.1.0`, CI goes green, release artifacts appear on GitHub Releases page.

---

## Task 6.2: PWA Support

**Objective:** Make the web UI installable as a Progressive Web App with offline support.

**Files:**
- Create: `web/public/manifest.json`
- Create: `web/public/sw.js`
- Modify: `web/index.html` (add manifest link + service worker registration)
- Modify: `web/vite.config.ts` (add PWA plugin or manual config)

**Step 1: Create web app manifest**

```json
{
  "name": "Empirical Astrology",
  "short_name": "Empirical",
  "description": "Cross-system empirical astrology engine",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0a0a0f",
  "theme_color": "#0a0a0f",
  "icons": [
    {
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

**Step 2: Create service worker**

```javascript
// web/public/sw.js
const CACHE = 'empirical-v1';
const ASSETS = [
  '/',
  '/index.html',
  // Vite injects hashed asset paths at build time
];

self.addEventListener('install', (e) => {
  e.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(ASSETS))
  );
});

self.addEventListener('fetch', (e) => {
  // Network-first for API calls, cache-first for static assets
  if (e.request.url.includes('/api/')) {
    e.respondWith(
      fetch(e.request).catch(() => caches.match(e.request))
    );
  } else {
    e.respondWith(
      caches.match(e.request).then((r) => r || fetch(e.request))
    );
  }
});
```

**Step 3: Wire into index.html**

Add to `<head>`:
```html
<link rel="manifest" href="/manifest.json">
```

Add before `</body>`:
```html
<script>
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
  }
</script>
```

**Step 4: Generate app icons**

Use a simple SVG-based icon generator or create a minimal 192x192 and 512x512 PNG. The icon should be the Empirical "E" or a simple astrological symbol.

**Verification:** Build, serve, open Chrome DevTools → Application → Manifest (no errors), Service Worker (registered). Lighthouse PWA audit passes.

---

## Task 6.3: Documentation Site

**Objective:** Create a proper documentation site with getting started, API reference, and examples.

**Files:**
- Create: `docs/` directory structure
- Create: `docs/index.md` (landing page)
- Create: `docs/getting-started.md`
- Create: `docs/api-reference.md`
- Create: `docs/cli-reference.md`
- Create: `docs/web-gui.md`
- Create: `docs/astrology-systems.md`
- Modify: `README.md` (expand, link to docs)

**Step 1: Create docs directory structure**

```
docs/
  index.md              # Landing page with quick start
  getting-started.md    # Installation, first chart, basic usage
  api-reference.md      # All API endpoints with curl examples
  cli-reference.md      # All CLI subcommands
  web-gui.md            # Web interface walkthrough
  astrology-systems.md  # Western, Vedic, Koiné, BaZi — what they are
```

**Step 2: Write docs/index.md**

Landing page with:
- One-paragraph description
- Quick start (3 commands: build, serve, open browser)
- Screenshot of the web UI
- Links to each doc section

**Step 3: Write docs/getting-started.md**

- Prerequisites (Go 1.25+, Node 22+)
- Clone + build
- First chart
- Web UI tour
- CLI examples

**Step 4: Write docs/api-reference.md**

- All endpoints with method, path, request body, response shape, curl example
- Pull from existing MANUAL.md but restructure for readability

**Step 5: Write docs/cli-reference.md**

- All subcommands with flags and examples
- Pull from existing MANUAL.md

**Step 6: Write docs/web-gui.md**

- Screenshots of each tab
- How to save charts, use tags, export
- Settings walkthrough

**Step 7: Write docs/astrology-systems.md**

- Brief explanation of each system
- What the recover protocol measures
- Link to the paper

**Step 8: Expand README.md**

Add sections:
- Badges (CI, license, Go version)
- Quick start (3 commands)
- Screenshot
- Links to full docs
- Community section

**Verification:** All docs render correctly on GitHub. Links work. README badges show green.

---

## Task 6.4: Community Infrastructure

**Objective:** Add contribution guide, issue templates, PR template, and code of conduct.

**Files:**
- Create: `CONTRIBUTING.md`
- Create: `.github/ISSUE_TEMPLATE/bug_report.md`
- Create: `.github/ISSUE_TEMPLATE/feature_request.md`
- Create: `.github/ISSUE_TEMPLATE/chart_data.md`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`
- Create: `CODE_OF_CONDUCT.md`

**Step 1: Create CONTRIBUTING.md**

Sections:
- Development setup (clone, build, test)
- Project structure overview
- How to run tests
- How to add a new interpretation
- How to add a new API endpoint
- PR checklist

**Step 2: Create issue templates**

`bug_report.md`:
```yaml
name: Bug Report
description: Something is broken
labels: [bug]
body:
  - type: textarea
    attributes:
      label: What happened?
  - type: textarea
    attributes:
      label: Steps to reproduce
  - type: input
    attributes:
      label: Version (git sha or release)
  - type: textarea
    attributes:
      label: Expected behavior
```

`feature_request.md`:
```yaml
name: Feature Request
description: Suggest an improvement
labels: [enhancement]
body:
  - type: textarea
    attributes:
      label: What would you like?
  - type: textarea
    attributes:
      label: Why is this valuable?
```

`chart_data.md`:
```yaml
name: Chart Data Submission
description: Submit verified birth data for a public figure
labels: [chart-data]
body:
  - type: input
    attributes:
      label: Name
  - type: input
    attributes:
      label: Birth date (YYYY-MM-DD)
  - type: input
    attributes:
      label: Birth time (HH:MM, 24h)
  - type: input
    attributes:
      label: Birth location (city, country)
  - type: input
    attributes:
      label: Source (Rodden rating, birth certificate, etc.)
```

**Step 3: Create PR template**

```markdown
## What
[Brief description]

## Testing
- [ ] `go test ./...` passes
- [ ] `cd web && npx tsc --noEmit` passes
- [ ] `cd web && npx vitest run` passes
- [ ] `cd web && npm run build` succeeds
- [ ] Manual verification: [describe what you tested]

## Screenshots
[If UI change, add before/after]
```

**Step 4: Create CODE_OF_CONDUCT.md**

Use Contributor Covenant v2.1 (standard for open source).

**Verification:** Create a test issue using each template. Create a test PR using the template. All render correctly.

---

## Task 6.5: Plugin System Foundation

**Objective:** Define a Go interface for interpretation plugins and build the loading infrastructure. This is the foundation — actual community plugins come later.

**Files:**
- Create: `plugin/plugin.go` (interface definition)
- Create: `plugin/registry.go` (plugin registry)
- Create: `plugin/loader.go` (filesystem plugin loader)
- Create: `plugin/plugin_test.go`
- Modify: `internal/dignity/` (wire plugin registry into interpretation pipeline)

**Step 1: Define plugin interface**

```go
// plugin/plugin.go
package plugin

import "github.com/aj-nt/empirical/internal/dignity"

// Interpreter is the interface for interpretation plugins.
// Each plugin provides interpretations for a specific domain
// (e.g., planet-in-sign, aspect, house placement).
type Interpreter interface {
    // Name returns the plugin's unique identifier.
    Name() string
    // Version returns the plugin's semantic version.
    Version() string
    // Description returns a human-readable description.
    Description() string
    // PlanetInSign returns an interpretation for a planet in a sign.
    // Returns empty string if this plugin doesn't cover this combination.
    PlanetInSign(planet, sign string) string
    // AspectInterpretation returns an interpretation for an aspect between two planets.
    AspectInterpretation(p1, p2, aspectType string, orb float64) string
    // HousePlacement returns an interpretation for a planet in a house.
    HousePlacement(planet string, house int) string
}
```

**Step 2: Create plugin registry**

```go
// plugin/registry.go
package plugin

import "sync"

var (
    registry   = make(map[string]Interpreter)
    registryMu sync.RWMutex
)

func Register(p Interpreter) {
    registryMu.Lock()
    defer registryMu.Unlock()
    registry[p.Name()] = p
}

func Get(name string) (Interpreter, bool) {
    registryMu.RLock()
    defer registryMu.RUnlock()
    p, ok := registry[name]
    return p, ok
}

func List() []Interpreter {
    registryMu.RLock()
    defer registryMu.RUnlock()
    result := make([]Interpreter, 0, len(registry))
    for _, p := range registry {
        result = append(result, p)
    }
    return result
}
```

**Step 3: Create filesystem loader**

```go
// plugin/loader.go
package plugin

import (
    "fmt"
    "os"
    "path/filepath"
    "plugin" // Go's built-in plugin package
)

// LoadDir loads all .so plugins from a directory.
// Plugins must export a symbol "Plugin" of type Interpreter.
func LoadDir(dir string) (int, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return 0, nil // No plugins dir is fine
        }
        return 0, err
    }

    loaded := 0
    for _, entry := range entries {
        if entry.IsDir() || filepath.Ext(entry.Name()) != ".so" {
            continue
        }
        path := filepath.Join(dir, entry.Name())
        p, err := plugin.Open(path)
        if err != nil {
            return loaded, fmt.Errorf("loading %s: %w", entry.Name(), err)
        }
        sym, err := p.Lookup("Plugin")
        if err != nil {
            return loaded, fmt.Errorf("symbol 'Plugin' not found in %s", entry.Name())
        }
        interp, ok := sym.(Interpreter)
        if !ok {
            return loaded, fmt.Errorf("symbol 'Plugin' in %s is not an Interpreter", entry.Name())
        }
        Register(interp)
        loaded++
    }
    return loaded, nil
}
```

**Step 4: Wire into interpretation pipeline**

In `internal/dignity/`, add a fallback chain: check plugin registry first, fall back to built-in interpretations. This is a one-line change in the interpretation lookup functions.

**Step 5: Add CLI flag**

Add `--plugin-dir` flag to the server command:
```go
pluginDir := flag.String("plugin-dir", "", "directory containing .so plugin files")
```

**Verification:** Write a test plugin that implements `Interpreter`, compile it as a `.so`, load it via `LoadDir`, verify `Get()` returns it, verify interpretation pipeline uses it.

---

## Task 6.6: Chart Database Foundation

**Objective:** Create the infrastructure for a community-maintained chart database — a JSON file of verified birth data for public figures, with a CLI subcommand to query it.

**Files:**
- Create: `charts/` directory
- Create: `charts/README.md` (submission guide)
- Create: `charts/people.json` (initial dataset — 10-20 verified charts)
- Create: `cmd/recover/chartdb.go` (CLI subcommand)
- Modify: `cmd/recover/main.go` (register subcommand)

**Step 1: Define chart data schema**

```json
{
  "name": "Albert Einstein",
  "birth": {
    "year": 1879,
    "month": 3,
    "day": 14,
    "hour": 11,
    "minute": 30,
    "tz_offset": 1,
    "lat": 48.401,
    "lng": 9.989
  },
  "source": "Birth certificate (Rodden AA)",
  "category": "science",
  "notes": "Ulm, Germany"
}
```

**Step 2: Create initial dataset**

10-20 verified charts with Rodden AA or A ratings:
- Albert Einstein (AA)
- Marie Curie (AA)
- Carl Jung (AA)
- Winston Churchill (A)
- Frida Kahlo (AA)
- Alan Turing (A)
- etc.

**Step 3: Create CLI subcommand**

```go
// cmd/recover/chartdb.go
// empirical chartdb list                    — list all charts
// empirical chartdb show <name>            — show one chart
// empirical chartdb recover <name>          — run recover on one chart
// empirical chartdb search <query>          — search by name/category
```

**Step 4: Embed chart data**

Use `//go:embed charts/*.json` to bundle the chart database into the binary.

**Verification:** `empirical chartdb list` shows all charts. `empirical chartdb show "Einstein"` shows birth data. `empirical chartdb recover "Einstein"` runs the recover protocol.

---

## Task 6.7: Final Polish & Release

**Objective:** Version bump, changelog, final README polish, tag v1.0.0.

**Files:**
- Create: `CHANGELOG.md`
- Modify: `README.md` (final polish)
- Modify: `web/package.json` (version bump)

**Step 1: Write CHANGELOG.md**

Summarize all 6 phases of work. Keep it high-level — one paragraph per phase.

**Step 2: Final README polish**

- Add demo GIF/screenshot
- Add "Why Empirical?" section
- Add comparison table vs TimePassages/Solar Fire
- Add community section with links

**Step 3: Version bump**

Set version to `1.0.0` in:
- `web/package.json`
- Go module (optional, Go doesn't use version in go.mod)

**Step 4: Tag and release**

```bash
git tag v1.0.0
git push origin v1.0.0
```

CI builds and publishes release artifacts via GoReleaser.

**Verification:** GitHub Releases page shows v1.0.0 with binaries for macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64). Homebrew tap has the formula. `brew install aj-nt/tap/empirical` works.

---

## Execution Order

Tasks are mostly independent. Recommended order:

1. **6.1** (CI/CD) — first, so everything else gets tested automatically
2. **6.4** (Community) — parallel with 6.1, no dependencies
3. **6.2** (PWA) — independent
4. **6.3** (Docs) — independent, can start after 6.1
5. **6.5** (Plugin System) — independent, but CI must be green first
6. **6.6** (Chart DB) — independent
7. **6.7** (Release) — last, after everything else

**Total estimated sessions:** 5-7 (some tasks are small and can be batched)

---

## What We're NOT Building (from the master plan)

- **Mobile companion app** — PWA (6.2) covers this. A separate React Native app is scope creep for v1.
- **Chart marketplace with user accounts** — Chart database (6.6) is a JSON file + CLI. A web marketplace with auth is a v2 feature.
- **Multi-window support** — Browser tabs handle this.
- **Video tutorials** — Community will create content.
- **Astro Clock (real-time sky)** — Nice to have, not v1.
