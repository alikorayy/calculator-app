import {
  MAX_EXPRESSION_LENGTH,
  OPERATOR_CHARACTERS,
  OPERATOR_SYMBOLS,
  SYMBOLS,
} from '../constants';
import type { InputKey } from '../types';

/** What {@link applyInputKey} needs to know besides the expression. */
export interface InputContext {
  /** The display shows the result of the last calculation, not an expression being typed. */
  showingResult: boolean;
  /** A previous answer exists, so "Ans" can be used. */
  hasAnswer: boolean;
}

/**
 * Returns the expression after a key press, applying the input rules that keep
 * the expression well formed (validation for user experience; the backend
 * still validates everything):
 *
 * - a second operator replaces the first; `−` after × ÷ ^ starts a negative number;
 * - a number has at most one decimal point, and `.` alone becomes `0.`;
 * - `)`, `%`, x² and 1/x only follow an operand; `)` only closes an open `(`;
 * - after a result, an operator continues from it as "Ans" and anything else
 *   starts a new expression;
 * - keys that would make the expression longer than the backend accepts are ignored.
 *
 * Ignored keys return the expression unchanged.
 *
 * @param expression - The current expression.
 * @param key - The key that was pressed.
 * @param context - Whether a result is shown and whether "Ans" is available.
 * @returns The new expression.
 */
export function applyInputKey(expression: string, key: InputKey, context: InputContext): string {
  const base = context.showingResult ? startAfterResult(key) : expression;
  const next = insertKey(base, key, context.hasAnswer);
  return next.length > MAX_EXPRESSION_LENGTH ? expression : next;
}

/** Endings of a finished operand: a new operand after them needs a × in between. */
const CLOSED_OPERAND_ENDINGS = [
  SYMBOLS.closeParen,
  SYMBOLS.percent,
  SYMBOLS.square,
  SYMBOLS.reciprocal,
  SYMBOLS.ans,
];

/**
 * Removes the last element of the expression: "Ans", "√(" and "⁻¹" go as a
 * whole, anything else one character at a time.
 *
 * @param expression - The current expression.
 * @returns The expression without its last element.
 */
export function deleteLast(expression: string): string {
  for (const unit of [SYMBOLS.ans, SYMBOLS.sqrt, SYMBOLS.reciprocal]) {
    if (expression.endsWith(unit)) {
      return expression.slice(0, -unit.length);
    }
  }
  return expression.slice(0, -1);
}

/**
 * Counts parentheses that are opened but not yet closed.
 *
 * @param expression - Any expression.
 * @returns The number of unclosed "(" (never negative).
 */
export function countOpenParentheses(expression: string): number {
  let open = 0;
  for (const character of expression) {
    if (character === SYMBOLS.openParen) {
      open++;
    } else if (character === SYMBOLS.closeParen && open > 0) {
      open--;
    }
  }
  return open;
}

/**
 * Whether the expression ends with something that can be followed by an
 * operator: a number, ")", "%", "²", "⁻¹" or "Ans".
 *
 * @param expression - Any expression.
 * @returns True when the last element is an operand.
 */
export function endsWithOperand(expression: string): boolean {
  return /[\d.]$/.test(expression) || endsWithClosedOperand(expression);
}

function endsWithClosedOperand(expression: string): boolean {
  return CLOSED_OPERAND_ENDINGS.some((ending) => expression.endsWith(ending));
}

/** After a result, operators continue from it as "Ans"; other keys start over. */
function startAfterResult(key: InputKey): string {
  const continuesResult =
    key in OPERATOR_SYMBOLS || key === 'percent' || key === 'square' || key === 'reciprocal';
  return continuesResult ? SYMBOLS.ans : '';
}

function insertKey(expression: string, key: InputKey, hasAnswer: boolean): string {
  switch (key) {
    case 'decimal':
      return insertDecimal(expression);
    case 'subtract':
      return insertMinus(expression);
    case 'add':
    case 'multiply':
    case 'divide':
    case 'power':
      return insertOperator(expression, OPERATOR_SYMBOLS[key]);
    case 'percent':
    case 'square':
    case 'reciprocal':
      return endsWithOperand(expression) ? expression + SYMBOLS[key] : expression;
    case 'closeParen':
      return countOpenParentheses(expression) > 0 && endsWithOperand(expression)
        ? expression + SYMBOLS.closeParen
        : expression;
    case 'openParen':
      return expression + SYMBOLS.openParen;
    case 'sqrt':
      return expression + SYMBOLS.sqrt;
    case 'ans':
      return hasAnswer ? insertAnswer(expression) : expression;
    default:
      return insertDigit(expression, key);
  }
}

/** "Ans" after any operand, even a number, means multiplication: 2 then Ans is 2×Ans. */
function insertAnswer(expression: string): string {
  const separator = endsWithOperand(expression) ? OPERATOR_SYMBOLS.multiply : '';
  return expression + separator + SYMBOLS.ans;
}

/** A lone leading "0" is replaced instead of producing "05". */
function insertDigit(expression: string, digit: string): string {
  if (currentNumber(expression) === SYMBOLS.zero) {
    return expression.slice(0, -1) + digit;
  }
  return withImplicitMultiply(expression) + digit;
}

function insertDecimal(expression: string): string {
  const number = currentNumber(expression);
  if (number.includes(SYMBOLS.decimal)) {
    return expression;
  }
  if (number === '') {
    return withImplicitMultiply(expression) + SYMBOLS.zero + SYMBOLS.decimal;
  }
  return expression + SYMBOLS.decimal;
}

/** + × ÷ ^ need an operand on their left; a trailing operator is replaced. */
function insertOperator(expression: string, symbol: string): string {
  const withoutOperators = trimTrailingOperators(expression);
  if (withoutOperators === '' || withoutOperators.endsWith(SYMBOLS.openParen)) {
    return expression;
  }
  return withoutOperators + symbol;
}

/**
 * − may start a negative number: at the start, after "(" and after × ÷ ^, so
 * it is appended there. It replaces a trailing + and is not doubled.
 */
function insertMinus(expression: string): string {
  const last = expression.at(-1);
  if (last === OPERATOR_SYMBOLS.subtract || last === '-') {
    return expression;
  }
  if (last === OPERATOR_SYMBOLS.add) {
    return expression.slice(0, -1) + OPERATOR_SYMBOLS.subtract;
  }
  return expression + OPERATOR_SYMBOLS.subtract;
}

/**
 * Inserts × when a new operand directly follows a finished operand such as
 * ")", "%", "²" or "Ans", so "Ans" then "5" means Ans × 5 rather than a longer number.
 */
function withImplicitMultiply(expression: string): string {
  return endsWithClosedOperand(expression) ? expression + OPERATOR_SYMBOLS.multiply : expression;
}

/** The digits and decimal point at the end of the expression, e.g. "12." in "3+12.". */
function currentNumber(expression: string): string {
  return /[\d.]*$/.exec(expression)?.[0] ?? '';
}

function trimTrailingOperators(expression: string): string {
  let end = expression.length;
  while (end > 0 && OPERATOR_CHARACTERS.has(expression.charAt(end - 1))) {
    end--;
  }
  return expression.slice(0, end);
}
