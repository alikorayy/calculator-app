import { BASIC_KEYS, CLEAR_ARIA_LABELS, FUNCTION_KEYS, LABELS } from '../../constants';
import type { ClearLabel, KeyDefinition, KeyId } from '../../types';
import { Key } from '../Key/Key';
import styles from './Keypad.module.css';

/** Props of {@link Keypad}. */
export interface KeypadProps {
  /** Current label of the clear key. */
  clearLabel: ClearLabel;
  /** While a calculation runs, "=" is disabled. */
  isEvaluating: boolean;
  /** Called with the id of every pressed key. */
  onKey: (key: KeyId) => void;
}

/**
 * The function pad and the basic pad. On wide screens they sit side by side;
 * when there is not enough room the function pad wraps above the basic pad,
 * keeping the digits within thumb reach on phones.
 *
 * @param props - {@link KeypadProps}
 * @returns Both pads.
 */
export function Keypad({ clearLabel, isEvaluating, onKey }: KeypadProps) {
  const renderKey = (definition: KeyDefinition) => {
    const shown =
      definition.id === 'clear'
        ? { ...definition, label: clearLabel, ariaLabel: CLEAR_ARIA_LABELS[clearLabel] }
        : definition;
    return (
      <Key
        key={definition.id}
        definition={shown}
        disabled={definition.id === 'equals' && isEvaluating}
        onPress={() => {
          onKey(definition.id);
        }}
      />
    );
  };

  return (
    <div className={styles.pads}>
      <div className={styles.functionPad} role="group" aria-label={LABELS.functionPad}>
        {FUNCTION_KEYS.map(renderKey)}
      </div>
      <div className={styles.basicPad} role="group" aria-label={LABELS.basicPad}>
        {BASIC_KEYS.map(renderKey)}
      </div>
    </div>
  );
}
