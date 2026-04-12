package middleware

import "net/http"

// SecurityHeaders returns middleware that sets OWASP-recommended security
// headers on every response.
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			// HSTS: force HTTPS for 1 year, include subdomains
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			// Prevent MIME-type sniffing
			h.Set("X-Content-Type-Options", "nosniff")
			// Prevent framing (clickjacking)
			h.Set("X-Frame-Options", "DENY")
			// Restrict content sources -- API-only service, no scripts/styles
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			// Referrer policy: send origin only for same-origin requests
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			// Disable browser features not needed by an API
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			// Signal to caches that this content should not be cached
			h.Set("Cache-Control", "no-store")

			next.ServeHTTP(w, r)
		})
	}
}
