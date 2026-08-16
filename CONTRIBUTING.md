# Contributing to Empirical

Thank you for your interest in contributing to Empirical! This guide covers everything you need to get started.

## Development Setup

### Prerequisites

- **Go** 1.25+ ([download](https://go.dev/dl/))
- **Node.js** 20+ and **npm** ([download](https://nodejs.org/))
- **Git**

### Clone & Build

```bash
git clone https://github.com/aj-nt/empirical.git
cd empirical

# Build the Go binary
go build ./...

# Or build and run the server
go run ./cmd/regulus
```

### Frontend Development

```bash
cd web
npm ci
npm run dev        # Vite dev server with hot reload
```

The frontend dev server proxies API requests to the Go backend (default `localhost:8080`).

### Running Both Together

1. Start the Go backend: `go run ./cmd/regulus`
2. In another terminal, start the frontend: `cd web && npm run dev`
3. Open the Vite dev server URL (usually `http://localhost:5173`)

## Project Structure

```
empirical/
├── cmd/                    # Go command-line tools and server entry points
│   ├── regulus/           # Main server (API + static files)
│   ├── verify_reading/    # Verify factual claims in chart readings
│   └── ...                # Other utility commands
├── internal/               # Go library packages
│   ├── server/            # HTTP handlers and routing
│   ├── dignity/           # Essential dignity calculations & interpretations
│   ├── systems/           # Astrology system implementations
│   │   ├── western/       # Tropical Western astrology
│   │   ├── vedic/         # Vedic/Jyotish astrology
│   │   ├── bazi/          # Chinese BaZi (Four Pillars)
│   │   ├── koine/         # Hellenistic astrology
│   │   └── draconic/      # Draconic astrology
│   ├── swe/               # Swiss Ephemeris wrapper
│   ├── geocode/           # Geocoding & timezone resolution
│   ├── comparison/        # Chart comparison logic
│   └── ...                # Other packages
├── web/                    # React + TypeScript frontend
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── lib/           # API client, types, utilities
│   │   ├── hooks/         # React hooks
│   │   └── styles/        # CSS
│   └── e2e/               # Playwright end-to-end tests
├── ephe/                   # Embedded ephemeris data files
└── docs/                   # Documentation and research paper
```

## Running Tests

### Go Tests

```bash
# Run all Go tests
go test ./...

# Run tests for a specific package
go test ./internal/dignity/...

# Run with verbose output
go test -v ./...
```

### Frontend Tests

```bash
cd web

# Unit tests (Vitest)
npx vitest run

# Watch mode
npx vitest

# TypeScript type check
npx tsc --noEmit

# Build production bundle
npm run build

# End-to-end tests (Playwright)
npx playwright test
```

## How to Add a New Interpretation

Interpretations live in `internal/dignity/` and related packages. Each interpretation type has its own file:

1. **Find the right file** — e.g., planet-in-house interpretations go in `internal/dignity/planet_in_house.go`, aspect interpretations in `internal/dignity/aspects.go`.

2. **Add your interpretation data** — most interpretation functions use a template or lookup table. Find the relevant map/switch and add your entry.

3. **Add tests** — add a test case covering the new interpretation in the corresponding `_test.go` file.

4. **Verify** — run `go test ./internal/dignity/...` to confirm nothing is broken.

## How to Add a New API Endpoint

1. **Create a handler** in `internal/server/` — add a new handler function following the existing patterns:

   ```go
   func (s *Server) handleYourFeature(w http.ResponseWriter, r *http.Request) {
       // Parse request, call business logic, write JSON response
   }
   ```

2. **Register the route** in `internal/server/routes.go` (or the relevant routes file):

   ```go
   mux.HandleFunc("GET /api/your-feature", s.handleYourFeature)
   ```

3. **Add frontend API client** — update `web/src/lib/api.ts` with a typed function for the new endpoint.

4. **Add tests** — both Go handler tests and, if applicable, frontend component tests.

## Pull Request Checklist

Before submitting a PR, verify the following:

- [ ] `go test ./...` — all Go tests pass
- [ ] `cd web && npx tsc --noEmit` — TypeScript is clean
- [ ] `cd web && npx vitest run` — all frontend unit tests pass
- [ ] `cd web && npm run build` — production build succeeds
- [ ] Manual verification — the feature works as expected in the browser

## Code Style

- **Go**: Follow standard `gofmt` formatting. Run `go vet ./...` before committing.
- **TypeScript/React**: Follow the existing patterns in the codebase. The project uses functional components with hooks.

## Reporting Issues

- **Bugs**: Use the [Bug Report](../../.github/ISSUE_TEMPLATE/bug_report.yml) template.
- **Features**: Use the [Feature Request](../../.github/ISSUE_TEMPLATE/feature_request.yml) template.
- **Birth data**: Use the [Chart Data Submission](../../.github/ISSUE_TEMPLATE/chart_data.yml) template for verified public figure birth data.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.