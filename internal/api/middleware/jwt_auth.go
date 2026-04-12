package middleware

import (
	"context"
	"net/http"
	"strings"
)

// Claims holds the extracted JWT claims after validation.
type Claims struct {
	Subject string
	Roles   []string
}

// claimsContextKey is the context key for storing JWT claims.
type claimsContextKey struct{}

// ClaimsFromContext extracts Claims from the request context.
// Returns nil if no claims are present (e.g., unauthenticated request).
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsContextKey{}).(*Claims)
	return claims
}

// JWTValidator defines the interface for JWT token validation.
// Implementations may use JWKS, static keys, or test doubles.
type JWTValidator interface {
	// Validate parses and validates the raw JWT token string.
	// Returns the extracted claims on success, or an error if the
	// token is invalid, expired, or malformed.
	Validate(ctx context.Context, tokenString string) (*Claims, error)
}

// JWTAuth returns middleware that extracts and validates a Bearer token
// from the Authorization header. On success, it stores the claims in the
// request context. On failure, it returns 401 Unauthorized.
//
// Paths listed in skipPaths bypass authentication (e.g., /health, /ready).
func JWTAuth(validator JWTValidator, skipPaths map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for exempt paths
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Unauthorized","status":401,"message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"Unauthorized","status":401,"message":"invalid authorization scheme"}`, http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := validator.Validate(r.Context(), tokenString)
			if err != nil {
				http.Error(w, `{"error":"Unauthorized","status":401,"message":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
