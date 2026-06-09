package middleware

import (
	"net/http"
)

// systemName scopes composite role keys to THIS service. Authentik delivers
// roles as groups named "<system>:<role>" (svc-people-context ADR-029); a
// guard here is satisfied only by roles scoped to this system. Foreign roles
// (e.g. "social-care:admin") never grant access to analysis-bi endpoints.
const systemName = "analysis-bi"

// RoleGuard returns middleware that checks if the authenticated user holds at
// least one of the required roles. Returns 403 Forbidden otherwise.
//
// Role keys are SYSTEM-SCOPED, mirroring the system-scoped RBAC of
// svc-people-context (ADR-029, v0.3.1):
//   - a composite role satisfies the guard only when scoped to this system:
//     "analysis-bi:analyst" satisfies require "analyst", but a foreign
//     "social-care:analyst" does NOT.
//   - "admin" (bare) or "analysis-bi:admin" bypasses every check for this
//     system; a foreign "<other>:admin" does NOT grant admin here.
//   - the global "superadmin" bypasses all checks across every system.
func RoleGuard(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// superadmin (global) and admin scoped to this system bypass checks.
			for _, role := range claims.Roles {
				if role == "superadmin" {
					next.ServeHTTP(w, r)
					return
				}
				if role == "admin" || role == systemName+":admin" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Required-role match, scoped to this system: a bare "<role>" or
			// "analysis-bi:<role>" satisfies; foreign "<other>:<role>" does not.
			for _, required := range requiredRoles {
				for _, role := range claims.Roles {
					if role == required || role == systemName+":"+required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			writeJSONError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}
