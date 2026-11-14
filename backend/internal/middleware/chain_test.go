package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChain_FirstMiddlewareIsOutermost(t *testing.T) {
	var order []string
	trace := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+" in")
				next.ServeHTTP(w, r)
				order = append(order, name+" out")
			})
		}
	}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "handler") })

	Chain(handler, trace("A"), trace("B")).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got, want := strings.Join(order, ", "), "A in, B in, handler, B out, A out"; got != want {
		t.Errorf("order = %s, want %s", got, want)
	}
}

func TestChain_WithoutMiddlewareReturnsTheHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
}
