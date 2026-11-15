import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { HINTS } from '../../constants';
import { FeatureHint } from './FeatureHint';

describe('FeatureHint', () => {
  it('shows what each button does', () => {
    render(
      <FeatureHint
        items={[
          { icon: 'history', ...HINTS.history },
          { icon: 'moon', ...HINTS.theme },
        ]}
      />,
    );

    expect(screen.getByText(HINTS.history.label)).toBeInTheDocument();
    expect(screen.getByText(HINTS.history.description)).toBeInTheDocument();
    expect(screen.getByText(HINTS.theme.label)).toBeInTheDocument();
    expect(screen.getByText(HINTS.theme.description)).toBeInTheDocument();
  });

  it('is hidden from assistive technology, since the buttons already have names', () => {
    const { container } = render(<FeatureHint items={[{ icon: 'history', ...HINTS.history }]} />);

    expect(container.firstElementChild).toHaveAttribute('aria-hidden', 'true');
  });
});
