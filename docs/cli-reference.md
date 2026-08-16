# CLI Reference

All subcommands of the `empirical` binary.

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON instead of text |
| `--sidereal` | Use sidereal zodiac (Lahiri ayanamsa) |
| `--house-system` | House system (default: placidus) |

## Chart Commands

### `recover`

Cross-system convergence report.

```bash
empirical recover <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng>
```

Options: `--json`, `--sidereal`, `--house-system`

### `transit`

Transit analysis for a date range.

```bash
empirical transit <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng> <start_date> <end_date>
```

### `solar-return`

Solar return chart for a target year.

```bash
empirical solar-return <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng> <target_year>
```

### `progressions`

Secondary progressions.

```bash
empirical progressions <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng> <target_date>
```

### `directions`

Primary directions.

```bash
empirical directions <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng> <target_date>
```

### `profections`

Annual profections.

```bash
empirical profections <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng> <target_date>
```

### `firdaria`

Firdaria periods.

```bash
empirical firdaria <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng>
```

### `zodiacal-releasing`

Zodiacal releasing from the Lot of Fortune.

```bash
empirical zodiacal-releasing <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng>
```

## Relationship Commands

### `synastry`

Synastry between two charts.

```bash
empirical synastry <name1> <y1> <m1> <d1> <h1> <min1> <tz1> <lat1> <lng1> <name2> <y2> <m2> <d2> <h2> <min2> <tz2> <lat2> <lng2>
```

### `composite`

Composite (midpoint) chart.

```bash
empirical composite <name1> <y1> <m1> <d1> <h1> <min1> <tz1> <lat1> <lng1> <name2> <y2> <m2> <d2> <h2> <min2> <tz2> <lat2> <lng2>
```

### `draconic`

Draconic synastry.

```bash
empirical draconic <name1> <y1> <m1> <d1> <h1> <min1> <tz1> <lat1> <lng1> <name2> <y2> <m2> <d2> <h2> <min2> <tz2> <lat2> <lng2>
```

## Other Commands

### `serve`

Start the HTTP server and web UI.

```bash
empirical serve [port]
```

Default port: 5000.

### `ephemeris`

Planetary positions for a date range.

```bash
empirical ephemeris <start_date> <end_date> [step_days]
```

### `astrocartography`

Planetary line map data.

```bash
empirical astrocartography <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng>
```

### `mundane`

Mundane astrology for a national chart.

```bash
empirical mundane <country_code> <start_date> <end_date>
```

Country codes: US, GB, FR, DE, JP, CN, IN, RU, BR, etc. (67 total).

### `chartdb`

Query the embedded chart database.

```bash
empirical chartdb list                     # List all charts
empirical chartdb show <name>              # Show one chart
empirical chartdb recover <name>           # Run recover on one chart
empirical chartdb search <query>           # Search by name/category
```

### `interpretation`

Full natal interpretation.

```bash
empirical interpretation <name> <year> <month> <day> <hour> <minute> <tz_offset> <lat> <lng>
```

Options: `--system western|koine` (default: western)

## Examples

```bash
# Quick natal chart
empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901

# Sidereal natal
empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901 --sidereal

# JSON output for scripting
empirical recover AJ 1969 2 15 23 10 -8 47.038 -122.901 --json | jq '.planets[] | {name: .name, sign: .sign}'

# Transits for July 2026
empirical transit AJ 1969 2 15 23 10 -8 47.038 -122.901 2026-07-01 2026-07-31

# Start the web server
empirical serve 5000
```
