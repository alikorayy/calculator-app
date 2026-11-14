package middleware

import "net/http"

// Chain wraps h in the given middleware. The first middleware listed is the
// outermost: it sees the request first and the response last.
//
//	Chain(h, A, B, C) == A(B(C(h)))
func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
