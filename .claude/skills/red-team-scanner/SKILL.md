---
name: red-team-scanner
description: >
  Offensive security pentester for Go vulnerability detection.
  Hunts exploitable vulnerabilities with PoCs and CVSS scoring.
  Use when performing penetration testing or vulnerability assessment.
user_invocable: true
---

# Red Team Scanner -- Go SAST

## Attack Vectors (Go specific)

### 1. SQL Injection
- Search for `fmt.Sprintf` or string concatenation near SQL queries
- Check if user input flows to query strings without `$1`, `$2` parameterization
- Verify pgx is used with parameterized queries consistently
- Watch for dynamic table/column names from user input

### 2. Broken Authentication
- JWT claim verification (iss, aud, exp, nbf)
- Algorithm restriction (reject `none`, verify against JWKS)
- API key timing attacks (constant-time comparison)
- Empty bearer token handling
- NATS consumer authentication

### 3. Anonymization Bypass
- Can events be crafted to bypass PII suppression?
- Are all PII fields enumerated and stripped? (patientId, actorId, memberId, victimId, caregiverId, professionalId)
- Can patientId hash be reversed? (weak salt, predictable input)
- Can quasi-identifier combinations uniquely identify individuals?
- Can temporal correlation across monthly snapshots reveal identity?

### 4. K-Anonymity Violations
- Can API query parameters be crafted to extract groups smaller than K=5?
- Is K-anonymity check applied at the database query level or only at the API response level?
- Can pagination be used to isolate small groups?
- Can multiple queries with different filters be combined to narrow down individuals?

### 5. Security Misconfiguration
- PostgreSQL TLS disabled
- NATS without auth/TLS
- Container running as root
- Default/fallback credentials in code
- Debug logging in production

### 6. Data Leakage via Export
- PII in Parquet file metadata (author, created_by)
- PII in ODS document properties
- PII in FHIR Bundle (Patient.identifier, Patient.name, Patient.address)
- PII in XML comments or processing instructions
- PII in CSV headers or comments
- DBF/DBC memo fields containing raw data

### 7. SSRF
- Any HTTP client calls with user-controlled input
- JWKS URL manipulation
- FHIR reference URLs

### 8. Business Logic Flaws
- Period manipulation to access closed/future periods
- Dataset parameter injection
- Format parameter path traversal
- Negative offset/limit for pagination bypass

### 9. Go-Specific Vulnerabilities
- Goroutine leaks (unbounded goroutine creation)
- `defer` in loops (resource accumulation)
- Type assertion panics (`v.(T)` without ok check)
- Template injection (if any template rendering exists)
- Integer overflow in pagination calculations
- Slice bounds out of range panics

### 10. Dependency Vulnerabilities
- Go modules with known CVEs (`govulncheck`)
- Outdated dependencies
- CGo dependencies (if DBC encoder uses CGo)

### 11. Event Injection
- NATS messages without authentication
- No schema validation on inbound events
- Malformed JSON causing panics in deserialization

## PoC Requirements
Every finding MUST include:
- Exact file and line number
- Description of the vulnerability
- Proof of Concept (curl command, code snippet, or attack scenario)
- CVSS score
- Remediation with Go code

## Scoring: Security Score XX/100
