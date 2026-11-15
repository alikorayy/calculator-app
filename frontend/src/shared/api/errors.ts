import type { CalculationErrorCode } from './types';

/**
 * Every way an API call can fail, as a discriminated union on `kind`, so callers
 * handle each case explicitly instead of inspecting status codes:
 *
 * - `calculation`: the expression cannot be calculated (422); `message` comes from
 *   the server and is meant for the user, `position` points at a syntax error;
 * - `invalid-request`: the request itself was rejected (400, 404, 405, 413, 415),
 *   which means a bug in the client;
 * - `rate-limited`: too many requests (429), retry after `retryAfterSeconds`;
 * - `server`: the server failed or answered something unexpected (5xx, bad body);
 * - `network`: the server could not be reached;
 * - `timeout`: no answer within the client's time limit;
 * - `aborted`: the caller cancelled the request, e.g. because a newer one started.
 *   It is not shown to the user.
 */
export type ApiError =
  | { kind: 'calculation'; code: CalculationErrorCode; message: string; position?: number }
  | { kind: 'invalid-request'; status: number; message: string }
  | { kind: 'rate-limited'; message: string; retryAfterSeconds?: number }
  | { kind: 'server'; status?: number; message: string }
  | { kind: 'network'; message: string }
  | { kind: 'timeout'; message: string }
  | { kind: 'aborted' };

/** Outcome of an API call: the decoded data or a typed error. Calls never throw. */
export type ApiResult<T> = { ok: true; data: T } | { ok: false; error: ApiError };
