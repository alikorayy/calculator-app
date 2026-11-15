import { setupServer } from 'msw/node';

import { handlers } from './handlers';

/**
 * The mock API server for tests. It intercepts `fetch` in Node, so components
 * and hooks run their real HTTP code against predictable responses.
 */
export const server = setupServer(...handlers);
