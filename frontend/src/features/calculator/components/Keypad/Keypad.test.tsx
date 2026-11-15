import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { BASIC_KEYS, CLEAR_ARIA_LABELS, FUNCTION_KEYS, LABELS } from '../../constants';
import { Keypad } from './Keypad';

describe('Keypad', () => {
  it('renders every key of both pads in its group', () => {
    render(<Keypad clearLabel="AC" isEvaluating={false} onKey={vi.fn()} />);

    const functionPad = screen.getByRole('group', { name: LABELS.functionPad });
    const basicPad = screen.getByRole('group', { name: LABELS.basicPad });
    expect(within(functionPad).getAllByRole('button')).toHaveLength(FUNCTION_KEYS.length);
    expect(within(basicPad).getAllByRole('button')).toHaveLength(BASIC_KEYS.length);
  });

  it('reports the id of the pressed key', async () => {
    const user = userEvent.setup();
    const onKey = vi.fn();
    render(<Keypad clearLabel="AC" isEvaluating={false} onKey={onKey} />);

    await user.click(screen.getByRole('button', { name: '7' }));
    await user.click(screen.getByRole('button', { name: 'Multiply' }));
    await user.click(screen.getByRole('button', { name: 'Square root' }));
    await user.click(screen.getByRole('button', { name: 'Delete last character' }));

    expect(onKey.mock.calls).toEqual([['7'], ['multiply'], ['sqrt'], ['backspace']]);
  });

  it.each(['AC', 'CE'] as const)('labels the clear key %s', (label) => {
    render(<Keypad clearLabel={label} isEvaluating={false} onKey={vi.fn()} />);

    expect(screen.getByRole('button', { name: CLEAR_ARIA_LABELS[label] })).toHaveTextContent(label);
  });

  it('disables "=" while a calculation runs', () => {
    render(<Keypad clearLabel="AC" isEvaluating onKey={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Calculate' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '7' })).toBeEnabled();
  });
});
