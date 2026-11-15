import { DISPLAY_SIZE_THRESHOLDS, KEYBOARD_KEYS } from '../constants';
import type { DisplaySize, KeyId } from '../types';

/**
 * Chooses the display text size so that longer expressions still fit.
 *
 * @param length - Number of characters shown.
 * @returns The text size to use.
 */
export function displaySizeFor(length: number): DisplaySize {
  if (length > DISPLAY_SIZE_THRESHOLDS.xs) {
    return 'xs';
  }
  if (length > DISPLAY_SIZE_THRESHOLDS.s) {
    return 's';
  }
  if (length > DISPLAY_SIZE_THRESHOLDS.m) {
    return 'm';
  }
  return 'l';
}

/**
 * Maps a physical keyboard key to the calculator key it presses.
 *
 * @param key - `KeyboardEvent.key`, e.g. "7", "*", "Enter" or "Backspace".
 * @returns The calculator key, or `undefined` for keys the calculator ignores.
 */
export function keyForKeyboardKey(key: string): KeyId | undefined {
  return Object.hasOwn(KEYBOARD_KEYS, key) ? KEYBOARD_KEYS[key] : undefined;
}
