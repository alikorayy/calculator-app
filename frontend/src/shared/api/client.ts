import {
  API_ERROR_MESSAGES,
  HEADER_ACCEPT,
  HEADER_CONTENT_TYPE,
  HTTP_METHOD_POST,
  HTTP_STATUS,
  JSON_CONTENT_TYPE,
  REQUEST_TIMEOUT_MS,
  RETRY_AFTER_HEADER,
} from '../constants/api';
import type { ApiError, ApiResult } from './errors';
import { isRecord } from './guards';
import { CALCULATION_ERROR_CODES, type CalculationErrorCode, type ErrorResponse } from './types';

/** Abort reason used when the time limit is reached, to tell it apart from cancellation. */
const TIMEOUT_ABORT_REASON = 'timeout';

/** Options of a single request. */
export interface RequestOptions {
  /** Cancels the request, e.g. when a newer request replaces it. */
  signal?: AbortSignal;
  /** Overrides the default time limit, in milliseconds. */
  timeoutMs?: number;
}

/**
 * Sends `body` as JSON with POST and decodes the answer. It never throws: every
 * outcome, including network failures, time-outs and cancellation, comes back
 * as an {@link ApiResult}.
 *
 * @param path - Path on the same origin, e.g. "/api/v1/evaluate".
 * @param body - Value to send; it is serialised with `JSON.stringify`.
 * @param isExpected - Type guard that checks the shape of a successful answer;
 *   anything else is reported as a server error rather than trusted blindly.
 * @param options - Cancellation signal and time limit.
 * @returns The decoded data, or a typed error.
 */
export async function postJson<T>(
  path: string,
  body: unknown,
  isExpected: (value: unknown) => value is T,
  options: RequestOptions = {},
): Promise<ApiResult<T>> {
  const { signal, timeoutMs = REQUEST_TIMEOUT_MS } = options;

  // One controller aborts the request for either reason: the time limit or the
  // caller. The abort reason tells them apart afterwards.
  const controller = new AbortController();
  const timer = setTimeout(() => {
    controller.abort(TIMEOUT_ABORT_REASON);
  }, timeoutMs);
  const abortFromCaller = () => {
    controller.abort();
  };
  if (signal?.aborted) {
    controller.abort();
  }
  signal?.addEventListener('abort', abortFromCaller, { once: true });

  try {
    const response = await fetch(new URL(path, window.location.origin), {
      method: HTTP_METHOD_POST,
      headers: { [HEADER_CONTENT_TYPE]: JSON_CONTENT_TYPE, [HEADER_ACCEPT]: JSON_CONTENT_TYPE },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
    const payload = await readJson(response);
    if (!response.ok) {
      return failure(toApiError(response, payload));
    }
    return isExpected(payload)
      ? { ok: true, data: payload }
      : failure({ kind: 'server', status: response.status, message: API_ERROR_MESSAGES.server });
  } catch {
    if (controller.signal.reason === TIMEOUT_ABORT_REASON) {
      return failure({ kind: 'timeout', message: API_ERROR_MESSAGES.timeout });
    }
    if (controller.signal.aborted) {
      return failure({ kind: 'aborted' });
    }
    return failure({ kind: 'network', message: API_ERROR_MESSAGES.network });
  } finally {
    clearTimeout(timer);
    signal?.removeEventListener('abort', abortFromCaller);
  }
}

/**
 * Reads the body as JSON. A body that is not JSON (for example an HTML error
 * page from a proxy) yields `undefined`; an aborted read still throws, so the
 * caller can tell cancellation apart.
 */
async function readJson(response: Response): Promise<unknown> {
  const text = await response.text();
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return undefined;
  }
}

/** Maps a non-2xx response to the matching {@link ApiError}. */
function toApiError(response: Response, payload: unknown): ApiError {
  const { status } = response;
  const envelope = isErrorResponse(payload) ? payload.error : undefined;

  if (
    status === HTTP_STATUS.UNPROCESSABLE_ENTITY &&
    envelope &&
    isCalculationErrorCode(envelope.code)
  ) {
    return {
      kind: 'calculation',
      code: envelope.code,
      message: envelope.message,
      position: envelope.position,
    };
  }
  if (status === HTTP_STATUS.TOO_MANY_REQUESTS) {
    return {
      kind: 'rate-limited',
      message: API_ERROR_MESSAGES.rateLimited,
      retryAfterSeconds: parseRetryAfter(response.headers.get(RETRY_AFTER_HEADER)),
    };
  }
  if (status >= HTTP_STATUS.BAD_REQUEST && status < HTTP_STATUS.SERVER_ERROR) {
    return { kind: 'invalid-request', status, message: API_ERROR_MESSAGES.invalidRequest };
  }
  return { kind: 'server', status, message: API_ERROR_MESSAGES.server };
}

/** Parses a Retry-After header given in seconds; other forms are ignored. */
function parseRetryAfter(value: string | null): number | undefined {
  if (value === null) {
    return undefined;
  }
  const seconds = Number.parseInt(value, 10);
  return Number.isNaN(seconds) || seconds < 0 ? undefined : seconds;
}

function failure<T>(error: ApiError): ApiResult<T> {
  return { ok: false, error };
}

function isErrorResponse(value: unknown): value is ErrorResponse {
  if (!isRecord(value) || !isRecord(value.error)) {
    return false;
  }
  const { code, message, position } = value.error;
  return (
    typeof code === 'string' &&
    typeof message === 'string' &&
    (position === undefined || typeof position === 'number')
  );
}

function isCalculationErrorCode(code: string): code is CalculationErrorCode {
  return (CALCULATION_ERROR_CODES as readonly string[]).includes(code);
}
