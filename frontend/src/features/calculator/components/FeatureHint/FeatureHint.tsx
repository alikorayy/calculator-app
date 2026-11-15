import { Icon } from '../../../../shared/components/Icon/Icon';
import type { IconName } from '../../../../shared/components/Icon/icons';
import styles from './FeatureHint.module.css';

/** One line of the hint: which button it is about and what it does. */
export interface FeatureHintItem {
  icon: IconName;
  label: string;
  description: string;
}

/** Props of {@link FeatureHint}. */
export interface FeatureHintProps {
  items: readonly FeatureHintItem[];
}

/**
 * A small speech bubble with an arrow that points up at the buttons it
 * explains. It is positioned under the element it is placed in, so that
 * element must be `position: relative` and as wide as the buttons.
 *
 * It only repeats what the buttons' accessible names already say, so it is
 * hidden from assistive technology, and it never takes clicks: the buttons it
 * overlaps stay usable.
 *
 * @param props - {@link FeatureHintProps}
 * @returns The bubble.
 */
export function FeatureHint({ items }: FeatureHintProps) {
  return (
    <div className={styles.hint} aria-hidden="true">
      {items.map((item) => (
        <p key={item.label} className={styles.item}>
          <Icon name={item.icon} />
          <span className={styles.label}>{item.label}</span>
          <span>{item.description}</span>
        </p>
      ))}
    </div>
  );
}
