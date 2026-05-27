# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Service Overview

`svc-analysis-bi` is a **descriptive analytics service** (Go) for ACDG Brasil. It consumes domain events from `svc-social-care` via NATS JetStream, anonymizes PII, materializes data into a PostgreSQL star schema, and exposes REST endpoints for indicators and multi-format exports.

**Status:** Scaffold/template project — ADR, handbook, agents, and skills are defined; source code implementation follows the TDD pipeline.

## Stack

- **Go** (binary, ~10MB Docker image)
- **chi** — HTTP router
- **pgx v5** — PostgreSQL driver (dedicated analytical instance)
- **nats.go** — NATS JetStream consumer (durable, at-least-once)
- **Export:** CSV/JSON/XML (stdlib), Parquet (segmentio/parquet-go), DBF (go-dbf), DBC (LZ77/CGo), ODS (excelize), FHIR Bundle (manual BR Core structs)

## Commands (when go.mod exists)

```bash
go build ./cmd/server/          # Build
go test ./...                   # All tests
go test ./internal/domain/...   # Domain tests only
go test -run TestName ./path/   # Single test
go vet ./...                    # Static analysis
go test -race ./...             # Race detector
```

## Architecture

### Data Flow

```
svc-social-care → Outbox Relay → NATS JetStream
    → Event Consumer → Anonymizer → Materializer → PostgreSQL (analytical)
                                                         ↑
                                             API Layer (/indicators/*, /export/{format})
                                             Export Pipeline (8 encoders)
```

### Module Structure

```
cmd/server/main.go              — entrypoint
internal/
  domain/                       — pure logic: anonymization, K-anonymity, geography, indicators
  ingestion/                    — pipeline: consumer → anonymize → materialize (goroutines + channels)
  store/                        — pgx repositories, dimension/fact CRUD, migrations
  api/                          — chi router, handlers, middleware (JWT, rate limit, security headers)
  export/                       — 8 format encoders implementing a shared Encoder interface
configs/                        — env parsing, IBGE mesoregion CSV
migrations/                     — forward-only SQL
```

### Key Design Decisions (ADR-001)

- **Star schema** with 10 dimension tables and 7+ fact tables. Monthly grain via `dim_period`.
- **K-anonymity K=5** on (age_band, sex, mesoregion). Groups below threshold are suppressed at query time.
- **Anonymization at ingestion:** PII is never stored. patientId → SHA-256 (salted one-way hash). Names, CPF, NIS, CNS, exact addresses are discarded. birthDate → 5-year age bands. CEP → IBGE mesoregion.
- **Idempotent processing:** `event_processing_log` + `event_dlq` tables for dedup and dead-letter tracking.
- **Carry-forward job:** On the 1st of each month, copies latest patient snapshots to new period for continuity.
- **API-only service** (`svc-*` prefix). No UI — visual interface is a future `app-*` project.

### API Surface

```
GET /api/v1/indicators/{demographics,epidemiological,socioeconomic,protection,care}
GET /api/v1/export/{csv,json,xml,parquet,dbf,dbc,ods,fhir}
GET /api/v1/metadata/{datasets,formats,regions}
GET /health    GET /ready
```

Response envelope: `{ "data": [...], "meta": { timestamp, period, k_threshold, suppressed_groups, total_records } }`

### NATS Events Consumed

18 event types from `social-care.events.<EventType>` covering demographics, socioeconomic, epidemiological, care, and protection domains.

## Go Conventions

- Error handling: errors as values, no panics in domain/application. `panic` recovery only allowed in HTTP handlers (infra-implementer).
- Concurrency: ingestion pipeline uses goroutines + channels for stage processing.
- Table-driven tests with subtests as default pattern.
- Domain layer is pure — no I/O, no database imports.

## Multi-Agent Pipeline

Implementation follows a fail-first pipeline via `.pipeline/<ticket>/` folders:

1. `domain-architect` → type contracts (001-contracts/)
2. `test-writer` → failing tests from contracts only (002-tests/)
3. `domain-modeler` → pure domain logic (003-domain/ + src)
4. `application-orchestrator` → ingestion pipeline wiring (003-application/ + src)
5. `infra-implementer` → HTTP, pgx, NATS adapter, exports (003-infra/ + src)
6. `code-reviewer` → architectural audit (004-code-review/)
7. `go-quality-checker` → Go idioms audit (005-ts-quality/)
8. `integration-validator` → build + test + race (006-integration/)

Each agent writes a `REPORT.md` with a Public API section consumed by downstream agents. Max 3 review rounds before user escalation. One ticket = one atomic unit.

## Compliance

- **LGPD** (Lei 13.709/2018): data minimization, purpose limitation, ROPA traceability via processing logs.
- **FHIR R4 / BR Core**: read-only Bundle generation for RNDS-compatible exports (Patient, Observation, Condition, Encounter).
- **DataSUS compatibility**: DBF/DBC formats for TABWIN integration.

## Reference

- ADR-001 (full design): `handbook/principles/ADR-001-service-design.md`
- Handbook root: `handbook/README.md`
- Agent definitions: `.claude/agents/`
- Skill definitions: `.claude/skills/`

## Reference Network — consulta fria (especialistas externos)

Para FATOS de documentação de tecnologias (sintaxe, versão exata, comportamento), não responda de memória nem chute: consulte o especialista **EXTERNO read-only**, que cita a doc oficial offline (`infra/reference/`) ou recusa. Divisão: você (interno) conhece o código e **decide**; ele (externo) só entrega o **fato citado** — nunca vê seu código.

Invocação: delegue isolado via `subagent_type: "acdg-ref:ref-<tech>"`, ou direto `/acdg-ref:ref-<tech> <pergunta>`.

| Dúvida sobre… | Consulte |
|---|---|
| NATS/JetStream: durable consumer, ack, subjects, redelivery | `ref-nats` |
| SQL, tipos, funções, índices, star-schema queries (PostgreSQL) | `ref-postgresql` |

Ainda **fora da rede** (P2/P3): Go, chi, pgx, Parquet, FHIR, LGPD.

Regras: passe a pergunta como **texto** (não mande "olhe meu arquivo X" — ele recusa). Se retornar `NÃO ENCONTRADO`, não invente: escale ou peça download da doc. Detalhes: `infra/reference-network/README.md`.
