import { describe, expect, it } from 'vitest';

import { INITIAL_CALCULATOR_STATE } from '../hooks/calculatorReducer';
import { DISPLAY_SIZE_THRESHOLDS, TOP_LINE } from '../constants';
import type { CalculatorState } from '../types';
import { displaySizeFor, keyForKeyboardKey, topLineFor } from './display';

describe('topLineFor', () => {
  const answer = { result: '29', approximate: false, expression: '50−7×3' };

  it.each<[string, CalculatorState, string]>([
    ['no answer yet', INITIAL_CALCULATOR_STATE, 'Ans = 0'],
    ['a previous answer', { ...INITIAL_CALCULATOR_STATE, answer }, 'Ans = 29'],
    [
      'a running calculation',
      { ...INITIAL_CALCULATOR_STATE, status: 'evaluating' },
      TOP_LINE.calculating,
    ],
    [
      'an exact result',
      { ...INITIAL_CALCULATOR_STATE, status: 'result', answer, evaluatedExpression: '50−7×3' },
      '50−7×3 =',
    ],
    [
      'a rounded result',
      {
        ...INITIAL_CALCULATOR_STATE,
        status: 'result',
        answer: { ...answer, approximate: true },
        evaluatedExpression: '1÷3',
      },
      '1÷3 ≈',
    ],
    [
      'an error',
      { ...INITIAL_CALCULATOR_STATE, status: 'error', error: { message: 'Cannot divide by zero' } },
      'Cannot divide by zero',
    ],
  ])('shows %s', (_, state, expected) => {
    expect(topLineFor(state)).toBe(expected);
  });
});

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
