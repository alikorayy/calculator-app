import { useEffect } from 'react';

import { useStoredState } from './useStoredState';

/**
 * Controls a hint that is shown on the user's first visit only. The hint stays
 * until the first interaction anywhere on the page (a click, tap or key press),
 * and that is remembered in local storage, so it never comes back.
 *
 * @param storageKey - Local storage key that remembers that the hint was seen.
 * @returns Whether the hint should be shown right now.
 */
export function useFirstVisitHint(storageKey: string): { visible: boolean } {
  const [seen, setSeen] = useStoredState<boolean>('local', storageKey, false, isBoolean);

  useEffect(() => {
    if (seen) {
      return;
    }
    const dismiss = () => {
      setSeen(true);
    };
    // Capture phase, so the hint goes away even if a component stops the event.
    window.addEventListener('pointerdown', dismiss, { capture: true });
    window.addEventListener('keydown', dismiss, { capture: true });
    return () => {
      window.removeEventListener('pointerdown', dismiss, { capture: true });
      window.removeEventListener('keydown', dismiss, { capture: true });
    };
  }, [seen, setSeen]);

  return { visible: !seen };
}

function isBoolean(value: unknown): value is boolean {
  return typeof value === 'boolean';
}
