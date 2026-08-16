# Chart Database

Community-maintained collection of verified birth data for public figures.

## Rodden Rating System

- **AA** — Birth certificate or birth record in hand
- **A** — From memory of person, family, or friend (accurate)
- **B** — Biography or autobiography
- **C** — Caution: no source, conflicting sources
- **DD** — Dirty data: unverified, often from online sources

## Submitting a Chart

1. Verify the birth data against at least one reliable source
2. Include the Rodden rating and source citation
3. Submit via the [Chart Data issue template](https://github.com/aj-nt/empirical/issues/new?template=chart_data.yml)
4. A maintainer will verify and merge into `charts/people.json`

## Format

```json
{
  "name": "Full Name",
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

## Categories

- `science` — Scientists, mathematicians, inventors
- `arts` — Artists, musicians, writers, actors
- `politics` — Political figures, activists
- `philosophy` — Philosophers, spiritual teachers
- `sports` — Athletes
- `business` — Entrepreneurs, business leaders
- `history` — Historical figures
