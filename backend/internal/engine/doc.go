// Package engine evaluates calculator expressions with exact arithmetic.
//
// It is pure domain logic: it performs no I/O, holds no global state and knows
// nothing about HTTP. Every failure is reported as an *Error with a stable Code.
package engine
