import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useStoredState } from './useStoredState';

const KEY = 'test.value';

function isNumber(value: unknown): value is number {
  return typeof value === 'number';
}

describe('useStoredState', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('starts from the fallback when nothing is stored', () => {
    const { result } = renderHook(() => useStoredState('local', KEY, 7, isNumber));

    expect(result.current[0]).toBe(7);
  });

  it('restores a valid stored value and saves changes', () => {
    window.localStorage.setItem(KEY, '3');
    const { result } = renderHook(() => useStoredState('local', KEY, 7, isNumber));
    expect(result.current[0]).toBe(3);

    act(() => {
      result.current[1](5);
    });

    expect(window.localStorage.getItem(KEY)).toBe('5');
  });

  it('uses session storage when asked to', () => {
    const { result } = renderHook(() => useStoredState('session', KEY, 1, isNumber));

    act(() => {
      result.current[1](2);
    });

    expect(window.sessionStorage.getItem(KEY)).toBe('2');
    expect(window.localStorage.getItem(KEY)).toBeNull();
  });

  it.each([
    ['not JSON', '{oops'],
    ['the wrong type', '"three"'],
  ])('ignores a stored value that is %s', (_, raw) => {
    window.localStorage.setItem(KEY, raw);

    const { result } = renderHook(() => useStoredState('local', KEY, 7, isNumber));

    expect(result.current[0]).toBe(7);
  });

  it('keeps working when storage is unavailable', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('full', 'QuotaExceededError');
    });

    const { result } = renderHook(() => useStoredState('local', KEY, 7, isNumber));
    act(() => {
      result.current[1](8);
    });

    expect(result.current[0]).toBe(8);
  });
});
