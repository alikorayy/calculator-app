# Prompts

I worked with Claude Code (in VS Code) throughout. These are the prompts that
shaped a decision, not every message. The originals mixed Turkish and English;
here they are in English and trimmed, with the meaning kept. Each entry ends
with what came out of it.

## 1. Understand the task first

> Here's the case study. Just read it and make sure you understand it, don't
> build anything yet. Ignore the 2–4 hour estimate, I want this done properly.
> The bar is correctness, not what's quickest.

→ A short list of open questions before any code: evaluation model, number
precision, service boundaries, how far validation should go.

## 2. Pick a reference for calculator behaviour

> The domain logic is the most important part. Find well-known calculators and
> let's behave like them. It must never break on edge cases. Tell me what we
> take as the reference and list the edge cases.

> I checked percent on Google and the Windows 11 calculator myself, both give
> the same result as you said. Use that.

→ Windows Calculator (its engine is open source and uses exact rationals) as
the main reference, Google as the second. Every edge case (`0.1 + 0.2`, `5 ÷ 0`,
`0 ÷ 0`, `0^0`, `√(−4)`, huge numbers, `50 + 10%`) became a table test.

## 3. Expression evaluation vs. immediate execution

> Google evaluates the whole expression on `=`. The basic Windows calculator
> applies each operator as you go. So `50 − 7 × 3` gives 29 on one and 129 on
> the other. Also, if we call the backend on every operator and someone holds a
> key down, that's a lot of requests, we'd need a rate limit anyway. Which one
> is right? List the pros and cons, think about how a fintech would want it.

> Should we also offer the Windows mode behind a switch, or is that wasted work?

→ Expression evaluation: one request per `=`, standard precedence, rounding
only at the end. The switch was dropped; two semantics would double the state
logic and the tests for no real gain. I asked for the rejected option and why
to be written up in the README.

I also asked a second model for an unbiased opinion on the same question,
without the earlier context. It reached the same conclusion.

## 4. Numbers, precision and grammar

> Decide these within the same frame: `math/big` or `shopspring/decimal` (this
> also decides whether `2^0.5` is possible), how many significant digits we
> show, the rounding mode, when we switch to scientific notation, how square
> root is written, implicit multiplication like `2(3)`, decimal separator.

→ `big.Rat` for exact `+ − × ÷`, 256-bit `big.Float` only for roots and
fractional powers, 16 significant digits, `√(…)` in function form only,
implicit multiplication allowed, `.` as the only separator. `decimal` was
rejected because its division is not exact.

## 5. Errors vs. preventing bad input

> You said `2++3` returns a syntax error with a position. But Google and Windows
> don't even let you type `++`, a second operator just replaces the first. Why
> would we return an error?

→ Two layers. The frontend prevents invalid input as you type, like those
calculators. The backend still validates everything and reports the position,
because the API is public and any client can call it.

## 6. How many services

> They ask for "a backend microservice", not a microservices architecture,
> right? How many services do we actually need? And what am I missing for that
> one service besides middleware and rate limiting?

> Do we need `/readyz`? Isn't `/healthz` enough here?

> If we put nginx in front, we don't need to open both ports, and we don't
> need CORS either? In a real microservices setup, would that be an API gateway?

→ One stateless service with internal packages (`engine`, `httpapi`,
`middleware`, `config`). Only `/healthz`, since there are no dependencies to be
ready for. nginx as the single origin, so no CORS. Authentication left out on
purpose: it's a public utility. The gateway goes under "next steps".

## 7. Rules for the AI before any code

> Before we write code I want rules for the assistant, one generic file and one
> per layer. Mine: best practices and clean code, doc comments on every method,
> unit tests with every change, no mock data inside test bodies (fixtures in a
> mocks folder), no hardcoded values (constants per package or file), keep the
> structure. Add yours, and push back where you disagree.

→ `CLAUDE.md`, `backend/CLAUDE.md`, `frontend/CLAUDE.md`. The assistant pushed
back on Go folders like `controllers/services/utils`; we used idiomatic package
names instead, and `testdata/` for Go fixtures.

## 8. Understand the engine before trusting it

> Show me the same calculation with float64 and then with big.Rat, I want to see
> the difference myself.

> Show me the tree that comes out of the parser for `50 − 7 × 3`, then walk me
> through how the evaluator goes over it. Also set up a debugger so I can step
> through `50 * 7 - 3`.

→ Built the engine step by step (tokenizer, parser, evaluator, formatter), each
with its tests, after confirming float64 really fails (`0.7 + 0.1` →
`0.7999999999999999`).

## 9. Limits

> What's our overflow limit and how do we detect it? What happens if someone
> pastes a 150-digit number times another huge number? Why is one of them
> "infinity" and the other shows as `1.8e+184`?

→ Limits on length, digits per number and nesting, plus a range check on every
intermediate result (not only the final one), estimated cheaply before the
exact value is computed. A fuzz test then found one more crash, which was fixed.

## 10. Rate limiting behind a proxy

> Do we limit per person or per endpoint? How do we even see the client's IP?
> And if the client writes the `X-Forwarded-For` header itself, what does nginx
> do with it?

→ A per-IP token bucket. The API reads `X-Forwarded-For` only from trusted
proxies, and nginx overwrites the header instead of appending to it. Verified
later in Docker: 60 requests with 60 forged IPs were still limited as one client.

## 11. Frontend from a design file

> I designed the UI in Claude Design. Use it for the look only, don't paste it.
> Split it into a page and reusable components following the frontend rules.
> It has to be responsive. Keep it to 6–7 steps and show me the important parts:
> hooks, architecture rules, React features.

> Why is `kind: 'timeout'` hardcoded? Shouldn't that be a constant?

→ A page container plus presentational components, logic in hooks
(`useReducer` state machine, request cancellation with `AbortController`), pure
input rules in `utils/`. On `kind`: literal tags of a discriminated union stay
literal because the compiler checks them; this was added to the rules as the
only exception.

## 12. Bugs found by hand

> On the keyboard, `5%0` becomes `5%×0`, but clicking the buttons gives "cannot
> divide by zero". Why?

> If I click into the display and type `100%7`, I get a syntax error. Without
> focusing it, the same keys give `100%×7`. These should behave the same.

→ The first one was my mistake (I was pressing `%` for divide). The second was a
real bug: typing at the end of the display now goes through the same input
rules as the keypad, and editing inside the text stays free.

## 13. Documentation that gets read

> The README is too detailed and doesn't show the core logic. Nobody will read
> 250 lines. Under the architecture, show one example, `50 − 7 × 3`, through
> tokenize, parse (with the tree), evaluate with big.Rat and format, then the
> response. Then the main features in one line each. Decisions and curl
> examples can go further down.

> The Mermaid diagrams render tiny and can't be read without a preview.

→ The README now leads with how to run it, the architecture, the calculation
flow and a features table. Diagrams are plain text so they read the same
everywhere.
