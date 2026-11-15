# Architecture

Two containers on a private network. The browser only talks to nginx, which
serves the built React app and forwards `/api` to the Go service; the API itself
is never published. In local development the Vite dev server takes nginx's
place and proxies `/api` the same way, so there is no CORS in either setup.

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

- **web** is stateless: the history and the theme live in the browser's storage.
- **api** is stateless and has no dependencies, so `/healthz` is both liveness
  and readiness. It trusts `X-Forwarded-For` only from the compose network.
- **engine** is pure Go with no HTTP, logging or state; it is the only code
  that calculates.
