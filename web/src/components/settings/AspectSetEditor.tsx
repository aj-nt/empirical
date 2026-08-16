import { useState } from 'react';
import { ASPECT_SET_PRESETS, ALL_ASPECT_TYPES, type AspectSetDef } from '../../lib/types';

interface Props {
  aspectSet: AspectSetDef;
  onSave: (set: AspectSetDef) => void;
}

const STORAGE_KEY = 'empirical-aspect-set';

export function loadAspectSet(): AspectSetDef {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) return JSON.parse(stored);
  } catch {}
  return { ...ASPECT_SET_PRESETS.modern };
}

export function saveAspectSet(set: AspectSetDef): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(set));
}

export function AspectSetEditor({ aspectSet, onSave }: Props) {
  const [set, setSet] = useState<AspectSetDef>({ ...aspectSet });
  const [preset, setPreset] = useState('custom');

  const applyPreset = (key: string) => {
    setPreset(key);
    setSet({ ...ASPECT_SET_PRESETS[key] });
  };

  const toggleAspect = (aspect: string) => {
    const next = { ...set.aspects };
    if (next[aspect] !== undefined) {
      delete next[aspect];
    } else {
      next[aspect] = aspect === 'Conjunction' || aspect === 'Opposition' ? 8 : 6;
    }
    setSet({ ...set, aspects: next });
    setPreset('custom');
  };

  const setOrb = (aspect: string, orb: number) => {
    setSet({ ...set, aspects: { ...set.aspects, [aspect]: orb } });
    setPreset('custom');
  };

  const handleSave = () => {
    saveAspectSet(set);
    onSave(set);
  };

  return (
    <div className="space-y-4">
      {/* Presets */}
      <div>
        <label className="block text-xs text-muted mb-1">Preset</label>
        <div className="flex gap-1">
          {Object.entries(ASPECT_SET_PRESETS).map(([key, val]) => (
            <button
              key={key}
              onClick={() => applyPreset(key)}
              className={`px-2 py-1 text-xs rounded ${
                preset === key
                  ? 'bg-accent text-white'
                  : 'bg-bg text-muted border border-border hover:text-text'
              }`}
            >
              {val.name}
            </button>
          ))}
          <button
            onClick={() => setPreset('custom')}
            className={`px-2 py-1 text-xs rounded ${
              preset === 'custom'
                ? 'bg-accent text-white'
                : 'bg-bg text-muted border border-border hover:text-text'
            }`}
          >
            Custom
          </button>
        </div>
      </div>

      {/* Aspect Toggles */}
      <div>
        <label className="block text-xs text-muted mb-1">Enabled Aspects</label>
        <div className="space-y-1 max-h-48 overflow-y-auto">
          {ALL_ASPECT_TYPES.map((aspect) => {
            const enabled = set.aspects[aspect] !== undefined;
            const orb = set.aspects[aspect] || 0;
            return (
              <div key={aspect} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={enabled}
                  onChange={() => toggleAspect(aspect)}
                  className="shrink-0"
                />
                <span className={`text-xs flex-1 ${enabled ? 'text-text' : 'text-muted'}`}>
                  {aspect}
                </span>
                {enabled && (
                  <input
                    type="number"
                    min={0.5}
                    max={15}
                    step={0.5}
                    value={orb}
                    onChange={(e) => setOrb(aspect, parseFloat(e.target.value) || 0)}
                    className="w-14 bg-bg border border-border rounded px-1 py-0.5 text-xs text-text text-right focus:border-accent focus:outline-none"
                  />
                )}
                {enabled && <span className="text-xs text-muted w-4">°</span>}
              </div>
            );
          })}
        </div>
      </div>

      <button
        onClick={handleSave}
        className="w-full px-3 py-1.5 text-sm rounded bg-accent text-white hover:opacity-90"
      >
        Apply Aspect Set
      </button>
    </div>
  );
}
