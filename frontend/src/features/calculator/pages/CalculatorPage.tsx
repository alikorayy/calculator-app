import { useState } from 'react';

import { IconButton } from '../../../shared/components/IconButton/IconButton';
import { useTheme } from '../../../shared/hooks/useTheme';
import { Brand } from '../components/Brand/Brand';
import { Display } from '../components/Display/Display';
import { HistoryDialog } from '../components/HistoryDialog/HistoryDialog';
import { Keypad } from '../components/Keypad/Keypad';
import { ThemeToggle } from '../components/ThemeToggle/ThemeToggle';
import { LABELS } from '../constants';
import { useCalculator } from '../hooks/useCalculator';
import { useHistory } from '../hooks/useHistory';
import { useKeyboard } from '../hooks/useKeyboard';
import type { HistoryEntry } from '../types';
import { reusableResult } from '../utils/evaluation';
import styles from './CalculatorPage.module.css';

/**
 * The calculator page. It is the only place that knows both the hooks and the
 * components: hooks provide state and actions, and the page hands them to
 * presentational components as props (data down, events up).
 *
 * @returns The page.
 */
export function CalculatorPage() {
  const history = useHistory();
  const calculator = useCalculator({ onResult: history.addEntry });
  const { theme, toggleTheme } = useTheme();
  const [historyOpen, setHistoryOpen] = useState(false);

  // While the history dialog is open, the keyboard belongs to it.
  useKeyboard(calculator.pressKey, !historyOpen);

  const pickFromHistory = (entry: HistoryEntry, part: 'expression' | 'result') => {
    calculator.restoreExpression(part === 'expression' ? entry.expression : reusableResult(entry));
    setHistoryOpen(false);
  };

  const { state } = calculator;
  return (
    <main className={styles.page}>
      <div className={styles.content}>
        <Brand />
        <Display
          expression={state.expression}
          topLine={calculator.topLine}
          hasError={state.error !== null}
          errorPosition={state.error?.position}
          onExpressionChange={calculator.typeExpression}
          onSubmit={() => {
            calculator.pressKey('equals');
          }}
          actions={
            <>
              <IconButton
                icon="history"
                label={LABELS.history}
                onClick={() => {
                  setHistoryOpen(true);
                }}
              />
              <ThemeToggle theme={theme} onToggle={toggleTheme} />
            </>
          }
        />
        <Keypad
          clearLabel={calculator.clearLabel}
          isEvaluating={calculator.isEvaluating}
          onKey={calculator.pressKey}
        />
      </div>
      <HistoryDialog
        open={historyOpen}
        entries={history.entries}
        onClose={() => {
          setHistoryOpen(false);
        }}
        onClear={history.clearHistory}
        onPick={pickFromHistory}
      />
    </main>
  );
}
