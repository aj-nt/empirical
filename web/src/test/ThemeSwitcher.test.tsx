import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeProvider, ThemeSwitcher } from '../lib/theme';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

Object.defineProperty(window, 'matchMedia', {
  value: vi.fn().mockImplementation(() => ({
    matches: false,
    media: '',
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe('ThemeSwitcher', () => {
  beforeEach(() => {
    localStorageMock.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('renders all 4 theme buttons', () => {
    render(
      <ThemeProvider>
        <ThemeSwitcher />
      </ThemeProvider>,
    );
    expect(screen.getByTitle('Dark')).toBeInTheDocument();
    expect(screen.getByTitle('Light')).toBeInTheDocument();
    expect(screen.getByTitle('Sepia')).toBeInTheDocument();
    expect(screen.getByTitle('High Contrast')).toBeInTheDocument();
  });

  it('highlights the active theme', () => {
    render(
      <ThemeProvider>
        <ThemeSwitcher />
      </ThemeProvider>,
    );
    const darkBtn = screen.getByTitle('Dark');
    expect(darkBtn.className).toContain('bg-accent');
  });

  it('switches theme on click', () => {
    render(
      <ThemeProvider>
        <ThemeSwitcher />
      </ThemeProvider>,
    );
    const lightBtn = screen.getByTitle('Light');
    fireEvent.click(lightBtn);

    // Light button should now be active
    expect(lightBtn.className).toContain('bg-accent');
    // Dark button should no longer be active
    const darkBtn = screen.getByTitle('Dark');
    expect(darkBtn.className).not.toContain('bg-accent');
  });

  it('persists theme to localStorage on click', () => {
    render(
      <ThemeProvider>
        <ThemeSwitcher />
      </ThemeProvider>,
    );
    fireEvent.click(screen.getByTitle('Sepia'));
    expect(localStorageMock.setItem).toHaveBeenCalledWith('empirical-theme', 'sepia');
  });

  it('sets data-theme attribute on document', () => {
    render(
      <ThemeProvider>
        <ThemeSwitcher />
      </ThemeProvider>,
    );
    fireEvent.click(screen.getByTitle('High Contrast'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('high-contrast');
  });
});
