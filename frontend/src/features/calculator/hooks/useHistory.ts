import { useCallback } from 'react';

import { isRecord } from '../../../shared/api/guards';
import { useStoredState } from '../../../shared/hooks/useStoredState';
import { HISTORY_STORAGE_KEY, MAX_HISTORY_ENTRIES } from '../constants';
import type { HistoryEntry } from '../types';

/** What {@link useHistory} returns. */
export interface HistoryControls {
  /** Finished calculations, newest first. */
  entries: HistoryEntry[];
  /** Records a calculation at the top of the history. */
  addEntry: (entry: Omit<HistoryEntry, 'id'>) => void;
  /** Removes every entry. */
  clearHistory: () => void;
}

/**
 * The calculation history of this browser tab. It lives only on the client, in
 * session storage, so it survives a reload but not closing the tab, and the
 * backend stays stateless. At most {@link MAX_HISTORY_ENTRIES} are kept.
 *
 * @returns The entries and functions to add and clear them.
 */
export function useHistory(): HistoryControls {
  const [entries, setEntries] = useStoredState<HistoryEntry[]>(
    'session',
    HISTORY_STORAGE_KEY,
    [],
    isHistoryEntries,
  );

  const addEntry = useCallback(
    (entry: Omit<HistoryEntry, 'id'>) => {
      setEntries((previous) =>
        [{ id: createId(), ...entry }, ...previous].slice(0, MAX_HISTORY_ENTRIES),
      );
    },
    [setEntries],
  );

  const clearHistory = useCallback(() => {
    setEntries([]);
  }, [setEntries]);

  return { entries, addEntry, clearHistory };
}

/**
 * A unique id for list rendering. `crypto.randomUUID` is not used because
 * browsers only offer it on secure origins (HTTPS or localhost).
 */
function createId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function isHistoryEntries(value: unknown): value is HistoryEntry[] {
  return (
    Array.isArray(value) &&
    value.every(
      (entry) =>
        isRecord(entry) &&
        typeof entry.id === 'string' &&
        typeof entry.expression === 'string' &&
        typeof entry.result === 'string',
    )
  );
}
