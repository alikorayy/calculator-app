import { useState } from 'react';

import { IconButton } from '../../../shared/components/IconButton/IconButton';
import { useFirstVisitHint } from '../../../shared/hooks/useFirstVisitHint';
import { useTheme } from '../../../shared/hooks/useTheme';
import { Brand } from '../components/Brand/Brand';
import { Display } from '../components/Display/Display';
import { FeatureHint, type FeatureHintItem } from '../components/FeatureHint/FeatureHint';
import { HistoryDialog } from '../components/HistoryDialog/HistoryDialog';
import { Keypad } from '../components/Keypad/Keypad';
import { ThemeToggle } from '../components/ThemeToggle/ThemeToggle';
import { HINT_STORAGE_KEY, HINTS, LABELS } from '../constants';
import { useCalculator } from '../hooks/useCalculator';
import { useHistory } from '../hooks/useHistory';
import { useKeyboard } from '../hooks/useKeyboard';
import type { HistoryEntry } from '../types';
import { reusableExpression, reusableResult } from '../utils/evaluation';
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
  const { visible: hintVisible } = useFirstVisitHint(HINT_STORAGE_KEY);

  // The theme icon shows the theme the button switches to, and so does the hint.
  const hintItems: readonly FeatureHintItem[] = [
    { icon: 'history', ...HINTS.history },
    { icon: theme === 'dark' ? 'sun' : 'moon', ...HINTS.theme },
  ];

  // While the history dialog is open, the keyboard belongs to it.
  useKeyboard(calculator.pressKey, !historyOpen);

  const pickFromHistory = (entry: HistoryEntry, part: 'expression' | 'result') => {
    calculator.restoreExpression(
      part === 'expression' ? reusableExpression(entry) : reusableResult(entry),
    );
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
          actionsHint={hintVisible ? <FeatureHint items={hintItems} /> : null}
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
