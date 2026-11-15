import { EVALUATE_PATH } from '../constants/api';
import { postJson, type RequestOptions } from './client';
import type { ApiResult } from './errors';
import { isRecord } from './guards';
import type { CalculationResponse, EvaluateRequest } from './types';

/**
 * Asks the server to evaluate a calculator expression such as "50 - 7 × 3".
 * The frontend never calculates; this is the only way results are obtained.
 *
 * @param expression - The expression exactly as shown to the user.
 * @param options - Cancellation signal and time limit.
 * @returns The result, or a typed error (see {@link ApiResult}).
 */
export function evaluateExpression(
  expression: string,
  options?: RequestOptions,
): Promise<ApiResult<CalculationResponse>> {
  const body: EvaluateRequest = { expression };
  return postJson(EVALUATE_PATH, body, isCalculationResponse, options);
}

/**
 * Checks that a value has the shape of a successful calculation.
 *
 * @param value - Any decoded JSON value.
 * @returns Whether `value` is a {@link CalculationResponse}.
 */
export function isCalculationResponse(value: unknown): value is CalculationResponse {
  return (
    isRecord(value) && typeof value.result === 'string' && typeof value.approximate === 'boolean'
  );
}
