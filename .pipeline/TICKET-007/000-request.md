# TICKET-007: Final Wiring — API Handlers + cmd/server/main.go

## Scope

Conectar todos os componentes implementados em tickets anteriores num servidor funcional:

1. **`internal/api/handlers/indicators.go`** — Handler para GET /api/v1/indicators/{axis}
   - Parseia query params (period_start, period_end, mesoregion, granularity, top)
   - Chama IndicatorStore.Query{Demographics,Epidemiological,Socioeconomic,Protection,Care}
   - Retorna Response envelope com K-anonymity metadata

2. **`internal/api/handlers/export.go`** — Handler para GET /api/v1/export/{format}
   - Parseia query params (dataset, period_start, period_end, mesoregion)
   - Busca dados via IndicatorStore
   - Codifica com o Encoder correto do Registry
   - Seta Content-Type e Content-Disposition headers

3. **`internal/api/handlers/metadata.go`** — Handler para GET /api/v1/metadata/{resource}
   - Retorna datasets disponíveis, formatos suportados, regiões

4. **`internal/api/router.go`** — Atualizar para usar handlers reais em vez de placeholders 501
   - Injetar IndicatorStore e export.Registry via RouterDeps

5. **`cmd/server/main.go`** — Entrypoint completo:
   - Load config → connect DB → run migrations → create stores
   - Create pipeline (ingestion) → create router → start HTTP server
   - Graceful shutdown via signal handling

6. **Migração para chi/v5** (OPCIONAL, se viável):
   - O TICKET-004 usou net/http.ServeMux por precaução
   - chi/v5 está no go.mod — migrar router.go se chi for mais expressivo
   - Decisão: manter ServeMux se equivalente, migrar se chi adiciona valor

## Dependencies

- TICKET-001 (domain): COMPLETE
- TICKET-002 (store foundation): COMPLETE
- TICKET-003 (ingestion pipeline): COMPLETE
- TICKET-004 (API foundation): COMPLETE
- TICKET-005 (indicator store): COMPLETE
- TICKET-006 (export encoders): COMPLETE

## Constraints

- cmd/server/main.go é o ÚNICO arquivo que importa todos os pacotes
- Handlers importam store e export — boundaries OK (adapters can import everything)
- Graceful shutdown: context + signal.NotifyContext
- No PII em logs ou error messages
- Panic recovery APENAS no middleware (já implementado)
