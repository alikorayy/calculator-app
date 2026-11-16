import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { HISTORY_STORAGE_KEY, MAX_HISTORY_ENTRIES } from '../constants';
import { useHistory } from './useHistory';

const ENTRY = { expression: '50−7×3', result: '29', approximate: false, sentExpression: '50−7×3' };

describe('useHistory', () => {
  it('adds entries newest first, each with its own id', () => {
    const { result } = renderHook(() => useHistory());

    act(() => {
      result.current.addEntry(ENTRY);
      result.current.addEntry({ ...ENTRY, expression: '1÷3' });
    });

    expect(result.current.entries.map((entry) => entry.expression)).toEqual(['1÷3', '50−7×3']);
    expect(new Set(result.current.entries.map((entry) => entry.id)).size).toBe(2);
  });

  it(`keeps at most ${String(MAX_HISTORY_ENTRIES)} entries`, () => {
    const { result } = renderHook(() => useHistory());

    act(() => {
      for (let i = 0; i <= MAX_HISTORY_ENTRIES; i++) {
        result.current.addEntry({ ...ENTRY, expression: String(i) });
      }
    });

    expect(result.current.entries).toHaveLength(MAX_HISTORY_ENTRIES);
    expect(result.current.entries[0]?.expression).toBe(String(MAX_HISTORY_ENTRIES));
  });

  it('is kept in session storage, so it survives a reload of the tab', () => {
    const first = renderHook(() => useHistory());
    act(() => {
      first.result.current.addEntry(ENTRY);
    });
    first.unmount();

    const reloaded = renderHook(() => useHistory());

    expect(reloaded.result.current.entries).toMatchObject([ENTRY]);
    expect(window.sessionStorage.getItem(HISTORY_STORAGE_KEY)).toContain(ENTRY.expression);
  });

  it('starts empty when the stored history has an older shape', () => {
    window.sessionStorage.setItem(
      HISTORY_STORAGE_KEY,
      JSON.stringify([{ id: 'old', expression: '1+1', result: '2' }]),
    );

    const { result } = renderHook(() => useHistory());

    expect(result.current.entries).toEqual([]);
  });

  it('can be cleared', () => {
    const { result } = renderHook(() => useHistory());
    act(() => {
      result.current.addEntry(ENTRY);
    });

    act(() => {
      result.current.clearHistory();
    });

    expect(result.current.entries).toEqual([]);
  });

  it('ignores stored data that is not a valid history', () => {
    window.sessionStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify([{ expression: 1 }]));

    const { result } = renderHook(() => useHistory());

    expect(result.current.entries).toEqual([]);
  });
});
