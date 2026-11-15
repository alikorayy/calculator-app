import { useEffect, useId, useRef, type KeyboardEvent, type ReactNode } from 'react';

import { LABELS } from '../../constants';
import { displaySizeFor } from '../../utils/display';
import styles from './Display.module.css';

/** Props of {@link Display}. */
export interface DisplayProps {
  /** The expression being typed, or the last result. */
  expression: string;
  /** The small line above the input. */
  topLine: string;
  /** Whether `topLine` is an error message. */
  hasError: boolean;
  /** Character position of a syntax error; the input highlights that character. */
  errorPosition?: number;
  /** Controls shown at the start of the top line (history, theme). */
  actions: ReactNode;
  /** Called when the user edits the text directly. */
  onExpressionChange: (expression: string) => void;
  /** Called when the user presses Enter in the input. */
  onSubmit: () => void;
}

/**
 * The calculator display: a top line for the previous answer, the expression
 * behind the result or an error, and an editable input for the expression.
 * The top line is a live region, so screen readers announce results and errors.
 *
 * @param props - {@link DisplayProps}
 * @returns The display panel.
 */
export function Display({
  expression,
  topLine,
  hasError,
  errorPosition,
  actions,
  onExpressionChange,
  onSubmit,
}: DisplayProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const topLineId = useId();

  // Select the character a syntax error points at, so the user sees what to fix.
  useEffect(() => {
    const input = inputRef.current;
    if (errorPosition === undefined || !input) {
      return;
    }
    input.focus();
    input.setSelectionRange(errorPosition, errorPosition + 1);
  }, [errorPosition, expression]);

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      onSubmit();
    }
  };

  return (
    <section className={styles.display}>
      <div className={styles.topRow}>
        {actions}
        <p id={topLineId} className={styles.topLine} data-error={hasError} aria-live="polite">
          {topLine}
        </p>
      </div>
      <input
        ref={inputRef}
        className={styles.input}
        data-size={displaySizeFor(expression.length)}
        data-error={hasError}
        value={expression}
        placeholder={LABELS.expressionPlaceholder}
        aria-label={LABELS.expression}
        aria-describedby={topLineId}
        aria-invalid={hasError}
        // The on-screen keypad replaces the phone keyboard; desktop typing still works.
        inputMode="none"
        autoComplete="off"
        autoCorrect="off"
        spellCheck={false}
        onChange={(event) => {
          onExpressionChange(event.target.value);
        }}
        onKeyDown={handleKeyDown}
      />
    </section>
  );
}
