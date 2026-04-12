package middleware

import "net/http"

// RoleGuard returns middleware that checks if the authenticated user has
// at least one of the required roles. Returns 403 Forbidden if not.
// The "admin" role always bypasses role checks.
func RoleGuard(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Admin role bypasses all checks
			for _, role := range claims.Roles {
				if role == "admin" {
					next.ServeHTTP(w, r)
					return
				}
			}

			for _, required := range requiredRoles {
				for _, role := range claims.Roles {
					if role == required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			writeJSONError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}
