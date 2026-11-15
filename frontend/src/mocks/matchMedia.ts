type ChangeListener = () => void;

const listeners = new Set<ChangeListener>();
let prefersDark = false;

/**
 * Installs a controllable `window.matchMedia` (jsdom has none). Every query
 * reports the simulated system colour scheme set with {@link setSystemDarkMode}.
 */
export function installMatchMedia(): void {
  window.matchMedia = (query: string) =>
    ({
      media: query,
      get matches() {
        return prefersDark;
      },
      onchange: null,
      addEventListener: (_type: string, listener: ChangeListener) => listeners.add(listener),
      removeEventListener: (_type: string, listener: ChangeListener) => listeners.delete(listener),
      addListener: (listener: ChangeListener) => listeners.add(listener),
      removeListener: (listener: ChangeListener) => listeners.delete(listener),
      dispatchEvent: () => true,
    }) as unknown as MediaQueryList;
}

/**
 * Simulates the operating system switching its colour scheme and notifies
 * every subscribed media query listener.
 *
 * @param dark - Whether the system now prefers a dark colour scheme.
 */
export function setSystemDarkMode(dark: boolean): void {
  prefersDark = dark;
  for (const listener of listeners) {
    listener();
  }
}

/** Resets the simulated system to a light colour scheme with no listeners. */
export function resetMatchMedia(): void {
  prefersDark = false;
  listeners.clear();
}
