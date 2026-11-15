import { ICONS, type IconName } from './icons';
import styles from './Icon.module.css';

/** Props of {@link Icon}. */
export interface IconProps {
  /** Which icon to draw. */
  name: IconName;
}

/**
 * A decorative stroke icon that takes the current text colour. It is hidden from
 * assistive technology; the control that contains it carries the accessible name.
 *
 * @param props - {@link IconProps}
 * @returns An inline SVG element.
 */
export function Icon({ name }: IconProps) {
  return (
    <svg
      className={styles.icon}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      {ICONS[name]}
    </svg>
  );
}
