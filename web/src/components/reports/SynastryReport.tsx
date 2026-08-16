import { useState, useEffect } from 'react';
import type { BirthData, SynastryResponse, CompositeResponse, DraconicSynastryFullResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { planetGlyph, signGlyph } from '../../lib/astrology';

interface Props {
  chartA: BirthData;
  chartB: BirthData;
  nameA: string;
  nameB: string;
}

const SIGNS = ['Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo', 'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces'];

export function SynastryReport({ chartA, chartB, nameA, nameB }: Props) {
  const [synastry, setSynastry] = useState<SynastryResponse | null>(null);
  const [composite, setComposite] = useState<CompositeResponse | null>(null);
  const [draconic, setDraconic] = useState<DraconicSynastryFullResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    Promise.all([
      api.synastry(chartA, chartB, 3),
      api.composite(chartA, chartB),
      api.draconicSynastryFull(chartA, chartB, 3),
    ])
      .then(([s, c, d]) => {
        setSynastry(s);
        setComposite(c);
        setDraconic(d);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : 'Failed to load');
        setLoading(false);
      });
  }, [chartA.name, chartB.name]);

  if (loading) return <p className="text-yellow text-sm">Loading report...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!synastry || !composite) return null;

  const topAspects = [...synastry.aspects]
    .sort((a, b) => a.orb - b.orb)
    .slice(0, 10);

  const compositePlanets = Object.entries(composite.planets)
    .map(([name, lon]) => ({ name, lon }))
    .sort((a, b) => a.lon - b.lon);

  // Count draconic bridges: aspects that appear in all three layers
  const dracBridges = draconic
    ? draconic.drac_to_drac.filter((a) => a.orb < 2)
    : [];

  return (
    <div className="print-area space-y-6 text-sm">
      {/* Header */}
      <div className="text-center border-b border-border pb-4 avoid-break">
        <h1 className="text-xl font-bold text-text">Relationship Report</h1>
        <p className="text-muted mt-1">
          {nameA} & {nameB}
        </p>
        <p className="text-xs text-muted mt-1">
          Generated {new Date().toLocaleDateString()}
        </p>
      </div>

      {/* Synastry Aspects */}
      <div className="avoid-break">
        <h2 className="text-base font-semibold text-accent mb-2">Synastry Aspects</h2>
        <p className="text-xs text-muted mb-2">
          {synastry.aspects.length} inter-aspects between the two charts
        </p>
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left py-1 text-muted">Aspect</th>
              <th className="text-left py-1 text-muted">Planet A</th>
              <th className="text-left py-1 text-muted">Planet B</th>
              <th className="text-right py-1 text-muted">Orb</th>
            </tr>
          </thead>
          <tbody>
            {synastry.aspects.map((a, i) => (
              <tr key={i} className="border-b border-border/50">
                <td className="py-1">{a.aspect}</td>
                <td className="py-1">{planetGlyph(a.planet1)} {a.planet1}</td>
                <td className="py-1">{planetGlyph(a.planet2)} {a.planet2}</td>
                <td className="py-1 text-right">{a.orb.toFixed(1)}°</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Top 10 Closest Aspects */}
      <div className="avoid-break page-break">
        <h2 className="text-base font-semibold text-accent mb-2">Strongest Connections</h2>
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left py-1 text-muted">#</th>
              <th className="text-left py-1 text-muted">Aspect</th>
              <th className="text-left py-1 text-muted">Connection</th>
              <th className="text-right py-1 text-muted">Orb</th>
            </tr>
          </thead>
          <tbody>
            {topAspects.map((a, i) => (
              <tr key={i} className="border-b border-border/50">
                <td className="py-1 text-muted">{i + 1}</td>
                <td className="py-1">{a.aspect}</td>
                <td className="py-1">
                  {planetGlyph(a.planet1)} {a.planet1} – {planetGlyph(a.planet2)} {a.planet2}
                </td>
                <td className="py-1 text-right font-mono">{a.orb.toFixed(2)}°</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Composite Chart */}
      <div className="avoid-break page-break">
        <h2 className="text-base font-semibold text-accent mb-2">Composite Chart</h2>
        <p className="text-xs text-muted mb-2">
          Midpoint positions of the relationship
        </p>
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left py-1 text-muted">Planet</th>
              <th className="text-right py-1 text-muted">Longitude</th>
              <th className="text-left py-1 text-muted">Sign</th>
            </tr>
          </thead>
          <tbody>
            {compositePlanets.map((p) => {
              const signIdx = Math.floor(p.lon / 30);
              const signLon = p.lon % 30;
              return (
                <tr key={p.name} className="border-b border-border/50">
                  <td className="py-1">
                    {planetGlyph(p.name)} {p.name}
                  </td>
                  <td className="py-1 text-right font-mono">{p.lon.toFixed(2)}°</td>
                  <td className="py-1">
                    {signGlyph(SIGNS[signIdx])} {SIGNS[signIdx]} {signLon.toFixed(1)}°
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Draconic Bridges */}
      {dracBridges.length > 0 && (
        <div className="avoid-break page-break">
          <h2 className="text-base font-semibold text-accent mb-2">Draconic Bridges</h2>
          <p className="text-xs text-muted mb-2">
            {dracBridges.length} close draconic-to-draconic aspects (orb &lt; 2°)
          </p>
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-1 text-muted">Aspect</th>
                <th className="text-left py-1 text-muted">Connection</th>
                <th className="text-right py-1 text-muted">Orb</th>
              </tr>
            </thead>
            <tbody>
              {dracBridges.map((a, i) => (
                <tr key={i} className="border-b border-border/50">
                  <td className="py-1">{a.aspect}</td>
                  <td className="py-1">
                    {planetGlyph(a.planet1)} {a.planet1} – {planetGlyph(a.planet2)} {a.planet2}
                  </td>
                  <td className="py-1 text-right font-mono">{a.orb.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Print Button */}
      <div className="no-print flex justify-center pt-4">
        <button
          onClick={() => window.print()}
          className="px-4 py-2 text-sm rounded bg-accent text-white hover:opacity-90 print-keep"
        >
          Print Report
        </button>
      </div>
    </div>
  );
}
