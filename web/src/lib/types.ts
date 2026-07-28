// ── Birth Data ──
export interface BirthData {
  name: string;
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
  tz_offset: number;
  lat: number;
  lng: number;
}

// ── Chart Request (extends BirthData with display options) ──
export interface ChartRequest extends BirthData {
  house_system?: string;
  sidereal?: boolean;
  show_aspects?: boolean;
  outer_planets?: boolean;
  orb?: number;
  highlight_patterns?: boolean;
  pattern_orb?: number;
  system?: string;
}

// ── Saved Chart (stored in IndexedDB) ──
export interface SavedChart {
  id?: number;
  name: string;
  birthData: BirthData;
  tags: string[];
  notes: string;
  createdAt: string;
  updatedAt: string;
}

// ── Planet Position ──
export interface PlanetPosition {
  planet: string;
  sign: string;
  sign_lon: number;
  house: number;
  lon: number;
  speed_deg_per_day: number;
  retrograde: boolean;
  declination: number;
  lat: number;
  dist_au: number;
}

// ── Aspect ──
export interface Aspect {
  planet1: string;
  planet2: string;
  aspect: string;
  orb: number;
  angle_degrees: number;
}

// ── Pattern ──
export interface Pattern {
  name: string;
  planets: string[];
  description: string;
}

// ── Transit Hit ──
export interface TransitHit {
  date: string;
  transit_planet: string;
  natal_planet: string;
  aspect: string;
  orb: number;
  start_date?: string;
  end_date?: string;
}

// ── Transit Response ──
export interface TransitResponse {
  name: string;
  start_date: string;
  end_date: string;
  transits: TransitHit[];
  sky_weather: TransitHit[];
  midpoints?: TransitMidpointHit[];
}

export interface TransitMidpointHit {
  date: string;
  transit_planet: string;
  natal_pair: [string, string];
  orb: number;
}

// ── Interpretation ──
export interface InterpretationResponse {
  planet_signs: string[];
  planet_houses: string[];
  aspects: string[];
  patterns: string[];
  lunar_phase?: string;
  retrogrades?: string[];
  antiscia?: string[];
  antiscia_contacts?: string[];
  mutual_receptions?: string[];
  decans?: string[];
  terms?: string[];
  voc_moon?: string;
  sect?: string;
  declinations?: string[];
  contraparallels?: string[];
  star_aspects?: string[];
  chart_ruler?: string;
  final_dispositor?: string;
  weighted_aspects?: string[];
  key_midpoints?: string[];
  key_star_aspects?: string[];
  angular_planets?: string[];
}

// ── Directions ──
export interface DirectionsResponse {
  name: string;
  age_years: number;
  directed_asc: number;
  directed_mc: number;
  asc_aspects: Aspect[];
  mc_aspects: Aspect[];
}

// ── Astrocartography ──
export interface AstroCartographyLine {
  planet: string;
  angle: string;
  points: { lat: number; lon: number }[];
}

export interface AstroCartographyResponse {
  name: string;
  gmst: number;
  lines: AstroCartographyLine[];
}

// ── Electional ──
export interface ElectionalResult {
  date: string;
  day: string;
  score: number;
  moon_sign: string;
  moon_house: number;
  merc_sign: string;
  good: string[];
  bad: string[];
}

export interface ElectionalResponse {
  name: string;
  start_date: string;
  end_date: string;
  results: ElectionalResult[];
}

// ── Synastry ──
export interface SynastryResponse {
  name1: string;
  name2: string;
  aspects: Aspect[];
}

// ── Composite ──
export interface CompositeResponse {
  name1: string;
  name2: string;
  planets: PlanetPosition[];
  aspects: Aspect[];
  patterns: Pattern[];
}

// ── Fixed Stars ──
export interface StarConjunction {
  star: string;
  planet: string;
  aspect: string;
  orb: number;
  meaning: string;
}

export interface StarsResponse {
  name: string;
  conjunctions: StarConjunction[];
}

// ── Draconic ──
export interface DraconicResponse {
  name: string;
  offset: number;
  planets: Record<string, number>;
  sign_shifts: { planet: string; tropical_sign: string; draconic_sign: string }[];
  bridges: { planet1: string; planet2: string; aspect: string; orb: number }[];
}

// ── Solar Return ──
export interface SolarReturnResponse {
  name: string;
  target_year: number;
  return_date: string;
  planets: PlanetPosition[];
  aspects: Aspect[];
  patterns: Pattern[];
}

// ── Firdaria ──
export interface FirdariaPeriod {
  planet: string;
  start_date: string;
  end_date: string;
  sub_periods?: FirdariaPeriod[];
}

