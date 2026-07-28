import type {
  BirthData,
  ChartRequest,
  TransitResponse,
  InterpretationResponse,
  DirectionsResponse,
  AstroCartographyResponse,
  ElectionalResponse,
  SynastryResponse,
  CompositeResponse,
  StarsResponse,
  DraconicResponse,
  SolarReturnResponse,
  FirdariaResponse,
  TraditionalResponse,
  ComparisonReport,
  VedicNatalReport,
  BaZiFourPillars,
  AspectDef,
  RelocationResponse,
  MansionConvergenceResponse,
  ArabicPartsResponse,
  HarmonicResponse,
  DivisionalResponse,
  ParansResponse,
  DeclinationResponse,
  UranianResponse,
  ProgressedResponse,
  ProgressedCrossResponse,
  DraconicTransitsCrossResponse,
  StarsCrossResponse,
  BaseChartResponse,
  ResearchMetrics,
  ResearchBaseline,
  BatchAnalysisResponse,
} from './types';

const BASE = '';

async function post<T>(path: string, body: unknown): Promise<T> {
  const r = await fetch(BASE + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const text = await r.text().catch(() => r.statusText);
    throw new Error(`${r.status}: ${text}`);
  }
  return r.json();
}

async function postText(path: string, body: unknown): Promise<string> {
  const r = await fetch(BASE + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const text = await r.text().catch(() => r.statusText);
    throw new Error(`${r.status}: ${text}`);
  }
  return r.text();
}

function bd(data: BirthData): ChartRequest {
  return { ...data, house_system: 'P', orb: 3 };
}

