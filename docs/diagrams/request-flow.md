# Request flow

What happens when the user presses `=` on `50 − 7 × 3`. The frontend sends the
whole expression in one request; the backend parses and evaluates it and
returns the result as a string. Every middleware can end the request early with
the same JSON error envelope.

```text
FRONTEND

  User presses "="
     │
     ▼
  CalculatorPage · useCalculator · useEvaluate
     │   closes open parentheses, replaces Ans,
     │   aborts a request that is still running
     ▼
  shared/api · evaluateExpression
     │   POST /api/v1/evaluate   {"expression": "50−7×3"}
     ▼
  nginx
     │   X-Forwarded-For = real client IP
     ▼
BACKEND

  Middleware chain, in this order          can answer early with
     │   1. request ID
     │   2. logging
     │   3. panic recovery                 500 INTERNAL_ERROR
     │   4. security headers
     │   5. rate limit                     429 RATE_LIMITED
     │   6. body size limit                413 PAYLOAD_TOO_LARGE
     │   7. timeout                        503 TIMEOUT
     ▼
  httpapi · evaluate handler
     │   strict JSON decoding              400 INVALID_REQUEST
     │                                     415 UNSUPPORTED_MEDIA_TYPE
     ▼
  engine.Evaluate("50−7×3")
     │   result                            422 SYNTAX_ERROR, DIVISION_BY_ZERO, ...
     ▼
  200 {"result": "29", "approximate": false}

BACK IN THE FRONTEND

  The API client turns every response into a typed result (it never throws)
     │
     ▼
  The display shows "29" with "50−7×3 =" above it,
  and the calculation is added to the history
```

Errors the user caused come back as `422` with a message to show; a syntax
error also carries the character position, which the display highlights.
Network failures and time-outs are mapped to short messages by the API client.
