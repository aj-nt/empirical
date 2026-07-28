import { describe, it, expect } from 'vitest';
import {
  planetGlyph,
  signGlyph,
  aspectGlyph,
  planetColor,
  aspectColor,
  signColor,
  sortPlanets,
  formatDegree,
  formatOrb,
  formatDate,
  addDays,
  today,
  daysFromNow,
  PLANET_GLYPHS,
  SIGN_GLYPHS,
  ASPECT_GLYPHS,
  PLANET_COLORS,
  DIGNITY_LABELS,
  HOUSE_SYSTEMS,
} from '../lib/astrology';

// ── Glyph Lookups ──

describe('planetGlyph', () => {
  it('returns glyph for known planets', () => {
    expect(planetGlyph('Sun')).toBe('☉');
    expect(planetGlyph('Moon')).toBe('☽');
    expect(planetGlyph('Mars')).toBe('♂');
    expect(planetGlyph('Pluto')).toBe('♇');
  });

  it('returns first 2 chars for unknown planets', () => {
    expect(planetGlyph('Zeus')).toBe('Ze');
    expect(planetGlyph('Hades')).toBe('Ha');
  });

  it('handles empty string', () => {
    expect(planetGlyph('')).toBe('');
  });
});

describe('signGlyph', () => {
  it('returns glyph for known signs', () => {
    expect(signGlyph('Aries')).toBe('♈');
    expect(signGlyph('Pisces')).toBe('♓');
  });

  it('returns input for unknown signs', () => {
    expect(signGlyph('Unknown')).toBe('Unknown');
  });
});

describe('aspectGlyph', () => {
  it('returns glyph for known aspects', () => {
    expect(aspectGlyph('conjunction')).toBe('☌');
    expect(aspectGlyph('trine')).toBe('△');
    expect(aspectGlyph('square')).toBe('□');
  });

  it('returns input for unknown aspects', () => {
    expect(aspectGlyph('parallel')).toBe('parallel');
  });
});

// ── Color Lookups ──

describe('planetColor', () => {
  it('returns color for known planets', () => {
    expect(planetColor('Sun')).toBe('#d2753b');
    expect(planetColor('Mars')).toBe('#f85149');
  });

  it('returns fallback for unknown', () => {
    expect(planetColor('Unknown')).toBe('#8b949e');
  });
});

describe('aspectColor', () => {
  it('returns red for hard aspects', () => {
    expect(aspectColor('conjunction')).toBe('#f85149');
    expect(aspectColor('square')).toBe('#f85149');
    expect(aspectColor('opposition')).toBe('#f85149');
  });

  it('returns green for trine', () => {
    expect(aspectColor('trine')).toBe('#3fb950');
  });

  it('returns blue for sextile', () => {
    expect(aspectColor('sextile')).toBe('#58a6ff');
  });
});

describe('signColor', () => {
  it('returns color for each element group', () => {
    expect(signColor('Aries')).toBe('#f85149'); // fire
    expect(signColor('Taurus')).toBe('#3fb950'); // earth
    expect(signColor('Gemini')).toBe('#d2991d'); // air
    expect(signColor('Cancer')).toBe('#a371f7'); // water
  });
});

// ── Sorting ──

describe('sortPlanets', () => {
  it('sorts by traditional order', () => {
    const input = ['Pluto', 'Sun', 'Moon', 'Mars'];
    expect(sortPlanets(input)).toEqual(['Sun', 'Moon', 'Mars', 'Pluto']);
  });

  it('puts unknown planets at end', () => {
    const input = ['Sun', 'Zeus', 'Moon'];
    const result = sortPlanets(input);
    expect(result[0]).toBe('Sun');
    expect(result[1]).toBe('Moon');
    expect(result[2]).toBe('Zeus');
  });

  it('does not mutate input', () => {
    const input = ['Pluto', 'Sun'];
    const copy = [...input];
    sortPlanets(input);
    expect(input).toEqual(copy);
  });
});

// ── Formatting ──

describe('formatDegree', () => {
  it('formats 0° as 0°00′ ♈ Aries', () => {
    const result = formatDegree(0);
    expect(result).toContain('0°00′');
    expect(result).toContain('♈');
    expect(result).toContain('Aries');
  });

  it('formats 15.5° as 15°30′', () => {
    const result = formatDegree(15.5);
    expect(result).toContain('15°30′');
  });

  it('wraps at 360°', () => {
    const result = formatDegree(360);
    expect(result).toContain('0°00′');
  });

  it('handles 29°59′ Pisces', () => {
    const result = formatDegree(359.983);
    expect(result).toContain('♓');
    expect(result).toContain('Pisces');
  });
});

describe('formatOrb', () => {
  it('formats positive orb', () => {
    expect(formatOrb(3.5)).toBe('3°30′');
  });

  it('formats negative orb as absolute', () => {
    expect(formatOrb(-1.25)).toBe('1°15′');
  });

  it('formats zero', () => {
    expect(formatOrb(0)).toBe('0°00′');
  });
});

// ── Date Utilities ──

describe('formatDate', () => {
  it('formats date as YYYY-MM-DD', () => {
    expect(formatDate(new Date('2026-07-26'))).toBe('2026-07-26');
  });

  it('pads single-digit months and days', () => {
    expect(formatDate(new Date('2026-01-05'))).toBe('2026-01-05');
  });
});

describe('addDays', () => {
  it('adds days correctly', () => {
    expect(addDays('2026-07-26', 5)).toBe('2026-07-31');
  });

  it('crosses month boundary', () => {
    expect(addDays('2026-07-30', 3)).toBe('2026-08-02');
  });

  it('handles negative days', () => {
    expect(addDays('2026-07-26', -5)).toBe('2026-07-21');
  });
});

describe('today', () => {
  it('returns YYYY-MM-DD format', () => {
    const result = today();
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

describe('daysFromNow', () => {
  it('returns future date', () => {
    const result = daysFromNow(7);
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

// ── Data Integrity ──

describe('data integrity', () => {
  it('PLANET_GLYPHS has all required entries', () => {
    const required = ['Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn'];
    for (const p of required) {
      expect(PLANET_GLYPHS[p]).toBeTruthy();
    }
  });

  it('SIGN_GLYPHS has all 12 signs', () => {
    expect(Object.keys(SIGN_GLYPHS)).toHaveLength(12);
  });

  it('ASPECT_GLYPHS has major aspects', () => {
    const major = ['conjunction', 'opposition', 'trine', 'square', 'sextile'];
    for (const a of major) {
      expect(ASPECT_GLYPHS[a]).toBeTruthy();
    }
  });

  it('DIGNITY_LABELS has all 5 dignities', () => {
    expect(Object.keys(DIGNITY_LABELS)).toHaveLength(5);
  });

  it('HOUSE_SYSTEMS has 5 systems', () => {
    expect(Object.keys(HOUSE_SYSTEMS)).toHaveLength(5);
  });

  it('color maps have matching keys with glyph maps', () => {
    for (const p of Object.keys(PLANET_GLYPHS)) {
      if (p !== 'Vertex' && p !== 'ASC' && p !== 'MC' && p !== 'DSC' && p !== 'IC' && p !== 'PartOfFortune') {
        expect(PLANET_COLORS[p]).toBeTruthy();
      }
    }
  });
});
