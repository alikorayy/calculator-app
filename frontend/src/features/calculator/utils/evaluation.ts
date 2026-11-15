import { OPERATOR_CHARACTERS, SYMBOLS } from '../constants';
import type { Answer } from '../types';
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
 * @param expression - The expression as shown, possibly containing "Ans".
 * @param answer - The previous answer, if any.
 * @returns The expression to send to the server.
 */
export function substituteAnswer(expression: string, answer: Answer | null): string {
  if (answer === null || !expression.includes(SYMBOLS.ans)) {
    return expression;
  }
  const exact = !answer.approximate && PLAIN_NUMBER.test(answer.result);
  const replacement = `(${exact ? answer.result : answer.expression})`;
  return expression.replaceAll(SYMBOLS.ans, replacement);
}
