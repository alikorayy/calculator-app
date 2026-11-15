import { useEffect, useEffectEvent } from 'react';

import type { KeyId } from '../types';
import { keyForKeyboardKey } from '../utils/display';

/**
 * Lets the physical keyboard operate the calculator: digits, operators,
 * parentheses, Enter or "=", Backspace and Escape.
 *
 * The same rules apply whether or not the display input has focus: text typed
 * at the end of the input goes through the calculator keys, exactly like the
 * keypad. Editing inside the text (caret in the middle, a selection, pasting) is
 * left to the input, so free editing keeps working.
 *
 * Keys are also left alone when a modifier is held (so shortcuts like Ctrl+C
 * keep working), while an input method is composing text, when another handler
 * already dealt with the key, and when Enter activates a focused button.
 *
 * @param onKey - Called with the calculator key for each relevant key press.
 * @param enabled - Turn off while something else, such as a dialog, has the keyboard.
 */
export function useKeyboard(onKey: (key: KeyId) => void, enabled = true): void {
  // useEffectEvent always calls the latest onKey without re-subscribing the
  // listener on every render.
  const handleKey = useEffectEvent((event: KeyboardEvent) => {
    if (
      event.defaultPrevented ||
      event.isComposing ||
      event.ctrlKey ||
      event.metaKey ||
      event.altKey ||
      isEditingInsideText(event.target)
    ) {
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

/**
 * Whether the key is meant to edit text that the browser should handle itself:
 * the caret is not at the very end of a text field, or part of the text is selected.
 * With the caret at the end, typing only appends, which the calculator keys do too.
 */
function isEditingInsideText(target: EventTarget | null): boolean {
  if (target instanceof HTMLTextAreaElement) {
    return true;
  }
  if (!(target instanceof HTMLInputElement)) {
    return false;
  }
  const { selectionStart, selectionEnd, value } = target;
  return selectionStart !== value.length || selectionEnd !== value.length;
}
