---
name: appsec-code-reviewer
description: >
  Defensive secure code review for Go/chi/pgx. 10-point security checklist
  adapted for Go concurrency, pgx, chi middleware, anonymization, and LGPD compliance.
  Use when performing security-focused code reviews.
user_invocable: true
---

# AppSec Code Reviewer -- Go

## 10-Point Security Checklist

### 1. Input Validation
- [ ] All query parameters validated (period YYYY-MM, mesoregion codes, format enum)
- [ ] Pagination bounded (max limit, non-negative offset)
- [ ] Export format validated against allowlist
- [ ] Dataset parameter validated against allowlist
- [ ] No user input flows to SQL without parameterization

### 2. Output Safety
- [ ] No PII in any API response
- [ ] K-anonymity enforced on all indicator queries
- [ ] `suppressed_groups` in response meta
- [ ] No stack traces in production responses
- [ ] Content-Type explicit on all responses

### 3. Authentication & Authorization
- [ ] JWT verifies `exp`, `iss`, `aud`
- [ ] Auth middleware on all protected routes
- [ ] API key uses constant-time comparison
- [ ] Health/ready endpoints are public

### 4. Data Protection (LGPD)
- [ ] No PII in logs (no patient IDs, no CPF, no names)
- [ ] Anonymization pipeline strips ALL PII
- [ ] `PATIENT_HASH_SALT` not logged or exposed
- [ ] DB connection uses TLS
- [ ] NATS connection authenticated + encrypted
- [ ] Export files contain only anonymized, K-anonymous data

### 5. SQL Safety
- [ ] pgx parameterized queries (`$1`, `$2`)
- [ ] No `fmt.Sprintf` for SQL construction
- [ ] Dynamic table/column names from allowlist only

### 6. Dependency Health
- [ ] `go.sum` committed (integrity verification)
- [ ] `govulncheck` clean
- [ ] Dependencies from trusted publishers
- [ ] No known CVEs

### 7. Security Headers
- [ ] HSTS, X-Content-Type-Options, X-Frame-Options
- [ ] Cache-Control: no-store
- [ ] No X-Build-Version in production

### 8. Error Handling
- [ ] No swallowed errors (`if err != nil { }` with no body)
- [ ] No `panic` in production code
- [ ] Error wrapping with context
- [ ] Errors mapped to HTTP status codes
- [ ] No sensitive info in error messages

### 9. Concurrency Safety
- [ ] No goroutine leaks (context cancellation respected)
- [ ] `defer` for cleanup (connections, mutexes, files)
- [ ] Shared state protected by sync.Mutex or channels
- [ ] NATS consumer handles concurrent delivery safely
- [ ] Graceful shutdown waits for in-flight work

### 10. Anonymization Integrity
- [ ] patientId hashed with per-environment salt
- [ ] All PII fields suppressed (actorId, memberId, victimId, etc.)
- [ ] Age generalized to 5-year bands
- [ ] CEP generalized to mesoregion
- [ ] K-anonymity check before API output
- [ ] No PII in export file metadata

## Issue Format
```markdown
### [SEVERITY] Short description
**File**: `path/to/file.go:42`
**Category**: SQL Safety

**Before** (insecure):
(code)

**After** (secure):
(code)

**Why it matters**: ...
```

## Verdict: APPROVED or NEEDS FIXES
