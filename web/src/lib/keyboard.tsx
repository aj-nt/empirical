import { useEffect, useState, useCallback } from 'react';

type View = 'wheel' | 'biwheel' | 'triwheel' | 'natal' | 'transits' | 'synastry' | 'maps' | 'reports' | 'ephemeris' | 'calendar' | 'research' | 'interpretation' | 'predictive' | 'form';

const TAB_ORDER: View[] = [
  'wheel', 'biwheel', 'triwheel', 'natal', 'transits', 'synastry',
  'maps', 'reports', 'ephemeris', 'calendar', 'research', 'interpretation', 'predictive',
];

interface ShortcutHandlers {
  onNewChart: () => void;
  onSearch: () => void;
  onPrint: () => void;
  onSettings: () => void;
  onTabSwitch: (view: View) => void;
  currentView: View;
}

export function useKeyboardShortcuts(handlers: ShortcutHandlers) {
  const [showHelp, setShowHelp] = useState(false);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Don't intercept when typing in inputs
    const target = e.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT') {
      return;
    }

    const mod = e.metaKey || e.ctrlKey;

    // Ctrl+N: New chart
    if (mod && e.key === 'n') {
      e.preventDefault();
      handlers.onNewChart();
      return;
    }

    // Ctrl+S: Focus search
    if (mod && e.key === 's') {
      e.preventDefault();
      handlers.onSearch();
      return;
    }

    // Ctrl+P: Print
    if (mod && e.key === 'p') {
      e.preventDefault();
      handlers.onPrint();
      return;
    }

    // Ctrl+,: Settings
    if (mod && e.key === ',') {
      e.preventDefault();
      handlers.onSettings();
      return;
    }

    // ?: Show help
    if (e.key === '?' && !mod) {
      e.preventDefault();
      setShowHelp((prev) => !prev);
      return;
    }

    // Escape: Close help
    if (e.key === 'Escape' && showHelp) {
      setShowHelp(false);
      return;
    }

    // 1-9: Switch tabs
    const num = parseInt(e.key);
    if (num >= 1 && num <= 9 && !mod) {
      e.preventDefault();
      const idx = num - 1;
      if (idx < TAB_ORDER.length) {
        handlers.onTabSwitch(TAB_ORDER[idx]);
      }
      return;
    }
  }, [handlers, showHelp]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return { showHelp, setShowHelp };
}

export function KeyboardHelp({ onClose }: { onClose: () => void }) {
  const shortcuts = [
    { keys: 'Ctrl+N', desc: 'New chart' },
    { keys: 'Ctrl+S', desc: 'Focus search' },
    { keys: 'Ctrl+P', desc: 'Print current view' },
    { keys: 'Ctrl+,', desc: 'Settings' },
    { keys: '1–9', desc: 'Switch tabs' },
    { keys: '?', desc: 'Show/hide this help' },
    { keys: 'Esc', desc: 'Close help / cancel' },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <h2 className="text-sm font-semibold text-text">Keyboard Shortcuts</h2>
          <button onClick={onClose} className="text-muted hover:text-text text-lg leading-none">✕</button>
        </div>
        <div className="p-4 space-y-2">
          {shortcuts.map((s) => (
            <div key={s.keys} className="flex justify-between text-sm">
              <kbd className="px-2 py-0.5 rounded bg-bg border border-border text-xs text-accent font-mono">
                {s.keys}
              </kbd>
              <span className="text-muted">{s.desc}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
