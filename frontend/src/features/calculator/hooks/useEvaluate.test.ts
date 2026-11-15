import { renderHook } from '@testing-library/react';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { EXPRESSIONS, RESULTS } from '../../../mocks/fixtures';
import { server } from '../../../mocks/server';
import { EVALUATE_PATH } from '../../../shared/constants/api';
import { useEvaluate } from './useEvaluate';

const SLOW_RESPONSE_MS = 200;

describe('useEvaluate', () => {
  it('returns the server result', async () => {
    const { result } = renderHook(() => useEvaluate());

    expect(await result.current(EXPRESSIONS.exact)).toEqual({ ok: true, data: RESULTS.exact });
  });

  it('cancels the previous request when a new one starts', async () => {
    server.use(
      http.post(EVALUATE_PATH, async () => {
        await delay(SLOW_RESPONSE_MS);
        return HttpResponse.json(RESULTS.exact);
      }),
    );
    const { result } = renderHook(() => useEvaluate());

    const first = result.current(EXPRESSIONS.exact);
    const second = result.current(EXPRESSIONS.exact);

    expect(await first).toEqual({ ok: false, error: { kind: 'aborted' } });
    expect(await second).toEqual({ ok: true, data: RESULTS.exact });
  });

  it('cancels a pending request when the component unmounts', async () => {
    server.use(
      http.post(EVALUATE_PATH, async () => {
        await delay(SLOW_RESPONSE_MS);
        return HttpResponse.json(RESULTS.exact);
      }),
    );
    const { result, unmount } = renderHook(() => useEvaluate());

    const pending = result.current(EXPRESSIONS.exact);
    unmount();

    expect(await pending).toEqual({ ok: false, error: { kind: 'aborted' } });
  });

  it('keeps the same function between renders', () => {
    const { result, rerender } = renderHook(() => useEvaluate());
    const first = result.current;

    rerender();

    expect(result.current).toBe(first);
  });
});
