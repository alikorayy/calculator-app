import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent, { type UserEvent } from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import { ERRORS, EXPRESSIONS, RESULTS } from '../../../mocks/fixtures';
import { CLOSE_LABEL } from '../../../shared/constants/ui';
import { LABELS } from '../constants';
import { CalculatorPage } from './CalculatorPage';

/** Clicks keypad keys by their accessible names. */
async function click(user: UserEvent, ...names: string[]) {
  for (const name of names) {
    await user.click(screen.getByRole('button', { name }));
  }
}

function display() {
  return screen.getByRole('textbox', { name: LABELS.expression });
}

describe('CalculatorPage', () => {
  it('calculates with the keypad and shows the result with its expression', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await click(user, '5', '0', 'Subtract', '7', 'Multiply', '3');
    expect(display()).toHaveValue(EXPRESSIONS.exact);

    await click(user, 'Calculate');

    await waitFor(() => {
      expect(display()).toHaveValue(RESULTS.exact.result);
    });
    // The top line describes the display, which is what a screen reader reads with it.
    expect(display()).toHaveAccessibleDescription(`${EXPRESSIONS.exact} =`);
  });

  it('works with the keyboard and keeps chained calculations exact', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await user.keyboard('1/3{Enter}');
    await waitFor(() => {
      expect(display()).toHaveValue(RESULTS.rounded.result);
    });
    expect(display()).toHaveAccessibleDescription(`${EXPRESSIONS.rounded} ≈`);

    await user.keyboard('*3{Enter}');
    await waitFor(() => {
      expect(display()).toHaveValue(RESULTS.one.result);
    });
  });

  it('shows calculation errors and marks the display', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await click(user, '5', 'Divide', '0', 'Calculate');

    await waitFor(() => {
      expect(display()).toHaveAccessibleDescription(ERRORS.divisionByZero.error.message);
    });
    expect(display()).toHaveAttribute('aria-invalid', 'true');
  });

  it('keeps finished calculations in the history and reuses a picked result', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);
    await click(user, '5', '0', 'Subtract', '7', 'Multiply', '3', 'Calculate');
    await waitFor(() => {
      expect(display()).toHaveValue(RESULTS.exact.result);
    });
    await click(user, 'All clear');

    await click(user, LABELS.history);
    const dialog = screen.getByRole('dialog', { name: LABELS.history });
    await user.click(within(dialog).getByRole('button', { name: RESULTS.exact.result }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(display()).toHaveValue(RESULTS.exact.result);
  });

  it('calculates an expression typed into the display when Enter is pressed', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await user.type(display(), `${EXPRESSIONS.exact}{Enter}`);

    await waitFor(() => {
      expect(display()).toHaveValue(RESULTS.exact.result);
    });
  });

  it('closes the history with its close button', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await click(user, LABELS.history);
    await click(user, CLOSE_LABEL);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('ignores the keyboard while the history is open', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);

    await click(user, LABELS.history);
    await user.keyboard('7');

    expect(display()).toHaveValue('');
  });

  it('switches the theme', async () => {
    const user = userEvent.setup();
    render(<CalculatorPage />);
    expect(document.documentElement.dataset.theme).toBe('light');

    await click(user, LABELS.switchToDark);

    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(screen.getByRole('button', { name: LABELS.switchToLight })).toBeInTheDocument();
  });
});
