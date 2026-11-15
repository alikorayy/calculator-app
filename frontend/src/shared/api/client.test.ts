import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { RESULTS } from '../../mocks/fixtures';
import { server } from '../../mocks/server';
import { API_ERROR_MESSAGES } from '../constants/api';
import { isCalculationResponse } from './calculatorApi';
import { postJson } from './client';

const TEST_PATH = '/api/v1/test';
const SHORT_TIMEOUT_MS = 20;
const SLOW_RESPONSE_MS = 1_000;

function post(options?: Parameters<typeof postJson>[3]) {
  return postJson(TEST_PATH, { expression: '1' }, isCalculationResponse, options);
}

describe('postJson', () => {
  it('reports a time-out when the server does not answer in time', async () => {
    server.use(
      http.post(TEST_PATH, async () => {
        await delay(SLOW_RESPONSE_MS);
        return HttpResponse.json(RESULTS.exact);
      }),
    );

    const result = await post({ timeoutMs: SHORT_TIMEOUT_MS });

    expect(result).toEqual({
      ok: false,
      error: { kind: 'timeout', message: API_ERROR_MESSAGES.timeout },
    });
  });

  it('reports cancellation by the caller as aborted', async () => {
    server.use(
      http.post(TEST_PATH, async () => {
        await delay(SLOW_RESPONSE_MS);
        return HttpResponse.json(RESULTS.exact);
      }),
    );
    const controller = new AbortController();

    const pending = post({ signal: controller.signal });
    controller.abort();

    expect(await pending).toEqual({ ok: false, error: { kind: 'aborted' } });
  });

  it('does not send a request whose signal is already aborted', async () => {
    let requests = 0;
    server.use(
      http.post(TEST_PATH, () => {
        requests++;
        return HttpResponse.json(RESULTS.exact);
      }),
    );

    const result = await post({ signal: AbortSignal.abort() });

    expect(result).toEqual({ ok: false, error: { kind: 'aborted' } });
    expect(requests).toBe(0);
  });

  it('reports a network failure', async () => {
    server.use(http.post(TEST_PATH, () => HttpResponse.error()));

    const result = await post();

    expect(result).toEqual({
      ok: false,
      error: { kind: 'network', message: API_ERROR_MESSAGES.network },
    });
  });

  it('treats a non-JSON error page (e.g. from a proxy) as a server error', async () => {
    server.use(
      http.post(TEST_PATH, () => new HttpResponse('<html>Bad Gateway</html>', { status: 502 })),
    );

    const result = await post();

    expect(result).toEqual({
      ok: false,
      error: { kind: 'server', status: 502, message: API_ERROR_MESSAGES.server },
    });
  });

  it('treats a successful answer with an unexpected shape as a server error', async () => {
    server.use(http.post(TEST_PATH, () => HttpResponse.json({ value: 29 })));

    const result = await post();

    expect(result).toEqual({
      ok: false,
      error: { kind: 'server', status: 200, message: API_ERROR_MESSAGES.server },
    });
  });
});
