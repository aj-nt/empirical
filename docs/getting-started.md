# Getting Started

## Prerequisites

- **Go 1.25+** — [install](https://go.dev/dl/)
- **Node.js 22+** — [install](https://nodejs.org/) (for web UI development)
- **Swiss Ephemeris data** — downloaded automatically on first run to `~/.cache/empirical/ephe/`

## Installation

### From source

```bash
git clone https://github.com/aj-nt/empirical.git
cd empirical
go build -buildvcs=false -o empirical ./cmd/recover/
```

### From Homebrew (macOS)

```bash
brew install aj-nt/tap/empirical
```

### From GitHub Releases

Download the latest binary for your platform from the [releases page](https://github.com/aj-nt/empirical/releases).

## First Chart

### Web UI

```bash
./empirical serve 5000
```

Open `http://localhost:5000`. Click **+ New Chart**, enter birth data, and the chart wheel renders instantly.

### CLI

```bash
# Natal chart
./empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901

# With JSON output
./empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901 --json

# Sidereal (Lahiri)
./empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901 --sidereal

# Transits for today
./empirical transit AJ 1969 2 15 23 10 -8 47.038 -122.901 2026-07-30
```

## Project Structure

```
empirical/
├── cmd/recover/          # CLI entry point + HTTP server
├── internal/
│   ├── dignity/          # Interpretation engine, templates, transit computation
│   ├── server/           # HTTP API handlers
│   ├── swe/              # Swiss Ephemeris CGo wrapper
│   ├── geocode/          # City geocoding (embedded GeoNames DB)
│   └── mundane/          # Mundane astrology (67 national charts)
├── web/                  # React + TypeScript frontend
│   ├── src/
│   │   ├── components/   # UI components (chart, synastry, reports, settings)
│   │   ├── lib/          # API client, IndexedDB, types, astrology utils
│   │   └── test/         # Vitest tests + MSW API mocks
│   └── public/           # Static assets (PWA manifest, icons)
├── docs/                 # Documentation
├── charts/               # Community chart database
└── plugin/               # Plugin system interface
```

## Development

```bash
# Backend
go build -buildvcs=false -o empirical ./cmd/recover/
go test ./...

# Frontend
cd web
npm ci
npm run dev          # Vite dev server (proxies /api to :5000)
npx vitest run       # Unit tests
npx tsc --noEmit     # Type check
npm run build        # Production build
```

## Next Steps

- [API Reference](api-reference.md) — all endpoints with curl examples
- [CLI Reference](cli-reference.md) — all subcommands
- [Web GUI](web-gui.md) — interface walkthrough
