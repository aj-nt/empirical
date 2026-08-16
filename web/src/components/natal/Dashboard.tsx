import type { InterpretationResponse, TraditionalResponse } from '../../lib/types';

interface DashboardProps {
  interp: InterpretationResponse;
  trad: TraditionalResponse | null;
}

const ELEMENT_EMOJI: Record<string, string> = {
  Fire: '🔥', Air: '🌬️', Water: '🌊', Earth: '⛰️',
};
const MODALITY_EMOJI: Record<string, string> = {
  Cardinal: '🟡', Fixed: '🟠', Mutable: '🔵',
};
const HOUSE_TYPE_EMOJI: Record<string, string> = {
  Angular: '🏠', Succedent: '🟢', Cadent: '🟣',
};

function parseElementBalance(interp: InterpretationResponse): Record<string, number> {
  const counts: Record<string, number> = { Fire: 0, Air: 0, Water: 0, Earth: 0 };
  for (const s of interp.planet_signs ?? []) {
    for (const el of Object.keys(counts)) {
      if (s.includes(el.toLowerCase())) {
        counts[el]++;
        break;
      }
    }
  }
  return counts;
}

function parseModalityBalance(interp: InterpretationResponse): Record<string, number> {
  const counts: Record<string, number> = { Cardinal: 0, Fixed: 0, Mutable: 0 };
  for (const s of interp.planet_signs ?? []) {
    for (const mod of Object.keys(counts)) {
      if (s.includes(mod.toLowerCase())) {
        counts[mod]++;
        break;
      }
    }
  }
  return counts;
}

function parseHouseBalance(interp: InterpretationResponse): Record<string, number> {
  const counts: Record<string, number> = { Angular: 0, Succedent: 0, Cadent: 0 };
  for (const s of interp.planet_houses ?? []) {
    const m = s.match(/in (?:the )?(\d+)(?:st|nd|rd|th)? house/);
    if (!m) continue;
    const h = parseInt(m[1]);
    if (h === 1 || h === 4 || h === 7 || h === 10) counts.Angular++;
    else if (h === 2 || h === 5 || h === 8 || h === 11) counts.Succedent++;
    else counts.Cadent++;
  }
  return counts;
}

export function Dashboard({ interp, trad }: DashboardProps) {
  const elements = parseElementBalance(interp);
  const modalities = parseModalityBalance(interp);
  const houses = parseHouseBalance(interp);
  const retroCount = (trad?.retrogrades ?? []).filter(r => r.retrograde).length;
  const lunarPhase = trad?.lunar_phase;

  return (
    <div className="bg-surface border border-border rounded-lg p-3">
      <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-muted">
        {/* Elements */}
        <span>
          {Object.entries(elements).map(([el, count]) => (
            <span key={el} className="mr-2" title={el}>
              {ELEMENT_EMOJI[el] ?? ''} {count}
            </span>
          ))}
        </span>
        <span className="text-border">|</span>
        {/* Modalities */}
        <span>
          {Object.entries(modalities).map(([mod, count]) => (
            <span key={mod} className="mr-2" title={mod}>
              {MODALITY_EMOJI[mod] ?? ''} {count}
            </span>
          ))}
        </span>
        <span className="text-border">|</span>
        {/* House types */}
        <span>
          {Object.entries(houses).map(([ht, count]) => (
            <span key={ht} className="mr-2" title={ht}>
              {HOUSE_TYPE_EMOJI[ht] ?? ''} {count}
            </span>
          ))}
        </span>
        {retroCount > 0 && (
          <>
            <span className="text-border">|</span>
            <span className="text-purple">℞ {retroCount} retrograde</span>
          </>
        )}
        {lunarPhase && (
          <>
            <span className="text-border">|</span>
            <span>🌙 {lunarPhase.name} ({lunarPhase.phase_index}/8)</span>
          </>
        )}
      </div>
    </div>
  );
}
