---
name: api-security-guardian
description: >
  API Security expert for Go/chi REST APIs. Covers input validation, rate limiting,
  CORS, security headers, error handling, K-anonymity enforcement, and abuse protection.
  Use when auditing or hardening API endpoints.
user_invocable: true
---

# API Security Guardian -- chi Router

## Security Dimensions

### 1. Transport Security
- All connections (DB, NATS) use TLS
- HSTS header present (`Strict-Transport-Security: max-age=31536000; includeSubDomains`)

### 2. Input Validation
- Query parameters validated: period format (YYYY-MM), mesoregion code, format enum
- Pagination parameters bounded: max limit (e.g., 100), non-negative offset
- Export format validated against allowlist: csv, json, xml, parquet, dbf, dbc, ods, fhir
- Dataset parameter validated against allowlist: full, demographics, epidemiological, socioeconomic, protection, care
- No user input flows to SQL without parameterization

### 3. Authentication
- JWT or API key middleware on all indicator and export endpoints
- Claims verified: `exp`, `iss`, `aud`
- Empty bearer token rejected early
- API key validated with constant-time comparison

### 4. Rate Limiting
- Per-IP rate limit on all endpoints
- Per-API-key rate limit on authenticated endpoints
- Stricter limits on export endpoints (expensive operations)
- 429 Too Many Requests with Retry-After header

### 5. CORS
- Explicit restrictive config (or deny-all for API-only service)
- No wildcard `*` with credentials

### 6. Security Headers
```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Cache-Control", "no-store")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

### 7. Error Handling
- Safe error message to client, full context to `slog.Logger` only
- No stack traces in production responses
- No build version or internal IDs in response headers
- Consistent error response format: `{"error": {"code": "...", "message": "..."}}`

### 8. Response Safety
- K-anonymity enforced: suppress groups with count < K=5
- `meta.suppressed_groups` in response for transparency
- `meta.k_threshold` always present (value: 5)
- No PII in any response -- ever
- Explicit Content-Type per response (application/json, text/csv, etc.)
- Content-Disposition with safe filenames on export

### 9. Export Security
- Streaming for large datasets (io.Writer based, not full dataset in memory)
- No PII in file metadata (author, comments, custom properties)
- Filename sanitization (no path traversal via format parameter)
- Timeout on export generation to prevent resource exhaustion

## Middleware Stack Order
```
chi.Recoverer            -> Panic recovery
middleware.RequestID      -> Request ID
middleware.RealIP         -> Real IP extraction
SecurityHeaders           -> Security headers on all responses
RateLimit                -> Throttle abusive clients
CORS                     -> Restrictive CORS policy
Logger                   -> Structured request logging
// JWTAuth is per-route group, not global (health/ready are public)
```
