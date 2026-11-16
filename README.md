# Calculator

[![CI](https://github.com/alikorayy/calculator-app/actions/workflows/ci.yml/badge.svg)](https://github.com/alikorayy/calculator-app/actions/workflows/ci.yml)

A full-stack calculator: a React + TypeScript frontend and a Go microservice
that evaluates whole expressions with exact arithmetic.

## Quick start

1. Start everything:

   ```sh
   docker compose up --build
   ```

2. When the build finishes and the containers are healthy, open
   **<http://localhost:3000>** in your browser.

**Optional `.env` file:** the app runs with default values out of the box. To
change settings such as the port, log level or rate limits, copy
[`.env.example`](.env.example) to `.env` and edit the values; `docker compose up`
picks it up automatically.

Stop it with `docker compose down`.

### Run without Docker

Requires Go 1.25 and Node.js 22.12 or newer. Start the backend first, in its
own terminal:

```sh
cd backend
go run ./cmd/calculator-api        # API on http://localhost:8080
```

Then the frontend, in a second terminal:

```sh
cd frontend
npm install
npm run dev                        # app on http://localhost:5173
```

The Vite dev server forwards `/api` to the backend on port 8080, so the browser
still talks to a single origin. In this setup, send the curl examples below to
`localhost:8080` instead of `localhost:3000`.

## Architecture

```text
                           ┌─────────┐
                           │ Browser │
                           └────┬────┘
                                │ http://localhost:3000
                                ▼
┌─ docker compose · private network 172.28.0.0/24 ───────────────┐
│                                                                │
│  ┌─ web · nginx · published on :3000 ───────────────────────┐  │
│  │                                                          │  │
│  │ GET  /         →  static files                           │  │
│  │                   index.html, JS, CSS, fonts             │  │
│  │                                                          │  │
│  │ POST /api/*    →  reverse proxy to the api service       │  │
│  │                   X-Forwarded-For = real client IP       │  │
│  │                                                          │  │
│  └────────────────────────────┬─────────────────────────────┘  │
│                               │ HTTP                           │
│                               ▼                                │
│  ┌─ api · calculator-api · internal :8080 only ─────────────┐  │
│  │                                                          │  │
│  │ middleware   request ID, logging, panic recovery,        │  │
│  │              security headers, rate limit,               │  │
│  │              body size limit, timeout                    │  │
│  │     │                                                    │  │
│  │     ▼                                                    │  │
│  │ httpapi      strict JSON decoding, error envelope        │  │
│  │     │                                                    │  │
│  │     ▼                                                    │  │
│  │ engine       exact arithmetic, the only code that        │  │
│  │              calculates (no HTTP, no state)              │  │
│  │                                                          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

The browser only talks to nginx, which serves the app and forwards `/api` to the
Go service. The API is never published, so there is no CORS. **All calculation
happens in the backend**; the frontend only builds the expression and shows the
result.

## How a calculation works (Backend Calculation Flow)

The frontend sends the whole expression in one request. The backend's engine
turns it into a result in four steps:

```text
  POST /api/v1/evaluate     {"expression": "50 - 7 × 3"}
        │
        ▼
  1. TOKENIZE    Split the text into numbers and operators.
                 [50] [−] [7] [×] [3]
        │
        ▼
  2. PARSE       Arrange the pieces as a tree. × binds tighter than −,
                 so it sits deeper and is calculated first.

                         −
                        / \
                      50   ×          =  50 − (7 × 3)
                          / \
                         7   3
        │
        ▼
  3. EVALUATE    Calculate from the bottom of the tree up (a depth-first,
                 post-order walk: children before their parent), with
                 exact fractions (Go's big.Rat), never floating point.
                 7 × 3   = 21
                 50 − 21 = 29
        │
        ▼
  4. FORMAT      Round once, at the very end, to 16 significant digits,
                 and say whether anything was rounded.
        │
        ▼
  200 {"result": "29", "approximate": false}
```

Why exact fractions: with floating point, `0.7 + 0.1` gives
`0.7999999999999999`. With fractions it is `7/10 + 1/10 = 8/10`, shown as `0.8`.
A result like `1 ÷ 3` is rounded only when it is shown:
`{"result": "0.3333333333333333", "approximate": true}`.

If any step fails, the response says what went wrong and where:
`2 + * 3` → `422 {"error": {"code": "SYNTAX_ERROR", "message": "Unexpected \"*\"", "position": 4}}`.

## Features

| Feature | What it means |
|---------|---------------|
| **Real calculator rules** | Precedence and parentheses, powers, square roots and percent as on Google's and Windows' calculators: `50 + 10% = 55`, `200 × 10% = 20` |
| **Exact results** | `0.7 + 0.1 = 0.8` and `1 ÷ 3 × 3 = 1`; rounded results are marked with `≈` |
| **Validation on both sides** | The frontend prevents invalid input as you type (one operator at a time, one decimal point, parentheses close on `=`). The backend checks everything again, because any client can call the API |
| **Rate limit** | Each IP address may send 10 requests per second (short bursts of 20). Beyond that the API answers `429` with `Retry-After`. The IP cannot be faked: nginx overwrites `X-Forwarded-For` |
| **Health check** | `GET /healthz` returns `200` while the service is alive. Docker uses it, and the web container starts only once the API is healthy |
| **Never crashes** | Input limits (256 characters, 32 digits per number, 32 nested parentheses), a 4 KiB body limit and time-outs; fuzz tests and panic recovery |
| **Clear errors** | Every error has the same JSON shape. `422` means "the maths cannot be done" (`DIVISION_BY_ZERO`, `SYNTAX_ERROR` with the character position, which the UI highlights) |
| **Secure containers** | Non-root, read-only file systems, no Linux capabilities, strict Content-Security-Policy |
| **Stateless** | No database. History (last 50, per tab) and theme live in the browser |
| **Tested** | Backend engine 99.5 %, handlers and middleware 100 %; frontend 98.6 % |

---

The sections below are the details.

## API

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/evaluate` | Evaluate an expression; this is what the UI uses |
| POST | `/api/v1/calculate` | Apply one operation (`add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percent`) to its operands |
| GET | `/healthz` | Health check (inside the network) |

```sh
curl -s localhost:3000/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression": "50 - 7 * 3"}'
# 200 {"result":"29","approximate":false}

curl -s localhost:3000/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation": "divide", "operands": ["1", "3"]}'
# 200 {"result":"0.3333333333333333","approximate":true}

curl -s localhost:3000/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression": "5 / 0"}'
# 422 {"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero"}}
```

| Status | Meaning | Codes |
|--------|---------|-------|
| 400 | The request itself is broken (a client bug) | `INVALID_REQUEST` |
| 413 / 415 | Body too large / not JSON | `PAYLOAD_TOO_LARGE`, `UNSUPPORTED_MEDIA_TYPE` |
| 422 | The request is fine but the maths cannot be done | `SYNTAX_ERROR`, `DIVISION_BY_ZERO`, `UNDEFINED_RESULT`, `DOMAIN_ERROR`, `OVERFLOW`, `LIMIT_EXCEEDED` |
| 429 | Rate limited, with `Retry-After` | `RATE_LIMITED` |
| 404 / 405 / 500 / 503 | Unknown route / wrong method / unexpected error / time-out | `NOT_FOUND`, `METHOD_NOT_ALLOWED`, `INTERNAL_ERROR`, `TIMEOUT` |

The full contract is in [`backend/api/openapi.yaml`](backend/api/openapi.yaml).

## Design decisions

### Assumptions

- **Modern calculator behaviour.** The Google and Windows calculators are the
  reference: whole expressions with standard precedence (`50 − 7 × 3 = 29`),
  and Google's percent rule. After `+` or `−`, a percentage is taken of
  everything to its left, and the result is one value for what follows:
  `50 + 10% = 55`, `50 + 10% × 2 = 110`, while `200 × 10% = 20` and
  `50 + (10%) = 50.1`. This is the one place where `×` does not bind tighter
  than `+`, on purpose.
- **Display precision.** Results are rounded once, to 16 significant digits as
  on the Windows calculator; very large or small values use scientific notation
  (`1e+20`).
- **Input format.** `.` is the decimal separator; `×`/`*`, `÷`/`/` and `−`/`-`
  are both accepted; square root is written as a function, `√(…)`; `2(3)` means
  `2 × 3`.
- **Reasonable limits.** Expressions of up to 256 characters, 32 digits per
  number and 32 nested parentheses are enough for a calculator; anything
  longer is rejected with a clear error instead of being computed.
- **Public and stateless.** No accounts or authentication; no data is stored on
  the server. History and theme stay in the user's browser.

### Evaluation model: considered alternatives

A calculator can work in two ways, and they give different answers for the same
keys. For `50 − 7 × 3`:

| Model | How it works | Result |
|-------|--------------|--------|
| **Immediate execution** (basic pocket calculators) | Each operator applies to the running total: `50 − 7 = 43`, then `43 × 3` | `129` |
| **Expression evaluation** (Google, Windows scientific, phones) | The whole expression is evaluated with standard precedence | `29` |

Immediate execution would have meant one request per operator, a running total
kept on the client, rounding between every step, and calculator logic split
between the browser and the server. **Chosen: expression evaluation.** One
request per calculation, the result users expect, and the backend as the single
source of truth for syntax and arithmetic. The cost is a parser on the server,
kept small and covered by limits and fuzz tests. `/api/v1/calculate` still
offers single operations for API clients, on the same engine.

### Exact arithmetic

`+ − × ÷` and integer powers use exact fractions (`math/big.Rat`) and have no
error at all. Square roots and fractional powers have no exact fraction; they
are computed with 256-bit `big.Float` (about 77 digits), so rounding to 16 digits
at the end hides the approximation (`√2 × √2` shows `2`, marked approximate).
Considered: `shopspring/decimal`, the usual choice for money; its division stops
at a fixed number of digits, so `1 ÷ 3 × 3` would not be exactly `1`.

The frontend keeps this exactness across chained calculations: when `Ans` is a
rounded result, it sends the expression that produced it, so `1 ÷ 3 =` followed
by `× 3 =` evaluates `(1÷3)×3` and shows `1`.

### One service, one origin, no authentication

The task asks for one backend microservice, and splitting a calculator further
would only add network hops. The service is stateless and has no dependencies,
so `/healthz` serves as both liveness and readiness. nginx gives the browser a
single origin, which removes the need for CORS. The calculator is a public
utility, so there is no authentication.

## Testing and quality

| | Backend | Frontend |
|---|---|---|
| Tests | Table-driven unit tests, `httptest` for handlers and middleware, fuzz tests for the engine, an end-to-end test of the middleware chain | Vitest + Testing Library; the API mocked with MSW at the network level; integration tests that click and type like a user |
| Static checks | `go vet`, golangci-lint, race detector | TypeScript strict, ESLint (type-checked, React hooks, accessibility), Prettier |

```sh
cd backend && go test -cover ./...
cd frontend && npm run lint && npm run coverage
```

CI runs all of this, plus a `docker compose` smoke test through nginx, on every
push and pull request.

**Coverage reports** are HTML pages and are not committed. Create them locally
with `make cover` in `backend/` (writes `coverage.html`) and `npm run coverage`
in `frontend/` (writes `coverage/index.html`). Every CI run also keeps both as
downloadable artifacts (`backend-coverage`, `frontend-coverage`): open the run
under **Actions** and scroll to **Artifacts**.

## Repository layout

```
backend/            Go service: cmd/ (wiring), internal/ (engine, httpapi, middleware, config), api/openapi.yaml
frontend/           React app: src/ (app, features/calculator, shared, mocks), nginx/ (container config)
docs/diagrams/      Diagrams in plain text
.github/workflows/  CI: backend, frontend and Docker checks on every push
compose.yaml        Both services with one command; settings in .env.example
PROMPTS.md          AI prompts behind the main decisions
CLAUDE.md           Coding rules for AI assistants (plus backend/ and frontend/ ones)
```

More reading: [`backend/README.md`](backend/README.md) (engine, configuration,
limits), [`frontend/README.md`](frontend/README.md) (structure, behaviour,
accessibility), [request flow](docs/diagrams/request-flow.md) and
[engine pipeline](docs/diagrams/engine-pipeline.md) diagrams.

## Limitations and next steps

- **Rate limits are per instance.** Several API instances would need a shared
  store (Redis) or an API gateway in front.
- **History is per browser tab.** A server-side history would be a separate
  service with its own storage, keeping the calculator itself stateless.
- **Scaling out:** the static app would move to a CDN, and nginx would become an
  API gateway handling TLS, authentication and rate limiting centrally.
