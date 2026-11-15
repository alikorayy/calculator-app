# calculator-web

The frontend of the calculator: a React + TypeScript single-page app built with
Vite. It builds the expression the user types and shows the result; every
calculation is done by the backend. See the [root README](../README.md) for the
overall architecture and the design decisions; this file is for working on the
frontend itself.

Requires Node.js 22.12 or newer.

## Quick start

The dev server forwards `/api` to the backend, so start the backend first:

```sh
# terminal 1
cd backend && go run ./cmd/calculator-api        # :8080

# terminal 2
cd frontend
npm install
npm run dev                                      # http://localhost:5173
```

Set `API_PROXY_TARGET` to proxy to a backend elsewhere.

## Scripts

| Command            | What it does                                              |
| ------------------ | --------------------------------------------------------- |
| `npm run dev`      | Dev server with hot module replacement                    |
| `npm run build`    | Type-check, then build the production bundle into `dist/` |
| `npm run preview`  | Serve the production build locally                        |
| `npm test`         | Run all tests once                                        |
| `npm run coverage` | Tests with a coverage report (`coverage/index.html`)      |
| `npm run lint`     | ESLint                                                    |
| `npm run format`   | Prettier                                                  |

## Layout

```
src/
  app/                    App shell, design tokens (tokens.css), base styles
  features/calculator/
    pages/                CalculatorPage: the only place that combines hooks and components
    components/           Presentational components (Display, Keypad, Key, HistoryDialog, ...)
    hooks/                useCalculator, useEvaluate, useHistory, useKeyboard, calculatorReducer
    utils/                Pure functions: input rules, Ans substitution, display helpers
    constants.ts, types.ts
  shared/
    api/                  Typed HTTP client and the calculator API
    components/           Icon, IconButton, Dialog
    hooks/                useTheme, useStoredState
    constants/            Cross-cutting constants
  mocks/                  Test fixtures and MSW request handlers
```

Data flows down and events flow up: the page reads state from hooks and passes
it to components as props; components report user actions through callbacks and
know nothing about state, hooks or the API.

```
CalculatorPage
  → useCalculator ─ useReducer(calculatorReducer)   state machine: editing / evaluating / result / error
      → useEvaluate                                 cancels superseded and unmounted requests
          → evaluateExpression → postJson → fetch   shared/api
```

## Behaviour

**No arithmetic in the frontend.** The app only builds an expression string and
displays what the server returns. Key presses go through pure input rules that
keep the expression well formed, as on Google's and Windows' calculators: a
second operator replaces the first, a number has one decimal point, `)`, `%`,
x² and 1/x need an operand before them, open parentheses close on `=`. Text typed
directly into the display is not restricted; the backend validates everything
and a syntax error highlights the offending character.

**Ans stays exact.** After `=`, an operator continues from the result as `Ans`.
A rounded or scientific result is never sent back as its rounded text: `Ans` is
replaced by the expression that produced it, so `1 ÷ 3 =` followed by `× 3 =`
sends `(1÷3)×3` and shows exactly `1`. Rounded results are marked with `≈`.

**Errors.** The API client never throws; every outcome is a typed result
(calculation error, invalid request, rate limited, server, network, time-out,
cancelled). Calculation errors show the server's message; the others show a
short user-facing text.

**History and theme** live only in the browser: the history in session storage
(last 50, cleared when the tab closes), the chosen theme in local storage. The
theme follows the operating system until the user picks one.

## Responsive design and accessibility

- Mobile first, from 320 px wide: the function pad sits next to the basic pad
  when both fit and wraps above it otherwise; key height scales with the
  dynamic viewport height; padding respects the safe-area insets of notched phones.
- On phones the display does not open the system keyboard (the keypad replaces
  it); on desktops the keyboard works everywhere: digits, operators, `Enter`,
  `Backspace`, `Escape`.
- Semantic buttons with spoken names (`×` is "Multiply"), a live region that
  announces results and errors, visible focus, a native `<dialog>` for the
  history, and reduced motion when the system asks for it.
- Colours are design tokens defined once with `light-dark()`, so the right theme
  shows before any JavaScript runs.

## Testing

Vitest with Testing Library and user-event; tests sit next to the code and query
by role, label or text. The API is mocked with MSW at the network level, so the
real client code runs in every test; a request without a matching handler fails
the test. Pure utilities and the reducer have exhaustive table tests; the page
has integration tests that click and type like a user.

## Notes

- Versions are pinned. TypeScript is on 6.0 and ESLint on 9 because
  typescript-eslint and eslint-plugin-jsx-a11y do not support the newer majors
  yet; jsdom is on 28 because newer versions need Node 22.22.
- `overrides` in `package.json` moves typescript-eslint's `minimatch` to a
  patched version (a development-only advisory; `npm audit` is clean).
- On Windows, run the scripts from PowerShell or cmd. Git Bash can report the
  drive letter in lower case (`c:`), which makes Vitest load itself twice.
