import type { ReactNode } from 'react';

/**
 * SVG contents of every icon, drawn on a 24×24 grid with strokes (no fills).
 * The shapes follow the Lucide icon set (ISC license).
 */
export const ICONS = {
  history: <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8M3 3v5h5M12 7v5l4 2" />,
  sun: (
    <>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2m-7.07-17.07 1.41 1.41m11.32 11.32 1.41 1.41M2 12h2m16 0h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
    </>
  ),
  moon: <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />,
  close: <path d="M18 6 6 18M6 6l12 12" />,
  backspace: (
    <>
      <path d="M10 5a2 2 0 0 0-1.344.519l-6.328 5.74a1 1 0 0 0 0 1.481l6.328 5.741A2 2 0 0 0 10 19h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2z" />
      <path d="m12 9 6 6M18 9l-6 6" />
    </>
  ),
} satisfies Record<string, ReactNode>;

/** Name of an available icon. */
export type IconName = keyof typeof ICONS;
