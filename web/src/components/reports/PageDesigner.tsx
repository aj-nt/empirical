import { useState, useEffect, useCallback, useRef } from 'react';
import type { BirthData, InterpretationResponse, TraditionalResponse } from '../../lib/types';
import { api } from '../../lib/api';

// ── Block Types ──

type BlockType = 'wheel' | 'planet_table' | 'aspect_table' | 'text' | 'pattern_list';

interface PageBlock {
  id: string;
  type: BlockType;
  title: string;
  content: string; // for text blocks
}

interface PageDesign {
  name: string;
  blocks: PageBlock[];
}

const BLOCK_LABELS: Record<BlockType, string> = {
  wheel: 'Chart Wheel',
  planet_table: 'Planet Table',
  aspect_table: 'Aspect Table',
  text: 'Text Block',
  pattern_list: 'Pattern List',
};

let blockId = 0;
function newBlock(type: BlockType): PageBlock {
  return { id: `b${++blockId}`, type, title: BLOCK_LABELS[type], content: '' };
}

// ── Templates ──

const TEMPLATES: Record<string, PageDesign> = {
  'natal-brief': {
    name: 'Natal Brief',
    blocks: [
      { id: 't1', type: 'wheel', title: 'Chart Wheel', content: '' },
      { id: 't2', type: 'planet_table', title: 'Planets in Signs & Houses', content: '' },
      { id: 't3', type: 'aspect_table', title: 'Major Aspects', content: '' },
      { id: 't4', type: 'pattern_list', title: 'Aspect Patterns', content: '' },
    ],
  },
  'natal-full': {
    name: 'Natal Full Report',
    blocks: [
      { id: 't1', type: 'text', title: 'Introduction', content: 'Natal chart report for {name}, born {date} in {location}.' },
      { id: 't2', type: 'wheel', title: 'Chart Wheel', content: '' },
      { id: 't3', type: 'planet_table', title: 'Planets in Signs & Houses', content: '' },
      { id: 't4', type: 'aspect_table', title: 'Major Aspects', content: '' },
      { id: 't5', type: 'pattern_list', title: 'Aspect Patterns', content: '' },
      { id: 't6', type: 'text', title: 'Notes', content: 'Additional notes and observations.' },
    ],
  },
};

// ── Component ──

interface PageDesignerProps {
  data: BirthData;
}

