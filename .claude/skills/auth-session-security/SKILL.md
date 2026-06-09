---
name: auth-session-security
description: >
  Identity & Access Management expert for Go/chi with Authentik OIDC.
  Covers JWT verification, API key validation, NATS authentication, audit trail.
  Use when auditing or implementing authentication/authorization.
user_invocable: true
---

# Auth & Session Security -- Go / chi + Authentik OIDC

## JWT Verification

### Required Claims
```go
type Claims struct {
    Subject  string   `json:"sub"`
    Issuer   string   `json:"iss"`
    Audience []string `json:"aud"`
    Expires  int64    `json:"exp"`
    NotBefore int64   `json:"nbf,omitempty"`
    // Authentik delivers roles as groups "<system>:<role>" (e.g.
    // "analysis-bi:analyst") + the global "superadmin" in the `groups` claim.
    Groups   []string `json:"groups,omitempty"`
}

func (c *Claims) Validate(expectedIssuer, expectedAudience string) error {
    if time.Now().Unix() > c.Expires {
        return ErrTokenExpired
    }
    if c.Issuer != expectedIssuer {
        return ErrInvalidIssuer
    }
    if !containsAudience(c.Audience, expectedAudience) {
        return ErrInvalidAudience
    }
    return nil
}
```

### Authentik Configuration
- **Issuer**: `<AUTHENTIK_URL>/application/o/<slug>/` (e.g. `https://auth.acdg-bv.org.br/application/o/<slug>/`)
- **JWKS**: `<AUTHENTIK_URL>/application/o/<slug>/jwks/`
- **Signing alg**: RS256 (the validator is RS256-only)
- **Roles claim**: `groups` — values are `<system>:<role>` (e.g. `analysis-bi:analyst`) plus the global `superadmin`
- **RBAC is system-scoped**: only `analysis-bi:*` (and bare `admin`/`superadmin`) grant access here; a foreign `social-care:admin` does not (see `RoleGuard`).

### JWT Middleware Pattern
```go
func JWTAuth(jwksURL, issuer, audience string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            claims, err := verifyJWT(token, jwksURL, issuer, audience)
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), claimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## API Key Authentication (for programmatic access)
- API keys for automated consumers (research tools, data pipelines)
- Keys stored hashed (SHA-256), never plaintext
- Constant-time comparison to prevent timing attacks
- Scoped to read-only operations (indicators + export)
- Rate-limited per key

## Route Protection
- `/health`, `/ready` -- public (no auth)
- `/api/v1/indicators/*` -- JWT or API key required
- `/api/v1/export/*` -- JWT or API key required
- `/api/v1/metadata/*` -- JWT or API key required (or public, depending on sensitivity)

## Infrastructure Auth
- PostgreSQL: TLS required (`sslmode=require` in connection string)
- NATS: nkey or token authentication + TLS
- No fallback credentials in code -- fail-fast if env vars missing
- `PATIENT_HASH_SALT` is a secret -- treat as credential

## NATS Consumer Auth
- Dedicated credentials for the consumer (not shared with API)
- Events validated against expected schema before processing
- Consumer uses durable subscription (at-least-once delivery)
- No unauthenticated event injection possible (NATS server enforces auth)
