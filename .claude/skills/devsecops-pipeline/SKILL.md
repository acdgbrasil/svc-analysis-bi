---
name: devsecops-pipeline
description: >
  DevSecOps for Go infrastructure. Covers Docker, CI/CD (GitHub Actions),
  Go modules dependency security, secrets management, and supply chain.
  Use when auditing infrastructure, Dockerfile, CI/CD, or dependencies.
user_invocable: true
---

# DevSecOps Pipeline -- Go

## 6 Pillars

### 1. Go Module Dependency Security
- `go.sum` committed (integrity verification)
- `go mod verify` passes
- `govulncheck ./...` clean (no known vulnerabilities)
- Dependencies from trusted publishers
- Dependabot configured for gomod ecosystem
- Regular `go get -u` for patch updates

### 2. Docker Security
```dockerfile
# Required patterns:
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /server /server
COPY --from=build /app/configs/ibge_mesoregions.csv /configs/
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s CMD ["/server", "-health"]
ENTRYPOINT ["/server"]
```

Checklist:
- [ ] Base image pinned (not `:latest`)
- [ ] Non-root USER directive (distroless nonroot or explicit)
- [ ] Multi-stage build (builder + minimal runtime)
- [ ] CGO_ENABLED=0 for static binary (unless DBC needs CGo)
- [ ] No secrets in ENV/ARG
- [ ] `.dockerignore` excludes `.env`, `.git`, `vendor/`, `*_test.go`
- [ ] HEALTHCHECK defined
- [ ] Binary stripped (`-ldflags="-s -w"`)

### 3. CI/CD Pipeline (GitHub Actions)
- [ ] Actions pinned by SHA (not tag)
- [ ] `go vet ./...` before build
- [ ] `govulncheck ./...` in pipeline
- [ ] `go test -race -cover` runs BEFORE image push
- [ ] Security scanning step (Trivy container scan)
- [ ] Least privilege permissions on jobs
- [ ] Secrets in GitHub Secrets only

### 4. Secrets Management
- [ ] No secrets in source code (grep for patterns)
- [ ] No fallback credentials compiled into binary
- [ ] `PATIENT_HASH_SALT` treated as secret (never hardcoded)
- [ ] `.env` in `.gitignore`
- [ ] Pre-commit hooks (gitleaks)
- [ ] Different secrets per environment (dev/stg/prod)
- [ ] Bitwarden Secret Manager for prod secrets

### 5. Supply Chain
- [ ] SBOM generation (syft or cyclonedx-gomod)
- [ ] Container image signing (cosign/Sigstore)
- [ ] Immutable image tags (`sha-<commit>`, `vX.Y.Z`)
- [ ] `:latest` only on main, production uses digest
- [ ] `go.sum` provides checksum verification

### 6. Monitoring & Alerting
- [ ] Structured logging with `slog` (not `fmt.Println` or `log`)
- [ ] Health/readiness probes (K8s compatible)
- [ ] Graceful shutdown handler (SIGTERM/SIGINT)
- [ ] NATS consumer monitoring (failed event delivery, DLQ size)
- [ ] Export generation monitoring (timeouts, errors)
