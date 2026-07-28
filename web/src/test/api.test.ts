import { describe, it, expect, beforeEach, vi } from 'vitest';
import { api } from '../lib/api';
import type { BirthData } from '../lib/types';

const mockBirthData: BirthData = {
  name: 'Test',
  year: 1990,
  month: 6,
  day: 15,
  hour: 8,
  minute: 30,
  tz_offset: -4,
  lat: 40.7128,
  lng: -74.006,
};

describe('API client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('chart() returns SVG string', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve('<svg>test</svg>'),
    } as Response);

    const result = await api.chart(mockBirthData, { house_system: 'P' });
    expect(result).toBe('<svg>test</svg>');
    expect(fetch).toHaveBeenCalledWith(
      '/api/chart',
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('interpretation() returns structured data', async () => {
    const mockResponse = {
      planet_signs: ['Sun in Gemini'],
      planet_houses: ['Sun in the 9th house'],
      aspects: ['Sun conjunction Mercury (0.5°)'],
      patterns: ['Stellium in Gemini'],
    };
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockResponse),
    } as Response);

    const result = await api.interpretation(mockBirthData, 'western', 3);
    expect(result.planet_signs).toEqual(['Sun in Gemini']);
    expect(result.patterns).toEqual(['Stellium in Gemini']);
  });

  it('transits() returns transit data', async () => {
    const mockResponse = {
      name: 'Test',
      start_date: '2026-08-01',
      end_date: '2026-08-31',
      transits: [
        {
          transit_planet: 'Jupiter',
          aspect: 'trine',
          natal_planet: 'Sun',
          date: '2026-08-08',
          orb: 0.5,
        },
      ],
      sky_weather: [],
    };
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockResponse),
    } as Response);

    const result = await api.transits(mockBirthData, '2026-08-01', '2026-08-31');
    expect(result.transits).toHaveLength(1);
    expect(result.transits[0].transit_planet).toBe('Jupiter');
  });

  it('throws on non-OK response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: () => Promise.resolve('Internal Server Error'),
    } as Response);

    await expect(
      api.interpretation(mockBirthData, 'western', 3),
    ).rejects.toThrow('500: Internal Server Error');
  });

  it('throws on network error', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('Network error'));

    await expect(
      api.interpretation(mockBirthData, 'western', 3),
    ).rejects.toThrow('Network error');
  });

  it('researchMetrics() calls correct endpoint', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    } as Response);

    await api.researchMetrics(mockBirthData);
    expect(fetch).toHaveBeenCalledWith(
      '/api/research-metrics',
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('researchBaseline() calls correct endpoint', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    } as Response);

    await api.researchBaseline('cross_system_sign_agreement', 500, 0);
    expect(fetch).toHaveBeenCalledWith(
      '/api/research-baseline',
      expect.objectContaining({ method: 'POST' }),
    );
  });

  it('batchAnalysis() calls correct endpoint', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    } as Response);

    await api.batchAnalysis([mockBirthData]);
    expect(fetch).toHaveBeenCalledWith(
      '/api/batch-analysis',
      expect.objectContaining({ method: 'POST' }),
    );
  });
});
