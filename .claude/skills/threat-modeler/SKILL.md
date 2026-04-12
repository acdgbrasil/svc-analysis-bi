---
name: threat-modeler
description: >
  Threat Modeling expert using STRIDE + DFD methodology with DREAD scoring.
  Evaluates OWASP Top 10 compliance and ASVS levels.
  Use when analyzing system security architecture or performing risk assessment.
user_invocable: true
---

# Threat Modeler -- STRIDE + DREAD

## Methodology

### 1. Data Flow Diagram (DFD)
Map the system with Mermaid diagrams showing:
- **External Entities**: API consumers, IdPs, svc-social-care (event source)
- **Processes**: NATS consumer, anonymizer, materializer, API handlers, export encoders
- **Data Stores**: PostgreSQL (analytical), NATS JetStream (event stream)
- **Data Flows**: NATS events (PII), anonymized data, SQL queries, HTTP responses, export files
- **Trust Boundaries**: NATS/ingestion (PII boundary), ingestion/store (anonymization boundary), store/API (K-anonymity boundary), API/client (data egress)

### 2. STRIDE Per Element

| Threat | Question | Applies To |
|--------|----------|------------|
| **S**poofing | Can identity be faked? | API consumers, NATS publisher |
| **T**ampering | Can data be modified? | NATS events, analytical store, export files |
| **R**epudiation | Can actions be denied? | API access, event processing |
| **I**nfo Disclosure | Can PII leak? | Anonymization bypass, export metadata, logs |
| **D**enial of Service | Can it be overloaded? | Export generation, NATS backlog, API queries |
| **E**levation of Privilege | Can access be escalated? | API key scope, JWT role manipulation |

### 3. DREAD Scoring (1-10 each)
- **D**amage: How bad is it? (PII leakage of rare disease patients = catastrophic)
- **R**eproducibility: How easy to reproduce?
- **E**xploitability: How easy to exploit?
- **A**ffected Users: How many affected? (all patients in the system)
- **D**iscoverability: How easy to discover?

Score = average of all 5 dimensions.

### 4. OWASP Top 10 (2021) Compliance
For each category (A01-A10): Compliant / Partial / Non-Compliant / N/A

### 5. Risk Matrix
High/Medium/Low Likelihood vs High/Medium/Low Impact -> Priority

## analysis-bi Specific Context
- Handles anonymized data derived from PHI/PII under LGPD
- Rare diseases = small populations = high re-identification risk
- K-anonymity K=5 is the defense, but quasi-identifier combinations can be unique
- NATS events contain PII (patientId, actorId) that MUST be stripped before storage
- Export formats (FHIR, Parquet, ODS) have metadata fields that could leak info
- Analytical store should NEVER contain PII -- this is the Iron Frontier
- Edge hardware deployment (private network, but shared K3s)

## Critical Trust Boundaries
1. **NATS -> Anonymizer**: PII enters here, must be stripped completely
2. **Anonymizer -> Store**: Only anonymized, generalized data crosses this boundary
3. **Store -> API**: K-anonymity filter must suppress small groups
4. **API -> Client**: Only K-anonymous aggregated data leaves the system
