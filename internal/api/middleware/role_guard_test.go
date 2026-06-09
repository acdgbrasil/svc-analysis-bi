package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleGuard_MatchingRole(t *testing.T) {
	handler := RoleGuard("analyst", "exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "user-1",
		Roles:   []string{"analyst"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRoleGuard_NoMatchingRole(t *testing.T) {
	handler := RoleGuard("analyst", "exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "user-1",
		Roles:   []string{"viewer"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if msg, _ := body["message"].(string); msg != "insufficient permissions" {
		t.Errorf("message = %q, want %q", msg, "insufficient permissions")
	}
}

func TestRoleGuard_NoClaims(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRoleGuard_AdminBypassesAll(t *testing.T) {
	handler := RoleGuard("exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "admin-user",
		Roles:   []string{"admin"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — admin should bypass role checks", rr.Code, http.StatusOK)
	}
}

func TestRoleGuard_MultipleRequiredAnyMatch(t *testing.T) {
	handler := RoleGuard("analyst", "exporter", "viewer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "user-2",
		Roles:   []string{"exporter"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRoleGuard_EmptyRoles(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "user-noroles",
		Roles:   []string{},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// ─── System-scoped composite role matching tests ────────────────

func TestRoleGuard_ScopedCompositeRoleMatches(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "user-1",
		Roles:   []string{"analysis-bi:analyst"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — own-system composite role should match", rr.Code, http.StatusOK)
	}
}

func TestRoleGuard_CompositeAdminBypassesAll(t *testing.T) {
	handler := RoleGuard("exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "admin-user",
		Roles:   []string{"analysis-bi:admin"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — own-system composite admin should bypass", rr.Code, http.StatusOK)
	}
}

func TestRoleGuard_SuperadminBypassesAll(t *testing.T) {
	handler := RoleGuard("exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "super-user",
		Roles:   []string{"superadmin"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — superadmin should bypass all checks", rr.Code, http.StatusOK)
	}
}

// ─── System-scoped RBAC: cross-system rejection ─────────────────
//
// Roles are scoped to their own system (svc-people-context ADR-029). A role
// granted for another service must never confer access on analysis-bi. Only
// the global "superadmin" crosses system boundaries.

func TestRoleGuard_ForeignSystemAdminDenied(t *testing.T) {
	handler := RoleGuard("exporter")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "foreign-admin",
		Roles:   []string{"social-care:admin"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d — foreign-system admin must not bypass", rr.Code, http.StatusForbidden)
	}
}

func TestRoleGuard_ForeignSystemRoleDenied(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "foreign-analyst",
		Roles:   []string{"social-care:analyst"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d — foreign-system role must not match", rr.Code, http.StatusForbidden)
	}
}

// TestRoleGuard_PrefixCollisionDenied guards against a regression to prefix/
// suffix matching: a system whose name merely starts with "analysis-bi"
// (e.g. "analysis-bi2") must NOT satisfy an analysis-bi guard. Exact equality
// on "<systemName>:<role>" is required.
func TestRoleGuard_PrefixCollisionDenied(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "collision-user",
		Roles:   []string{"analysis-bi2:analyst"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d — prefix-collision system must not match", rr.Code, http.StatusForbidden)
	}
}

// TestRoleGuard_CompositeSuperadminDenied ensures only the bare global
// "superadmin" crosses system boundaries; a composite "<other>:superadmin" is
// just a foreign role and must NOT bypass.
func TestRoleGuard_CompositeSuperadminDenied(t *testing.T) {
	handler := RoleGuard("analyst")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsContextKey{}, &Claims{
		Subject: "fake-super",
		Roles:   []string{"social-care:superadmin"},
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d — composite superadmin must not bypass", rr.Code, http.StatusForbidden)
	}
}
