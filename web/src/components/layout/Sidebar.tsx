import { useState, useEffect, useRef } from 'react';
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
  const [tagFilter, setTagFilter] = useState('');
  const [editingTag, setEditingTag] = useState<number | null>(null);
  const [tagInput, setTagInput] = useState('');
  const [selectMode, setSelectMode] = useState(false);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [bulkTagInput, setBulkTagInput] = useState('');
  const [showBulkTag, setShowBulkTag] = useState(false);
  const [showImport, setShowImport] = useState(false);
  const [importJson, setImportJson] = useState('');
  const tagInputRef = useRef<HTMLInputElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadCharts();
  }, []);

  useEffect(() => {
    if (editingTag !== null && tagInputRef.current) {
      tagInputRef.current.focus();
    }
  }, [editingTag]);

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

  const allTags = [...new Set(charts.flatMap((c) => c.tags))].sort();

  const filteredCharts = tagFilter
    ? charts.filter((c) => c.tags.includes(tagFilter))
    : charts;

  async function addTag(chart: SavedChart, tag: string) {
    const trimmed = tag.trim();
    if (!trimmed || chart.tags.includes(trimmed)) return;
    const newTags = [...chart.tags, trimmed];
    if (chart.id) {
      await chartDB.update(chart.id, { tags: newTags });
      loadCharts();
    }
  }

  async function removeTag(chart: SavedChart, tag: string) {
    if (!chart.id) return;
    const newTags = chart.tags.filter((t) => t !== tag);
    await chartDB.update(chart.id, { tags: newTags });
    loadCharts();
  }

  function handleTagKeyDown(e: React.KeyboardEvent, chart: SavedChart) {
    if (e.key === 'Enter') {
      addTag(chart, tagInput);
      setTagInput('');
      setEditingTag(null);
    } else if (e.key === 'Escape') {
      setTagInput('');
      setEditingTag(null);
    }
  }

  function toggleSelect(id: number) {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  }

  function toggleSelectAll() {
    if (selected.size === filteredCharts.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(filteredCharts.map((c) => c.id!)));
    }
  }

  async function bulkDelete() {
    if (selected.size === 0) return;
    if (!confirm(`Delete ${selected.size} chart(s)? This cannot be undone.`)) return;
    for (const id of selected) {
      await chartDB.remove(id);
      if (activeChart?.id === id) onDeleteChart(id);
    }
    setSelected(new Set());
    setSelectMode(false);
    loadCharts();
  }

  async function bulkAddTag() {
    const tag = bulkTagInput.trim();
    if (!tag || selected.size === 0) return;
    for (const id of selected) {
      const chart = charts.find((c) => c.id === id);
      if (chart && !chart.tags.includes(tag)) {
        await chartDB.update(id, { tags: [...chart.tags, tag] });
      }
    }
    setBulkTagInput('');
    setShowBulkTag(false);
    loadCharts();
  }

  async function duplicateChart(chart: SavedChart) {
    if (!chart.id) return;
    const { id, createdAt, updatedAt, ...rest } = chart;
    await chartDB.add({
      ...rest,
      name: `${rest.name} (copy)`,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    });
    loadCharts();
  }

  async function handleExport() {
    const json = await chartDB.exportAll();
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `empirical-charts-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  async function handleImport() {
    try {
      const count = await chartDB.importAll(importJson);
      alert(`Imported ${count} chart(s).`);
      setImportJson('');
      setShowImport(false);
      loadCharts();
    } catch (e) {
      alert('Invalid JSON file.');
    }
  }

  function handleFileImport(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      setImportJson(reader.result as string);
      setShowImport(true);
    };
    reader.readAsText(file);
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

      {/* Tag Filter */}
      {allTags.length > 0 && (
        <div className="px-3 pb-2">
          <select
            value={tagFilter}
            onChange={(e) => setTagFilter(e.target.value)}
            className="w-full bg-bg border border-border rounded px-2 py-1 text-xs text-muted focus:border-accent focus:outline-none"
          >
            <option value="">All tags</option>
            {allTags.map((tag) => (
              <option key={tag} value={tag}>{tag}</option>
            ))}
          </select>
        </div>
      )}

      {/* Bulk Actions Bar */}
      {selectMode && (
        <div className="px-3 py-2 border-b border-border space-y-1">
          <div className="flex items-center gap-2">
            <button onClick={toggleSelectAll} className="text-xs text-muted hover:text-text">
              {selected.size === filteredCharts.length ? 'Deselect all' : 'Select all'}
            </button>
            <span className="text-xs text-muted">{selected.size} selected</span>
          </div>
          <div className="flex gap-1">
            <button
              onClick={bulkDelete}
              disabled={selected.size === 0}
              className="flex-1 px-2 py-1 text-xs rounded bg-red/20 text-red hover:bg-red/30 disabled:opacity-30"
            >
              Delete
            </button>
            <button
              onClick={() => setShowBulkTag(!showBulkTag)}
              disabled={selected.size === 0}
              className="flex-1 px-2 py-1 text-xs rounded bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30"
            >
              Tag
            </button>
            <button
              onClick={() => { setSelectMode(false); setSelected(new Set()); }}
              className="px-2 py-1 text-xs rounded bg-bg text-muted border border-border hover:text-text"
            >
              Done
            </button>
          </div>
          {showBulkTag && (
            <div className="flex gap-1">
              <input
                type="text"
                value={bulkTagInput}
                onChange={(e) => setBulkTagInput(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') bulkAddTag(); }}
                placeholder="tag name..."
                className="flex-1 bg-bg border border-border rounded px-2 py-1 text-xs text-text focus:border-accent focus:outline-none"
              />
              <button
                onClick={bulkAddTag}
                className="px-2 py-1 text-xs rounded bg-accent text-white hover:opacity-90"
              >
                Add
              </button>
            </div>
          )}
        </div>
      )}

      {/* Chart List */}
      <div className="flex-1 overflow-y-auto px-2">
        {filteredCharts.map((chart) => (
          <div
            key={chart.id}
            className={`px-3 py-2 rounded cursor-pointer mb-1 text-sm ${
              activeChart?.id === chart.id && !selectMode
                ? 'bg-accent/20 text-accent'
                : 'hover:bg-surface text-text'
            }`}
          >
            <div className="flex items-center gap-2">
              {selectMode && (
                <input
                  type="checkbox"
                  checked={selected.has(chart.id!)}
                  onChange={() => toggleSelect(chart.id!)}
                  onClick={(e) => e.stopPropagation()}
                  className="shrink-0"
                />
              )}
              <span
                className="text-lg shrink-0"
                onClick={() => { if (!selectMode) onSelectChart(chart); }}
              >
                {planetGlyph('Sun')}
              </span>
              <div
                className="flex-1 min-w-0"
                onClick={() => { if (!selectMode) onSelectChart(chart); }}
              >
                <div className="truncate font-medium">{chart.name}</div>
                <div className="text-xs text-muted truncate">
                  {chart.birthData.year}-{String(chart.birthData.month).padStart(2, '0')}-
                  {String(chart.birthData.day).padStart(2, '0')}
                </div>
              </div>
              {!selectMode && (
                <div className="flex items-center gap-0.5 shrink-0">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      duplicateChart(chart);
                    }}
                    className="text-muted hover:text-accent text-xs px-0.5"
                    title="Duplicate"
                  >
                    ⧉
                  </button>
                  {activeChart?.id === chart.id && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        if (chart.id) onDeleteChart(chart.id);
                      }}
                      className="text-muted hover:text-red text-xs px-0.5"
                      title="Delete"
                    >
                      ✕
                    </button>
                  )}
                </div>
              )}
            </div>

            {/* Tags */}
            {!selectMode && (
              <div className="flex flex-wrap gap-1 mt-1" onClick={(e) => e.stopPropagation()}>
                {chart.tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-xs bg-accent/10 text-accent"
                  >
                    {tag}
                    <button
                      onClick={() => removeTag(chart, tag)}
                      className="hover:text-red leading-none"
                      title={`Remove tag "${tag}"`}
                    >
                      ×
                    </button>
                  </span>
                ))}
                {editingTag === chart.id ? (
                  <input
                    ref={tagInputRef}
                    type="text"
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyDown={(e) => handleTagKeyDown(e, chart)}
                    onBlur={() => {
                      if (tagInput.trim()) addTag(chart, tagInput);
                      setTagInput('');
                      setEditingTag(null);
                    }}
                    placeholder="tag..."
                    className="w-16 bg-bg border border-border rounded px-1 py-0 text-xs text-text focus:border-accent focus:outline-none"
                  />
                ) : (
                  <button
                    onClick={() => {
                      setEditingTag(chart.id!);
                      setTagInput('');
                    }}
                    className="text-muted hover:text-accent text-xs px-1"
                    title="Add tag"
                  >
                    + tag
                  </button>
                )}
              </div>
            )}
          </div>
        ))}
        {filteredCharts.length === 0 && (
          <p className="text-muted text-sm text-center py-8">
            {tagFilter ? `No charts tagged "${tagFilter}"` : 'No charts yet. Click + to add one.'}
          </p>
        )}
      </div>

      {/* Import Modal */}
      {showImport && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setShowImport(false)}>
          <div className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-md mx-4 p-4" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-sm font-semibold text-text mb-2">Import Charts</h3>
            <textarea
              value={importJson}
              onChange={(e) => setImportJson(e.target.value)}
              placeholder="Paste JSON here..."
              rows={8}
              className="w-full bg-bg border border-border rounded px-3 py-2 text-xs text-text font-mono focus:border-accent focus:outline-none mb-3"
            />
            <div className="flex gap-2">
              <button onClick={handleImport} className="flex-1 px-3 py-1.5 text-sm rounded bg-accent text-white hover:opacity-90">Import</button>
              <button onClick={() => setShowImport(false)} className="flex-1 px-3 py-1.5 text-sm rounded bg-bg text-muted border border-border hover:text-text">Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Theme Switcher */}
      <div className="px-3 py-2 border-t border-border">
        <ThemeSwitcher />
      </div>

      {/* Bottom Actions */}
      <div className="p-3 border-t border-border space-y-2">
        <button
          onClick={onNewChart}
          className="w-full bg-accent text-white rounded py-2 text-sm font-semibold hover:opacity-90"
        >
          + New Chart
        </button>
        <div className="flex gap-1">
          <button
            onClick={() => setSelectMode(!selectMode)}
            className={`flex-1 px-2 py-1 text-xs rounded border ${
              selectMode
                ? 'bg-accent/20 text-accent border-accent'
                : 'bg-bg text-muted border-border hover:text-text'
            }`}
          >
            {selectMode ? 'Selecting' : 'Select'}
          </button>
          <button
            onClick={handleExport}
            className="flex-1 px-2 py-1 text-xs rounded bg-bg text-muted border border-border hover:text-text"
          >
            Export
          </button>
          <button
            onClick={() => fileInputRef.current?.click()}
            className="flex-1 px-2 py-1 text-xs rounded bg-bg text-muted border border-border hover:text-text"
          >
            Import
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json"
            onChange={handleFileImport}
            className="hidden"
          />
        </div>
      </div>
    </aside>
  );
}
