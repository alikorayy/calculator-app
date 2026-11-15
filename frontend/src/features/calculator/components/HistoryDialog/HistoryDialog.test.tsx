import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { HISTORY_ENTRIES } from '../../../../mocks/fixtures';
import { LABELS } from '../../constants';
import { HistoryDialog, type HistoryDialogProps } from './HistoryDialog';

function renderHistory(props: Partial<HistoryDialogProps> = {}) {
  const handlers = { onClose: vi.fn(), onClear: vi.fn(), onPick: vi.fn() };
  render(<HistoryDialog open entries={HISTORY_ENTRIES} {...handlers} {...props} />);
  return handlers;
}

describe('HistoryDialog', () => {
  it('lists every entry as an expression and a result button', () => {
    renderHistory();

    for (const entry of HISTORY_ENTRIES) {
      expect(screen.getByRole('button', { name: `${entry.expression} =` })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: entry.result })).toBeInTheDocument();
    }
  });

  it('reports which part of which entry was picked', async () => {
    const user = userEvent.setup();
    const { onPick } = renderHistory();
    const [first] = HISTORY_ENTRIES;

    await user.click(screen.getByRole('button', { name: `${first.expression} =` }));
    await user.click(screen.getByRole('button', { name: first.result }));

    expect(onPick.mock.calls).toEqual([
      [first, 'expression'],
      [first, 'result'],
    ]);
  });

  it('can be cleared', async () => {
    const user = userEvent.setup();
    const { onClear } = renderHistory();

    await user.click(screen.getByRole('button', { name: LABELS.clearHistory }));

    expect(onClear).toHaveBeenCalledOnce();
  });

  it('explains the empty state and offers no clear button', () => {
    renderHistory({ entries: [] });

    expect(screen.getByText(LABELS.historyEmpty)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: LABELS.clearHistory })).not.toBeInTheDocument();
  });

  it('is hidden while closed', () => {
    renderHistory({ open: false });

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });
});
