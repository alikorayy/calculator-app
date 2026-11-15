import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { LABELS } from '../../constants';
import { Display, type DisplayProps } from './Display';

function renderDisplay(props: Partial<DisplayProps> = {}) {
  const handlers = { onExpressionChange: vi.fn(), onSubmit: vi.fn() };
  render(
    <Display
      expression="12+3"
      topLine="Ans = 0"
      hasError={false}
      actions={<button type="button">History</button>}
      {...handlers}
      {...props}
    />,
  );
  return handlers;
}

describe('Display', () => {
  it('shows the expression, the top line and the actions', () => {
    renderDisplay();

    expect(screen.getByRole('textbox', { name: LABELS.expression })).toHaveValue('12+3');
    expect(screen.getByText('Ans = 0')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'History' })).toBeInTheDocument();
  });

  it('describes the input with the top line, which is announced when it changes', () => {
    renderDisplay();

    const topLine = screen.getByText('Ans = 0');
    expect(topLine).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByRole('textbox')).toHaveAccessibleDescription('Ans = 0');
  });

  it('reports typing and Enter', async () => {
    const user = userEvent.setup();
    const { onExpressionChange, onSubmit } = renderDisplay({ expression: '' });

    await user.type(screen.getByRole('textbox'), '5');
    await user.keyboard('{Enter}');

    expect(onExpressionChange).toHaveBeenCalledWith('5');
    expect(onSubmit).toHaveBeenCalledOnce();
  });

  it('marks an error and selects the character it points at', () => {
    renderDisplay({
      expression: '2+*3',
      topLine: 'Unexpected "*"',
      hasError: true,
      errorPosition: 2,
    });

    const input = screen.getByRole<HTMLInputElement>('textbox');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(input).toHaveFocus();
    expect([input.selectionStart, input.selectionEnd]).toEqual([2, 3]);
  });

  it('does not open the phone keyboard, because the keypad replaces it', () => {
    renderDisplay();

    expect(screen.getByRole('textbox')).toHaveAttribute('inputmode', 'none');
  });
});
