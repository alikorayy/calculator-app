import type { IconName } from '../../shared/components/Icon/icons';

/** A digit key. */
export type DigitKey = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9';

/** Keys of binary operators placed between two operands. */
export type BinaryOperatorKey = 'add' | 'subtract' | 'multiply' | 'divide' | 'power';

/** Keys that change the expression being typed. */
export type InputKey =
  | DigitKey
  | BinaryOperatorKey
  | 'decimal'
  | 'percent'
  | 'openParen'
  | 'closeParen'
  | 'sqrt'
  | 'square'
  | 'reciprocal'
  | 'ans';

/** Every key of the calculator, on the keypad or the keyboard. */
export type KeyId = InputKey | 'backspace' | 'clear' | 'equals';

/** The previous answer, used where the expression contains "Ans". */
export interface Answer {
  /** The result as returned by the server, e.g. "29" or "0.3333333333333333". */
  result: string;
  /** Whether `result` was rounded by the server. */
  approximate: boolean;
  /** The expression that was sent to the server to get this result. */
  expression: string;
}

/** Text size of the display, from largest to smallest. */
export type DisplaySize = 'l' | 'm' | 's' | 'xs';

/**
 * Phase of the calculator:
 * - `editing`: an expression is being typed;
 * - `evaluating`: waiting for the server;
 * - `result`: the display shows the last result;
 * - `error`: the last calculation failed and the expression can be corrected.
 */
export type CalculatorStatus = 'editing' | 'evaluating' | 'result' | 'error';

/** An error shown to the user. */
export interface CalculatorError {
  message: string;
  /** Character position of a syntax error in the displayed expression. */
  position?: number;
}

/** Complete state of the calculator. */
export interface CalculatorState {
  status: CalculatorStatus;
  /** Text in the display: the expression being typed, or the last result. */
  expression: string;
  /** The last successful answer, used for "Ans". */
  answer: Answer | null;
  /** The expression that produced the result on display, shown as "<expression> =". */
  evaluatedExpression: string | null;
  error: CalculatorError | null;
}

/** A finished calculation kept in the history. */
export interface HistoryEntry {
  id: string;
  /** The expression as the user saw it; it may contain "Ans". */
  expression: string;
  result: string;
  /** Whether `result` was rounded by the server. */
  approximate: boolean;
  /** The expression that was sent to the server, with "Ans" replaced. */
  sentExpression: string;
}

/** Label of the clear key: clear the entry, or clear everything. */
export type ClearLabel = 'CE' | 'AC';

/** Visual style of a key. */
export type KeyVariant = 'number' | 'operator' | 'equals' | 'clear' | 'percent' | 'function';

/** Everything needed to render one key of a keypad. */
export interface KeyDefinition {
  id: KeyId;
  /** Text on the key; ignored when `icon` is set. */
  label: string;
  /** Icon shown instead of the label. */
  icon?: IconName;
  /** Spoken name when the label alone is unclear, e.g. "Multiply" for ×. */
  ariaLabel?: string;
  variant: KeyVariant;
}
