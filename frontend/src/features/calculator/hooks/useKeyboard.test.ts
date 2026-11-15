import { fireEvent, renderHook } from '@testing-library/react';
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

  describe('in a text field', () => {
    /** Adds a focused text input with the caret placed at `caret`, and removes it after the test. */
    function focusedInput(value: string, caret: [number, number]) {
      const input = document.createElement('input');
      input.value = value;
      document.body.append(input);
      input.focus();
      input.setSelectionRange(...caret);
      return input;
    }

    it('handles keys typed at the end as calculator keys, like the keypad', async () => {
      const user = userEvent.setup();
      const onKey = vi.fn();
      const input = focusedInput('12', [2, 2]);
      renderHook(() => {
        useKeyboard(onKey);
      });

      await user.keyboard('5%');

      expect(onKey.mock.calls).toEqual([['5'], ['percent']]);
      expect(input.value).toBe('12');
      input.remove();
    });

    it.each([
      ['the caret is in the middle', [1, 1]],
      ['part of the text is selected', [0, 2]],
    ] as const)('leaves editing to the field when %s', async (_, caret) => {
      const user = userEvent.setup();
      const onKey = vi.fn();
      const input = focusedInput('12', [...caret]);
      renderHook(() => {
        useKeyboard(onKey);
      });

      await user.keyboard('5');

      expect(onKey).not.toHaveBeenCalled();
      expect(input.value).not.toBe('12');
      input.remove();
    });

    it('ignores a key that another handler already dealt with', async () => {
      const user = userEvent.setup();
      const onKey = vi.fn();
      const input = focusedInput('', [0, 0]);
      input.addEventListener('keydown', (event) => {
        event.preventDefault();
      });
      renderHook(() => {
        useKeyboard(onKey);
      });

      await user.keyboard('{Enter}');

      expect(onKey).not.toHaveBeenCalled();
      input.remove();
    });
  });

  it('ignores keys while an input method is composing text', () => {
    const onKey = vi.fn();
    renderHook(() => {
      useKeyboard(onKey);
    });

    fireEvent.keyDown(window, { key: '7', isComposing: true });

    expect(onKey).not.toHaveBeenCalled();
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
