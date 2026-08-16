import { useState, useEffect } from 'react';
import type { BirthData, InterpretationResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { ElementBar } from './ElementBar';
import { ModalityBar } from './ModalityBar';
import { TransitInterpretation } from './TransitInterpretation';
import { TextManager } from './TextManager';

interface Props {
  data: BirthData;
  system?: 'western' | 'koiné';
}

export function InterpretationPanel({ data, system = 'western' }: Props) {
  const [interp, setInterp] = useState<InterpretationResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [subTab, setSubTab] = useState<'natal' | 'transits' | 'texts'>('natal');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.interpretation(data, system, 3)
      .then(setInterp)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng, system]);

  if (loading) return <p className="text-yellow text-sm p-4">Loading interpretation...</p>;
  if (error) return <p className="text-red text-sm p-4">{error}</p>;
  if (!interp) return <p className="text-muted text-sm p-4">No interpretation available.</p>;

  return (
    <div className="flex flex-col h-full">
      {/* Sub-tab bar */}
      <div className="flex gap-1 px-4 py-2 border-b border-border shrink-0">
        {(['natal', 'transits', 'texts'] as const).map(tab => (
          <button
            key={tab}
            onClick={() => setSubTab(tab)}
            className={`px-3 py-1 text-xs rounded ${
              subTab === tab ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'
            }`}
          >
            {tab === 'natal' ? 'Natal' : tab === 'transits' ? 'Transits' : 'Texts'}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {subTab === 'transits' ? (
          <TransitInterpretation data={data} />
        ) : subTab === 'texts' ? (
          <TextManager data={data} system={system} />
        ) : (
          <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      {/* Chart Overview */}
      <Section title="Chart Overview" defaultOpen={true}>
        <div className="space-y-1 text-sm">
          {interp.sect && <p>{interp.sect}</p>}
          {interp.lunar_phase && (
            <p>Lunar Phase: {interp.lunar_phase} ({interp.lunar_phase_angle?.toFixed(1)}°)</p>
          )}
          {interp.chart_ruler && (
            <p>
              Chart Ruler: <strong>{interp.chart_ruler}</strong>
              {interp.chart_ruler_traditional && interp.chart_ruler_traditional !== interp.chart_ruler && (
                <> (traditional: {interp.chart_ruler_traditional})</>
              )}
              {interp.chart_ruler_sign && <> in {interp.chart_ruler_sign}</>}
              {interp.chart_ruler_house && <> — House {interp.chart_ruler_house}</>}
              {interp.chart_ruler_dignity && <> ({interp.chart_ruler_dignity})</>}
            </p>
          )}
          {interp.final_dispositor && (
            <p>Final Dispositor: <strong>{interp.final_dispositor}</strong>
              {interp.final_dispositor_traditional && interp.final_dispositor_traditional !== interp.final_dispositor && (
                <> (traditional: {interp.final_dispositor_traditional})</>
              )}
            </p>
          )}
          {interp.angular_planets && interp.angular_planets.length > 0 && (
            <p>Angular Planets: {interp.angular_planets.join(', ')}</p>
          )}
          {interp.voc_moon && <p>{interp.voc_moon}</p>}
        </div>
      </Section>

      {/* Planet in Sign */}
      <Section title="Planets in Signs" count={interp.planet_signs.length}>
        {interp.planet_signs.map((text, i) => (
          <p key={i} className="text-sm mb-2 leading-relaxed">{text}</p>
        ))}
      </Section>

      {/* Planet in House */}
      <Section title="Planets in Houses" count={interp.planet_houses.length}>
        {interp.planet_houses.map((text, i) => (
          <p key={i} className="text-sm mb-2 leading-relaxed">{text}</p>
        ))}
      </Section>

      {/* Weighted Aspects */}
      {interp.weighted_aspects && interp.weighted_aspects.length > 0 && (
        <Section title="Top Aspects" count={interp.weighted_aspects.length}>
          {interp.weighted_aspects.slice(0, 20).map((wa, i) => (
            <p key={i} className="text-sm mb-1">
              <span className="text-muted">{wa.planet1} {wa.aspect} {wa.planet2}</span>
              {' '}orb {wa.orb.toFixed(1)}° — weight {wa.weight.toFixed(1)}
            </p>
          ))}
        </Section>
      )}

      {/* All Aspects */}
      <Section title="All Aspects" count={interp.aspects.length}>
        {interp.aspects.map((text, i) => (
          <p key={i} className="text-sm mb-1 leading-relaxed">{text}</p>
        ))}
      </Section>

      {/* Patterns */}
      {interp.patterns && interp.patterns.length > 0 && (
        <Section title="Patterns" count={interp.patterns.length}>
          {interp.patterns.map((text, i) => (
            <p key={i} className="text-sm mb-2 leading-relaxed">{text}</p>
          ))}
        </Section>
      )}

      {/* Element Balance */}
      {interp.element_balance && (
        <Section title="Element Balance">
          <ElementBar balance={interp.element_balance} />
        </Section>
      )}

      {/* Modality Balance */}
      {interp.modality_balance && (
        <Section title="Modality Balance">
          <ModalityBar balance={interp.modality_balance} />
        </Section>
      )}

      {/* Hemisphere */}
      {interp.hemisphere && (
        <Section title="Hemisphere Emphasis">
          <div className="text-sm space-y-1">
            <p>Above horizon: {interp.hemisphere.above} planets</p>
            <p>Below horizon: {interp.hemisphere.below} planets</p>
            <p>Eastern (rising): {interp.hemisphere.east} planets</p>
            <p>Western (setting): {interp.hemisphere.west} planets</p>
          </div>
        </Section>
      )}

      {/* Rulership Chains */}
      {interp.rulership_chains && Object.keys(interp.rulership_chains).length > 0 && (
        <Section title="Rulership Chains">
          {Object.entries(interp.rulership_chains).map(([house, chain]) => (
            <p key={house} className="text-sm mb-1">
              <span className="text-muted">House {house}:</span>{' '}
              {chain.join(' → ')}
            </p>
          ))}
        </Section>
      )}

      {/* Dispositor Trees */}
      {interp.dispositor_trees && Object.keys(interp.dispositor_trees).length > 0 && (
        <Section title="Dispositor Trees">
          {Object.entries(interp.dispositor_trees).map(([planet, chain]) => (
            <p key={planet} className="text-sm mb-1">
              <span className="text-muted">{planet}:</span>{' '}
              {chain.join(' → ')}
            </p>
          ))}
        </Section>
      )}

      {/* Midpoints */}
      {interp.midpoints && interp.midpoints.length > 0 && (
        <Section title="Midpoints" count={interp.midpoints.length}>
          {interp.midpoints.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Key Midpoints */}
      {interp.key_midpoints && interp.key_midpoints.length > 0 && (
        <Section title="Key Midpoints (personal planets, ≤0.5°)" count={interp.key_midpoints.length}>
          {interp.key_midpoints.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Fixed Star Contacts */}
      {interp.stars && interp.stars.length > 0 && (
        <Section title="Fixed Star Contacts" count={interp.stars.length}>
          {interp.stars.map((text, i) => (
            <p key={i} className="text-sm mb-2 leading-relaxed">{text}</p>
          ))}
        </Section>
      )}

      {/* Key Star Aspects */}
      {interp.key_star_aspects && interp.key_star_aspects.length > 0 && (
        <Section title="Key Star Aspects (personal planets, ≤1°)" count={interp.key_star_aspects.length}>
          {interp.key_star_aspects.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Declinations */}
      {interp.declinations && interp.declinations.length > 0 && (
        <Section title="Declination Parallels" count={interp.declinations.length}>
          {interp.declinations.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Contraparallels */}
      {interp.contraparallels && interp.contraparallels.length > 0 && (
        <Section title="Contraparallels" count={interp.contraparallels.length}>
          {interp.contraparallels.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Antiscia */}
      {interp.antiscia && interp.antiscia.length > 0 && (
        <Section title="Antiscia" count={interp.antiscia.length}>
          {interp.antiscia.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Antiscia Contacts */}
      {interp.antiscia_contacts && interp.antiscia_contacts.length > 0 && (
        <Section title="Antiscia Contacts" count={interp.antiscia_contacts.length}>
          {interp.antiscia_contacts.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Mutual Receptions */}
      {interp.mutual_receptions && interp.mutual_receptions.length > 0 && (
        <Section title="Mutual Receptions" count={interp.mutual_receptions.length}>
          {interp.mutual_receptions.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Retrogrades */}
      {interp.retrogrades && interp.retrogrades.length > 0 && (
        <Section title="Retrogrades" count={interp.retrogrades.length}>
          {interp.retrogrades.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Decans */}
      {interp.decans && interp.decans.length > 0 && (
        <Section title="Decans (Faces)" count={interp.decans.length}>
          {interp.decans.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}

      {/* Terms */}
      {interp.terms && interp.terms.length > 0 && (
        <Section title="Egyptian Terms" count={interp.terms.length}>
          {interp.terms.map((text, i) => (
            <p key={i} className="text-sm mb-1">{text}</p>
          ))}
        </Section>
      )}
    </div>
        )}
      </div>
    </div>
  );
}

function Section({
  title,
  children,
  count,
  defaultOpen = false,
}: {
  title: string;
  children: React.ReactNode;
  count?: number;
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-4 py-2 bg-surface text-sm font-semibold hover:bg-bg-hover transition-colors"
      >
        <span>
          {title}
          {count !== undefined && <span className="text-muted ml-2 text-xs">({count})</span>}
        </span>
        <span className="text-muted text-xs">{open ? '▾' : '▸'}</span>
      </button>
      {open && <div className="px-4 py-3 border-t border-border">{children}</div>}
    </div>
  );
}
