import { describe, expect, it } from 'vitest';

import { MAX_EXPRESSION_LENGTH } from '../constants';
import type { InputKey } from '../types';
import {
  applyInputKey,
  countOpenParentheses,
  deleteLast,
  endsWithOperand,
  type InputContext,
} from './expressionInput';

const TYPING: InputContext = { showingResult: false, hasAnswer: false };
const TYPING_WITH_ANSWER: InputContext = { showingResult: false, hasAnswer: true };
const SHOWING_RESULT: InputContext = { showingResult: true, hasAnswer: true };

/** Presses the keys in order, starting from `start`; only the first press sees `context`. */
function press(keys: InputKey[], context: InputContext = TYPING, start = ''): string {
  return keys.reduce(
    (expression, key, index) =>
      applyInputKey(expression, key, index === 0 ? context : { ...context, showingResult: false }),
    start,
  );
}

describe('applyInputKey', () => {
  describe('digits', () => {
    it.each<[InputKey[], string]>([
      [['1', '2'], '12'],
      [['0', '5'], '5'],
      [['0', '0'], '0'],
      [['1', '0'], '10'],
      [['1', 'add', '0', '5'], '1+5'],
      [['decimal', '0', '5'], '0.05'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys)).toBe(expected);
    });
  });

  describe('decimal point', () => {
    it.each<[InputKey[], string]>([
      [['decimal'], '0.'],
      [['1', 'decimal', 'decimal'], '1.'],
      [['1', 'decimal', '5', 'decimal'], '1.5'],
      [['1', 'add', 'decimal'], '1+0.'],
      [['1', 'decimal', '5', 'add', '2', 'decimal', '5'], '1.5+2.5'],
      [['openParen', '1', 'closeParen', 'decimal'], '(1)×0.'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys)).toBe(expected);
    });
  });

  describe('operators', () => {
    it.each<[InputKey[], string]>([
      [['add'], ''],
      [['multiply'], ''],
      [['1', 'add', 'multiply'], '1×'],
      [['1', 'multiply', 'subtract'], '1×−'],
      [['1', 'multiply', 'subtract', 'add'], '1+'],
      [['1', 'add', 'subtract'], '1−'],
      [['subtract'], '−'],
      [['subtract', 'subtract'], '−'],
      [['openParen', 'multiply'], '('],
      [['openParen', 'subtract', '2'], '(−2'],
      [['2', 'power', '3'], '2^3'],
      [['2', 'power', 'subtract', '1'], '2^−1'],
      [['5', 'decimal', 'add'], '5.+'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys)).toBe(expected);
    });

    it('replaces a typed ASCII operator too', () => {
      expect(applyInputKey('2*', 'add', TYPING)).toBe('2+');
    });
  });

  describe('percent, x² and 1/x need an operand on their left', () => {
    it.each<[InputKey[], string]>([
      [['percent'], ''],
      [['5', '0', 'percent'], '50%'],
      [['5', 'add', 'percent'], '5+'],
      [['5', 'percent', '3'], '5%×3'],
      [['3', 'square'], '3^2'],
      [['add', 'square'], ''],
      [['4', 'reciprocal'], '4^(−1)'],
      [['openParen', '2', 'add', '3', 'closeParen', 'square'], '(2+3)^2'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys)).toBe(expected);
    });
  });

  describe('parentheses and square root', () => {
    it.each<[InputKey[], string]>([
      [['closeParen'], ''],
      [['openParen', '2', 'closeParen', 'closeParen'], '(2)'],
      [['openParen', '2', 'add', 'closeParen'], '(2+'],
      [['2', 'openParen'], '2('],
      [['sqrt', '9', 'closeParen'], '√(9)'],
      [['2', 'sqrt'], '2√('],
      [['openParen', 'openParen', '1', 'closeParen', 'closeParen'], '((1))'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys)).toBe(expected);
    });
  });

  describe('Ans', () => {
    it('is ignored when there is no previous answer', () => {
      expect(press(['ans'])).toBe('');
    });

    it.each<[InputKey[], string]>([
      [['ans'], 'Ans'],
      [['2', 'ans'], '2×Ans'],
      [['ans', '5'], 'Ans×5'],
      [['ans', 'decimal'], 'Ans×0.'],
      [['ans', 'square'], 'Ans^2'],
    ])('%j → %s', (keys, expected) => {
      expect(press(keys, TYPING_WITH_ANSWER)).toBe(expected);
    });
  });

  describe('after a result', () => {
    it.each<[InputKey[], string]>([
      [['add', '2'], 'Ans+2'],
      [['subtract', '1'], 'Ans−1'],
      [['percent'], 'Ans%'],
      [['square'], 'Ans^2'],
      [['reciprocal'], 'Ans^(−1)'],
      [['7'], '7'],
      [['decimal'], '0.'],
      [['openParen'], '('],
      [['sqrt'], '√('],
      [['ans'], 'Ans'],
    ])('%j continues or starts over: %s', (keys, expected) => {
      expect(press(keys, SHOWING_RESULT, '29')).toBe(expected);
    });
  });

  it('ignores keys that would make the expression too long for the backend', () => {
    const longest = '1'.repeat(MAX_EXPRESSION_LENGTH);

    expect(applyInputKey(longest, '1', TYPING)).toBe(longest);
    expect(applyInputKey(longest.slice(1), '1', TYPING)).toBe(longest);
  });
});

describe('deleteLast', () => {
  it.each([
    ['12', '1'],
    ['1', ''],
    ['', ''],
    ['Ans', ''],
    ['2×Ans', '2×'],
    ['√(', ''],
    ['2√(', '2'],
    ['4^(−1)', '4^(−1'],
  ])('%s → %s', (expression, expected) => {
    expect(deleteLast(expression)).toBe(expected);
  });
});

describe('countOpenParentheses', () => {
  it.each([
    ['', 0],
    ['(1', 1],
    ['((1)', 1],
    ['√(√(16', 2],
    ['(1)', 0],
    [')(', 1],
  ])('%s → %i', (expression, expected) => {
    expect(countOpenParentheses(expression)).toBe(expected);
  });
});

describe('endsWithOperand', () => {
  it.each([
    ['12', true],
    ['1.', true],
    ['(1)', true],
    ['50%', true],
    ['2×Ans', true],
    ['', false],
    ['1+', false],
    ['(', false],
    ['√(', false],
  ])('%s → %s', (expression, expected) => {
    expect(endsWithOperand(expression)).toBe(expected);
  });
});
