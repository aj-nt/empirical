import type { Aspect } from '../../lib/types';
import { planetGlyph, aspectColor } from '../../lib/astrology';

interface AspectGridProps {
  aspects: Aspect[];
  planets1: string[];
  planets2: string[];
}

function aspectAbbrev(aspect: string): string {
  const abbrevs: Record<string, string> = {
    conjunction: 'C',
    opposition: 'O',
    square: 'Sq',
    trine: 'Tr',
    sextile: 'Se',
    quincunx: 'Q',
    'semi-sextile': 'Ss',
    'semi-square': 'Ssq',
    sesquiquadrate: 'Sesq',
    quintile: 'Qu',
    'bi-quintile': 'Bq',
  };
  return abbrevs[aspect] || aspect.slice(0, 2);
}

export default function AspectGrid({ aspects, planets1, planets2 }: AspectGridProps) {
  // Build lookup map: "planet1|planet2" → Aspect
  const aspectMap = new Map<string, Aspect>();
  for (const a of aspects) {
    aspectMap.set(`${a.planet1}|${a.planet2}`, a);
  }

  return (
    <div className="overflow-x-auto">
      <table
        style={{
          borderCollapse: 'collapse',
          color: '#e0e0e0',
        }}
      >
        <thead>
          <tr>
            <th
              style={{
                border: '1px solid #333',
                padding: '4px 8px',
                fontSize: '0.75rem',
              }}
            ></th>
            {planets2.map((p2) => (
              <th
                key={p2}
                title={p2}
                style={{
                  border: '1px solid #333',
                  padding: '4px 8px',
                  fontSize: '0.85rem',
                  textAlign: 'center',
                }}
              >
                {planetGlyph(p2)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {planets1.map((p1) => (
            <tr key={p1}>
              <td
                title={p1}
                style={{
                  border: '1px solid #333',
                  padding: '4px 8px',
                  fontSize: '0.85rem',
                  textAlign: 'center',
                  fontWeight: 'bold',
                }}
              >
                {planetGlyph(p1)}
              </td>
              {planets2.map((p2) => {
                const aspect = aspectMap.get(`${p1}|${p2}`);
                if (!aspect) {
                  return (
                    <td
                      key={`${p1}|${p2}`}
                      style={{
                        border: '1px solid #333',
                        padding: '4px 8px',
                        textAlign: 'center',
                        fontSize: '0.7rem',
                      }}
                    ></td>
                  );
                }
                const color = aspectColor(aspect.aspect);
                return (
                  <td
                    key={`${p1}|${p2}`}
                    title={`${aspect.planet1} ${aspect.aspect} ${aspect.planet2} — orb ${aspect.orb}°`}
                    style={{
                      border: '1px solid #333',
                      padding: '4px 8px',
                      textAlign: 'center',
                      fontSize: '0.7rem',
                      backgroundColor: color + '33',
                    }}
                  >
                    {aspectAbbrev(aspect.aspect)}{aspect.orb}°
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}