import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { setSystemDarkMode } from '../../mocks/matchMedia';
import { THEME_STORAGE_KEY } from '../constants/ui';
import { useTheme } from './useTheme';

describe('useTheme', () => {
  it('follows the system theme until the user picks one', () => {
    setSystemDarkMode(true);
    const { result } = renderHook(() => useTheme());

    expect(result.current.theme).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');

    act(() => {
      setSystemDarkMode(false);
    });

    expect(result.current.theme).toBe('light');
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('toggles the theme, remembers the choice and stops following the system', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('light');

    act(() => {
      result.current.toggleTheme();
    });
    act(() => {
      setSystemDarkMode(false);
    });

    expect(result.current.theme).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(window.localStorage.getItem(THEME_STORAGE_KEY)).toBe('"dark"');
  });

  it('restores a stored choice', () => {
    window.localStorage.setItem(THEME_STORAGE_KEY, '"light"');
    setSystemDarkMode(true);

    const { result } = renderHook(() => useTheme());

    expect(result.current.theme).toBe('light');
  });
});
