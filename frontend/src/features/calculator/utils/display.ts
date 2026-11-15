import { DISPLAY_SIZE_THRESHOLDS, KEYBOARD_KEYS, TOP_LINE } from '../constants';
import type { CalculatorState, DisplaySize, KeyId } from '../types';

/**
 * The small line above the input. It shows, in order of priority, the error,
 * that a calculation is running, the expression behind the result on display
 * ("1/3 ≈" when the result was rounded), or the previous answer.
 *
 * @param state - Current calculator state.
 * @returns The text to show.
 */
export function topLineFor(state: CalculatorState): string {
  if (state.error) {
    return state.error.message;
  }
  if (state.status === 'evaluating') {
    return TOP_LINE.calculating;
  }
  if (state.status === 'result' && state.evaluatedExpression !== null) {
    const relation = state.answer?.approximate ? TOP_LINE.approximately : TOP_LINE.equals;
    return `${state.evaluatedExpression} ${relation}`;
  }
  return `${TOP_LINE.answerPrefix}${state.answer?.result ?? TOP_LINE.noAnswer}`;
}

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
