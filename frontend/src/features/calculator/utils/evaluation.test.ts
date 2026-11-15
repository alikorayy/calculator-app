import { describe, expect, it } from 'vitest';

import type { Answer } from '../types';
import { prepareForEvaluation, substituteAnswer } from './evaluation';

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
