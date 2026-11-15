import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { CLOSE_LABEL } from '../../constants/ui';
import { Dialog } from './Dialog';

const TITLE = 'History';

function renderDialog(open: boolean, onClose = vi.fn()) {
  const view = render(
    <Dialog
      open={open}
      title={TITLE}
      onClose={onClose}
      headerActions={<button type="button">Clear</button>}
    >
      <p>Dialog content</p>
    </Dialog>,
  );
  return { ...view, onClose };
}

describe('Dialog', () => {
  it('is not shown while closed', () => {
    renderDialog(false);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('shows its title as the accessible name, the header actions and the content when open', () => {
    renderDialog(true);

    expect(screen.getByRole('dialog', { name: TITLE })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: TITLE })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Clear' })).toBeInTheDocument();
    expect(screen.getByText('Dialog content')).toBeInTheDocument();
  });

  it('opens and closes when the open prop changes', () => {
    const { rerender, onClose } = renderDialog(false);

    rerender(
      <Dialog open title={TITLE} onClose={onClose}>
        content
      </Dialog>,
    );
    expect(screen.getByRole('dialog')).toBeInTheDocument();

    rerender(
      <Dialog open={false} title={TITLE} onClose={onClose}>
        content
      </Dialog>,
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('asks to close from the close button', async () => {
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);

    await user.click(screen.getByRole('button', { name: CLOSE_LABEL }));

    expect(onClose).toHaveBeenCalledOnce();
  });

  it('asks to close on Escape (the cancel event) instead of closing by itself', () => {
    const { onClose } = renderDialog(true);
    const dialog = screen.getByRole('dialog');

    const cancel = new Event('cancel', { cancelable: true });
    fireEvent(dialog, cancel);

    expect(onClose).toHaveBeenCalledOnce();
    expect(cancel.defaultPrevented).toBe(true);
  });

  it('asks to close on a backdrop click but not on a click inside', async () => {
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);

    await user.click(screen.getByText('Dialog content'));
    expect(onClose).not.toHaveBeenCalled();

    await user.click(screen.getByRole('dialog'));
    expect(onClose).toHaveBeenCalledOnce();
  });
});