// ── Core ──
export const api = {
  // Chart wheel SVG
  chart: (data: BirthData, opts?: Partial<ChartRequest>) =>
    postText('/api/chart', { ...bd(data), ...opts }),

  // Natal interpretation
  interpretation: (data: BirthData, system = 'western', orb = 3) =>
    post<InterpretationResponse>('/api/interpretation', {
      ...bd(data),
      system,
      orb,
    }),

  // Transits
  transits: (data: BirthData, startDate: string, endDate: string, orb = 3, sidereal = false) =>
    post<TransitResponse>('/api/transits', {
      ...bd(data),
      start_date: startDate,
      end_date: endDate,
      orb,
      sidereal,
    }),

  // Patterns
  patterns: (data: BirthData, orb = 5) =>
    post<{ patterns: { name: string; planets: string[]; description: string }[] }>(
      '/api/patterns',
      { ...bd(data), orb }
    ),

  // Aspect catalog
  aspectCatalog: () => post<AspectDef[]>('/api/aspect-catalog', {}),

  // Primary directions
  directions: (data: BirthData, age: number, orb = 1) =>
    post<DirectionsResponse>('/api/directions', { ...bd(data), age, orb }),

  // Synastry
  synastry: (a: BirthData, b: BirthData, orb = 5) =>
    post<SynastryResponse>('/api/synastry', {
      name1: a.name, year1: a.year, month1: a.month, day1: a.day,
      hour1: a.hour, min1: a.minute, tz1: a.tz_offset, lat1: a.lat, lng1: a.lng,
      name2: b.name, year2: b.year, month2: b.month, day2: b.day,
      hour2: b.hour, min2: b.minute, tz2: b.tz_offset, lat2: b.lat, lng2: b.lng,
      orb,
    }),

  // Composite
  composite: (a: BirthData, b: BirthData, orb = 3) =>
    post<CompositeResponse>('/api/composite', {
      name1: a.name, year1: a.year, month1: a.month, day1: a.day,
      hour1: a.hour, min1: a.minute, tz1: a.tz_offset, lat1: a.lat, lng1: a.lng,
      name2: b.name, year2: b.year, month2: b.month, day2: b.day,
      hour2: b.hour, min2: b.minute, tz2: b.tz_offset, lat2: b.lat, lng2: b.lng,
      orb,
    }),

  // Fixed stars
  stars: (data: BirthData, orb = 2) =>
    post<StarsResponse>('/api/stars', { ...bd(data), orb }),

  // Draconic
  draconic: (data: BirthData, orb = 3) =>
    post<DraconicResponse>('/api/draconic', { ...bd(data), orb }),

  // Solar return
  solarReturn: (data: BirthData, targetYear: number) =>
    post<SolarReturnResponse>('/api/solar-return', {
      ...bd(data),
      target_year: targetYear,
    }),

  // Firdaria
  firdaria: (data: BirthData) =>
    post<FirdariaResponse>('/api/firdaria', bd(data)),

  // Traditional
  traditional: (data: BirthData) =>
    post<TraditionalResponse>('/api/traditional', bd(data)),

  // Cross-system comparison
  compare: (data: BirthData, orb = 5) =>
    post<ComparisonReport>('/api/compare', { ...bd(data), orb }),

  // Vedic natal
  vedic: (data: BirthData) =>
    post<VedicNatalReport>('/api/system', { ...bd(data), system: 'vedic' }),

  // Ba Zi
  bazi: (data: BirthData) =>
    post<BaZiFourPillars>('/api/system', { ...bd(data), system: 'bazi' }),

  // Astrocartography
  astrocartography: (data: BirthData, latStep = 2, frame = 'tropical') =>
    post<AstroCartographyResponse>('/api/astrocartography', {
      ...bd(data),
      lat_step: latStep,
      frame,
    }),

  // Electional
  electional: (data: BirthData, startDate: string, endDate: string, orb = 3) =>
    post<ElectionalResponse>('/api/electional', {
      ...bd(data),
      start_date: startDate,
      end_date: endDate,
      orb,
    }),

  // Relocation
  relocation: (
    data: BirthData,
    locA: { name: string; lat: number; lng: number },
    locB: { name: string; lat: number; lng: number },
    targetDate: string
  ) =>
    post<RelocationResponse>('/api/relocation-compare', {
      ...bd(data),
      location_a: locA,
      location_b: locB,
      target_date: targetDate,
    }),

  // Mansion convergence
  mansionConvergence: (data: BirthData) =>
    post<MansionConvergenceResponse>('/api/mansion-convergence', bd(data)),

  // Arabic parts
  arabicParts: (data: BirthData, orb = 3) =>
    post<ArabicPartsResponse>('/api/arabic-parts', { ...bd(data), orb }),

  // Harmonic
  harmonic: (data: BirthData, harmonics: number[] = [4, 5, 7, 9], orb = 2) =>
    post<HarmonicResponse>('/api/harmonic', { ...bd(data), harmonics, orb }),

  // Divisional (Vedic)
  divisional: (data: BirthData) =>
    post<DivisionalResponse>('/api/divisional', bd(data)),

  // Parans
  parans: (data: BirthData, orb = 2) =>
    post<ParansResponse>('/api/parans', { ...bd(data), orb }),

  // Declination
  declination: (data: BirthData, orb = 1) =>
    post<DeclinationResponse>('/api/declination', { ...bd(data), orb }),

  // Uranian
  uranian: (data: BirthData) =>
    post<UranianResponse>('/api/uranian', bd(data)),

  // Progressed
  progressed: (data: BirthData, targetDate: string, orb = 3) =>
    post<ProgressedResponse>('/api/progressed', {
      ...bd(data),
      target_date: targetDate,
      orb,
    }),

  // Progressed cross-system
  progressedCross: (data: BirthData, targetDate: string, orb = 3) =>
    post<ProgressedCrossResponse>('/api/progressed-cross', {
      ...bd(data),
      target_date: targetDate,
      orb,
    }),

  // Draconic transits cross
  draconicTransitsCross: (data: BirthData, startDate: string, endDate: string, orb = 3) =>
    post<DraconicTransitsCrossResponse>('/api/draconic-transits-cross', {
      ...bd(data),
      start_date: startDate,
      end_date: endDate,
      orb,
    }),

  // Stars cross
  starsCross: (data: BirthData, orb = 2) =>
    post<StarsCrossResponse>('/api/stars-cross', { ...bd(data), orb }),

  // Base chart (raw physics)
  baseChart: (data: BirthData) =>
    post<BaseChartResponse>('/api/base-chart', bd(data)),

  // Research metrics
  researchMetrics: (data: BirthData) =>
    post<ResearchMetrics>('/api/research-metrics', bd(data)),

  // Research baseline
  researchBaseline: (metric: string, n = 1000, seed = 0) =>
    post<ResearchBaseline>('/api/research-baseline', { metric, n, seed }),

  // Batch analysis
  batchAnalysis: (charts: BirthData[]) =>
    post<BatchAnalysisResponse>('/api/batch-analysis', { charts }),

  // Natal HTML
  natalHTML: (data: BirthData, system = 'western', orb = 3) =>
    postText('/api/natal-html', { ...bd(data), system, orb }),

  // Transit HTML
  transitHTML: (
    data: BirthData,
    transitDate: { year: number; month: number; day: number; hour: number; minute: number; tz: number; lat: number; lng: number },
    system = 'western',
    orb = 3
  ) =>
    postText('/api/transit-html', {
      ...bd(data),
      transit_year: transitDate.year,
      transit_month: transitDate.month,
      transit_day: transitDate.day,
      transit_hour: transitDate.hour,
      transit_minute: transitDate.minute,
      transit_tz: transitDate.tz,
      transit_lat: transitDate.lat,
      transit_lng: transitDate.lng,
      system,
      orb,
    }),
};
