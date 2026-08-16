# API Reference

All endpoints accept POST with JSON bodies unless noted. Server runs at `http://localhost:5000` by default.

## Chart Computation

### `POST /api/chart`

Compute a natal chart.

```bash
curl -s -X POST http://localhost:5000/api/chart \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "AJ",
    "year": 1969, "month": 2, "day": 15,
    "hour": 23, "minute": 10,
    "tz_offset": -8,
    "lat": 47.038, "lng": -122.901,
    "house_system": "placidus",
    "ayanamsa": "tropical"
  }'
```

Returns: SVG string, planet positions, house cusps, aspects.

### `POST /api/bi-wheel`

Bi-wheel (natal + transits or natal + natal).

```bash
curl -s -X POST http://localhost:5000/api/bi-wheel \
  -H 'Content-Type: application/json' \
  -d '{
    "inner": { "year": 1969, "month": 2, "day": 15, "hour": 23, "minute": 10, "tz_offset": -8, "lat": 47.038, "lng": -122.901 },
    "outer": { "year": 2026, "month": 7, "day": 30, "hour": 12, "minute": 0, "tz_offset": 0, "lat": 51.507, "lng": -0.128 },
    "house_system": "placidus",
    "ayanamsa": "tropical"
  }'
```

### `POST /api/tri-wheel`

Tri-wheel (natal + secondary + tertiary).

Same structure as bi-wheel with `inner`, `middle`, `outer` birth data objects.

## Interpretation

### `POST /api/interpretation`

Full natal interpretation.

```bash
curl -s -X POST http://localhost:5000/api/interpretation \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"house_system":"placidus","ayanamsa":"tropical"}'
```

Returns: planet-in-sign, planet-in-house, aspects, patterns, dignities, dispositor tree, element balance, midpoints, star aspects.

### `POST /api/transits`

Transit interpretation for a date range.

```bash
curl -s -X POST http://localhost:5000/api/transits \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"start_date":"2026-07-01","end_date":"2026-07-31","house_system":"placidus","ayanamsa":"tropical"}'
```

## Predictive

### `POST /api/solar-return`

Solar return chart for a given year.

### `POST /api/progressions`

Secondary progressions.

### `POST /api/directions`

Primary directions.

### `POST /api/profections`

Annual profections.

### `POST /api/firdaria`

Firdaria periods.

### `POST /api/zodiacal-releasing`

Zodiacal releasing from the Lot of Fortune.

All predictive endpoints accept the same birth data JSON as `/api/chart` plus a target date.

## Relationship

### `POST /api/synastry`

Synastry between two charts.

```bash
curl -s -X POST http://localhost:5000/api/synastry \
  -H 'Content-Type: application/json' \
  -d '{
    "chart_a": {"year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901},
    "chart_b": {"year":1970,"month":6,"day":20,"hour":8,"minute":30,"tz_offset":-7,"lat":34.052,"lng":-118.244},
    "house_system": "placidus",
    "ayanamsa": "tropical"
  }'
```

Returns: 201+ aspects between all bodies, house overlays.

### `POST /api/composite`

Composite (midpoint) chart.

### `POST /api/draconic-synastry`

Draconic synastry comparison.

### `POST /api/draconic-synastry-full`

Full three-layer draconic comparison with bridges.

## Other

### `POST /api/recover`

Cross-system convergence report.

### `POST /api/astrocartography`

Planetary line map data.

### `POST /api/ephemeris`

Planetary positions for a date range.

### `POST /api/mundane`

Mundane astrology for a national chart.

### `GET /api/chartdb`

List charts in the embedded chart database.

## Common Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `house_system` | string | `"placidus"` | Placidus, Whole Sign, Equal, Koch, Porphyry, Regiomontanus, Alcabitius, Campanus |
| `ayanamsa` | string | `"tropical"` | Tropical, Lahiri, Fagan-Bradley, Raman, Krishnamurti |
| `orb` | number | `3` | Aspect orb in degrees |
