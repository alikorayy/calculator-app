import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Icon } from './Icon';
import { ICONS, type IconName } from './icons';

describe('Icon', () => {
  it.each(Object.keys(ICONS) as IconName[])(
    'draws the %s icon hidden from assistive technology',
    (name) => {
      const { container } = render(<Icon name={name} />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('aria-hidden', 'true');
      expect(svg?.childElementCount).toBeGreaterThan(0);
    },
  );
});
