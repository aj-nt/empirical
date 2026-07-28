import { useState, useEffect } from 'react';
import type { SavedChart } from '../../lib/types';
import { chartDB } from '../../lib/db';
import { planetGlyph } from '../../lib/astrology';
import { ThemeSwitcher } from '../../lib/theme';

interface SidebarProps {
  activeChart: SavedChart | null;
  onSelectChart: (chart: SavedChart) => void;
  onNewChart: () => void;
  onDeleteChart: (id: number) => void;
}

export function Sidebar({ activeChart, onSelectChart, onNewChart, onDeleteChart }: SidebarProps) {
  const [charts, setCharts] = useState<SavedChart[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    loadCharts();
  }, []);

  async function loadCharts() {
    const all = await chartDB.getAll();
    setCharts(all.sort((a: SavedChart, b: SavedChart) => b.updatedAt.localeCompare(a.updatedAt)));
  }

  async function handleSearch(q: string) {
    setSearch(q);
    if (!q.trim()) {
      loadCharts();
    } else {
      const results = await chartDB.search(q);
      setCharts(results);
    }
  }

  return (
    <aside className="w-64 bg-surface border-r border-border flex flex-col h-screen shrink-0">
      {/* Header */}
      <div className="p-4 border-b border-border">
        <h1 className="text-lg font-bold text-accent">Empirical</h1>
        <p className="text-xs text-muted">Astrology</p>
      </div>

      {/* Search */}
      <div className="p-3">
        <input
          type="text"
          placeholder="Search charts..."
          value={search}
          onChange={(e) => handleSearch(e.target.value)}
          className="w-full bg-bg border border-border rounded px-3 py-1.5 text-sm text-text placeholder-muted focus:border-accent focus:outline-none"
        />
      </div>

      {/* Chart List */}
      <div className="flex-1 overflow-y-auto px-2">
        {charts.map((chart) => (
          <div
            key={chart.id}
            onClick={() => onSelectChart(chart)}
            className={`px-3 py-2 rounded cursor-pointer mb-1 text-sm flex items-center gap-2 ${
              activeChart?.id === chart.id
                ? 'bg-accent/20 text-accent'
                : 'hover:bg-surface text-text'
            }`}
          >
            <span className="text-lg">{planetGlyph('Sun')}</span>
            <div className="flex-1 min-w-0">
              <div className="truncate font-medium">{chart.name}</div>
              <div className="text-xs text-muted truncate">
                {chart.birthData.year}-{String(chart.birthData.month).padStart(2, '0')}-
                {String(chart.birthData.day).padStart(2, '0')}
              </div>
            </div>
            {activeChart?.id === chart.id && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  if (chart.id) onDeleteChart(chart.id);
                }}
                className="text-muted hover:text-red text-xs px-1"
                title="Delete"
              >
                ✕
              </button>
            )}
          </div>
        ))}
        {charts.length === 0 && (
          <p className="text-muted text-sm text-center py-8">
            No charts yet. Click + to add one.
          </p>
        )}
      </div>

      {/* Theme Switcher */}
      <div className="px-3 py-2 border-t border-border">
        <ThemeSwitcher />
      </div>

      {/* New Chart Button */}
      <div className="p-3 border-t border-border">
        <button
          onClick={onNewChart}
          className="w-full bg-accent text-white rounded py-2 text-sm font-semibold hover:opacity-90"
        >
          + New Chart
        </button>
      </div>
    </aside>
  );
}
