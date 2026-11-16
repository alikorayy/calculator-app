import { useReducer } from 'react';

import type { CalculatorState, ClearLabel, HistoryEntry, KeyId } from '../types';
import { prepareForEvaluation, substituteAnswer } from '../utils/evaluation';
import { topLineFor } from '../utils/display';
import { calculatorReducer, clearLabelFor, INITIAL_CALCULATOR_STATE } from './calculatorReducer';
import { useEvaluate } from './useEvaluate';

/** Options of {@link useCalculator}. */
export interface UseCalculatorOptions {
  /** Called after every successful calculation, e.g. to add it to the history. */
  onResult?: (entry: Omit<HistoryEntry, 'id'>) => void;
}

/** What {@link useCalculator} returns: the state to display and the actions to call. */
export interface CalculatorControls {
  state: CalculatorState;
  /** Text above the input: "Ans = …", "<expression> =", "Calculating…" or the error. */
  topLine: string;
  clearLabel: ClearLabel;
  isEvaluating: boolean;
  /** Handles a key from the keypad or the keyboard. */
  pressKey: (key: KeyId) => void;
  /** Handles free typing in the input. */
  typeExpression: (expression: string) => void;
  /** Puts an expression or result from the history back into the display. */
  restoreExpression: (expression: string) => void;
}

/**
 * The calculator's behaviour, independent of how it looks. Keys are applied by
 * the pure input rules; "=" prepares the expression, replaces "Ans", sends it
 * to the server through {@link useEvaluate} and records the outcome.
 *
 * @param options - Callback for finished calculations.
 * @returns State to render and actions for the components.
 */
export function useCalculator({ onResult }: UseCalculatorOptions = {}): CalculatorControls {
  const [state, dispatch] = useReducer(calculatorReducer, INITIAL_CALCULATOR_STATE);
  const evaluate = useEvaluate();

  const submit = async () => {
    // A result on display is already calculated; sending its rounded or
    // scientific text again would lose precision or fail to parse.
    if (state.status === 'evaluating' || state.status === 'result') {
      return;
    }
    const shownExpression = prepareForEvaluation(state.expression);
    if (shownExpression === null) {
      return;
    }
    const sentExpression = substituteAnswer(shownExpression, state.answer);

    dispatch({ type: 'evaluationStarted' });
    const result = await evaluate(sentExpression);

    if (result.ok) {
      dispatch({
        type: 'evaluationSucceeded',
        shownExpression,
        sentExpression,
        response: result.data,
      });
      onResult?.({
        expression: shownExpression,
        result: result.data.result,
        approximate: result.data.approximate,
        sentExpression,
      });
      return;
    }
    if (result.error.kind === 'aborted') {
      dispatch({ type: 'evaluationCancelled' });
      return;
    }
    // A syntax error position refers to the sent text; it matches the display
    // only when no "Ans" was replaced.
    const position =
      result.error.kind === 'calculation' && sentExpression === shownExpression
        ? result.error.position
        : undefined;
    dispatch({ type: 'evaluationFailed', error: { message: result.error.message, position } });
  };

  const pressKey = (key: KeyId) => {
    switch (key) {
      case 'equals':
        void submit();
        return;
      case 'backspace':
        dispatch({ type: 'backspace' });
        return;
      case 'clear':
        dispatch({ type: 'clear' });
        return;
      default:
        dispatch({ type: 'input', key });
    }
  };

  return {
    state,
    topLine: topLineFor(state),
    clearLabel: clearLabelFor(state),
    isEvaluating: state.status === 'evaluating',
    pressKey,
    typeExpression: (expression) => {
      dispatch({ type: 'typed', expression });
    },
    restoreExpression: (expression) => {
      dispatch({ type: 'restore', expression });
    },
  };
}
