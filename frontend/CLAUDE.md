# Frontend Rules (React + TypeScript)

The calculator UI. It builds an expression from user input, sends it to the backend and displays the
result. The root `CLAUDE.md` also applies.

## Layout

```
frontend/src/
├── app/                      # application shell
│   ├── App.tsx
│   ├── providers.tsx         # only if providers are actually needed
│   ├── tokens.css            # design tokens (colours, spacing, typography)
│   └── global.css            # base styles
├── features/
│   └── calculator/
│       ├── pages/            # page containers: the only place that combines hooks and components
│       ├── components/       # presentational components (Display, Keypad, Key, ...)
│       ├── hooks/            # stateful logic (useCalculator, useEvaluate, ...)
│       ├── utils/            # pure functions (input rules, display formatting)
│       ├── constants.ts
│       └── types.ts
├── shared/
│   ├── api/                  # typed HTTP client, API types, error mapping
│   ├── components/           # generic UI building blocks (Icon, IconButton, Dialog)
│   ├── hooks/                # hooks shared across features
│   └── constants/            # cross-cutting constants
├── mocks/                    # test fixtures and MSW handlers
└── main.tsx
```

- New features go under `features/<name>/`. Code used by more than one feature goes under `shared/`.

## Architecture rules

- **No arithmetic in the frontend.** All calculation goes through the backend API. The frontend only
  builds the expression string and formats the result for display.
- All HTTP calls go through `src/shared/api`. Components and hooks never call `fetch` directly.
  The client handles timeouts, cancellation (`AbortController`) and maps errors to typed values.
- Components are presentational. State and logic live in hooks; pure logic lives in `utils/`
  so it can be unit-tested without rendering.
- Input validation in the frontend is for UX: invalid input is prevented as the user types
  (operator replacement, single decimal point, auto-closing parentheses). The backend remains the
  source of truth.

## Code style

- TypeScript in `strict` mode. No `any`; use `unknown` with type narrowing. No non-null assertions
  (`!`) without a clear reason.
- Function components and hooks only. Named exports only.
- Every exported function, hook and component has a JSDoc comment describing its purpose,
  parameters and return value.
- Constants live in the feature's `constants.ts`; cross-cutting ones in `src/shared/constants`.
  No hardcoded strings, numbers, colours or sizes in components. Exception: the literal tags of a
  discriminated union (e.g. `kind: 'timeout'`) stay literals, because the type checks them.
- Styling uses CSS Modules and the design tokens in `tokens.css`. No inline styles.
- Code passes ESLint and is formatted with Prettier. No `console.log`.

## Accessibility

- Use semantic elements (`<button>` for keys). Every icon-only key has an `aria-label`.
- The whole calculator is usable with the keyboard.
- The result area is announced to screen readers (`aria-live`).
- The layout works on mobile screens.

## Testing

- Vitest + React Testing Library + `@testing-library/user-event`.
- Tests are colocated with the code as `*.test.ts` / `*.test.tsx`.
- Test behaviour, not implementation: query by role, label or text, not by class names or test ids
  unless there is no accessible alternative.
- Mock the API with MSW. All fixtures and MSW handlers live in `src/mocks/`; tests import them
  instead of defining mock data inline.
- Pure utilities in `utils/` have exhaustive unit tests for their edge cases.
