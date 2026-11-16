# Engine pipeline

The four stages of `engine.Evaluate`. Values stay exact rationals from the
parser to the formatter; the only rounding happens once, at the very end.

```text
  "50 − 7 × 3"
       │
       ▼
  1. tokenize     string → tokens
       │          50   −   7   ×   3        (each with its character position)
       │          errors: SYNTAX_ERROR, LIMIT_EXCEEDED
       ▼
  2. parse        tokens → syntax tree      (see below)
       │          precedence becomes the shape of the tree
       │          errors: SYNTAX_ERROR with position, LIMIT_EXCEEDED
       ▼
  3. evaluate     syntax tree → exact fraction (big.Rat)
       │          √ and fractional powers via 256-bit big.Float
       │          errors: DIVISION_BY_ZERO, UNDEFINED_RESULT,
       │                  DOMAIN_ERROR, OVERFLOW
       ▼
  4. format       fraction → decimal text, rounded once to 16 significant digits
       │
       ▼
  {"result": "29", "approximate": false}
```

| Stage    | Guarantees |
|----------|------------|
| tokenize | ASCII and Unicode operators (`*` `×`, `/` `÷`, `-` `−`); at most 256 characters and 32 significant digits per number |
| parse    | Standard precedence, `^` right-associative, implicit multiplication; at most 32 nested parentheses |
| evaluate | `+ − × ÷` and integer powers are exact; every intermediate result is range-checked |
| format   | `−0` becomes `0`, trailing zeros are trimmed, scientific notation for very large or small values; `approximate` tells whether digits were dropped |

## The syntax tree

The parser builds an abstract syntax tree (AST): operators are nodes, numbers
are leaves, and precedence is the shape of the tree. Parentheses and spaces are
gone; `50−7×3` and `50 − (7 × 3)` give the same tree.

```text
            −            50 − 21 = 29
           / \
         50   ×          7 × 3 = 21
             / \
            7   3
```

The tree also decides the meaning of percent. In `50 + 10%` the whole right
operand of `+` is a percentage, so it means "10 % of 50"; in `200 × 10%` it is
simply `0.1`. This matches the Windows and Google calculators.

```text
            +            50 + 5 = 55
           / \
         50   %          10 % of 50 = 5
              │
              10
```

## Grammar

From the lowest precedence to the highest:

```text
expression := term (('+' | '-') term)*
term       := unary (('*' | '/' | implicit) unary)*
unary      := ('-' | '+') unary | power
power      := postfix ('^' unary)?
postfix    := primary ('%' | '²' | '⁻¹')*
primary    := NUMBER | '(' expression ')' | '√(' expression ')'
```
