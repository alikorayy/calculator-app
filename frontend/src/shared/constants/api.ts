/** Base path of the versioned API. Same origin: nginx in Docker, the Vite proxy in development. */
export const API_BASE_PATH = '/api/v1';

/** Path of the expression evaluation endpoint. */
export const EVALUATE_PATH = `${API_BASE_PATH}/evaluate`;

/**
 * How long the client waits for an answer. Longer than the server's own request
 * timeout (5 s), so the server's 503 normally arrives first.
 */
export const REQUEST_TIMEOUT_MS = 10_000;

/** HTTP status codes the client treats specially. */
export const HTTP_STATUS = {
  UNPROCESSABLE_ENTITY: 422,
  TOO_MANY_REQUESTS: 429,
  BAD_REQUEST: 400,
  SERVER_ERROR: 500,
} as const;

/** HTTP method of the calculation endpoints. */
export const HTTP_METHOD_POST = 'POST';

/** Header names sent with every request. */
export const HEADER_CONTENT_TYPE = 'Content-Type';
export const HEADER_ACCEPT = 'Accept';

/** Media type of request and response bodies. */
export const JSON_CONTENT_TYPE = 'application/json';

/** Header that tells how many seconds to wait after a 429. */
export const RETRY_AFTER_HEADER = 'Retry-After';

/** Messages shown to the user for failures that carry no message meant for them. */
export const API_ERROR_MESSAGES = {
  invalidRequest: 'Something went wrong with the request. Please try again.',
  rateLimited: 'Too many calculations at once. Please wait a moment.',
  server: 'The calculator service is having trouble. Please try again.',
  network: 'Cannot reach the calculator service. Check your connection.',
  timeout: 'The calculation took too long. Please try again.',
} as const;
