import { describe, it, expect } from 'vitest';
import { generateICal } from '../lib/export';

describe('generateICal', () => {
  it('generates valid iCal with VEVENTs', () => {
    const hits = [
      {
        transit_planet: 'Jupiter',
        aspect: 'trine',
        natal_planet: 'Sun',
        date: '2026-08-08',
        start_date: '2026-08-01',
        end_date: '2026-08-15',
        orb: 0.5,
      },
      {
        transit_planet: 'Saturn',
        aspect: 'square',
        natal_planet: 'Moon',
        date: '2026-08-17',
        start_date: '2026-08-10',
        end_date: '2026-08-25',
        orb: 1.2,
      },
    ];

    const ical = generateICal(hits, 'Test Chart');

    // Basic iCal structure
    expect(ical).toContain('BEGIN:VCALENDAR');
    expect(ical).toContain('END:VCALENDAR');
    expect(ical).toContain('VERSION:2.0');
    expect(ical).toContain('PRODID:-//Empirical Astrology//Transit Calendar//EN');

    // Events
    expect(ical).toContain('BEGIN:VEVENT');
    expect(ical).toContain('END:VEVENT');

    // First event
    expect(ical).toContain('SUMMARY:Jupiter trine Sun');
    expect(ical).toContain('DTSTART;VALUE=DATE:20260801');
    expect(ical).toContain('DTEND;VALUE=DATE:20260816'); // end_date + 1
    expect(ical).toContain('DESCRIPTION:Jupiter trine Sun (orb: 0.5°)');

    // Second event
    expect(ical).toContain('SUMMARY:Saturn square Moon');
    expect(ical).toContain('DTSTART;VALUE=DATE:20260810');
  });

  it('handles empty hits array', () => {
    const ical = generateICal([], 'Empty');
    expect(ical).toContain('BEGIN:VCALENDAR');
    expect(ical).toContain('END:VCALENDAR');
    expect(ical).not.toContain('BEGIN:VEVENT');
  });

  it('falls back to date when start_date/end_date missing', () => {
    const hits = [
      {
        transit_planet: 'Mars',
        aspect: 'conjunction',
        natal_planet: 'ASC',
        date: '2026-01-03',
        orb: 0.1,
      },
    ];
    const ical = generateICal(hits, 'Test');
    expect(ical).toContain('DTSTART;VALUE=DATE:20260103');
    expect(ical).toContain('DTEND;VALUE=DATE:20260104'); // date + 1
  });

  it('includes chart name in X-WR-CALNAME', () => {
    const ical = generateICal([], 'AJ Natal');
    expect(ical).toContain('X-WR-CALNAME:Transits for AJ Natal');
  });
});
