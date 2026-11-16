import {
  MAX_EXPRESSION_LENGTH,
  MAX_NESTING_DEPTH,
  OPERATOR_CHARACTERS,
  OPERATOR_SYMBOLS,
  SCIENTIFIC_BASE,
  SYMBOLS,
} from '../constants';
import type { Answer, HistoryEntry } from '../types';
import { countOpenParentheses } from './expressionInput';

/** A result that can be reused as written: a plain decimal number, no exponent. */
const PLAIN_NUMBER = /^-?\d+(\.\d+)?$/;

/**
 * Prepares an expression for "=": open parentheses are closed automatically,
 * like on Google's and Windows' calculators. An empty expression, or one that
 * ends with an operator or an open parenthesis, is not ready to evaluate.
 *
 * @param expression - The expression as shown in the display.
 * @returns The expression to evaluate, or `null` when "=" should do nothing.
 */
export function prepareForEvaluation(expression: string): string | null {
  const trimmed = expression.trim();
  const last = trimmed.at(-1);
  if (last === undefined || last === SYMBOLS.openParen || OPERATOR_CHARACTERS.has(last)) {
    return null;
  }
  return trimmed + SYMBOLS.closeParen.repeat(countOpenParentheses(trimmed));
}

/**
 * Replaces every "Ans" with the previous answer, in parentheses so it keeps its
 * meaning next to any operator (Ans² with Ans = −5 is 25, not −25).
 *
 * A rounded or scientific answer is not fed back as its rounded text, which
 * would lose precision: the expression that produced it is used instead. So
 * 1/3 = followed by × 3 = evaluates (1/3)×3 and gives exactly 1.
 *
 * Each such step nests the previous expression once more. When that would go
 * past the backend's length or nesting limit (after about 30 chained steps),
 * the displayed result is used instead: precise to 16 digits, like on other
 * calculators, and the chain starts short again.
 *
 * @param expression - The expression as shown, possibly containing "Ans".
 * @param answer - The previous answer, if any.
 * @returns The expression to send to the server.
 */
export function substituteAnswer(expression: string, answer: Answer | null): string {
  if (answer === null || !expression.includes(SYMBOLS.ans)) {
    return expression;
  }
  const source = isExactText(answer.result, answer.approximate) ? answer.result : answer.expression;
  const exact = expression.replaceAll(SYMBOLS.ans, `(${source})`);
  if (exact.length <= MAX_EXPRESSION_LENGTH && nestingDepth(exact) <= MAX_NESTING_DEPTH) {
    return exact;
  }
  return expression.replaceAll(SYMBOLS.ans, `(${typeableNumber(answer.result)})`);
}

/** Result text in scientific notation, e.g. "1.5e+20" or "-2e-7". */
const SCIENTIFIC_NUMBER = /^(-?[\d.]+)e([+-])(\d+)$/;

/**
 * Writes a result so the backend can read it back: "1.5e+20" becomes
 * "1.5×10^20" and "2e-7" becomes "2×10^−7"; plain numbers stay as they are.
 */
function typeableNumber(result: string): string {
  const match = SCIENTIFIC_NUMBER.exec(result);
  if (match === null) {
    return result;
  }
  const [, mantissa = '', sign, exponent = ''] = match;
  const exponentSign = sign === '-' ? OPERATOR_SYMBOLS.subtract : '';
  return `${mantissa}${OPERATOR_SYMBOLS.multiply}${SCIENTIFIC_BASE}${OPERATOR_SYMBOLS.power}${exponentSign}${exponent}`;
}

/** The deepest nesting of parentheses in an expression; "√(" opens one too. */
function nestingDepth(expression: string): number {
  let depth = 0;
  let deepest = 0;
  for (const character of expression) {
    if (character === SYMBOLS.openParen) {
      depth++;
      deepest = Math.max(deepest, depth);
    } else if (character === SYMBOLS.closeParen) {
      depth = Math.max(depth - 1, 0);
    }
  }
  return deepest;
}

/**
 * Text to put back into the display when a history result is picked, by the
 * same rule as "Ans": an exact plain number is reused as it is. A rounded
 * result ("0.3333333333333333") would lose precision and a scientific one
 * ("1e+64") cannot be typed back, so the expression that was sent to get it is
 * used instead, in parentheses.
 *
 * @param entry - The picked history entry.
 * @returns Text the backend accepts as part of an expression.
 */
export function reusableResult(entry: HistoryEntry): string {
  return isExactText(entry.result, entry.approximate) ? entry.result : `(${entry.sentExpression})`;
}

/**
 * Text to put back into the display when a history expression is picked. An
 * expression with "Ans" is restored as it was sent, because "Ans" may mean a
 * different value by now.
 *
 * @param entry - The picked history entry.
 * @returns The expression to restore.
 */
export function reusableExpression(entry: HistoryEntry): string {
  return entry.expression.includes(SYMBOLS.ans) ? entry.sentExpression : entry.expression;
}

/** Whether a result's text is its exact value and can be typed back as it is. */
function isExactText(result: string, approximate: boolean): boolean {
  return !approximate && PLAIN_NUMBER.test(result);
}
