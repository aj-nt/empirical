import { useState, useEffect } from 'react';
import type { BirthData, InterpretationResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
  system?: 'western' | 'koiné';
}

export function TextManager({ data, system = 'western' }: Props) {
  const [interp, setInterp] = useState<InterpretationResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState<string | null>(null);
  const [editText, setEditText] = useState('');
  const [customTexts, setCustomTexts] = useState<Record<string, string>>({});

  useEffect(() => {
    setLoading(true);
    setError('');
    api.interpretation(data, system, 3)
      .then(setInterp)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));

    // Load custom texts from localStorage
    try {
      const stored = localStorage.getItem(`empirical-texts-${system}`);
      if (stored) setCustomTexts(JSON.parse(stored));
    } catch {}
  }, [data.name, system]);

  const saveCustomText = (key: string, text: string) => {
    const updated = { ...customTexts, [key]: text };
    setCustomTexts(updated);
    localStorage.setItem(`empirical-texts-${system}`, JSON.stringify(updated));
    setEditing(null);
  };

  const deleteCustomText = (key: string) => {
    const updated = { ...customTexts };
    delete updated[key];
    setCustomTexts(updated);
    localStorage.setItem(`empirical-texts-${system}`, JSON.stringify(updated));
  };

  if (loading) return <p className="text-yellow text-sm p-4">Loading...</p>;
  if (error) return <p className="text-red text-sm p-4">{error}</p>;
  if (!interp) return <p className="text-muted text-sm p-4">No data available.</p>;

  const categories: { name: string; items: string[] }[] = [
    { name: 'Planets in Signs', items: interp.planet_signs },
    { name: 'Planets in Houses', items: interp.planet_houses },
    { name: 'Aspects', items: interp.aspects },
    { name: 'Patterns', items: interp.patterns || [] },
    { name: 'Fixed Stars', items: interp.stars || [] },
  ];

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="text-xs text-muted mb-2">
        Customize interpretation texts. Changes are saved locally and override defaults.
        <button
          onClick={() => {
            localStorage.removeItem(`empirical-texts-${system}`);
            setCustomTexts({});
          }}
          className="ml-2 text-red hover:underline"
        >
          Reset all
        </button>
      </div>

      {categories.map(cat => (
        <div key={cat.name} className="border border-border rounded-lg overflow-hidden">
          <div className="px-3 py-1.5 bg-surface text-sm font-medium">
            {cat.name} ({cat.items.length})
          </div>
          <div className="divide-y divide-border">
            {cat.items.map((text, i) => {
              const key = `${cat.name}:${i}`;
              const isCustom = key in customTexts;
              const displayText = customTexts[key] || text;
              const isEditing = editing === key;

              return (
                <div key={i} className="px-3 py-2">
                  {isEditing ? (
                    <div className="space-y-2">
                      <textarea
                        value={editText}
                        onChange={e => setEditText(e.target.value)}
                        className="w-full bg-bg border border-border rounded p-2 text-sm text-text resize-y"
                        rows={3}
                      />
                      <div className="flex gap-2">
                        <button
                          onClick={() => saveCustomText(key, editText)}
                          className="px-2 py-0.5 text-xs bg-accent text-white rounded"
                        >
                          Save
                        </button>
                        <button
                          onClick={() => setEditing(null)}
                          className="px-2 py-0.5 text-xs bg-surface text-muted border border-border rounded"
                        >
                          Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div>
                      <p className={`text-sm leading-relaxed ${isCustom ? 'text-accent' : 'text-text'}`}>
                        {displayText}
                      </p>
                      <div className="flex gap-2 mt-1">
                        <button
                          onClick={() => { setEditing(key); setEditText(displayText); }}
                          className="text-xs text-muted hover:text-accent"
                        >
                          {isCustom ? 'Edit' : 'Customize'}
                        </button>
                        {isCustom && (
                          <button
                            onClick={() => deleteCustomText(key)}
                            className="text-xs text-red hover:underline"
                          >
                            Revert to default
                          </button>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
