import { Icon } from '../../../../shared/components/Icon/Icon';
import type { KeyDefinition } from '../../types';
import styles from './Key.module.css';

/** Props of {@link Key}. */
export interface KeyProps {
  definition: KeyDefinition;
  /** Called when the key is pressed. */
  onPress: () => void;
  disabled?: boolean;
}

/**
 * One calculator key. Its look comes from the definition's variant (number,
 * operator, equals, ...), set as a data attribute that the stylesheet reads.
 *
 * @param props - {@link KeyProps}
 * @returns A button element.
 */
export function Key({ definition, onPress, disabled = false }: KeyProps) {
  const { label, ariaLabel, icon, variant } = definition;
  return (
    <button
      type="button"
      className={styles.key}
      data-variant={variant}
      aria-label={ariaLabel}
      disabled={disabled}
      onClick={onPress}
    >
      {icon ? <Icon name={icon} /> : label}
    </button>
  );
}