export function PageDesigner({ data }: PageDesignerProps) {
  const [blocks, setBlocks] = useState<PageBlock[]>([]);
  const [designName, setDesignName] = useState('My Report');
  const [interp, setInterp] = useState<InterpretationResponse | null>(null);
  const [trad, setTrad] = useState<TraditionalResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [svg, setSvg] = useState('');
  const previewRef = useRef<HTMLDivElement>(null);

  // Load chart data
  useEffect(() => {
    setLoading(true);
    setError('');
    Promise.all([
      api.interpretation(data, 'western', 3),
      api.traditional(data),
      api.chart(data, { house_system: 'placidus', show_aspects: true, outer_planets: true, highlight_patterns: true, pattern_orb: 3 }),
    ])
      .then(([i, t, s]) => {
        setInterp(i);
        setTrad(t);
        setSvg(s);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : 'Failed to load');
        setLoading(false);
      });
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  // Load saved design
  useEffect(() => {
    try {
      const saved = localStorage.getItem(`page-design-${data.name}`);
      if (saved) {
        const design: PageDesign = JSON.parse(saved);
        setDesignName(design.name);
        setBlocks(design.blocks);
      } else {
        setBlocks(TEMPLATES['natal-brief'].blocks.map((b) => ({ ...b })));
      }
    } catch {
      setBlocks(TEMPLATES['natal-brief'].blocks.map((b) => ({ ...b })));
    }
  }, [data.name]);

  const addBlock = useCallback((type: BlockType) => {
    setBlocks((prev) => [...prev, newBlock(type)]);
  }, []);

  const removeBlock = useCallback((id: string) => {
    setBlocks((prev) => prev.filter((b) => b.id !== id));
  }, []);

  const moveBlock = useCallback((id: string, dir: -1 | 1) => {
    setBlocks((prev) => {
      const idx = prev.findIndex((b) => b.id === id);
      if (idx < 0) return prev;
      const newIdx = idx + dir;
      if (newIdx < 0 || newIdx >= prev.length) return prev;
      const next = [...prev];
      [next[idx], next[newIdx]] = [next[newIdx], next[idx]];
      return next;
    });
  }, []);

  const updateBlock = useCallback((id: string, updates: Partial<PageBlock>) => {
    setBlocks((prev) => prev.map((b) => (b.id === id ? { ...b, ...updates } : b)));
  }, []);

  const saveDesign = useCallback(() => {
    const design: PageDesign = { name: designName, blocks };
    localStorage.setItem(`page-design-${data.name}`, JSON.stringify(design));
  }, [data.name, designName, blocks]);

  const loadTemplate = useCallback((key: string) => {
    const tmpl = TEMPLATES[key];
    if (tmpl) {
      setDesignName(tmpl.name);
      setBlocks(tmpl.blocks.map((b) => ({ ...b })));
    }
  }, []);

  const exportHTML = useCallback(() => {
    if (!previewRef.current) return;
    const html = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>${designName} — ${data.name}</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 800px; margin: 0 auto; padding: 2rem; color: #1f2328; background: #fff; }
  h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
  h2 { font-size: 1.1rem; color: #656d76; margin-top: 2rem; margin-bottom: 0.5rem; border-bottom: 1px solid #d0d7de; padding-bottom: 0.25rem; }
  table { width: 100%; border-collapse: collapse; margin: 0.5rem 0; }
  th, td { text-align: left; padding: 0.25rem 0.5rem; border-bottom: 1px solid #d0d7de; font-size: 0.9rem; }
  th { color: #656d76; font-weight: 600; }
  .chart { text-align: center; margin: 1rem 0; }
  .chart svg { max-width: 400px; }
  .text-block { line-height: 1.6; margin: 0.5rem 0; }
  @media print { body { padding: 0; } }
</style>
</head>
<body>
<h1>${designName}</h1>
<p style="color:#656d76">${data.name} · ${data.year}-${String(data.month).padStart(2, '0')}-${String(data.day).padStart(2, '0')} · ${data.lat.toFixed(2)}°, ${data.lng.toFixed(2)}°</p>
${previewRef.current.innerHTML}
</body>
</html>`;
    const blob = new Blob([html], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${data.name}-${designName.toLowerCase().replace(/\s+/g, '-')}.html`;
    a.click();
    URL.revokeObjectURL(url);
  }, [designName, data]);

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;

  return (
    <div className="flex h-full gap-4">
      {/* Left: Editor */}
      <div className="w-72 shrink-0 overflow-y-auto space-y-3">
        <div>
          <input
            type="text"
            value={designName}
            onChange={(e) => setDesignName(e.target.value)}
            className="w-full px-2 py-1 text-sm bg-background border border-border rounded text-text"
            placeholder="Design name"
          />
        </div>

        {/* Templates */}
        <div className="bg-surface border border-border rounded-lg p-3">
          <h3 className="text-xs font-semibold text-muted mb-2">Templates</h3>
          <div className="flex flex-wrap gap-1">
            {Object.entries(TEMPLATES).map(([key, tmpl]) => (
              <button
                key={key}
                onClick={() => loadTemplate(key)}
                className="px-2 py-1 text-xs rounded bg-background text-muted border border-border hover:text-text"
              >
                {tmpl.name}
              </button>
            ))}
          </div>
        </div>

        {/* Add Blocks */}
        <div className="bg-surface border border-border rounded-lg p-3">
          <h3 className="text-xs font-semibold text-muted mb-2">Add Block</h3>
          <div className="flex flex-wrap gap-1">
            {(Object.entries(BLOCK_LABELS) as [BlockType, string][]).map(([type, label]) => (
              <button
                key={type}
                onClick={() => addBlock(type)}
                className="px-2 py-1 text-xs rounded bg-background text-muted border border-border hover:text-text"
              >
                + {label}
              </button>
            ))}
          </div>
        </div>

        {/* Block List */}
        <div className="bg-surface border border-border rounded-lg p-3">
          <h3 className="text-xs font-semibold text-muted mb-2">Layout ({blocks.length} blocks)</h3>
          <div className="space-y-1">
            {blocks.map((block, idx) => (
              <div key={block.id} className="flex items-center gap-1 bg-background rounded px-2 py-1 text-xs">
                <button
                  onClick={() => moveBlock(block.id, -1)}
                  disabled={idx === 0}
                  className="text-muted hover:text-text disabled:opacity-20"
                  title="Move up"
                >
                  ▲
                </button>
                <button
                  onClick={() => moveBlock(block.id, 1)}
                  disabled={idx === blocks.length - 1}
                  className="text-muted hover:text-text disabled:opacity-20"
                  title="Move down"
                >
                  ▼
                </button>
                <span className="flex-1 truncate text-text">{block.title}</span>
                <button
                  onClick={() => removeBlock(block.id)}
                  className="text-muted hover:text-red"
                  title="Remove"
                >
                  ✕
                </button>
              </div>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <button
            onClick={saveDesign}
            className="flex-1 px-3 py-1.5 text-sm rounded bg-accent text-white hover:opacity-90"
          >
            Save Design
          </button>
          <button
            onClick={exportHTML}
            className="flex-1 px-3 py-1.5 text-sm rounded bg-green text-white hover:opacity-90"
          >
            Export HTML
          </button>
        </div>
      </div>

      {/* Right: Preview */}
      <div className="flex-1 overflow-y-auto bg-white rounded-lg border border-border p-6" ref={previewRef}>
        <div style={{ color: '#1f2328', fontFamily: '-apple-system, BlinkMacSystemFont, sans-serif' }}>
          {blocks.map((block) => (
            <div key={block.id} className="mb-6">
              <h2 style={{ fontSize: '1.1rem', color: '#656d76', marginTop: 0, marginBottom: '0.5rem', borderBottom: '1px solid #d0d7de', paddingBottom: '0.25rem' }}>
                {block.type === 'text' ? (
                  <input
                    type="text"
                    value={block.title}
                    onChange={(e) => updateBlock(block.id, { title: e.target.value })}
                    className="bg-transparent border-none text-inherit font-semibold w-full"
                    style={{ color: '#656d76' }}
                  />
                ) : (
                  block.title
                )}
              </h2>
              {renderBlock(block, interp, trad, svg, updateBlock)}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function renderBlock(
  block: PageBlock,
  interp: InterpretationResponse | null,
  trad: TraditionalResponse | null,
  svg: string,
  updateBlock: (id: string, updates: Partial<PageBlock>) => void,
) {
  switch (block.type) {
    case 'wheel':
      return (
        <div className="chart" style={{ textAlign: 'center' }}>
          <div dangerouslySetInnerHTML={{ __html: svg }} />
        </div>
      );

    case 'planet_table':
      return (
        <table>
          <thead>
            <tr>
              <th>Planet</th>
              <th>Sign</th>
              <th>House</th>
              <th>Retrograde</th>
            </tr>
          </thead>
          <tbody>
            {interp?.planet_signs?.map((s, i) => {
              const parts = s.split(' in ');
              const planet = parts[0]?.trim() ?? '';
              const signHouse = parts[1] ?? '';
              const houseText = interp.planet_houses?.[i] ?? '';
              // Extract house number from "Planet in the Nth house: ..." or "Planet in house N: ..."
              const houseMatch = houseText.match(/in (?:the )?(\d+)(?:st|nd|rd|th)? house/i);
              const house = houseMatch ? houseMatch[1] : '—';
              const isRx = trad?.retrogrades?.find((r) => r.planet === planet)?.retrograde;
              return (
                <tr key={i}>
                  <td style={{ fontWeight: 600 }}>{planet}</td>
                  <td>{signHouse}</td>
                  <td>{house}</td>
                  <td>{isRx ? '℞' : ''}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      );

    case 'aspect_table':
      return (
        <table>
          <thead>
            <tr>
              <th>Planet A</th>
              <th>Aspect</th>
              <th>Planet B</th>
              <th>Orb</th>
            </tr>
          </thead>
          <tbody>
            {interp?.aspects?.map((a, i) => {
              // Parse "Sun conjunction Moon (0.5°)" format
              const match = a.match(/^(.+?)\s+(conjunction|opposition|trine|square|sextile|quincunx)\s+(.+?)\s+\((.+?)°\)$/);
              if (!match) return <tr key={i}><td colSpan={4}>{a}</td></tr>;
              return (
                <tr key={i}>
                  <td style={{ fontWeight: 600 }}>{match[1]}</td>
                  <td>{match[2]}</td>
                  <td>{match[3]}</td>
                  <td>{match[4]}°</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      );

    case 'pattern_list':
      return (
        <ul style={{ paddingLeft: '1.5rem', lineHeight: 1.8 }}>
          {interp?.patterns?.map((p, i) => (
            <li key={i}>{p}</li>
          ))}
        </ul>
      );

    case 'text':
      return (
        <textarea
          value={block.content}
          onChange={(e) => updateBlock(block.id, { content: e.target.value })}
          placeholder="Enter text..."
          className="text-block w-full bg-transparent border border-gray-200 rounded p-2 text-sm"
          style={{ minHeight: 80, lineHeight: 1.6, color: '#1f2328' }}
        />
      );

    default:
      return null;
  }
}
