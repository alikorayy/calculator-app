import type { BinaryOperatorKey, KeyId } from './types';

/** Text inserted into the expression by each binary operator key, as shown on the keypad. */
export const OPERATOR_SYMBOLS = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
} as const satisfies Record<BinaryOperatorKey, string>;

/** Text inserted by the other non-digit keys. The backend understands all of them. */
export const SYMBOLS = {
  decimal: '.',
  percent: '%',
  openParen: '(',
  closeParen: ')',
  sqrt: '√(',
  /** x² is "to the power of two". */
  square: '^2',
  /** 1/x is "to the power of minus one", so it applies to the operand before it. */
  reciprocal: '^(−1)',
  ans: 'Ans',
  zero: '0',
} as const;

/**
 * Characters that act as binary operators, including the ASCII forms a user
 * may type directly into the input.
 */
export const OPERATOR_CHARACTERS: ReadonlySet<string> = new Set([
  '+',
  '−',
  '-',
  '×',
  '*',
  '÷',
  '/',
  '^',
]);

/** Longest expression the backend accepts; keys that would exceed it are ignored. */
export const MAX_EXPRESSION_LENGTH = 256;

/** Texts of the line above the input. */
export const TOP_LINE = {
  calculating: 'Calculating…',
  equals: '=',
  /** Shown instead of "=" when the result was rounded. */
  approximately: '≈',
  answerPrefix: 'Ans = ',
  noAnswer: '0',
} as const;

/** Session storage key of the calculation history. */
export const HISTORY_STORAGE_KEY = 'calculator.history';

/** How many calculations the history keeps; older ones are dropped. */
export const MAX_HISTORY_ENTRIES = 50;

/** Expression lengths above which the display text gets smaller. */
export const DISPLAY_SIZE_THRESHOLDS = { m: 12, s: 18, xs: 26 } as const;

/** Physical keyboard keys (KeyboardEvent.key) and the calculator key each one presses. */
export const KEYBOARD_KEYS: Readonly<Record<string, KeyId>> = {
  '0': '0',
  '1': '1',
  '2': '2',
  '3': '3',
  '4': '4',
  '5': '5',
  '6': '6',
  '7': '7',
  '8': '8',
  '9': '9',
  '.': 'decimal',
  ',': 'decimal',
  '+': 'add',
  '-': 'subtract',
  '*': 'multiply',
  x: 'multiply',
  '/': 'divide',
  '^': 'power',
  '%': 'percent',
  '(': 'openParen',
  ')': 'closeParen',
  Enter: 'equals',
  '=': 'equals',
  Backspace: 'backspace',
  Escape: 'clear',
  Delete: 'clear',
};