export interface FirdariaResponse {
  name: string;
  periods: FirdariaPeriod[];
}

// ── Traditional ──
export interface DispositorNode {
  planet: string;
  sign: string;
  dispositor: string;
  is_final: boolean;
  in_loop: boolean;
}

export interface DispositorTree {
  nodes: DispositorNode[];
  final_dispositors: string[];
  mutual_receptions: string[][] | null;
}

export interface LunarPhase {
  name: string;
  angle_deg: number;
  phase_index: number;
}

export interface RetrogradeInfo {
  planet: string;
  retrograde: boolean;
  speed_deg_per_day: number;
}

export interface AntiscionInfo {
  planet: string;
  longitude: number;
  antiscion: number;
  antiscion_sign: string;
  contra_antiscion: number;
  contra_antiscion_sign: string;
}

export interface DecanInfo {
  planet: string;
  sign: string;
  decan: number;
  ruler: string;
}

export interface TermInfo {
  planet: string;
  sign: string;
  term: number;
  ruler: string;
}

export interface VOCMoonInfo {
  void_of_course: boolean;
  moon_sign: string;
  moon_lon: number;
  next_sign: string;
  degrees_to_next_sign: number;
}

export interface TraditionalResponse {
  name: string;
  lunar_phase: LunarPhase;
  retrogrades: RetrogradeInfo[];
  antiscia: AntiscionInfo[];
  decans: DecanInfo[];
  terms: TermInfo[];
  dispositor_tree: DispositorTree;
  void_of_course_moon: VOCMoonInfo;
}

// ── Compare ──
export interface ComparisonItem {
  planet: string;
  systems: Record<string, string | number>;
  agrees: boolean;
}

export interface ComparisonReport {
  name: string;
  systems: string[];
  planet_signs: ComparisonItem[];
  planet_houses: ComparisonItem[];
  dignity_comparison: ComparisonItem[];
  summary: {
    sign_agreement: number;
    house_agreement: number;
    dignity_agreement: number;
    total_planets: number;
  };
}

// ── Vedic ──
export interface VedicNatalReport {
  name: string;
  ayanamsa: number;
  ascendant: { sidereal_lon: number; sidereal_sign: string; nakshatra: string; nakshatra_pada: number; nakshatra_ruler: string };
  house_lords: Record<string, string>;
  planets: VedicPlanet[];
  aspects: { planet1: string; planet2: string; aspect: string; orb: number }[];
  yogas: string[];
  vargas: Record<string, Record<string, string>>;
  dasha: VedicDasha[];
  antardasha: VedicDasha[];
  signal_count: number;
  total_planets: number;
}

export interface VedicPlanet {
  planet: string;
  sidereal_lon: number;
  sidereal_sign: string;
  nakshatra: string;
  nakshatra_pada: number;
  nakshatra_ruler: string;
  nakshatra_lord_house: number;
  navamsha_sign: string;
  house: number;
  dignity: string;
  western_dignity: string;
  convergence: string;
  retrograde: boolean;
  combust: boolean;
}

export interface VedicDasha {
  planet: string;
  start: string;
  end: string;
  years: number;
  antardasha?: VedicDasha[];
}

// ── Ba Zi ──
export interface BaZiFourPillars {
  name: string;
  year_pillar: string;
  month_pillar: string;
  day_pillar: string;
  hour_pillar: string;
  day_master: string;
}

// ── Aspect Catalog ──
export interface AspectDef {
  Name: string;
  AngleDegrees: number;
  Universality: string;
}

// ── Relocation ──
export interface RelocationResponse {
  name: string;
  location_a: { name: string; lat: number; lng: number };
  location_b: { name: string; lat: number; lng: number };
  house_shifts: { planet: string; house_a: number; house_b: number }[];
}

// ── Mansion Convergence ──
export interface MansionConvergenceResponse {
  name: string;
  ayanamsa: number;
  planets: { planet: string; tropical_lon: number; sidereal_lon: number; nakshatra: string; nakshatra_num: number; xiu: string; xiu_num: number; xiu_pinyin: string; converges: boolean }[];
  converging: number;
  total: number;
}

// ── Arabic Parts ──
export interface ArabicPartsResponse {
  name: string;
  ayanamsa: number;
  is_day: boolean;
  tropical: { part: string; lon: number; sign: string; sign_num: number }[];
  sidereal: { part: string; lon: number; sign: string; sign_num: number }[];
  sign_survivors: number;
  total: number;
  aspects: { part: string; planet: string; aspect: string; orb: number }[];
}

