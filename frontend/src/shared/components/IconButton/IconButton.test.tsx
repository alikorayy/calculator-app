import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { IconButton } from './IconButton';

describe('IconButton', () => {
  it('uses the label as its accessible name and tooltip', () => {
    render(<IconButton icon="history" label="History" onClick={vi.fn()} />);

    const button = screen.getByRole('button', { name: 'History' });
    expect(button).toHaveAttribute('title', 'History');
  });

  it('calls onClick when clicked or activated with the keyboard', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<IconButton icon="close" label="Close" onClick={onClick} />);

    await user.click(screen.getByRole('button', { name: 'Close' }));
    await user.keyboard('{Enter}');
    await user.keyboard(' ');

    expect(onClick).toHaveBeenCalledTimes(3);
  });
});
