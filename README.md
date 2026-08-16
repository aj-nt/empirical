# Empirical Astrology

Cross-system **signal recovery**, not distillation. Nine astrological systems treated as degraded transmissions of one original computational grammar — convergence is triangulation, divergence is stereo information.

Single Go binary. Zero runtime dependencies. Same JPL DE ephemeris data NASA uses for spacecraft navigation (Swiss Ephemeris via CGo).

## Quick Start

```bash
git clone https://github.com/aj-nt/empirical.git
cd empirical
go build -buildvcs=false -o empirical ./cmd/recover/
./empirical serve 5000
```

Open `http://localhost:5000` in your browser.

## What it does

Runs the **recover** protocol across all systems and produces a convergence report: which placements agree across tropical and sidereal traditions, which diverge, and what that tells you.

```
$ empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901

Dignity Convergence Report — AJ
Ayanamsa: 23.4259 deg (Lahiri)

Planet     Trop Sign      Sid Sign       Western        Vedic          Verdict
------------------------------------------------------------------------------
Sun        Aquarius       Aquarius       detriment      peregrine      NOISE
Moon       Aquarius       Capricorn      peregrine      peregrine      SIGNAL
Mercury    Aquarius       Capricorn      peregrine      peregrine      SIGNAL
Venus      Aries          Pisces         detriment      exalted        NOISE
Mars       Scorpio        Scorpio        domicile       own sign       SIGNAL
Jupiter    Libra          Virgo          peregrine      peregrine      SIGNAL
Saturn     Aries          Pisces         fall           peregrine      NOISE

Signal: 4/7 (57%)
```

## Features

- **Chart wheel** — natal, bi-wheel (transits), tri-wheel
- **Interpretation** — Western and Koiné (Hellenistic) systems
- **Predictive tools** — solar returns, progressions, directions, profections, firdaria, zodiacal releasing
- **Relationship astrology** — synastry, composite, draconic synastry
- **Astrocartography** — planetary line maps
- **Research tools** — ephemeris, transit calendar, mundane astrology (67 national charts)
- **Professional features** — tag management, bulk operations, backup/restore, print layout, custom aspect sets, keyboard shortcuts

## Documentation

- [Getting Started](docs/getting-started.md)
- [API Reference](docs/api-reference.md)
- [CLI Reference](docs/cli-reference.md)
- [Web GUI](docs/web-gui.md)
- [Astrology Systems](docs/astrology-systems.md)

## Community

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Report a Bug](https://github.com/aj-nt/empirical/issues/new?template=bug_report.yml)
- [Request a Feature](https://github.com/aj-nt/empirical/issues/new?template=feature_request.yml)

## License

Engine: MIT. Paper: CC-BY 4.0.
