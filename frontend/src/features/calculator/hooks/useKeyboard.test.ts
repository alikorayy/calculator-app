import { renderHook } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { useKeyboard } from './useKeyboard';

describe('useKeyboard', () => {
  it('maps keys to calculator keys', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    renderHook(() => {
      useKeyboard(onKey);
    });

    await user.keyboard('7*{Enter}{Backspace}{Escape}');

    expect(onKey.mock.calls).toEqual([['7'], ['multiply'], ['equals'], ['backspace'], ['clear']]);
  });

  it('ignores unrelated keys and shortcuts with modifiers', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    renderHook(() => {
      useKeyboard(onKey);
    });

    await user.keyboard('a{Shift}{Control>}c{/Control}');

    expect(onKey).not.toHaveBeenCalled();
  });

  it('leaves typing in a text field to the field', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    const input = document.createElement('input');
    document.body.append(input);
    renderHook(() => {
      useKeyboard(onKey);
    });

    await user.click(input);
    await user.keyboard('12');

    expect(onKey).not.toHaveBeenCalled();
    input.remove();
  });

  it('lets Enter activate a focused button instead of meaning "="', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    const button = document.createElement('button');
    document.body.append(button);
    renderHook(() => {
      useKeyboard(onKey);
    });

    button.focus();
    await user.keyboard('{Enter}');

    expect(onKey).not.toHaveBeenCalled();
    button.remove();
  });

  it('does nothing while disabled', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    renderHook(() => {
      useKeyboard(onKey, false);
    });

    await user.keyboard('7');

    expect(onKey).not.toHaveBeenCalled();
  });

  it('always calls the latest callback', async () => {
    const user = userEvent.setup();
    const first = vi.fn();
    const second = vi.fn();
    const { rerender } = renderHook(
      ({ onKey }) => {
        useKeyboard(onKey);
      },
      { initialProps: { onKey: first } },
    );

    rerender({ onKey: second });
    await user.keyboard('7');

    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledWith('7');
  });
});
