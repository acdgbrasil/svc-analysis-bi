# TICKET-003: Ingestion Pipeline — Consumer, Anonymize, Materialize

## Scope

Implementar o `internal/ingestion/` — o pipeline que consome eventos NATS, anonimiza PII e materializa dados no star schema PostgreSQL:

1. **consumer.go** — NATS JetStream consumer (durable, at-least-once), deserializa JSON, valida schema, dedup via EventStore
2. **handler.go** — Event handler registry: mapeia EventType → handler function, orquestra o pipeline por evento
3. **anonymize.go** — Stage de anonimização: aplica domain.HashPatientID, domain.GeneralizeAge, domain.GeneralizeIncome, descarta PII (actorId, memberId, victimId, caregiverId, professionalId)
4. **materialize.go** — Stage de materialização: resolve dimension IDs via DimensionStores, faz UPSERT nas fact tables via FactStore
5. **pipeline.go** — Orquestração goroutines + channels: consumer → anonymize → materialize com graceful shutdown

## Dependencies

- TICKET-001 domain types (complete): HashPatientID, GeneralizeAge, GeneralizeIncome, GeographyLookup, EventTypes, event structs
- TICKET-002 infra foundation (complete): EventStore (dedup/DLQ), DimensionStores, pgx pool, config

## Constraints

- Pipeline usa goroutines + channels para stage processing (ADR-001 §3.2)
- Idempotent: dedup via event_processing_log antes de processar
- Dead letter: eventos com erro vão para event_dlq
- Anonymization at ingestion: PII NUNCA persiste (LGPD Art. 46)
- patientId → SHA-256 hash (domain.HashPatientID com salt do config)
- Campos descartados: actorId, memberId, victimId, caregiverId, professionalId
- birthDate → age band 5 anos (domain.GeneralizeAge)
- CEP → mesoregion IBGE (domain.GeographyLookup)
- totalFamilyIncome → income band (domain.GeneralizeIncome)
- Graceful shutdown via context cancellation
- No business logic na ingestion layer — chama funções do domain
- At-least-once delivery: ack NATS message somente após persistência bem-sucedida

## Source

ADR-001 sections: 3 (Architecture), 5 (Domain Events), 7 (Anonymization Flow)
