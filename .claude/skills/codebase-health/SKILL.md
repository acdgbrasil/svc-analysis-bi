---
name: codebase-health
description: >
  Codebase health check comparing CLAUDE.md declarations vs actual code reality.
  Detects drift: folders that don't exist, import boundary violations, patterns
  declared but not followed, dead code, missing test coverage.
  Trigger for: "health check", "audit", "check code quality", "verify architecture".
user_invocable: true
---

# Codebase Health Check -- analysis-bi (Go)

## What to Check

### 1. Folder Structure Drift
Compare CLAUDE.md / ADR-001 declared structure vs actual project contents.
Flag: missing directories, unexpected files, misplaced code.
Expected structure:
```
cmd/server/main.go
internal/domain/
internal/ingestion/
internal/store/
internal/api/
internal/export/
configs/
migrations/
go.mod
Dockerfile
Makefile
```

### 2. Import Boundary Violations
- `internal/domain/` must NOT import from any other `internal/` package
- `internal/ingestion/` must NOT import from `internal/api/` or `internal/export/`
- `internal/store/` must NOT import from `internal/api/` or `internal/export/`
- `internal/export/` must NOT import from `internal/api/` or `internal/store/`
- `internal/api/` can import from `internal/store/` and `internal/export/`
- `cmd/` imports from `internal/` only for wiring

Scan: `import` blocks in every .go file.

### 3. Prohibited Patterns
- `panic` in non-test production code (except init functions for truly unrecoverable cases)
- `log.Fatal` or `os.Exit` outside of `main()`
- `fmt.Sprintf` near SQL queries (indicates SQL injection risk)
- `interface{}` or `any` without type assertion/narrowing
- `_ = err` (swallowed errors)
- Hardcoded credentials or secrets
- `print`/`println` instead of `slog.Logger` in production code

### 4. Goroutine Safety
- All goroutines have cancellation path via `context.Context`
- Shared mutable state protected by `sync.Mutex` or channels
- No data races (verified by `go test -race`)
- `defer` used for resource cleanup

### 5. Error Handling Chain
- Domain errors: sentinel values or typed error structs, never panic
- Ingestion errors: wrapped with context, events NAK'd on failure
- Store errors: wrapped with context, transactions rolled back on failure
- API errors: safe message to client, full context to structured log
- No empty error checks

### 6. Test Coverage Map
- `internal/domain/` tests in same package or `_test` package
- `internal/ingestion/` tests with mock dependencies
- `internal/store/` tests (integration or mock)
- `internal/api/` tests (httptest)
- `internal/export/` tests for each encoder

### 7. Security Posture Quick Scan
- JWT claims verified (iss, aud, exp)
- Auth middleware on all protected endpoints
- No PII in analytical store
- K-anonymity enforced on indicator queries
- SQL parameterized (no fmt.Sprintf for queries)
- No hardcoded credentials
- Anonymization pipeline strips all PII

### 8. Anonymization Integrity
- patientId hashed with per-environment salt
- actorId, memberId, victimId, caregiverId, professionalId discarded
- Age generalized to 5-year bands
- CEP generalized to IBGE mesoregion
- Income generalized to bands (relative to minimum wage)
- K-anonymity check runs before data reaches API

## Output Format
```markdown
# Codebase Health Report
**Date**: YYYY-MM-DD
**Score**: XX/100

| Category | Status | Issues |
|----------|--------|--------|
| Structure | OK/DRIFT | ... |
| Imports | OK/VIOLATION | ... |
| Patterns | OK/VIOLATION | ... |
| Goroutines | OK/WARNING | ... |
| Errors | OK/GAP | ... |
| Tests | XX% | ... |
| Security | OK/WARNING | ... |
| Anonymization | OK/WARNING | ... |

## Violations (by severity)
### HIGH
### MEDIUM
### LOW

## Recommendations
```
