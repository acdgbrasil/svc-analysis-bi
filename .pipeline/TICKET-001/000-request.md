# TICKET-001: Domain Foundation

## Scope

Implementar os tipos e funções puras do `internal/domain/`:

1. **anonymizer.go** — Suppression (hash SHA-256 com salt) e generalization (birthDate→age band, CEP→mesoregion, income→band)
2. **k_anonymity.go** — Verificação K-anonymity K=5 sobre quasi-identifiers (age_band, sex, mesoregion)
3. **geography.go** — Mapeamento CEP→mesoregion IBGE
4. **indicator.go** — Tipos de indicadores (demographics, epidemiological, socioeconomic, protection, care)
5. **ageband.go** — Faixas etárias de 5 anos
6. **incomeband.go** — Faixas de renda em salários mínimos
7. **period.go** — Tipo Period (year, month) para série temporal

## Constraints

- Domínio puro: zero I/O, zero dependências externas
- Errors as values (nunca panic/throw)
- Immutable structs
- K=5 hardcoded como constante
- SHA-256 hash do patientId com salt parametrizável
- Faixas etárias: 0-4, 5-9, 10-14, ..., 75-79, 80+
- Faixas de renda: 0-0.5SM, 0.5-1SM, 1-2SM, 2-3SM, 3-5SM, 5+SM
- Geografia: CEP (8 dígitos) → mesoregion code (4 dígitos IBGE)

## Source

ADR-001 sections: 7 (Anonymization), 8 (Time Series), 9 (Indicators), 2.5 (K-anonymity), 2.7 (Geography)
