import { act, renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';

import { ERRORS, EXPRESSIONS, RESULTS } from '../../../mocks/fixtures';
import { server } from '../../../mocks/server';
import { EVALUATE_PATH, HTTP_STATUS } from '../../../shared/constants/api';
import { TOP_LINE } from '../constants';
import type { KeyId } from '../types';
import { useCalculator } from './useCalculator';

/** Presses the keys one by one, as a user would. */
function press(result: { current: ReturnType<typeof useCalculator> }, keys: KeyId[]) {
  for (const key of keys) {
    act(() => {
      result.current.pressKey(key);
    });
  }
}

/** Keys that build EXPRESSIONS.exact, "50−7×3". */
const EXACT_KEYS: KeyId[] = ['5', '0', 'subtract', '7', 'multiply', '3'];

describe('useCalculator', () => {
  it('starts empty, showing "Ans = 0"', () => {
    const { result } = renderHook(() => useCalculator());

    expect(result.current.state.expression).toBe('');
    expect(result.current.topLine).toBe(`${TOP_LINE.answerPrefix}${TOP_LINE.noAnswer}`);
    expect(result.current.clearLabel).toBe('AC');
  });

  it('evaluates on "=", shows the result and reports it', async () => {
    const onResult = vi.fn();
    const { result } = renderHook(() => useCalculator({ onResult }));

    press(result, [...EXACT_KEYS, 'equals']);
    expect(result.current.isEvaluating).toBe(true);
    expect(result.current.topLine).toBe(TOP_LINE.calculating);

    await waitFor(() => {
      expect(result.current.state.status).toBe('result');
    });
    expect(result.current.state.expression).toBe(RESULTS.exact.result);
    expect(result.current.topLine).toBe(`${EXPRESSIONS.exact} ${TOP_LINE.equals}`);
    expect(onResult).toHaveBeenCalledWith({
      expression: EXPRESSIONS.exact,
      result: RESULTS.exact.result,
      approximate: RESULTS.exact.approximate,
      sentExpression: EXPRESSIONS.exact,
    });
  });

  it('does nothing on "=" while a result is on display', async () => {
    const onResult = vi.fn();
    const { result } = renderHook(() => useCalculator({ onResult }));
    press(result, [...EXACT_KEYS, 'equals']);
    await waitFor(() => {
      expect(result.current.state.status).toBe('result');
    });

    press(result, ['equals']);

    expect(result.current.isEvaluating).toBe(false);
    expect(result.current.state.expression).toBe(RESULTS.exact.result);
    expect(onResult).toHaveBeenCalledTimes(1);
  });

  it('marks a rounded result with ≈ and reuses its exact expression for Ans', async () => {
    let lastSent = '';
    server.use(
      http.post(EVALUATE_PATH, async ({ request }) => {
        const body = (await request.clone().json()) as { expression: string };
        lastSent = body.expression;
        return undefined;
      }),
    );
    const { result } = renderHook(() => useCalculator());

    press(result, ['1', 'divide', '3', 'equals']);
    await waitFor(() => {
      expect(result.current.state.status).toBe('result');
    });
    expect(result.current.topLine).toBe(`${EXPRESSIONS.rounded} ${TOP_LINE.approximately}`);

    press(result, ['multiply', '3']);
    expect(result.current.state.expression).toBe('Ans×3');

    press(result, ['equals']);
    await waitFor(() => {
      expect(result.current.state.expression).toBe(RESULTS.one.result);
    });
    expect(lastSent).toBe(EXPRESSIONS.roundedTimesThree);
  });

  it('shows a calculation error and keeps its position for the display', async () => {
    const { result } = renderHook(() => useCalculator());

    act(() => {
      result.current.typeExpression(EXPRESSIONS.syntaxError);
    });
    press(result, ['equals']);

    await waitFor(() => {
      expect(result.current.state.status).toBe('error');
    });
    expect(result.current.topLine).toBe(ERRORS.syntaxError.error.message);
    expect(result.current.state.error?.position).toBe(ERRORS.syntaxError.error.position);
    expect(result.current.state.expression).toBe(EXPRESSIONS.syntaxError);
  });

  it('drops the error position when Ans was replaced, because it would not match the display', async () => {
    const { result } = renderHook(() => useCalculator());
    press(result, [...EXACT_KEYS, 'equals']);
    await waitFor(() => {
      expect(result.current.state.status).toBe('result');
    });

    server.use(
      http.post(EVALUATE_PATH, () =>
        HttpResponse.json(ERRORS.syntaxError, { status: HTTP_STATUS.UNPROCESSABLE_ENTITY }),
      ),
    );
    press(result, ['add', '1', 'equals']);

    await waitFor(() => {
      expect(result.current.state.status).toBe('error');
    });
    expect(result.current.state.error?.position).toBeUndefined();
  });

  it('shows the user-facing message for transport errors', async () => {
    server.use(http.post(EVALUATE_PATH, () => HttpResponse.error()));
    const { result } = renderHook(() => useCalculator());

    press(result, ['1', 'equals']);

    await waitFor(() => {
      expect(result.current.state.status).toBe('error');
    });
    expect(result.current.state.error?.position).toBeUndefined();
  });

  it('does nothing on "=" when the expression is not complete', () => {
    const { result } = renderHook(() => useCalculator());

    press(result, ['1', 'add', 'equals']);

    expect(result.current.state.status).toBe('editing');
    expect(result.current.state.expression).toBe('1+');
  });

  it('handles backspace, clear and restoring a history entry', () => {
    const { result } = renderHook(() => useCalculator());

    press(result, ['1', '2', 'backspace']);
    expect(result.current.state.expression).toBe('1');
    expect(result.current.clearLabel).toBe('CE');

    press(result, ['clear']);
    expect(result.current.state.expression).toBe('');

    act(() => {
      result.current.restoreExpression(EXPRESSIONS.rounded);
    });
    expect(result.current.state.expression).toBe(EXPRESSIONS.rounded);
  });
});
