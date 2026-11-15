import type { CalculationResponse, ErrorResponse } from '../shared/api/types';

/** Expressions the default mock handler knows, as the keypad builds them. */
export const EXPRESSIONS = {
  exact: '50−7×3',
  rounded: '1÷3',
  /** What "Ans×3" becomes after 1÷3 =, because the rounded answer is not reused. */
  roundedTimesThree: '(1÷3)×3',
  syntaxError: '2+*3',
  divisionByZero: '5÷0',
} as const;

/** Successful answers, matching real responses of the backend. */
export const RESULTS = {
  exact: { result: '29', approximate: false },
  rounded: { result: '0.3333333333333333', approximate: true },
  one: { result: '1', approximate: false },
} satisfies Record<string, CalculationResponse>;

/** Error envelopes, matching real responses of the backend. */
export const ERRORS = {
  syntaxError: { error: { code: 'SYNTAX_ERROR', message: 'Unexpected "*"', position: 2 } },
  divisionByZero: { error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero' } },
  invalidRequest: { error: { code: 'INVALID_REQUEST', message: 'Unknown field "expresion"' } },
  payloadTooLarge: {
    error: { code: 'PAYLOAD_TOO_LARGE', message: 'Request body is too large (maximum 4096 bytes)' },
  },
  rateLimited: { error: { code: 'RATE_LIMITED', message: 'Too many requests, please slow down' } },
  internal: { error: { code: 'INTERNAL_ERROR', message: 'An unexpected error occurred' } },
  timeout: { error: { code: 'TIMEOUT', message: 'The request took too long to process' } },
  unknownCalculationCode: { error: { code: 'NEW_CODE', message: 'Something new' } },
} satisfies Record<string, ErrorResponse>;

/** Seconds sent in the Retry-After header of a 429 fixture. */
export const RETRY_AFTER_SECONDS = 1;
