package middleware

import (
	"net/http"
	"strings"
)

// RoleGuard returns middleware that checks if the authenticated user has
// at least one of the required roles. Returns 403 Forbidden if not.
//
// Supports composite role keys: a JWT role "analysis-bi:analyst" satisfies
// a guard requiring "analyst" (suffix matching after ":").
// The "superadmin" role bypasses all checks.
// The "admin" role (simple or composite e.g. "analysis-bi:admin") bypasses all checks.
func RoleGuard(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// superadmin and admin bypass all checks
			for _, role := range claims.Roles {
				if role == "superadmin" {
					next.ServeHTTP(w, r)
					return
				}
				if role == "admin" || strings.HasSuffix(role, ":admin") {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check required roles with composite suffix matching
			for _, required := range requiredRoles {
				for _, role := range claims.Roles {
					if role == required || strings.HasSuffix(role, ":"+required) {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			writeJSONError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}
