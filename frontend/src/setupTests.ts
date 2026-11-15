import '@testing-library/jest-dom/vitest';

import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll } from 'vitest';

import { server } from './mocks/server';

// jsdom does not implement the <dialog> methods yet. This minimal stand-in
// toggles the `open` attribute, which is what the components and the
// accessibility queries rely on; focus trapping is left to real browsers.
if (!('showModal' in HTMLDialogElement.prototype)) {
  Object.assign(HTMLDialogElement.prototype, {
    showModal(this: HTMLDialogElement) {
      this.open = true;
    },
    close(this: HTMLDialogElement) {
      this.open = false;
    },
  });
}

// Any request without a matching handler fails the test, so no test can
// silently depend on a real network.
beforeAll(() => {
  server.listen({ onUnhandledFrame: 'error' });
});

afterEach(() => {
  server.resetHandlers();
  cleanup();
});

afterAll(() => {
  server.close();
});
