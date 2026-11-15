import { useCallback, useEffect, useSyncExternalStore } from 'react';

import { DARK_SCHEME_QUERY, THEME_STORAGE_KEY } from '../constants/ui';
import { useStoredState } from './useStoredState';

/** A colour theme. */
export type Theme = 'dark' | 'light';

/** What {@link useTheme} returns. */
export interface ThemeControls {
  /** The theme in effect: the user's choice, or the system's when there is none. */
  theme: Theme;
  /** Switches to the other theme and remembers the choice. */
  toggleTheme: () => void;
}

/**
 * The application's colour theme. Until the user picks one, it follows the
 * operating system (and keeps following it when that changes); a pick is
 * stored in local storage. The theme is applied as `data-theme` on `<html>`,
 * which the design tokens read.
 *
 * @returns The current theme and a function to toggle it.
 */
export function useTheme(): ThemeControls {
  const [chosen, setChosen] = useStoredState<Theme | null>(
    'local',
    THEME_STORAGE_KEY,
    null,
    isThemeOrNull,
  );
  const system = useSystemTheme();
  const theme = chosen ?? system;

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  const toggleTheme = useCallback(() => {
    setChosen(theme === 'dark' ? 'light' : 'dark');
  }, [theme, setChosen]);

  return { theme, toggleTheme };
}

/**
 * The operating system's colour scheme, kept up to date. `matchMedia` is an
 * external source React does not control, so it is read with
 * `useSyncExternalStore`, which subscribes to its changes safely.
 */
function useSystemTheme(): Theme {
  return useSyncExternalStore(subscribeToSystemTheme, getSystemTheme);
}

function subscribeToSystemTheme(onChange: () => void): () => void {
  const query = window.matchMedia(DARK_SCHEME_QUERY);
  query.addEventListener('change', onChange);
  return () => {
    query.removeEventListener('change', onChange);
  };
}

function getSystemTheme(): Theme {
  return window.matchMedia(DARK_SCHEME_QUERY).matches ? 'dark' : 'light';
}

function isThemeOrNull(value: unknown): value is Theme | null {
  return value === null || value === 'dark' || value === 'light';
}
