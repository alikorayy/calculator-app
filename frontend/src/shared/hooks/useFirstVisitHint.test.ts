import { act, renderHook } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import { useFirstVisitHint } from './useFirstVisitHint';

const KEY = 'test.hintSeen';

describe('useFirstVisitHint', () => {
  it('is visible on the first visit', () => {
    const { result } = renderHook(() => useFirstVisitHint(KEY));

    expect(result.current.visible).toBe(true);
  });

  it.each([
    [
      'a click anywhere',
      async (user: ReturnType<typeof userEvent.setup>) => user.click(document.body),
    ],
    ['a key press', async (user: ReturnType<typeof userEvent.setup>) => user.keyboard('a')],
  ])('goes away after %s and stays away', async (_, interact) => {
    const user = userEvent.setup();
    const { result } = renderHook(() => useFirstVisitHint(KEY));

    await act(async () => {
      await interact(user);
    });

    expect(result.current.visible).toBe(false);
    expect(window.localStorage.getItem(KEY)).toBe('true');
  });

  it('does not come back on the next visit', () => {
    window.localStorage.setItem(KEY, 'true');

    const { result } = renderHook(() => useFirstVisitHint(KEY));

    expect(result.current.visible).toBe(false);
  });

  it('comes back when the stored value is not a boolean', () => {
    window.localStorage.setItem(KEY, '"yes"');

    const { result } = renderHook(() => useFirstVisitHint(KEY));

    expect(result.current.visible).toBe(true);
  });

  it('stops listening once it has been seen', async () => {
    const user = userEvent.setup();
    const { result } = renderHook(() => useFirstVisitHint(KEY));
    await act(async () => {
      await user.click(document.body);
    });
    window.localStorage.setItem(KEY, 'false');

    await act(async () => {
      await user.click(document.body);
    });

    // The listeners were removed, so this click no longer writes anything.
    expect(window.localStorage.getItem(KEY)).toBe('false');
    expect(result.current.visible).toBe(false);
  });
});
