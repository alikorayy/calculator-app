import { Icon } from '../Icon/Icon';
import type { IconName } from '../Icon/icons';
import styles from './IconButton.module.css';

/** Props of {@link IconButton}. */
export interface IconButtonProps {
  /** The icon to show. */
  icon: IconName;
  /** Accessible name, also shown as a tooltip, e.g. "History". */
  label: string;
  /** Called when the button is activated by click, tap, Enter or Space. */
  onClick: () => void;
}

/**
 * A round, icon-only button. Because it has no visible text, `label` becomes
 * its accessible name for screen readers and its tooltip on hover. The hit area
 * is at least the minimum touch target size.
 *
 * @param props - {@link IconButtonProps}
 * @returns A button element.
 */
export function IconButton({ icon, label, onClick }: IconButtonProps) {
  return (
    <button
      type="button"
      className={styles.button}
      aria-label={label}
      title={label}
      onClick={onClick}
    >
      <Icon name={icon} />
    </button>
  );
}
