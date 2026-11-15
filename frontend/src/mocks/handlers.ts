import type { RequestHandler } from 'msw';

/**
 * Default MSW request handlers used by every test. Handlers for the calculator
 * API are added together with the API client; tests override them per case
 * with `server.use(...)`.
 */
export const handlers: RequestHandler[] = [];
