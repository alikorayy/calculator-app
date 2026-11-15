import { useEffect, useState, type Dispatch, type SetStateAction } from 'react';

/** Which browser storage to use: per tab (`session`) or permanent (`local`). */
export type StorageKind = 'session' | 'local';

/**
 * State that is saved in browser storage and restored on the next load.
 *
 * Storage is optional comfort, never a requirement: it can be unavailable
 * (private windows, blocked site data) or contain anything, so every access is
 * guarded and invalid stored values are ignored in favour of `fallback`.
 *
 * @param kind - Session storage (cleared when the tab closes) or local storage.
 * @param key - Storage key.
 * @param fallback - Value used when nothing valid is stored.
 * @param isValid - Type guard that checks a stored value before it is used.
 * @returns The value and its setter, like `useState`.
 */
export function useStoredState<T>(
  kind: StorageKind,
  key: string,
  fallback: T,
  isValid: (value: unknown) => value is T,
): [T, Dispatch<SetStateAction<T>>] {
  const [value, setValue] = useState<T>(() => readStoredValue(kind, key, fallback, isValid));

  useEffect(() => {
    writeStoredValue(kind, key, value);
  }, [kind, key, value]);

  return [value, setValue];
}

function getStorage(kind: StorageKind): Storage | undefined {
  try {
    return kind === 'session' ? window.sessionStorage : window.localStorage;
  } catch {
    return undefined;
  }
}

function readStoredValue<T>(
  kind: StorageKind,
  key: string,
  fallback: T,
  isValid: (value: unknown) => value is T,
): T {
  try {
    const raw = getStorage(kind)?.getItem(key);
    if (raw === null || raw === undefined) {
      return fallback;
    }
    const parsed: unknown = JSON.parse(raw);
    return isValid(parsed) ? parsed : fallback;
  } catch {
    return fallback;
  }
}

function writeStoredValue(kind: StorageKind, key: string, value: unknown): void {
  try {
    getStorage(kind)?.setItem(key, JSON.stringify(value));
  } catch {
    // Full or blocked storage only costs the saved value, not the app.
  }
}
