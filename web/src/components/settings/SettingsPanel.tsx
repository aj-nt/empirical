import { useState } from 'react';
import type { UserPreferences, AspectSetDef } from '../../lib/types';
import { DEFAULT_PREFERENCES } from '../../lib/types';
import { HOUSE_SYSTEMS } from '../../lib/houseSystems';
import { AYANAMSAS } from '../../lib/ayanamsas';
import { AspectSetEditor, loadAspectSet } from './AspectSetEditor';

interface Props {
  preferences: UserPreferences;
  onSave: (prefs: UserPreferences) => void;
  onClose: () => void;
}

const STORAGE_KEY = 'empirical-preferences';

export function loadPreferences(): UserPreferences {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      return { ...DEFAULT_PREFERENCES, ...parsed };
    }
  } catch {}
  return { ...DEFAULT_PREFERENCES };
}

export function savePreferences(prefs: UserPreferences): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
}

export function SettingsPanel({ preferences, onSave, onClose }: Props) {
  const [prefs, setPrefs] = useState<UserPreferences>({ ...preferences });
  const [tab, setTab] = useState<'general' | 'aspects'>('general');
  const [aspectSet, setAspectSet] = useState<AspectSetDef>(loadAspectSet);

  const handleSave = () => {
    savePreferences(prefs);
    onSave(prefs);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-md mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <div className="flex gap-2">
            <button
              onClick={() => setTab('general')}
              className={`text-sm ${tab === 'general' ? 'text-accent font-semibold' : 'text-muted hover:text-text'}`}
            >
              General
            </button>
            <button
              onClick={() => setTab('aspects')}
              className={`text-sm ${tab === 'aspects' ? 'text-accent font-semibold' : 'text-muted hover:text-text'}`}
            >
              Aspects
            </button>
          </div>
          <button onClick={onClose} className="text-muted hover:text-text text-lg leading-none">
            ✕
          </button>
        </div>

        {tab === 'general' ? (
          <>
            <div className="p-4 space-y-4">
              {/* Default House System */}
              <div>
                <label className="block text-xs text-muted mb-1">Default House System</label>
                <select
                  value={prefs.defaultHouseSystem}
                  onChange={(e) => setPrefs({ ...prefs, defaultHouseSystem: e.target.value })}
                  className="w-full bg-bg border border-border rounded px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                >
                  {HOUSE_SYSTEMS.map((hs) => (
                    <option key={hs.value} value={hs.value}>{hs.label}</option>
                  ))}
                </select>
              </div>

              {/* Default Ayanamsa */}
              <div>
                <label className="block text-xs text-muted mb-1">Default Ayanamsa</label>
                <select
                  value={prefs.defaultAyanamsa}
                  onChange={(e) => setPrefs({ ...prefs, defaultAyanamsa: e.target.value })}
                  className="w-full bg-bg border border-border rounded px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                >
                  {AYANAMSAS.map((a) => (
                    <option key={a.value} value={a.value}>{a.label}</option>
                  ))}
                </select>
              </div>

              {/* Default Orb */}
              <div>
                <label className="block text-xs text-muted mb-1">Default Orb (°)</label>
                <input
                  type="number"
                  min={1}
                  max={10}
                  step={0.5}
                  value={prefs.defaultOrb}
                  onChange={(e) => setPrefs({ ...prefs, defaultOrb: parseFloat(e.target.value) || 3 })}
                  className="w-full bg-bg border border-border rounded px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                />
              </div>

              {/* Theme */}
              <div>
                <label className="block text-xs text-muted mb-1">Theme</label>
                <div className="flex gap-2">
                  <button
                    onClick={() => setPrefs({ ...prefs, theme: 'dark' })}
                    className={`flex-1 px-3 py-2 text-sm rounded border ${
                      prefs.theme === 'dark'
                        ? 'bg-accent text-white border-accent'
                        : 'bg-bg text-muted border-border hover:text-text'
                    }`}
                  >
                    🌙 Dark
                  </button>
                  <button
                    onClick={() => setPrefs({ ...prefs, theme: 'light' })}
                    className={`flex-1 px-3 py-2 text-sm rounded border ${
                      prefs.theme === 'light'
                        ? 'bg-accent text-white border-accent'
                        : 'bg-bg text-muted border-border hover:text-text'
                    }`}
                  >
                    ☀️ Light
                  </button>
                </div>
              </div>
            </div>

            <div className="flex gap-2 px-4 py-3 border-t border-border">
              <button
                onClick={handleSave}
                className="flex-1 px-4 py-2 text-sm rounded bg-accent text-white hover:opacity-90"
              >
                Save
              </button>
              <button
                onClick={onClose}
                className="flex-1 px-4 py-2 text-sm rounded bg-bg text-muted border border-border hover:text-text"
              >
                Cancel
              </button>
            </div>
          </>
        ) : (
          <div className="p-4">
            <AspectSetEditor
              aspectSet={aspectSet}
              onSave={(newSet) => {
                setAspectSet(newSet);
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
