import { useState, useRef, useEffect } from 'react';
import type { BirthData, SavedChart } from '../../lib/types';
import { chartDB } from '../../lib/db';

interface ChartFormProps {
  initialData?: BirthData;
  initialName?: string;
  onSave: (chart: SavedChart) => void;
  onCancel: () => void;
}

interface CityResult {
  name: string;
  country: string;
  lat: number;
  lon: number;
  tz_offset: number;
}

export function ChartForm({ initialData, initialName, onSave, onCancel }: ChartFormProps) {
  const [name, setName] = useState(initialName || '');
  const [year, setYear] = useState(initialData?.year || new Date().getFullYear());
  const [month, setMonth] = useState(initialData?.month || 1);
  const [day, setDay] = useState(initialData?.day || 1);
  const [hour, setHour] = useState(initialData?.hour || 12);
  const [minute, setMinute] = useState(initialData?.minute || 0);
  const [tzOffset, setTzOffset] = useState(initialData?.tz_offset || 0);
  const [lat, setLat] = useState(initialData?.lat || 0);
  const [lng, setLng] = useState(initialData?.lng || 0);
  const [tags, setTags] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  // City search state
  const [cityQuery, setCityQuery] = useState('');
  const [cityResults, setCityResults] = useState<CityResult[]>([]);
  const [cityLoading, setCityLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [selectedCity, setSelectedCity] = useState<CityResult | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const searchTimeout = useRef<ReturnType<typeof setTimeout>>(undefined);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  async function searchCities(q: string) {
    setCityQuery(q);
    if (q.length < 2) {
      setCityResults([]);
      setShowDropdown(false);
      return;
    }
    // Debounce
    if (searchTimeout.current) clearTimeout(searchTimeout.current);
    searchTimeout.current = setTimeout(async () => {
      setCityLoading(true);
      try {
        const r = await fetch(`/api/geocode/search?q=${encodeURIComponent(q)}`);
        if (!r.ok) throw new Error('Search failed');
        const cities: CityResult[] = await r.json();
        setCityResults(cities);
        setShowDropdown(cities.length > 0);
      } catch {
        setCityResults([]);
      } finally {
        setCityLoading(false);
      }
    }, 200);
  }

  function selectCity(city: CityResult) {
    setSelectedCity(city);
    setCityQuery(`${city.name}, ${city.country}`);
    setLat(city.lat);
    setLng(city.lon);
    setTzOffset(city.tz_offset);
    setShowDropdown(false);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    setSaving(true);
    setError('');

    try {
      const birthData: BirthData = {
        name: name.trim(),
        year,
        month,
        day,
        hour,
        minute,
        tz_offset: tzOffset,
        lat,
        lng,
      };

      const tagList = tags
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean);

      const id = await chartDB.createFromBirthData(name.trim(), birthData, tagList);
      const chart = await chartDB.getById(id);
      if (chart) onSave(chart);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  }

  const inputClass =
    'w-full bg-bg border border-border rounded px-3 py-1.5 text-sm text-text focus:border-accent focus:outline-none';

  return (
    <div className="max-w-lg mx-auto p-6">
      <h2 className="text-xl font-bold text-accent mb-6">
        {initialData ? 'Edit Chart' : 'New Chart'}
      </h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red/10 border border-red/30 rounded px-3 py-2 text-sm text-red">
            {error}
          </div>
        )}

        {/* Name */}
        <div>
          <label className="block text-xs text-muted mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={inputClass}
            placeholder="e.g. AJ"
            autoFocus
          />
        </div>

        {/* Date */}
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="block text-xs text-muted mb-1">Year</label>
            <input
              type="number"
              value={year}
              onChange={(e) => setYear(Number(e.target.value))}
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Month</label>
            <input
              type="number"
              value={month}
              onChange={(e) => setMonth(Number(e.target.value))}
              min={1}
              max={12}
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Day</label>
            <input
              type="number"
              value={day}
              onChange={(e) => setDay(Number(e.target.value))}
              min={1}
              max={31}
              className={inputClass}
            />
          </div>
        </div>

        {/* Time */}
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="block text-xs text-muted mb-1">Hour (0-23)</label>
            <input
              type="number"
              value={hour}
              onChange={(e) => setHour(Number(e.target.value))}
              min={0}
              max={23}
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Minute</label>
            <input
              type="number"
              value={minute}
              onChange={(e) => setMinute(Number(e.target.value))}
              min={0}
              max={59}
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">TZ Offset (UTC)</label>
            <input
              type="number"
              value={tzOffset}
              onChange={(e) => setTzOffset(Number(e.target.value))}
              step={0.5}
              className={inputClass}
            />
            {selectedCity && (
              <p className="text-xs text-muted mt-0.5">
                Auto-set from {selectedCity.name}
              </p>
            )}
          </div>
        </div>

        {/* City Search */}
        <div ref={dropdownRef} className="relative">
          <label className="block text-xs text-muted mb-1">
            Birth Place (search city)
          </label>
          <input
            type="text"
            value={cityQuery}
            onChange={(e) => searchCities(e.target.value)}
            className={inputClass}
            placeholder="Type a city name..."
          />
          {cityLoading && (
            <span className="absolute right-3 top-7 text-xs text-muted">Searching...</span>
          )}
          {showDropdown && cityResults.length > 0 && (
            <div className="absolute z-10 w-full mt-1 bg-surface border border-border rounded shadow-lg max-h-48 overflow-y-auto">
              {cityResults.map((city, i) => (
                <button
                  key={i}
                  type="button"
                  onClick={() => selectCity(city)}
                  className="w-full text-left px-3 py-2 text-sm text-text hover:bg-bg flex justify-between items-center"
                >
                  <span>
                    {city.name}, {city.country}
                  </span>
                  <span className="text-xs text-muted">
                    {city.lat.toFixed(2)}, {city.lon.toFixed(2)} · UTC{city.tz_offset >= 0 ? '+' : ''}{city.tz_offset}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Location (manual override) */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs text-muted mb-1">Latitude</label>
            <input
              type="number"
              value={lat}
              onChange={(e) => setLat(Number(e.target.value))}
              step={0.001}
              className={inputClass}
              placeholder="47.038"
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Longitude</label>
            <input
              type="number"
              value={lng}
              onChange={(e) => setLng(Number(e.target.value))}
              step={0.001}
              className={inputClass}
              placeholder="-122.901"
            />
          </div>
        </div>

        {/* Tags */}
        <div>
          <label className="block text-xs text-muted mb-1">Tags (comma-separated)</label>
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            className={inputClass}
            placeholder="client, family, research"
          />
        </div>

        {/* Buttons */}
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={saving}
            className="flex-1 bg-accent text-white rounded py-2 text-sm font-semibold hover:opacity-90 disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Chart'}
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 bg-surface border border-border rounded py-2 text-sm text-muted hover:text-text"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
