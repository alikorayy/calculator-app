import { IconButton } from '../../../../shared/components/IconButton/IconButton';
import type { Theme } from '../../../../shared/hooks/useTheme';
import { LABELS } from '../../constants';

/** Props of {@link ThemeToggle}. */
export interface ThemeToggleProps {
  theme: Theme;
  onToggle: () => void;
}

/**
 * Switches between the dark and light theme. The icon and label describe the
 * theme it switches to.
 *
 * @param props - {@link ThemeToggleProps}
 * @returns An icon button.
 */
export function ThemeToggle({ theme, onToggle }: ThemeToggleProps) {
  const isDark = theme === 'dark';
  return (
    <IconButton
      icon={isDark ? 'sun' : 'moon'}
      label={isDark ? LABELS.switchToLight : LABELS.switchToDark}
      onClick={onToggle}
    />
  );
}
