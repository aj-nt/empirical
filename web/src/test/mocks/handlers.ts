import { http, HttpResponse } from 'msw';

/**
 * MSW handlers for Empirical API endpoints.
 * Returns deterministic data matching real API responses.
 * Used by both Vitest (jsdom) and Playwright (browser) tests.
 */

const CHART_SVG = `<svg viewBox="0 0 800 800" xmlns="http://www.w3.org/2000/svg">
  <rect width="800" height="800" fill="#ffffff"/>
  <circle cx="400" cy="400" r="340" fill="none" stroke="#000" stroke-width="2"/>
  <circle cx="400" cy="400" r="305" fill="none" stroke="#000" stroke-width="1"/>
  <circle cx="400" cy="400" r="180" fill="none" stroke="#000" stroke-width="1.5"/>
  <text x="400" y="60" fill="#000" font-size="18" text-anchor="middle">♌</text>
  <text x="740" y="400" fill="#000" font-size="18" text-anchor="middle">♏</text>
  <text x="400" y="740" fill="#000" font-size="18" text-anchor="middle">♒</text>
  <text x="60" y="400" fill="#000" font-size="18" text-anchor="middle">♉</text>
  <circle cx="400" cy="120" r="16" fill="none" stroke="#000" stroke-width="2"/>
  <text x="400" y="120" fill="#000" font-size="12" text-anchor="middle" dominant-baseline="central">☉</text>
  <circle cx="400" cy="680" r="16" fill="none" stroke="#000" stroke-width="2"/>
  <text x="400" y="680" fill="#000" font-size="12" text-anchor="middle" dominant-baseline="central">☽</text>
  <line x1="400" y1="120" x2="400" y2="680" stroke="#ff0000" stroke-width="1.5"/>
</svg>`;

const BI_WHEEL_SVG = `<svg viewBox="0 0 800 800" xmlns="http://www.w3.org/2000/svg">
  <rect width="800" height="800" fill="#ffffff"/>
  <circle cx="400" cy="400" r="340" fill="none" stroke="#000" stroke-width="2"/>
  <circle cx="400" cy="400" r="180" fill="none" stroke="#000" stroke-width="1.5"/>
  <circle cx="400" cy="400" r="120" fill="none" stroke="#ff0000" stroke-width="1"/>
  <text x="400" y="60" fill="#000" font-size="18" text-anchor="middle">♌</text>
  <circle cx="400" cy="120" r="16" fill="none" stroke="#000" stroke-width="2"/>
  <text x="400" y="120" fill="#000" font-size="12" text-anchor="middle" dominant-baseline="central">☉</text>
  <circle cx="400" cy="100" r="12" fill="none" stroke="#ff0000" stroke-width="2"/>
  <text x="400" y="100" fill="#ff0000" font-size="10" text-anchor="middle" dominant-baseline="central">♃</text>
  <line x1="400" y1="120" x2="400" y2="100" stroke="#00ff00" stroke-width="1.5"/>
</svg>`;

const TRI_WHEEL_SVG = `<svg viewBox="0 0 800 800" xmlns="http://www.w3.org/2000/svg">
  <rect width="800" height="800" fill="#ffffff"/>
  <circle cx="400" cy="400" r="340" fill="none" stroke="#000" stroke-width="2"/>
  <circle cx="400" cy="400" r="280" fill="none" stroke="#0066cc" stroke-width="1.5"/>
  <circle cx="400" cy="400" r="195" fill="none" stroke="#000" stroke-width="1.5"/>
  <circle cx="400" cy="400" r="370" fill="none" stroke="#cc0000" stroke-width="1"/>
  <text x="400" y="60" fill="#000" font-size="18" text-anchor="middle">♌</text>
  <circle cx="400" cy="195" r="16" fill="none" stroke="#000" stroke-width="2"/>
  <text x="400" y="195" fill="#000" font-size="12" text-anchor="middle" dominant-baseline="central">☉</text>
  <circle cx="400" cy="280" r="12" fill="none" stroke="#0066cc" stroke-width="2"/>
  <text x="400" y="280" fill="#0066cc" font-size="10" text-anchor="middle" dominant-baseline="central">☽</text>
  <circle cx="400" cy="370" r="12" fill="none" stroke="#cc0000" stroke-width="2"/>
  <text x="400" y="370" fill="#cc0000" font-size="10" text-anchor="middle" dominant-baseline="central">♃</text>
  <line x1="400" y1="195" x2="400" y2="280" stroke="#0066cc" stroke-width="1.5"/>
  <line x1="400" y1="195" x2="400" y2="370" stroke="#cc0000" stroke-width="1.5"/>
</svg>`;

