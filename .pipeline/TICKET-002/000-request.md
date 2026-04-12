# TICKET-002: Infrastructure Foundation — Config, Migrations, Store

## Scope

Implementar a base de infraestrutura que habilita todos os demais componentes:

1. **configs/config.go** — Parsing de variáveis de ambiente (DB, NATS, server, hash salt)
2. **migrations/** — SQL files para o star schema completo (10 dimension tables, 7 fact tables, 2 control tables)
3. **internal/store/postgres.go** — Pool de conexão pgx v5
4. **internal/store/migrations.go** — Runner de migrations (forward-only)
5. **Makefile** — Targets: build, test, lint, dev, docker-build, migrate
6. **Dockerfile** — Multi-stage build (scratch final)

## Dependencies

- TICKET-001 domain types (complete)
- ADR-001 sections 4 (Data Model), 10 (Project Structure)

## Constraints

- pgx v5 (latest secure version: v5.9.1)
- Migrations forward-only (no rollback)
- Config via environment variables only (12-factor)
- Zero PII in any migration or config
- Connection pool with sensible defaults (max 10 conns, idle timeout)