// ── Harmonic ──
export interface HarmonicEntry {
  harmonic: number;
  aspect_name: string;
  positions: Record<string, number>;
  conjunctions: { planet_a: string; planet_b: string; harmonic_lon_a: number; harmonic_lon_b: number; orb: number }[];
}

export interface HarmonicResponse {
  name: string;
  harmonics: HarmonicEntry[];
}

// ── Divisional ──
export interface DivisionalResponse {
  name: string;
  navamsha: PlanetPosition[];
  nakshatras: { planet: string; nakshatra: string; pada: number; ruler: string }[];
  dasha: VedicDasha[];
}

// ── Parans ──
export interface AngleContact {
  body: string;
  body_lon: number;
  angle: string;
  angle_lon: number;
  orb: number;
}

export interface ParanEntry {
  star: string;
  star_lon: number;
  planet: string;
  planet_lon: number;
  angle: string;
  angle_lon: number;
  star_orb: number;
  planet_orb: number;
}

export interface ParansResponse {
  name: string;
  angles: Record<string, number>;
  stars_on_angles: AngleContact[];
  planets_on_angles: AngleContact[];
  parans: ParanEntry[];
}

// ── Declination ──
export interface DeclinationResponse {
  name: string;
  parallels: string[];
  contraparallels: string[];
}

// ── Uranian ──
export interface UranianResponse {
  name: string;
  dial_positions: Record<string, number>;
  midpoint_pictures: string[];
  planetary_pictures: string[];
}

// ── Progressed ──
export interface ProgressedResponse {
  name: string;
  target_date: string;
  planets: PlanetPosition[];
  aspects: Aspect[];
  patterns: Pattern[];
}

// ── Progressed Cross ──
export interface ProgressedCrossResponse {
  name: string;
  target_date: string;
  age_years: number;
  ayanamsa: number;
  orb: number;
  survivors: { progressed_planet: string; natal_planet: string; aspect: string; orb: number }[];
  tropical_only: { progressed_planet: string; natal_planet: string; aspect: string; orb: number }[];
  sidereal_only: { progressed_planet: string; natal_planet: string; aspect: string; orb: number }[];
}

// ── Draconic Transits Cross ──
export interface DraconicTransitsCrossResponse {
  name: string;
  start_date: string;
  end_date: string;
  tropical_hits: TransitHit[];
  sidereal_hits: TransitHit[];
  survivors: TransitHit[];
  survival_rate: number;
}

// ── Stars Cross ──
export interface StarsCrossResponse {
  name: string;
  ayanamsa: number;
  orb: number;
  tropical: { star: string; star_lon: number; planet: string; planet_lon: number; orb: number; meaning: string }[];
  sidereal: { star: string; star_lon: number; planet: string; planet_lon: number; orb: number; meaning: string }[];
  survivors: { star: string; star_lon: number; planet: string; planet_lon: number; orb: number; meaning: string }[];
  tropical_only: { star: string; star_lon: number; planet: string; planet_lon: number; orb: number; meaning: string }[];
  sidereal_only: { star: string; star_lon: number; planet: string; planet_lon: number; orb: number; meaning: string }[];
  total_tropical: number;
  total_sidereal: number;
  total_survivors: number;
}

// ── Research Metrics ──
export interface ResearchMetrics {
  cross_system_sign_agreement: number;
  draconic_bridge_count: number;
  harmonic_conjunctions: Record<string, number>;
  paran_count: number;
  declination_parallel_count: number;
  arabic_parts_survivor_pct: number;
  mansion_convergence_count: number;
  stars_cross_survivor_pct: number;
  aspect_pattern_count: number;
}

// ── Research Baseline ──
export interface ResearchBaseline {
  metric: string;
  n: number;
  mean: number;
  median: number;
  std_dev: number;
  min: number;
  max: number;
  p5: number;
  p25: number;
  p50: number;
  p75: number;
  p95: number;
}

// ── Batch Analysis ──
export interface BatchChartResult {
  name: string;
  metrics: ResearchMetrics | null;
}

export interface BatchAnalysisResponse {
  charts: BatchChartResult[];
  aggregates: Record<string, ResearchBaseline>;
}

// ── Base Chart (raw physics) ──
export interface BaseChartResponse {
  name: string;
  jd: number;
  gmst: number;
  ayanamsa: number;
  tropical: Record<string, { lon: number; lat: number; speed: number; dist: number }>;
  sidereal: Record<string, { lon: number; lat: number; speed: number; dist: number }>;
  asc: number;
  mc: number;
  dsc: number;
  ic: number;
  north_node: number;
  south_node: number;
  houses: Record<string, number[]>;
  star_positions: Record<string, number>;
  declinations: Record<string, number>;
}
