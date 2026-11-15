import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { Key } from './Key';

describe('Key', () => {
  it('shows its label and uses the spoken name when there is one', () => {
    render(
      <Key
        definition={{ id: 'multiply', label: '×', ariaLabel: 'Multiply', variant: 'operator' }}
        onPress={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: 'Multiply' })).toHaveTextContent('×');
  });

  it('uses the label as its name when there is no spoken name', () => {
    render(<Key definition={{ id: '7', label: '7', variant: 'number' }} onPress={vi.fn()} />);

    expect(screen.getByRole('button', { name: '7' })).toBeInTheDocument();
  });

  it('shows an icon instead of the label when it has one', () => {
    render(
      <Key
        definition={{
          id: 'backspace',
          label: '',
          icon: 'backspace',
          ariaLabel: 'Delete',
          variant: 'function',
        }}
        onPress={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: 'Delete' }).querySelector('svg')).toBeInTheDocument();
  });

  it('calls onPress, unless disabled', async () => {
    const user = userEvent.setup();
    const onPress = vi.fn();
    const { rerender } = render(
      <Key definition={{ id: 'equals', label: '=', variant: 'equals' }} onPress={onPress} />,
    );

    await user.click(screen.getByRole('button', { name: '=' }));
    expect(onPress).toHaveBeenCalledOnce();

    rerender(
      <Key
        definition={{ id: 'equals', label: '=', variant: 'equals' }}
        onPress={onPress}
        disabled
      />,
    );
    await user.click(screen.getByRole('button', { name: '=' }));
    expect(onPress).toHaveBeenCalledOnce();
  });
});
