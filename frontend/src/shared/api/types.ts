/**
 * Types of the calculator API, mirroring `backend/api/openapi.yaml`.
 * Results are strings so that no precision is lost in JavaScript numbers.
 */

/** Body of `POST /api/v1/evaluate`. */
export interface EvaluateRequest {
  expression: string;
}

/** Body of every successful calculation. */
export interface CalculationResponse {
  /** The result rounded to 16 significant digits, e.g. "29" or "1e+64". */
  result: string;
  /** True when `result` was rounded, i.e. is not the exact value. */
  approximate: boolean;
}

/** Error codes of well-formed requests whose math cannot be done (HTTP 422). */
export const CALCULATION_ERROR_CODES = [
  'SYNTAX_ERROR',
  'DIVISION_BY_ZERO',
  'UNDEFINED_RESULT',
  'DOMAIN_ERROR',
  'OVERFLOW',
  'LIMIT_EXCEEDED',
] as const;

/** One of {@link CALCULATION_ERROR_CODES}. */
export type CalculationErrorCode = (typeof CALCULATION_ERROR_CODES)[number];

/** The error envelope the API uses for every error response. */
export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    /** For SYNTAX_ERROR, the 0-based character index of the problem. */
    position?: number;
  };
}
