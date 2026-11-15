import { describe, expect, it } from 'vitest';

import { DISPLAY_SIZE_THRESHOLDS } from '../constants';
import { displaySizeFor, keyForKeyboardKey } from './display';

describe('displaySizeFor', () => {
  it.each([
    [0, 'l'],
    [DISPLAY_SIZE_THRESHOLDS.m, 'l'],
    [DISPLAY_SIZE_THRESHOLDS.m + 1, 'm'],
    [DISPLAY_SIZE_THRESHOLDS.s, 'm'],
    [DISPLAY_SIZE_THRESHOLDS.s + 1, 's'],
    [DISPLAY_SIZE_THRESHOLDS.xs, 's'],
    [DISPLAY_SIZE_THRESHOLDS.xs + 1, 'xs'],
  ])('length %i → %s', (length, expected) => {
    expect(displaySizeFor(length)).toBe(expected);
  });
});

describe('keyForKeyboardKey', () => {
  it.each([
    ['7', '7'],
    ['.', 'decimal'],
    [',', 'decimal'],
    ['*', 'multiply'],
    ['x', 'multiply'],
    ['-', 'subtract'],
    ['Enter', 'equals'],
    ['=', 'equals'],
    ['Backspace', 'backspace'],
    ['Escape', 'clear'],
  ])('%s → %s', (key, expected) => {
    expect(keyForKeyboardKey(key)).toBe(expected);
  });

  it.each(['a', 'Shift', 'F5', 'toString', 'constructor'])('ignores %s', (key) => {
    expect(keyForKeyboardKey(key)).toBeUndefined();
  });
});
