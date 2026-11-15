import type { BinaryOperatorKey, ClearLabel, KeyDefinition, KeyId } from './types';

/**
 * Keys of the function pad. Only operations the backend can calculate are
 * offered; the frontend never calculates anything itself.
 */
export const FUNCTION_KEYS: readonly KeyDefinition[] = [
  { id: 'sqrt', label: '√', ariaLabel: 'Square root', variant: 'function' },
  { id: 'square', label: 'x²', ariaLabel: 'Square', variant: 'function' },
  { id: 'power', label: 'xʸ', ariaLabel: 'Power', variant: 'function' },
  { id: 'reciprocal', label: '1/x', ariaLabel: 'Reciprocal', variant: 'function' },
  { id: 'ans', label: 'Ans', ariaLabel: 'Previous answer', variant: 'function' },
  {
    id: 'backspace',
    label: '',
    icon: 'backspace',
    ariaLabel: 'Delete last character',
    variant: 'function',
  },
];

/** Keys of the basic pad, row by row. The clear key's label comes from the state. */
export const BASIC_KEYS: readonly KeyDefinition[] = [
  { id: 'openParen', label: '(', ariaLabel: 'Open parenthesis', variant: 'function' },
  { id: 'closeParen', label: ')', ariaLabel: 'Close parenthesis', variant: 'function' },
  { id: 'percent', label: '%', ariaLabel: 'Percent', variant: 'percent' },
  { id: 'clear', label: 'AC', variant: 'clear' },
  { id: '7', label: '7', variant: 'number' },
  { id: '8', label: '8', variant: 'number' },
  { id: '9', label: '9', variant: 'number' },
  { id: 'divide', label: '÷', ariaLabel: 'Divide', variant: 'operator' },
  { id: '4', label: '4', variant: 'number' },
  { id: '5', label: '5', variant: 'number' },
  { id: '6', label: '6', variant: 'number' },
  { id: 'multiply', label: '×', ariaLabel: 'Multiply', variant: 'operator' },
  { id: '1', label: '1', variant: 'number' },
  { id: '2', label: '2', variant: 'number' },
  { id: '3', label: '3', variant: 'number' },
  { id: 'subtract', label: '−', ariaLabel: 'Subtract', variant: 'operator' },
  { id: '0', label: '0', variant: 'number' },
  { id: 'decimal', label: '.', ariaLabel: 'Decimal point', variant: 'number' },
  { id: 'equals', label: '=', ariaLabel: 'Calculate', variant: 'equals' },
  { id: 'add', label: '+', ariaLabel: 'Add', variant: 'operator' },
];

/** Spoken names of the clear key for each of its labels. */
export const CLEAR_ARIA_LABELS: Readonly<Record<ClearLabel, string>> = {
  AC: 'All clear',
  CE: 'Clear entry',
};

/** Texts of the calculator interface. */
export const LABELS = {
  expression: 'Expression',
  expressionPlaceholder: '0',
  functionPad: 'Functions',
  basicPad: 'Keypad',
  history: 'History',
  clearHistory: 'Clear',
  historyEmpty:
    'No calculations yet. Results show up here, and you can tap an expression or result to reuse it.',
  switchToLight: 'Switch to light theme',
  switchToDark: 'Switch to dark theme',
} as const;

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
