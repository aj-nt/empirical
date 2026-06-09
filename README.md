# Empirical Astrology

Cross-system **signal recovery**, not distillation. Nine astrological systems treated as degraded transmissions of one original computational grammar — convergence is triangulation, divergence is stereo information.

Single Go binary. Zero runtime dependencies. Same JPL DE ephemeris data NASA uses for spacecraft navigation (Swiss Ephemeris via CGo).

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

Use `--json` for machine-readable output.

## Installing

Download the binary for your platform from [Releases](https://github.com/aj-nt/empirical/releases). No install step — it's self-contained. The binary extracts ephemeris data files to `~/.cache/empirical/ephe/` on first run.

## Building from source

Requires Go 1.26+ and the Swiss Ephemeris C library.

```bash
# Install Swiss Ephemeris C library
git clone https://github.com/aloistr/swisseph
cd swisseph && make libswe.a
sudo cp libswe.a /usr/local/lib/
sudo cp src/*.h /usr/local/include/

# Build
go build -o empirical ./cmd/recover
```

## The thesis

Modern astrology is a degraded transmission of an original computational system. Each surviving tradition (Western, Vedic, Chinese, Uranian, etc.) preserves a partial signal. By measuring where they agree and where they diverge, we can recover the original grammar — not by choosing a "correct" system, but by treating the variance itself as information.

This is falsifiable, not interpretive. The convergence scores are deterministic. The claims can be tested.

### Signal recovery phases

1. **Dignity convergence** — Western vs Vedic planetary dignity. Where they agree = signal.
2. **Aspect geometry** — Which angle families are universal? Only conjunction (0), trine (120), and opposition (180) survive all three traditions.
3. **House division** — 5 systems compared. Whole-sign houses show highest cross-system agreement.
4. **Timing layers** — Vedic dasha, Ba Zi luck pillars, Hellenistic profections as complementary timing frameworks.
5. **Node axis** — The lunar node axis is invariant across all coordinate systems.
6. **Zodiac comparison** — 9,738 synthetic charts confirm tropical and sidereal are symmetric views of the same grammar. Neither is "correct."

## License

MIT

## Status

In development. Phases 1-2 ported to Go, cross-validated against the Python reference implementation. Phases 3-6 in progress.
