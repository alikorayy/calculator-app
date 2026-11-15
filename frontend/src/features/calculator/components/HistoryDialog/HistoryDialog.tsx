import { Dialog } from '../../../../shared/components/Dialog/Dialog';
import { LABELS, TOP_LINE } from '../../constants';
import type { HistoryEntry } from '../../types';
import styles from './HistoryDialog.module.css';

/** Props of {@link HistoryDialog}. */
export interface HistoryDialogProps {
  open: boolean;
  /** Calculations to show, newest first. */
  entries: readonly HistoryEntry[];
  onClose: () => void;
  onClear: () => void;
  /** Called when the user picks an entry's expression or its result to reuse. */
  onPick: (entry: HistoryEntry, part: 'expression' | 'result') => void;
}

/**
 * Lists finished calculations. Each expression and each result is a button,
 * so either can be put back into the display with a click or the keyboard.
 *
 * @param props - {@link HistoryDialogProps}
 * @returns The history dialog.
 */
export function HistoryDialog({ open, entries, onClose, onClear, onPick }: HistoryDialogProps) {
  const clearButton =
    entries.length > 0 ? (
      <button type="button" className={styles.clear} onClick={onClear}>
        {LABELS.clearHistory}
      </button>
    ) : null;

  return (
    <Dialog open={open} title={LABELS.history} onClose={onClose} headerActions={clearButton}>
      {entries.length > 0 ? (
        <ul className={styles.list}>
          {entries.map((entry) => (
            <li key={entry.id} className={styles.entry}>
              <button
                type="button"
                className={styles.expression}
                onClick={() => {
                  onPick(entry, 'expression');
                }}
              >
                {entry.expression} {TOP_LINE.equals}
              </button>
              <button
                type="button"
                className={styles.result}
                onClick={() => {
                  onPick(entry, 'result');
                }}
              >
                {entry.result}
              </button>
            </li>
          ))}
        </ul>
      ) : (
        <p className={styles.empty}>{LABELS.historyEmpty}</p>
      )}
    </Dialog>
  );
}
