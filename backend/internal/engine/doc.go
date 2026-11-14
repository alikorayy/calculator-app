// Package engine evaluates calculator expressions with exact arithmetic.
//
// The public surface is small: Evaluate computes an expression such as
// "50 - 7 × 3", Calculate applies a single Operation to operands, both return
// a Result, and every failure is an *Error with a stable Code.
//
// Internally an expression goes through a pipeline: tokenize splits it into
// tokens, parse builds an abstract syntax tree that encodes operator
// precedence, evaluate walks the tree with exact rational arithmetic
// (*big.Rat), and format rounds the value once, for display.
//
// It is pure domain logic: it performs no I/O, holds no global state and knows
// nothing about HTTP.
package engine
