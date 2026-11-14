# Calculator – Project Rules

Full-stack calculator: React + TypeScript frontend (`frontend/`) and a single Go microservice (`backend/`).
These rules apply to the whole repository. Layer-specific rules live in `backend/CLAUDE.md` and
`frontend/CLAUDE.md` and take precedence for their layer.

## Language

- Everything in this repository is written in English: code, identifiers, comments, commit messages,
  documentation, UI text, error messages and test names. No other language anywhere.

## Architecture (do not change silently)

These decisions are deliberate. Changing any of them requires explicit agreement first.

- **Expression-based evaluation.** The frontend builds an expression string; the backend parses and
  evaluates it in a single request (`POST /api/v1/evaluate`). Standard operator precedence applies
  (`50 - 7 × 3 = 29`).
- **Exact arithmetic.** The backend computes with exact rationals and rounds only once, at output.
- **All calculation happens in the backend.** The frontend never performs arithmetic.
- **One backend service** (`calculator-api`), stateless. No database, no additional services.
- **No authentication.** The calculator is a public, stateless utility.
- **No CORS.** The browser only talks to one origin: nginx in Docker, the Vite dev proxy locally.
- **Two-layer validation.** The frontend prevents invalid input for UX; the backend validates
  everything for correctness and never trusts its clients.

See `README.md` and `docs/diagrams/` for the full picture.

## Code quality

- Follow the idiomatic best practices of each language. Write clean, readable code.
- Keep functions small and focused on one responsibility. Prefer clear names over comments.
- Document every exported/public function, type and component (see the layer rules for the format).
  Internal helpers get a comment only when their behaviour is not obvious.
- No hardcoded values. Magic numbers and strings live in named constants; runtime settings
  (ports, limits, timeouts) come from configuration with defaults defined as constants.
  Self-evident literals such as `0`, `1` or `""` are fine.
- No dead code, commented-out code or leftover debug output.

## Project structure

- Respect the existing structure. Do not add top-level folders, new dependencies or features
  without asking first.
- Every new dependency needs a short justification.

## Testing

- Every change ships with unit tests that cover the new behaviour and its edge cases.
- Test data follows the conventions of each layer (`testdata/` in Go, `src/mocks/` in the frontend).
  Do not paste large mock objects into test bodies.
- A task is done only when all tests and linters pass.

## Workflow

- Small, focused changes. One concern per commit.
- Commit messages follow Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `refactor:`, `chore:`).
- Never commit secrets, credentials or local environment files.
- When a meaningful design decision is made, add an entry to `PROMPTS.md` using its existing format.
