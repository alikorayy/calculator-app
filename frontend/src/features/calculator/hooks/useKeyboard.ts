import { useEffect, useEffectEvent } from 'react';

import type { KeyId } from '../types';
import { keyForKeyboardKey } from '../utils/display';

/**
 * Lets the physical keyboard operate the calculator: digits, operators,
 * parentheses, Enter or "=", Backspace and Escape.
 *
 * Keys are left alone when the user is typing in a text field (the input
 * handles its own editing), when a modifier is held (so shortcuts like Ctrl+C
 * keep working) and when Enter activates a focused button.
 *
 * @param onKey - Called with the calculator key for each relevant key press.
 * @param enabled - Turn off while something else, such as a dialog, has the keyboard.
 */
export function useKeyboard(onKey: (key: KeyId) => void, enabled = true): void {
  // useEffectEvent always calls the latest onKey without re-subscribing the
  // listener on every render.
  const handleKey = useEffectEvent((event: KeyboardEvent) => {
    if (event.ctrlKey || event.metaKey || event.altKey || isTypingTarget(event.target)) {
      return;
    }
    if (event.key === 'Enter' && event.target instanceof HTMLButtonElement) {
      return;
    }
    const key = keyForKeyboardKey(event.key);
    if (key) {
      event.preventDefault();
      onKey(key);
    }
  });

  useEffect(() => {
    if (!enabled) {
      return;
    }
    const listener = (event: KeyboardEvent) => {
      handleKey(event);
    };
    window.addEventListener('keydown', listener);
    return () => {
      window.removeEventListener('keydown', listener);
    };
  }, [enabled]);
}

function isTypingTarget(target: EventTarget | null): boolean {
  return target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement;
}
