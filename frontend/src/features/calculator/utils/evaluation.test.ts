import { describe, expect, it } from 'vitest';

import { HISTORY_ENTRIES } from '../../../mocks/fixtures';
import type { Answer } from '../types';
import {
  prepareForEvaluation,
  reusableExpression,
  reusableResult,
  substituteAnswer,
} from './evaluation';

const [scientific, rounded, exact, withAnswer] = HISTORY_ENTRIES;

describe('reusableResult', () => {
  it('reuses an exact plain result as it is', () => {
    expect(reusableResult(exact)).toBe(exact.result);
  });

  it('replaces a rounded result by its expression, so no precision is lost', () => {
    expect(reusableResult(rounded)).toBe(`(${rounded.sentExpression})`);
  });

  it('replaces a scientific result by its expression, which the backend can read', () => {
    expect(reusableResult(scientific)).toBe(`(${scientific.sentExpression})`);
  });

  it('uses the sent expression, because "Ans" may mean another value by now', () => {
    expect(reusableResult(withAnswer)).toBe(`(${withAnswer.sentExpression})`);
  });
});

describe('reusableExpression', () => {
  it('restores an expression as the user saw it', () => {
    expect(reusableExpression(exact)).toBe(exact.expression);
  });

  it('restores an expression with "Ans" as it was sent', () => {
    expect(reusableExpression(withAnswer)).toBe(withAnswer.sentExpression);
  });
});

describe('prepareForEvaluation', () => {
  it.each([
    ['', null],
    ['   ', null],
    ['2+', null],
    ['2×', null],
    ['2-', null],
    ['(', null],
    ['√(', null],
    ['2', '2'],
    [' 2 ', '2'],
    ['50−7×3', '50−7×3'],
    ['(2+3', '(2+3)'],
    ['√(√(16', '√(√(16))'],
    ['2^(−1', '2^(−1)'],
  ])('%j → %j', (expression, expected) => {
    expect(prepareForEvaluation(expression)).toBe(expected);
  });
});

describe('substituteAnswer', () => {
  const exact: Answer = { result: '29', approximate: false, expression: '50−7×3' };
  const negative: Answer = { result: '-5', approximate: false, expression: '2−7' };
  const rounded: Answer = { result: '0.3333333333333333', approximate: true, expression: '1/3' };
  const scientific: Answer = { result: '1e+20', approximate: false, expression: '10^20' };

  it.each<[string, Answer | null, string]>([
    ['2+3', exact, '2+3'],
    ['Ans+1', null, 'Ans+1'],
    ['Ans+1', exact, '(29)+1'],
    ['Ans^2', negative, '(-5)^2'],
    ['Ans×3', rounded, '(1/3)×3'],
    ['Ans×2', scientific, '(10^20)×2'],
    ['Ans×Ans', exact, '(29)×(29)'],
  ])('%s with %j → %s', (expression, answer, expected) => {
    expect(substituteAnswer(expression, answer)).toBe(expected);
  });
});
