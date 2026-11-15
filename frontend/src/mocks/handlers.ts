import { http, HttpResponse, type RequestHandler } from 'msw';

import { isRecord } from '../shared/api/guards';
import { HTTP_STATUS, EVALUATE_PATH } from '../shared/constants/api';
import { ERRORS, EXPRESSIONS, RESULTS } from './fixtures';

/** Answers of the default evaluate handler, keyed by expression. */
const EVALUATE_ANSWERS = new Map<string, () => Response>([
  [EXPRESSIONS.exact, () => HttpResponse.json(RESULTS.exact)],
  [EXPRESSIONS.rounded, () => HttpResponse.json(RESULTS.rounded)],
  [EXPRESSIONS.roundedTimesThree, () => HttpResponse.json(RESULTS.one)],
  [
    EXPRESSIONS.syntaxError,
    () => HttpResponse.json(ERRORS.syntaxError, { status: HTTP_STATUS.UNPROCESSABLE_ENTITY }),
  ],
  [
    EXPRESSIONS.divisionByZero,
    () => HttpResponse.json(ERRORS.divisionByZero, { status: HTTP_STATUS.UNPROCESSABLE_ENTITY }),
  ],
]);

/**
 * Default MSW request handlers used by every test. The evaluate handler answers
 * the expressions in {@link EXPRESSIONS} like the real backend; any other
 * expression gets a 500, so a test that forgot its fixture fails loudly. Tests
 * override handlers per case with `server.use(...)`.
 */
export const handlers: RequestHandler[] = [
  http.post(EVALUATE_PATH, async ({ request }) => {
    const body: unknown = await request.json();
    const expression = isRecord(body) && typeof body.expression === 'string' ? body.expression : '';
    const answer = EVALUATE_ANSWERS.get(expression);
    return answer
      ? answer()
      : HttpResponse.json(ERRORS.internal, { status: HTTP_STATUS.SERVER_ERROR });
  }),
];
