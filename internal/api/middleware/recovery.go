// Package middleware provides HTTP middleware for the analysis-bi API.
//
// recovery.go is the ONLY place in the codebase where panic recovery
// is allowed, per project conventions.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery returns middleware that recovers from panics, logs the stack
// trace, and returns a 500 Internal Server Error response.
// This is the ONLY place in the codebase where panic recovery is allowed.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error("panic recovered",
						"panic", fmt.Sprintf("%v", rec),
						"stack", string(stack),
						"method", r.Method,
						"path", r.URL.Path,
					)
					http.Error(w, `{"error":"Internal Server Error","status":500,"message":"internal error"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
