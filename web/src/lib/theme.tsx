import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';

export type Theme = 'dark' | 'light' | 'sepia' | 'high-contrast';

const THEME_KEY = 'empirical-theme';
const THEMES: { id: Theme; label: string; icon: string }[] = [
  { id: 'dark', label: 'Dark', icon: '🌙' },
  { id: 'light', label: 'Light', icon: '☀️' },
  { id: 'sepia', label: 'Sepia', icon: '📜' },
  { id: 'high-contrast', label: 'High Contrast', icon: '🔲' },
];

function getStoredTheme(): Theme {
  try {
    const stored = localStorage.getItem(THEME_KEY);
    if (stored && THEMES.some((t) => t.id === stored)) return stored as Theme;
  } catch { /* localStorage unavailable */ }
  return 'dark';
}

interface ThemeContextValue {
  theme: Theme;
  setTheme: (t: Theme) => void;
  themes: typeof THEMES;
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'dark',
  setTheme: () => {},
  themes: THEMES,
});

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    try { localStorage.setItem(THEME_KEY, t); } catch { /* ok */ }
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme, themes: THEMES }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  return useContext(ThemeContext);
}

export function ThemeSwitcher() {
  const { theme, setTheme, themes } = useTheme();

  return (
    <div className="flex items-center gap-1">
      {themes.map((t) => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          title={t.label}
          className={`px-2 py-1 text-xs rounded ${
            theme === t.id
              ? 'bg-accent text-white'
              : 'bg-surface text-muted border border-border hover:text-text'
          }`}
        >
          {t.icon}
        </button>
      ))}
    </div>
  );
}
