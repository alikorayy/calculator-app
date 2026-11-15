import {
  useEffect,
  useId,
  useRef,
  type MouseEvent,
  type ReactNode,
  type SyntheticEvent,
} from 'react';

import { CLOSE_LABEL } from '../../constants/ui';
import { IconButton } from '../IconButton/IconButton';
import styles from './Dialog.module.css';

/** Props of {@link Dialog}. */
export interface DialogProps {
  /** Whether the dialog is shown. */
  open: boolean;
  /** Heading of the dialog; also its accessible name. */
  title: string;
  /** Called when the user asks to close it: close button, Escape or a click outside. */
  onClose: () => void;
  /** Extra controls in the header, before the close button (e.g. a "Clear" action). */
  headerActions?: ReactNode;
  /** The dialog's content. */
  children: ReactNode;
}

/**
 * A modal dialog built on the native `<dialog>` element. While it is open the
 * browser keeps focus inside it, makes the rest of the page inert and shows a
 * backdrop. Whether it is open is controlled entirely by the `open` prop.
 *
 * @param props - {@link DialogProps}
 * @returns The dialog element (hidden while closed).
 */
export function Dialog({ open, title, onClose, headerActions, children }: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  // Keep the native element in sync with the `open` prop.
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) {
      return;
    }
    if (open && !dialog.open) {
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
    }
  }, [open]);

  // Escape fires "cancel"; let the parent decide by updating `open`.
  const handleCancel = (event: SyntheticEvent<HTMLDialogElement>) => {
    event.preventDefault();
    onClose();
  };

  // The content fills the dialog, so a click whose target is the dialog itself
  // landed on the backdrop around it.
  const handleClick = (event: MouseEvent<HTMLDialogElement>) => {
    if (event.target === event.currentTarget) {
      onClose();
    }
  };

  return (
    // Clicking the backdrop is a pointer shortcut; keyboard users close the
    // dialog with Escape or the close button.
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-noninteractive-element-interactions
    <dialog
      ref={dialogRef}
      className={styles.dialog}
      aria-labelledby={titleId}
      onCancel={handleCancel}
      onClick={handleClick}
    >
      <div className={styles.content}>
        <header className={styles.header}>
          <h2 id={titleId} className={styles.title}>
            {title}
          </h2>
          {headerActions}
          <IconButton icon="close" label={CLOSE_LABEL} onClick={onClose} />
        </header>
        <div className={styles.body}>{children}</div>
      </div>
    </dialog>
  );
}
