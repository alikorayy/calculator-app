import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { ERRORS, EXPRESSIONS, RESULTS, RETRY_AFTER_SECONDS } from '../../mocks/fixtures';
import { server } from '../../mocks/server';
import { API_ERROR_MESSAGES, EVALUATE_PATH, RETRY_AFTER_HEADER } from '../constants/api';
import { evaluateExpression, isCalculationResponse } from './calculatorApi';
import type { ErrorResponse } from './types';

/** Makes the evaluate endpoint answer with the given error body, status and headers. */
function answerWith(body: ErrorResponse, status: number, headers: Record<string, string> = {}) {
  server.use(http.post(EVALUATE_PATH, () => HttpResponse.json(body, { status, headers })));
}

describe('evaluateExpression', () => {
  it('sends the expression and returns the result', async () => {
    let sentBody: unknown;
    server.use(
      http.post(EVALUATE_PATH, async ({ request }) => {
        sentBody = await request.json();
        return HttpResponse.json(RESULTS.exact);
      }),
    );

    const result = await evaluateExpression(EXPRESSIONS.exact);

    expect(sentBody).toEqual({ expression: EXPRESSIONS.exact });
    expect(result).toEqual({ ok: true, data: RESULTS.exact });
  });

  it('turns a calculation error into a calculation error with the server message', async () => {
    const result = await evaluateExpression(EXPRESSIONS.divisionByZero);

    expect(result).toEqual({
      ok: false,
      error: {
        kind: 'calculation',
        code: 'DIVISION_BY_ZERO',
        message: ERRORS.divisionByZero.error.message,
        position: undefined,
      },
    });
  });

  it('keeps the position of a syntax error', async () => {
    const result = await evaluateExpression(EXPRESSIONS.syntaxError);

    expect(result).toMatchObject({
      ok: false,
      error: {
        kind: 'calculation',
        code: 'SYNTAX_ERROR',
        position: ERRORS.syntaxError.error.position,
      },
    });
  });

  it.each([
    [
      '429 with Retry-After',
      ERRORS.rateLimited,
      429,
      { [RETRY_AFTER_HEADER]: String(RETRY_AFTER_SECONDS) },
      {
        kind: 'rate-limited',
        message: API_ERROR_MESSAGES.rateLimited,
        retryAfterSeconds: RETRY_AFTER_SECONDS,
      },
    ],
    [
      '429 without Retry-After',
      ERRORS.rateLimited,
      429,
      {},
      {
        kind: 'rate-limited',
        message: API_ERROR_MESSAGES.rateLimited,
        retryAfterSeconds: undefined,
      },
    ],
    [
      '429 with an unreadable Retry-After',
      ERRORS.rateLimited,
      429,
      { [RETRY_AFTER_HEADER]: 'soon' },
      {
        kind: 'rate-limited',
        message: API_ERROR_MESSAGES.rateLimited,
        retryAfterSeconds: undefined,
      },
    ],
    [
      '400',
      ERRORS.invalidRequest,
      400,
      {},
      {
        kind: 'invalid-request',
        status: 400,
        message: API_ERROR_MESSAGES.invalidRequest,
      },
    ],
    [
      '413',
      ERRORS.payloadTooLarge,
      413,
      {},
      {
        kind: 'invalid-request',
        status: 413,
        message: API_ERROR_MESSAGES.invalidRequest,
      },
    ],
    [
      '422 with an unknown code',
      ERRORS.unknownCalculationCode,
      422,
      {},
      {
        kind: 'invalid-request',
        status: 422,
        message: API_ERROR_MESSAGES.invalidRequest,
      },
    ],
    [
      '500',
      ERRORS.internal,
      500,
      {},
      { kind: 'server', status: 500, message: API_ERROR_MESSAGES.server },
    ],
    [
      '503 timeout',
      ERRORS.timeout,
      503,
      {},
      { kind: 'server', status: 503, message: API_ERROR_MESSAGES.server },
    ],
  ])(
    'maps %s to a typed error with a user-facing message',
    async (_, body, status, headers, error) => {
      answerWith(body, status, headers);

      const result = await evaluateExpression(EXPRESSIONS.exact);

      expect(result).toEqual({ ok: false, error });
    },
  );
});

describe('isCalculationResponse', () => {
  it.each([
    [RESULTS.exact, true],
    [RESULTS.rounded, true],
    [{ result: 29, approximate: false }, false],
    [{ result: '29' }, false],
    [null, false],
    ['29', false],
  ])('%j → %s', (value, expected) => {
    expect(isCalculationResponse(value)).toBe(expected);
  });
});
