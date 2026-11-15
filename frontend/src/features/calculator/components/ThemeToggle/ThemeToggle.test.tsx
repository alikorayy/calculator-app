import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { LABELS } from '../../constants';
import { ThemeToggle } from './ThemeToggle';

describe('ThemeToggle', () => {
  it.each([
    ['dark', LABELS.switchToLight],
    ['light', LABELS.switchToDark],
  ] as const)('in the %s theme offers to switch to the other one', (theme, label) => {
    render(<ThemeToggle theme={theme} onToggle={vi.fn()} />);

    expect(screen.getByRole('button', { name: label })).toBeInTheDocument();
  });

  it('calls onToggle', async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    render(<ThemeToggle theme="dark" onToggle={onToggle} />);

    await user.click(screen.getByRole('button'));

    expect(onToggle).toHaveBeenCalledOnce();
  });
});