const INTERPRETATION = {
  planet_signs: [
    "Sun in Aquarius: individuality, vitality, core self, fixed air — detached, innovative, humanitarian. neutral — neither strengthened nor weakened by this sign.",
    "Moon in Aquarius: emotion, instinct, inner life, fixed air — detached, innovative, humanitarian. neutral — neither strengthened nor weakened by this sign.",
    "Mars in Scorpio: action, drive, assertion, fixed water — penetrating, transformative, intense. domicile — strengthened and at home in this sign.",
  ],
  planet_houses: [
    "Sun in the 4th house: vitality expressed through home, roots, family, the private self.",
    "Moon in the 4th house: emotion expressed through home, roots, family, the private self.",
    "Mars in the 1st house: action expressed through self-presentation, the body, first impressions.",
  ],
  aspects: [
    "Sun opposition Moon (orb 2.5°): Sun and Moon in contact — polarity and tension — a seesaw between two extremes.",
    "Mars trine Jupiter (orb 1.2°): Mars and Jupiter in contact — flow and ease — natural harmony, talent that comes without effort.",
    "Venus square Saturn (orb 0.8°): Venus and Saturn in contact — friction and growth — challenge that builds character.",
  ],
  patterns: [
    "Stellium in Aquarius involving Sun, Moon, Mercury: a Stellium in Aquarius configuration.",
    "T-Square involving Mars, Venus, Saturn: a T-Square configuration.",
  ],
};

const TRANSITS = {
  name: "AJ",
  start_date: "2026-07-01",
  end_date: "2026-07-31",
  transits: [
    {
      transit_planet: "Jupiter",
      aspect: "trine",
      natal_planet: "Sun",
      date: "2026-07-15",
      orb: 0.5,
    },
    {
      transit_planet: "Saturn",
      aspect: "square",
      natal_planet: "Mars",
      date: "2026-07-20",
      orb: 1.2,
    },
  ],
  sky_weather: [],
};

const PATTERNS = {
  name: "AJ",
  patterns: [
    {
      type: "stellium",
      planets: ["Sun", "Moon", "Mercury"],
      sign: "Aquarius",
    },
    {
      type: "t_square",
      planets: ["Mars", "Venus", "Saturn"],
    },
  ],
};

const RECOVER = {
  name: "AJ",
  phase1_dignity: {
    Planets: [
      { Planet: "Sun", TropSign: "Aquarius", Western: "peregrine" },
      { Planet: "Moon", TropSign: "Aquarius", Western: "peregrine" },
      { Planet: "Mars", TropSign: "Scorpio", Western: "domicile" },
    ],
  },
  phase3_houses: {
    Planets: [
      { Planet: "Sun", TropicalSign: "Aquarius", Placements: { placidus: 4 } },
      { Planet: "Moon", TropicalSign: "Aquarius", Placements: { placidus: 4 } },
      { Planet: "Mars", TropicalSign: "Scorpio", Placements: { placidus: 1 } },
    ],
  },
};

export const handlers = [
  // Chart SVG
  http.post('/api/chart', () => {
    return new HttpResponse(CHART_SVG, {
      headers: { 'Content-Type': 'image/svg+xml' },
    });
  }),

  // Bi-wheel SVG
  http.post('/api/bi-wheel', () => {
    return new HttpResponse(BI_WHEEL_SVG, {
      headers: { 'Content-Type': 'image/svg+xml' },
    });
  }),

  // Tri-wheel SVG
  http.post('/api/tri-wheel', () => {
    return new HttpResponse(TRI_WHEEL_SVG, {
      headers: { 'Content-Type': 'image/svg+xml' },
    });
  }),

  // Interpretation
  http.post('/api/interpretation', () => {
    return HttpResponse.json(INTERPRETATION);
  }),

  // Transits
  http.post('/api/transits', () => {
    return HttpResponse.json(TRANSITS);
  }),

  // Patterns
  http.post('/api/patterns', () => {
    return HttpResponse.json(PATTERNS);
  }),

  // Recover (natal data)
  http.post('/api/recover', () => {
    return HttpResponse.json(RECOVER);
  }),

  // Synastry
  http.post('/api/synastry', () => {
    return HttpResponse.json({ aspects: [] });
  }),

  // Composite
  http.post('/api/composite', () => {
    return HttpResponse.json({ planets: [] });
  }),

  // Solar return
  http.post('/api/solar-return', () => {
    return HttpResponse.json({ positions: [] });
  }),

  // Research metrics
  http.post('/api/research-metrics', () => {
    return HttpResponse.json({});
  }),

  // Research baseline
  http.post('/api/research-baseline', () => {
    return HttpResponse.json({});
  }),

  // Batch analysis
  http.post('/api/batch-analysis', () => {
    return HttpResponse.json({});
  }),
];
