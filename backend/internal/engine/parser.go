package engine

import (
	"fmt"
	"math/big"
)

// The parser is a recursive-descent parser. Each grammar rule below is one
// method, and a rule calls the rule of the next higher precedence level:
//
//	expression := term (('+' | '-') term)*
//	term       := primary (('*' | '/') primary)*
//	primary    := NUMBER | '(' expression ')'
//
// Because expression is built out of terms, '*' and '/' bind tighter than
// '+' and '-'. The loops build the tree left to right, so operators of the same
// precedence are left-associative: 10 - 4 - 3 = (10 - 4) - 3.

// parser holds the tokens of one expression and a cursor into them.
type parser struct {
	tokens []token
	pos    int
}

// parse turns expr into an AST, or returns a SYNTAX_ERROR pointing at the
// first token that does not fit the grammar.
func parse(expr string) (node, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}

	p := &parser{tokens: tokens}
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

// parseTerm handles '*' and '/', which bind tighter than '+' and '-'.
func (p *parser) parseTerm() (node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokenMultiply || p.peek().kind == tokenDivide {
		op := p.advance()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op.kind, left: left, right: right}
	}
	return left, nil
}

// parsePrimary handles a number or a parenthesized sub-expression.
func (p *parser) parsePrimary() (node, error) {
	t := p.peek()
	switch t.kind {
	case tokenNumber:
		p.advance()
		return parseNumber(t)
	case tokenLeftParen:
		p.advance()
		inner, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if closing := p.peek(); closing.kind != tokenRightParen {
			return nil, newSyntaxError(closing.pos, msgMissingClosingParen)
		}
		p.advance()
		return inner, nil
	default:
		return nil, unexpectedToken(t)
	}
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
