import { useState, useEffect } from 'react';
import type { BirthData, ComparisonReport } from '../../lib/types';
import { api } from '../../lib/api';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

export function CrossSystemComparison({ data }: { data: BirthData }) {
  const [report, setReport] = useState<ComparisonReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setReport(null);
    setError(null);

    api
      .compare(data)
      .then((result) => {
        if (!cancelled) setReport(result);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      });

    return () => {
      cancelled = true;
    };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (error) {
    return <p className="text-red text-sm">{error}</p>;
  }

  if (!report) {
    return <p className="text-yellow text-sm">Loading cross-system comparison…</p>;
  }

  const systems = report.systems;
  const displayNames: Record<string, string> = {
    koine: 'Koiné',
    western: 'Western',
    vedic: 'Vedic',
  };

  return (
    <div className="space-y-4">
      {/* Planet Signs */}
      <Section title="Planet Signs by System">
        <div className="grid grid-cols-3 gap-4">
          {systems.map((sys) => (
            <div key={sys}>
              <h4 className="text-sm font-bold text-text mb-1">{displayNames[sys] ?? sys}</h4>
              <ul className="text-sm text-text space-y-0.5">
                {report.planet_signs.map((item) => (
                  <li key={item.planet}>
                    {item.planet}: {item.systems[sys] ?? '—'}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </Section>

      {/* Planet Houses */}
      <Section title="Planet Houses by System">
        <div className="grid grid-cols-3 gap-4">
          {systems.map((sys) => (
            <div key={sys}>
              <h4 className="text-sm font-bold text-text mb-1">{displayNames[sys] ?? sys}</h4>
              <ul className="text-sm text-text space-y-0.5">
                {report.planet_houses.map((item) => (
                  <li key={item.planet}>
                    {item.planet}: House {item.systems[sys] ?? '—'}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </Section>

      {/* Dignity */}
      <Section title="Dignity by System">
        <div className="grid grid-cols-3 gap-4">
          {systems.map((sys) => (
            <div key={sys}>
              <h4 className="text-sm font-bold text-text mb-1">{displayNames[sys] ?? sys}</h4>
              <ul className="text-sm text-text space-y-0.5">
                {report.dignity_comparison.map((item) => (
                  <li key={item.planet}>
                    {item.planet}: {item.systems[sys] ?? '—'}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </Section>

      {/* Agreement Summary */}
      <Section title="System Agreement">
        <div className="grid grid-cols-3 gap-4 text-sm text-text">
          <div>
            <span className="text-muted">Sign agreement:</span>{' '}
            <span className="font-semibold">{(report.summary.sign_agreement * 100).toFixed(0)}%</span>
          </div>
          <div>
            <span className="text-muted">House agreement:</span>{' '}
            <span className="font-semibold">{(report.summary.house_agreement * 100).toFixed(0)}%</span>
          </div>
          <div>
            <span className="text-muted">Dignity agreement:</span>{' '}
            <span className="font-semibold">{(report.summary.dignity_agreement * 100).toFixed(0)}%</span>
          </div>
        </div>
      </Section>
    </div>
  );
}
