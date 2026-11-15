import { describe, expect, it } from 'vitest';

import { isRecord } from './guards';

describe('isRecord', () => {
  it.each([
    [{}, true],
    [{ result: '29' }, true],
    [[], true],
    [null, false],
    [undefined, false],
    ['text', false],
    [42, false],
  ])('%j → %s', (value, expected) => {
    expect(isRecord(value)).toBe(expected);
  });
});
