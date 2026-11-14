package engine

import (
	"fmt"
	"math/big"
)

// The parser is a recursive-descent parser. Each grammar rule below is one
// method, and a rule calls the rule of the next higher precedence level:
//
//	expression := term (('+' | '-') term)*
//	term       := unary (('*' | '/' | implicit) unary)*
//	unary      := ('-' | '+') unary | power
//	power      := postfix ('^' unary)?
//	postfix    := primary ('%')*
//	primary    := NUMBER | '(' expression ')' | SQRT '(' expression ')'
//
// Because expression is built out of terms, '*' and '/' bind tighter than
// '+' and '-'. The loops build the tree left to right, so operators of the same
// precedence are left-associative: 10 - 4 - 3 = (10 - 4) - 3.
//
// A sign in front of an operand is unary and binds tighter than '*' and '/':
// -2 * 3 = (-2) * 3. Signs can be repeated: --5 = 5.
//
// '^' binds tighter than a sign, so -2^2 = -(2^2) = -4, as on Google and in
// standard mathematical notation. Its exponent is parsed as a unary, which
// makes '^' right-associative (2^3^2 = 2^(3^2) = 512) and allows a signed
// exponent (2^-1 = 0.5).
//
// '%' is a postfix operator that binds tightest of all: 10%^2 = (10%)^2.
// What a percentage means depends on where it is used; see evaluateBinary.
//
// The square root (√ or sqrt) is used only in function form with parentheses,
// as on Google: √(9). This leaves no doubt about how much it covers.
//
// A multiplication sign may be left out before a parenthesis or a square root,
// and between a closing parenthesis and a number: 2(3+1), (1+2)(3+4), 2√(9),
// (2)3. Such an implicit multiplication has the same precedence as '*' and is
// evaluated left to right, as on Google: 6/2(1+2) = (6/2)*3 = 9. Two numbers
// in a row ("2 3") are still a syntax error.
//
// Parentheses (including those of a square root) may be nested at most
// maxNestingDepth levels deep, which also bounds the parser's recursion.

// parser holds the tokens of one expression and a cursor into them.
type parser struct {
	tokens []token
	pos    int
	// previous is the kind of the most recently consumed token.
	previous tokenKind
	// depth is the number of currently open parentheses.
	depth int
}

// parse turns expr into an AST, or returns a SYNTAX_ERROR pointing at the
// first token that does not fit the grammar.
func parse(expr string) (node, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}

	p := &parser{tokens: tokens, previous: tokenEOF}
	if p.peek().kind == tokenEOF {
		return nil, newSyntaxError(0, msgEmptyExpression)
	}

	tree, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if next := p.peek(); next.kind != tokenEOF {
		return nil, unexpectedToken(next)
	}
	return tree, nil
}

// peek returns the current token without consuming it.
func (p *parser) peek() token {
	return p.tokens[p.pos]
}

// advance consumes and returns the current token. The cursor never moves past
// the final EOF token, so peek is always safe.
func (p *parser) advance() token {
	t := p.tokens[p.pos]
	if t.kind != tokenEOF {
		p.pos++
	}
	p.previous = t.kind
	return t
}

// parseExpression handles '+' and '-', the lowest precedence level.
func (p *parser) parseExpression() (node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokenPlus || p.peek().kind == tokenMinus {
		op := p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.kind, left: left, right: right}
	}
	return left, nil
}

// parseTerm handles '*', '/' and implicit multiplication, which bind tighter
// than '+' and '-'.
func (p *parser) parseTerm() (node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		var op tokenKind
		switch {
		case p.peek().kind == tokenMultiply || p.peek().kind == tokenDivide:
			op = p.advance().kind
		case p.startsImplicitMultiplication():
			op = tokenMultiply
		default:
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op, left: left, right: right}
	}
}

// startsImplicitMultiplication reports whether the next token starts a new
// factor directly after the previous one, with the '*' left out.
func (p *parser) startsImplicitMultiplication() bool {
	switch p.peek().kind {
	case tokenLeftParen, tokenSqrt:
		return p.previous == tokenNumber || p.previous == tokenRightParen || p.previous == tokenPercent
	case tokenNumber:
		return p.previous == tokenRightParen
	default:
		return false
	}
}

// parseUnary handles a leading sign. '-' negates its operand; '+' changes
// nothing and is dropped from the tree.
func (p *parser) parseUnary() (node, error) {
	switch p.peek().kind {
	case tokenMinus:
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &negateNode{operand: operand}, nil
	case tokenPlus:
		p.advance()
		return p.parseUnary()
	default:
		return p.parsePower()
	}
}

// parsePower handles '^'. The exponent may itself contain '^' or a sign, so it
// is parsed by parseUnary; this recursion makes '^' right-associative.
func (p *parser) parsePower() (node, error) {
	base, err := p.parsePostfix()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tokenPower {
		return base, nil
	}
	op := p.advance()
	exponent, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	return &binaryNode{op: op.kind, left: base, right: exponent}, nil
}

// parsePostfix handles any number of '%' signs after an operand: 50%%.
func (p *parser) parsePostfix() (node, error) {
	operand, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokenPercent {
		p.advance()
		operand = &percentNode{operand: operand}
	}
	return operand, nil
}

// parsePrimary handles a number, a parenthesized sub-expression or a square root.
func (p *parser) parsePrimary() (node, error) {
	t := p.peek()
	switch t.kind {
	case tokenNumber:
		p.advance()
		return parseNumber(t)
	case tokenLeftParen:
		return p.parseParenthesized()
	case tokenSqrt:
		return p.parseSqrt()
	default:
		return nil, unexpectedToken(t)
	}
}

// parseParenthesized handles '(' expression ')'. The parentheses only shape
// the tree; they do not appear in it.
func (p *parser) parseParenthesized() (node, error) {
	opening := p.advance()
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxNestingDepth {
		return nil, newErrorAt(CodeLimitExceeded, opening.pos, fmt.Sprintf(msgNestedTooDeeply, maxNestingDepth))
	}

	inner, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if closing := p.peek(); closing.kind != tokenRightParen {
		return nil, newSyntaxError(closing.pos, msgMissingClosingParen)
	}
	p.advance()
	return inner, nil
}

// parseSqrt handles √(expression) and sqrt(expression).
func (p *parser) parseSqrt() (node, error) {
	sqrt := p.advance()
	if next := p.peek(); next.kind != tokenLeftParen {
		return nil, newSyntaxError(next.pos, fmt.Sprintf(msgExpectedParenAfter, sqrt.text))
	}
	operand, err := p.parseParenthesized()
	if err != nil {
		return nil, err
	}
	return &sqrtNode{operand: operand}, nil
}

// parseNumber converts a number token into an exact rational.
func parseNumber(t token) (node, error) {
	value, ok := new(big.Rat).SetString(t.text)
	if !ok {
		// The tokenizer only produces literals big.Rat accepts; this guards the invariant.
		return nil, newSyntaxError(t.pos, fmt.Sprintf(msgInvalidNumber, t.text))
	}
	return &numberNode{value: value}, nil
}

// unexpectedToken reports t as not allowed where it appears.
func unexpectedToken(t token) error {
	if t.kind == tokenEOF {
		return newSyntaxError(t.pos, msgUnexpectedEnd)
	}
	return newSyntaxError(t.pos, fmt.Sprintf(msgUnexpectedToken, t.text))
}
