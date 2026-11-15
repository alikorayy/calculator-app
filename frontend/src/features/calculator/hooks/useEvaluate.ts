import { useCallback, useEffect, useRef } from 'react';

import { evaluateExpression } from '../../../shared/api/calculatorApi';
import type { ApiResult } from '../../../shared/api/errors';
import type { CalculationResponse } from '../../../shared/api/types';

/**
 * Connects the API client to React. It returns a function that evaluates an
 * expression on the server and guarantees that only the latest request can
 * produce a result: starting a request cancels the previous one, and leaving
 * the page cancels whatever is still in flight. A cancelled request resolves
 * with an `aborted` error, which callers ignore.
 *
 * @returns A stable `evaluate(expression)` function.
 */
export function useEvaluate(): (expression: string) => Promise<ApiResult<CalculationResponse>> {
  // A ref, not state: changing the in-flight request must not re-render anything.
  const inFlight = useRef<AbortController | null>(null);

  useEffect(
    () => () => {
      inFlight.current?.abort();
    },
    [],
  );

  return useCallback(async (expression: string) => {
    inFlight.current?.abort();
    const controller = new AbortController();
    inFlight.current = controller;

    // The single place where the calculator reaches the backend.
    const result = await evaluateExpression(expression, { signal: controller.signal });

    if (inFlight.current === controller) {
      inFlight.current = null;
    }
    return result;
  }, []);
}
