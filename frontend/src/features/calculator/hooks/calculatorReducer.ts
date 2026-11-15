import type { CalculationResponse } from '../../../shared/api/types';
import type { CalculatorError, CalculatorState, ClearLabel, InputKey } from '../types';
import { applyInputKey, deleteLast } from '../utils/expressionInput';

/** Everything that can happen to the calculator. */
export type CalculatorAction =
  | { type: 'input'; key: InputKey }
  | { type: 'backspace' }
  | { type: 'clear' }
  | { type: 'typed'; expression: string }
  | { type: 'restore'; expression: string }
  | { type: 'evaluationStarted' }
  | {
      type: 'evaluationSucceeded';
      shownExpression: string;
      sentExpression: string;
      response: CalculationResponse;
    }
  | { type: 'evaluationFailed'; error: CalculatorError }
  | { type: 'evaluationCancelled' };

/** The calculator when the page opens. */
export const INITIAL_CALCULATOR_STATE: CalculatorState = {
  status: 'editing',
  expression: '',
  answer: null,
  evaluatedExpression: null,
  error: null,
};

/**
 * Computes the next calculator state. It is a pure function of the current
 * state and an action, so every transition can be tested without React.
 * While a request is in flight, editing actions are ignored so a late answer
 * never overwrites what the user typed.
 *
 * @param state - Current state.
 * @param action - What happened.
 * @returns The next state.
 */
export function calculatorReducer(
  state: CalculatorState,
  action: CalculatorAction,
): CalculatorState {
  if (state.status === 'evaluating' && isEditingAction(action)) {
    return state;
  }

  switch (action.type) {
    case 'input':
      return editing(
        state,
        applyInputKey(state.expression, action.key, {
          showingResult: state.status === 'result',
          hasAnswer: state.answer !== null,
        }),
      );
    case 'backspace':
      return editing(state, state.status === 'result' ? '' : deleteLast(state.expression));
    case 'clear':
      return clearLabelFor(state) === 'AC' ? INITIAL_CALCULATOR_STATE : editing(state, '');
    case 'typed':
    case 'restore':
      return editing(state, action.expression);
    case 'evaluationStarted':
      return { ...state, status: 'evaluating', error: null };
    case 'evaluationSucceeded':
      return {
        status: 'result',
        expression: action.response.result,
        answer: {
          result: action.response.result,
          approximate: action.response.approximate,
          expression: action.sentExpression,
        },
        evaluatedExpression: action.shownExpression,
        error: null,
      };
    case 'evaluationFailed':
      return { ...state, status: 'error', error: action.error };
    case 'evaluationCancelled':
      return { ...state, status: 'editing' };
  }
}

/**
 * The clear key shows "AC" (all clear: also forgets Ans) when there is nothing
 * being typed or a result is on display, and "CE" (clear entry) otherwise.
 *
 * @param state - Current state.
 * @returns The label, which also decides what the key does.
 */
export function clearLabelFor(state: CalculatorState): ClearLabel {
  return state.status === 'result' || state.expression === '' ? 'AC' : 'CE';
}

function editing(state: CalculatorState, expression: string): CalculatorState {
  return { ...state, status: 'editing', expression, error: null };
}

function isEditingAction(action: CalculatorAction): boolean {
  return ['input', 'backspace', 'clear', 'typed', 'restore'].includes(action.type);
}
