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
