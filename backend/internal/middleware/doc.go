// Package middleware contains the HTTP middleware that wraps the API: each
// function takes the next http.Handler and returns a handler that does its
// work before and/or after calling it.
package middleware
