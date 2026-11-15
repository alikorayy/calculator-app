import { describe, expect, it } from 'vitest';

import { RESULTS } from '../../../mocks/fixtures';
import type { CalculatorState } from '../types';
import {
  calculatorReducer,
  clearLabelFor,
  INITIAL_CALCULATOR_STATE,
  type CalculatorAction,
} from './calculatorReducer';

const EDITING: CalculatorState = { ...INITIAL_CALCULATOR_STATE, expression: '1+2' };

const SHOWING_RESULT: CalculatorState = {
  status: 'result',
  expression: '29',
  answer: { result: '29', approximate: false, expression: '50−7×3' },
  evaluatedExpression: '50−7×3',
  error: null,
};

const FAILED: CalculatorState = {
  ...EDITING,
  status: 'error',
  error: { message: 'Cannot divide by zero' },
};

const EVALUATING: CalculatorState = { ...EDITING, status: 'evaluating' };

function reduce(state: CalculatorState, ...actions: CalculatorAction[]): CalculatorState {
  return actions.reduce(calculatorReducer, state);
}

describe('calculatorReducer', () => {
  it('applies the input rules to keys', () => {
    expect(
      reduce(INITIAL_CALCULATOR_STATE, { type: 'input', key: '7' }, { type: 'input', key: 'add' }),
    ).toMatchObject({
      status: 'editing',
      expression: '7+',
    });
  });

  it('continues from a result with an operator', () => {
    expect(reduce(SHOWING_RESULT, { type: 'input', key: 'multiply' })).toMatchObject({
      status: 'editing',
      expression: 'Ans×',
      answer: SHOWING_RESULT.answer,
    });
  });

  it('deletes the last character, or the whole result', () => {
    expect(reduce(EDITING, { type: 'backspace' }).expression).toBe('1+');
    expect(reduce(SHOWING_RESULT, { type: 'backspace' })).toMatchObject({
      status: 'editing',
      expression: '',
    });
  });

  it('clears the entry with CE but keeps the answer', () => {
    const withAnswer = { ...EDITING, answer: SHOWING_RESULT.answer };

    expect(clearLabelFor(withAnswer)).toBe('CE');
    expect(reduce(withAnswer, { type: 'clear' })).toMatchObject({
      expression: '',
      answer: SHOWING_RESULT.answer,
    });
  });

  it('clears everything, including the answer, with AC', () => {
    expect(clearLabelFor(SHOWING_RESULT)).toBe('AC');
    expect(reduce(SHOWING_RESULT, { type: 'clear' })).toEqual(INITIAL_CALCULATOR_STATE);
    expect(clearLabelFor(INITIAL_CALCULATOR_STATE)).toBe('AC');
  });

  it('takes typed and restored text as the new expression', () => {
    expect(reduce(SHOWING_RESULT, { type: 'typed', expression: '2*3' })).toMatchObject({
      status: 'editing',
      expression: '2*3',
    });
    expect(reduce(EDITING, { type: 'restore', expression: '1÷3' }).expression).toBe('1÷3');
  });

  it('clears an error as soon as the expression is edited', () => {
    expect(reduce(FAILED, { type: 'input', key: '1' })).toMatchObject({
      status: 'editing',
      error: null,
    });
  });

  it('records a successful evaluation as the new answer', () => {
    const next = reduce(
      { ...EDITING, status: 'evaluating' },
      {
        type: 'evaluationSucceeded',
        shownExpression: '1÷3',
        sentExpression: '1÷3',
        response: RESULTS.rounded,
      },
    );

    expect(next).toEqual({
      status: 'result',
      expression: RESULTS.rounded.result,
      answer: { ...RESULTS.rounded, expression: '1÷3' },
      evaluatedExpression: '1÷3',
      error: null,
    });
  });

  it('keeps the expression for correction when the evaluation fails', () => {
    const error = { message: 'Unexpected "*"', position: 2 };

    expect(reduce(EVALUATING, { type: 'evaluationFailed', error })).toEqual({
      ...EDITING,
      status: 'error',
      error,
    });
  });

  it('returns to editing when the evaluation is cancelled', () => {
    expect(reduce(EVALUATING, { type: 'evaluationCancelled' }).status).toBe('editing');
  });

  it.each<CalculatorAction>([
    { type: 'input', key: '1' },
    { type: 'backspace' },
    { type: 'clear' },
    { type: 'typed', expression: '9' },
    { type: 'restore', expression: '9' },
  ])('ignores $type while a request is in flight', (action) => {
    expect(calculatorReducer(EVALUATING, action)).toBe(EVALUATING);
  });
});
