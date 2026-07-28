// ── Planet Glyphs (Unicode) ──
export const PLANET_GLYPHS: Record<string, string> = {
  Sun: '☉',
  Moon: '☽',
  Mercury: '☿',
  Venus: '♀',
  Mars: '♂',
  Jupiter: '♃',
  Saturn: '♄',
  Uranus: '♅',
  Neptune: '♆',
  Pluto: '♇',
  Chiron: '⚷',
  Ceres: '⚳',
  Pallas: '⚴',
  Juno: '⚵',
  Vesta: '⚶',
  Node: '☊',
  TrueNode: '☊',
  NorthNode: '☊',
  SouthNode: '☋',
  Lilith: '⚸',
  Eris: '⯰',
  Vertex: 'Vx',
  ASC: 'AS',
  MC: 'MC',
  DSC: 'DS',
  IC: 'IC',
  PartOfFortune: '⯃',
};

// ── Sign Glyphs ──
export const SIGN_GLYPHS: Record<string, string> = {
  Aries: '♈',
  Taurus: '♉',
  Gemini: '♊',
  Cancer: '♋',
  Leo: '♌',
  Virgo: '♍',
  Libra: '♎',
  Scorpio: '♏',
  Sagittarius: '♐',
  Capricorn: '♑',
  Aquarius: '♒',
  Pisces: '♓',
};

// ── Aspect Glyphs ──
export const ASPECT_GLYPHS: Record<string, string> = {
  conjunction: '☌',
  opposition: '☍',
  trine: '△',
  square: '□',
  sextile: '⚹',
  quincunx: '⚻',
  'semi-sextile': '⚺',
  'semi-square': '∠',
  sesquiquadrate: '⚼',
  quintile: 'Q',
  'bi-quintile': 'BQ',
  septile: 'S',
  novile: 'N',
};

// ── Sign Colors ──
export const SIGN_COLORS: Record<string, string> = {
  Aries: '#f85149',
  Taurus: '#3fb950',
  Gemini: '#d2991d',
  Cancer: '#a371f7',
  Leo: '#d2753b',
  Virgo: '#8b949e',
  Libra: '#58a6ff',
  Scorpio: '#f85149',
  Sagittarius: '#d2753b',
  Capricorn: '#3fb950',
  Aquarius: '#58a6ff',
  Pisces: '#a371f7',
};

// ── Aspect Colors ──
export const ASPECT_COLORS: Record<string, string> = {
  conjunction: '#f85149',
  opposition: '#f85149',
  square: '#f85149',
  trine: '#3fb950',
  sextile: '#58a6ff',
  quincunx: '#d2991d',
  'semi-sextile': '#8b949e',
  'semi-square': '#8b949e',
  sesquiquadrate: '#8b949e',
};

// ── Planet Colors ──
export const PLANET_COLORS: Record<string, string> = {
  Sun: '#d2753b',
  Moon: '#a371f7',
  Mercury: '#d2991d',
  Venus: '#3fb950',
  Mars: '#f85149',
  Jupiter: '#d2753b',
  Saturn: '#8b949e',
  Uranus: '#58a6ff',
  Neptune: '#a371f7',
  Pluto: '#f85149',
  Chiron: '#8b949e',
  Node: '#58a6ff',
  TrueNode: '#58a6ff',
  SouthNode: '#8b949e',
  NorthNode: '#58a6ff',
  Lilith: '#d2991d',
  Eris: '#d2753b',
  Ceres: '#3fb950',
  Pallas: '#d2991d',
  Juno: '#a371f7',
  Vesta: '#d2753b',
  // Major asteroids
  Astraea: '#8b949e',
  Hebe: '#3fb950',
  Iris: '#a371f7',
  Flora: '#d2991d',
  Metis: '#58a6ff',
  Hygiea: '#3fb950',
  Psyche: '#a371f7',
  Fortuna: '#d2991d',
  Proserpina: '#8b949e',
  Amphitrite: '#58a6ff',
  Pandora: '#d2753b',
  Mnemosyne: '#a371f7',
  Cybele: '#3fb950',
  Diana: '#58a6ff',
  Sappho: '#d2991d',
  Eros: '#f85149',
  // Distant objects
  Orcus: '#8b949e',
  Sedna: '#a371f7',
  Haumea: '#3fb950',
  Makemake: '#58a6ff',
  Gonggong: '#d2753b',
};

