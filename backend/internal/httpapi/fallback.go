package httpapi

import "net/http"

// withJSONFallback wraps mux so that its built-in 404 and 405 responses use the
// JSON error envelope instead of Go's plain-text defaults. Requests that match a
// registered route are passed through untouched.
func withJSONFallback(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(&fallbackWriter{ResponseWriter: w}, r)
	})
}

// fallbackWriter replaces a 404 or 405 written by the mux with the JSON error
// envelope and discards the mux's plain-text body. Any other status (such as a
// path-cleaning redirect) is written through unchanged. Headers the mux has
// already set, like Allow on a 405, are kept.
type fallbackWriter struct {
	http.ResponseWriter
	replaced bool
}

func (f *fallbackWriter) WriteHeader(status int) {
	switch status {
	case http.StatusNotFound:
		f.replaced = true
		writeError(f.ResponseWriter, status, codeNotFound, messageNotFound)
	case http.StatusMethodNotAllowed:
		f.replaced = true
		writeError(f.ResponseWriter, status, codeMethodNotAllowed, messageMethodNotAllowed)
	default:
		f.ResponseWriter.WriteHeader(status)
	}
}

func (f *fallbackWriter) Write(b []byte) (int, error) {
	if f.replaced {
		return len(b), nil
	}
	return f.ResponseWriter.Write(b)
}
