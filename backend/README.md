# calculator-api

The backend of the calculator: a single, stateless Go service that evaluates
calculator expressions with exact arithmetic and serves them over a small JSON
API. See the [root README](../README.md) for the overall architecture and the
design decisions; this file is for working on the backend itself.

Requires Go 1.25. The only third-party modules are
[`ALTree/bigfloat`](https://github.com/ALTree/bigfloat) (logarithm and exponent
for `big.Float`, needed for fractional powers) and
[`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate) (token
bucket for rate limiting).

## Quick start

```sh
go run ./cmd/calculator-api        # listens on :8080

curl -s localhost:8080/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression": "50 - 7 × 3"}'
# {"result":"29","approximate":false}
```

## Layout

```
cmd/calculator-api/   main: configuration, middleware chain, server, graceful shutdown
internal/
  engine/             pure domain logic: tokenizer → parser → evaluator → formatter
  httpapi/            handlers, strict JSON decoding, error envelope, engine error → HTTP status
  middleware/         request ID, logging, panic recovery, security headers,
                      rate limiting, body size limit, timeout
  config/             settings from environment variables, validated at startup
api/openapi.yaml      the API contract (OpenAPI 3.0)
```

`engine` knows nothing about HTTP and holds no state; it is the only place
where calculations happen. Its public surface is `Evaluate`, `Calculate`,
`Result` and the typed `Error`.

How an expression is evaluated:

1. **tokenize** splits it into numbers, operators and parentheses, keeping each
   token's character position for error messages.
2. **parse** builds an abstract syntax tree with a recursive-descent parser; the
   shape of the tree encodes operator precedence.
3. **evaluate** walks the tree with exact rationals (`*big.Rat`). Irrational
   results (`√2`, `2^0.5`) are computed with 256-bit `big.Float` and turned
   back into a fraction. Every intermediate result is range-checked.
4. **format** rounds once, to 16 significant digits, and reports whether the
   value was rounded.

## API

| Method | Path               | Purpose                                   |
|--------|--------------------|-------------------------------------------|
| POST   | `/api/v1/evaluate` | Evaluate an expression such as `50 + 10%` |
| POST   | `/api/v1/calculate`| Apply one operation to one or two operands|
| GET    | `/healthz`         | Liveness check                            |

The full contract, with every status code and example, is in
[`api/openapi.yaml`](api/openapi.yaml). Paste it into
[editor.swagger.io](https://editor.swagger.io) for an interactive view.

```sh
curl -s localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation": "divide", "operands": ["10", "4"]}'
# {"result":"2.5","approximate":false}

curl -s localhost:8080/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression": "2 + * 3"}'
# 422 {"error":{"code":"SYNTAX_ERROR","message":"Unexpected \"*\"","position":4}}
```

### Input limits

| Limit                           | Value                    | Error            |
|---------------------------------|--------------------------|------------------|
| Expression length               | 256 characters           | `LIMIT_EXCEEDED` |
| Significant digits per number   | 32                       | `LIMIT_EXCEEDED` |
| Nested parentheses              | 32 levels                | `LIMIT_EXCEEDED` |
| Magnitude of any result         | 1e-10000 … 1e+10000      | `OVERFLOW`       |

## Configuration

All settings come from environment variables. Invalid values stop the service
at startup with a clear message.

| Variable           | Default | Meaning |
|--------------------|---------|---------|
| `PORT`             | `8080`  | TCP port to listen on |
| `LOG_LEVEL`        | `info`  | `debug`, `info`, `warn` or `error` |
| `READ_TIMEOUT`     | `5s`    | Time to read a whole request |
| `WRITE_TIMEOUT`    | `10s`   | Time to write a response |
| `IDLE_TIMEOUT`     | `60s`   | Keep-alive time of an idle connection |
| `REQUEST_TIMEOUT`  | `5s`    | Time a handler may run before a 503; must be shorter than `WRITE_TIMEOUT` |
| `SHUTDOWN_TIMEOUT` | `10s`   | Time in-flight requests get to finish on SIGTERM |
| `MAX_BODY_BYTES`   | `4096`  | Largest accepted request body |
| `RATE_LIMIT_RPS`   | `10`    | Requests per second each client may sustain |
| `RATE_LIMIT_BURST` | `20`    | Requests each client may send at once |
| `TRUSTED_PROXIES`  | (none)  | Comma-separated IPs or CIDRs of reverse proxies whose `X-Forwarded-For` is trusted, e.g. `172.16.0.0/12` |

Rate limiting is per client IP (per /64 network for IPv6). Behind a reverse
proxy every request comes from the proxy's address, so the proxy's network must
be listed in `TRUSTED_PROXIES`; the client address is then the rightmost
untrusted entry of `X-Forwarded-For`. Limits are kept in memory, per instance.

## Development

The [Makefile](Makefile) wraps the common tasks; `make help` lists them. On
systems without `make`, run the commands it contains directly.

| Task | Command |
|------|---------|
| Run the service | `make run` |
| Run all tests | `make test` |
| Tests with the race detector (needs cgo) | `make test-race` |
| Coverage report (`coverage.html`) | `make cover` |
| Fuzz the engine (60 s per target) | `make fuzz` |
| Lint with golangci-lint | `make lint` |
| Build and run the Docker image | `make docker-build docker-run` |

Tests live next to the code. Table-driven tests cover every grammar rule and
error path; handlers and middleware are tested with `httptest`; the engine has
fuzz tests asserting that any input yields a result or a typed error, never a
panic. Inputs found by fuzzing are kept in `internal/engine/testdata/fuzz` and
re-run by `go test`. Coverage targets: `internal/engine` ≥ 95 %, overall ≥ 85 %.

## Docker

```sh
docker build -t calculator-api .
docker run --rm -p 8080:8080 calculator-api
```

The image is built in two stages and runs a static binary on
`distroless/static` as a non-root user. It contains no shell or curl, so the
container health check runs the binary itself: `calculator-api healthcheck`
requests `/healthz` on the local port and exits non-zero if it does not answer
200.