// ── Dignity Labels ──
export const DIGNITY_LABELS: Record<string, string> = {
  domicile: 'Domicile',
  exaltation: 'Exaltation',
  detriment: 'Detriment',
  fall: 'Fall',
  peregrine: 'Peregrine',
};

// ── Dignity Colors ──
export const DIGNITY_COLORS: Record<string, string> = {
  domicile: '#3fb950',
  exaltation: '#58a6ff',
  detriment: '#d2753b',
  fall: '#f85149',
  peregrine: '#8b949e',
};

// ── House System Labels ──
export const HOUSE_SYSTEMS: Record<string, string> = {
  P: 'Placidus',
  W: 'Whole Sign',
  K: 'Koch',
  E: 'Equal',
  O: 'Porphyry',
};

// ── Format degree as sign + degree + minute ──
export function formatDegree(lon: number): string {
  const signs = Object.keys(SIGN_GLYPHS);
  const signIdx = Math.floor(lon / 30);
  const sign = signs[signIdx] || '?';
  const deg = Math.floor(lon % 30);
  const min = Math.round((lon % 1) * 60);
  return `${deg}°${min.toString().padStart(2, '0')}′ ${SIGN_GLYPHS[sign]} ${sign}`;
}

// ── Format orb ──
export function formatOrb(orb: number): string {
  const deg = Math.floor(Math.abs(orb));
  const min = Math.round((Math.abs(orb) % 1) * 60);
  return `${deg}°${min.toString().padStart(2, '0')}′`;
}

// ── Get planet glyph ──
export function planetGlyph(name: string): string {
  return PLANET_GLYPHS[name] || name.slice(0, 2);
}

// ── Get sign glyph ──
export function signGlyph(sign: string): string {
  return SIGN_GLYPHS[sign] || sign;
}

// ── Get aspect glyph ──
export function aspectGlyph(aspect: string): string {
  return ASPECT_GLYPHS[aspect] || aspect;
}

// ── Get planet color ──
export function planetColor(name: string): string {
  return PLANET_COLORS[name] || '#8b949e';
}

// ── Get aspect color ──
export function aspectColor(aspect: string): string {
  return ASPECT_COLORS[aspect] || '#8b949e';
}

// ── Get sign color ──
export function signColor(sign: string): string {
  return SIGN_COLORS[sign] || '#8b949e';
}

// ── Planet sort order ──
const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars',
  'Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto',
  'Chiron', 'Ceres', 'Pallas', 'Juno', 'Vesta',
  'Astraea', 'Hebe', 'Iris', 'Flora', 'Metis',
  'Hygiea', 'Psyche', 'Fortuna', 'Proserpina',
  'Amphitrite', 'Pandora', 'Mnemosyne', 'Cybele',
  'Diana', 'Sappho', 'Eros',
  'Node', 'TrueNode', 'SouthNode', 'Lilith',
  'Eris', 'Makemake', 'Gonggong', 'Orcus', 'Sedna', 'Haumea',
  'Vertex', 'ASC', 'MC',
];

export function sortPlanets(planets: string[]): string[] {
  return [...planets].sort((a, b) => {
    const ai = PLANET_ORDER.indexOf(a);
    const bi = PLANET_ORDER.indexOf(b);
    // Unknown planets sort to the end
    const aIdx = ai === -1 ? PLANET_ORDER.length : ai;
    const bIdx = bi === -1 ? PLANET_ORDER.length : bi;
    return aIdx - bIdx;
  });
}

// ── Format date as YYYY-MM-DD ──
export function formatDate(date: Date): string {
  return date.toISOString().slice(0, 10);
}

// ── Add days to a date string ──
export function addDays(dateStr: string, days: number): string {
  const d = new Date(dateStr + 'T00:00:00');
  d.setDate(d.getDate() + days);
  return formatDate(d);
}

// ── Today as YYYY-MM-DD ──
export function today(): string {
  return formatDate(new Date());
}

// ── N days from now ──
export function daysFromNow(n: number): string {
  return addDays(today(), n);
}
