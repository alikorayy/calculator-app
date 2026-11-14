package middleware

import "net/http"

// SecurityHeaders sets headers that tell browsers to restrict what they do
// with the API's responses:
//
//   - X-Content-Type-Options: nosniff stops a browser from guessing that a
//     JSON response is HTML and running it;
//   - Content-Security-Policy forbids loading any resource from a response and
//     embedding it in a frame (clickjacking);
//   - Referrer-Policy: no-referrer keeps URLs from leaking to other sites;
//   - Cross-Origin-Resource-Policy: same-origin stops other sites from loading
//     the responses.
//
// Strict-Transport-Security is deliberately not set here: HTTPS is terminated
// in front of this service, and that layer owns it.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set(headerContentTypeOptions, valueNoSniff)
		header.Set(headerContentSecurityPolicy, valueContentSecurityPolicy)
		header.Set(headerReferrerPolicy, valueNoReferrer)
		header.Set(headerCrossOriginResourcePolicy, valueSameOrigin)
		next.ServeHTTP(w, r)
	})
}
