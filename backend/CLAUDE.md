# Backend Rules (Go)

`calculator-api`: a single, stateless Go microservice that evaluates calculator expressions.
The root `CLAUDE.md` also applies.

## Layout

```
backend/
├── cmd/calculator-api/main.go   # wiring only: load config, build dependencies, start server
├── internal/
│   ├── engine/                  # pure domain logic: tokenizer, parser, evaluator, formatter
│   ├── httpapi/                 # HTTP handlers, request/response types, error mapping
│   ├── middleware/              # request id, logging, recover, rate limit, body limit, timeout
│   └── config/                  # configuration from environment variables
├── api/openapi.yaml             # API contract
└── Dockerfile
```

- Packages are named after what they provide. There is no `utils`, `helpers` or `common` package;
  a helper lives in the package that uses it.
- Handlers (the "controllers") live in `internal/httpapi`; domain logic (the "services") lives in
  `internal/engine`.
- `main.go` contains no business logic.

## Engine

- The engine is pure: no `net/http`, no logging, no global state, no I/O.
- No `float64` for calculations. Use `*big.Rat` for exact arithmetic and `*big.Float` only for
  irrational operations (square root, non-integer exponents).
- The public surface is small: `Evaluate`, `Calculate`, `Result` and the typed `Error`.
- Every failure is returned as a typed `*engine.Error` with a stable error code. The engine never panics.
- Input limits (expression length, nesting depth, digits, exponent size) are named constants.

## HTTP layer

- Router: standard library `net/http` (method + path patterns). No web framework.
- Strict JSON decoding: `DisallowUnknownFields`, required `Content-Type: application/json`,
  explicit errors for empty or malformed bodies.
- Every error response uses the same envelope: `{"error": {"code": "...", "message": "..."}}`.
- Engine error codes are mapped to HTTP status codes in a single table: 400 for malformed requests,
  422 for valid requests whose math cannot be computed.
- Handlers propagate `r.Context()`.

## Code style

- Code is formatted with `gofmt`/`goimports` and passes `golangci-lint`.
- Every exported identifier has a godoc comment that starts with its name.
- Constants live in the package that owns them, typically in a `constants.go` file.
  There is no global constants package.
- Errors are wrapped with `%w` and inspected with `errors.Is` / `errors.As`.
- Never panic in library code. Panics are recovered only by the recover middleware.
- Dependencies are passed through constructors. Interfaces are small and defined by the consumer.
  No mutable package-level variables.
- Logging uses `log/slog` (structured JSON). No `fmt.Println` or `log.Printf`.
- Prefer the standard library. A new third-party module needs a justification.

## Testing

- Tests live next to the code in `_test.go` files.
- Use table-driven tests with `t.Run`. Small test cases belong in the test table; large or shared
  fixtures and golden files go under `testdata/`.
- Test HTTP handlers and middleware with `net/http/httptest`.
- Add fuzz tests (`go test -fuzz`) for the tokenizer/parser/evaluator: any input must produce either
  a result or a typed error, never a panic.
- Run tests with the race detector: `go test -race ./...`.
- Coverage targets: `internal/engine` ≥ 95%, overall ≥ 85%.
