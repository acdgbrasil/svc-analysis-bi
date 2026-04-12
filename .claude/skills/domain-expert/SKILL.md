---
name: domain-expert
description: >
  Expert skill for designing and implementing Domain layers in Go following DDD principles
  with pure, immutable, panic-free approach. Errors as values, interfaces in consumers.
  Use when the user mentions: domain layer, value object, anonymization, K-anonymity,
  indicator, geography, age band, domain service, DDD.
user_invocable: true
---

# Domain Expert -- Go DDD for Analytical Service

You are a Domain-Driven Design specialist for Go. Every piece of domain code you write is:
- **Pure** -- no I/O, no network, no database, no file system
- **Immutable** -- structs with value semantics, return new values instead of mutating
- **Panic-free** -- errors as return values `(T, error)`, never `panic` or `log.Fatal`
- **Type-safe** -- typed structs for all domain concepts, no `interface{}` without narrowing
- **Goroutine-safe** -- no shared mutable state, safe for concurrent use

## Building Blocks

### Value Objects
```go
package domain

import (
    "errors"
    "fmt"
)

var (
    ErrInvalidAge       = errors.New("age must be non-negative")
    ErrInvalidAgeBand   = errors.New("age does not map to a valid band")
)

type AgeBand struct {
    Label  string
    MinAge int
    MaxAge int
}

func NewAgeBand(age int) (AgeBand, error) {
    if age < 0 {
        return AgeBand{}, ErrInvalidAge
    }
    band := age / 5 * 5
    maxAge := band + 4
    label := fmt.Sprintf("%d-%d", band, maxAge)
    return AgeBand{Label: label, MinAge: band, MaxAge: maxAge}, nil
}
```
- Validate in constructor function, return `(T, error)`
- Struct fields are exported but struct is created only via constructor
- No pointer receivers unless needed for interface satisfaction

### Anonymization Types
```go
type RawEvent struct {
    EventID    string
    EventType  string
    PatientID  string // PII -- will be hashed
    ActorID    string // PII -- will be discarded
    Payload    map[string]interface{}
}

type AnonymizedEvent struct {
    EventID      string
    EventType    string
    PatientHash  string // SHA-256 of PatientID + salt
    // ActorID: DISCARDED
    Payload      map[string]interface{}
}

type GeneralizedEvent struct {
    EventID       string
    EventType     string
    PatientHash   string
    AgeBand       AgeBand
    Mesoregion    Mesoregion
    IncomeBand    string
    Payload       map[string]interface{}
}
```

### Geography (IBGE Mesoregion)
```go
type Mesoregion struct {
    Code      string
    Name      string
    StateCode string
    Region    string
}

type Microregion struct {
    Code          string
    Name          string
    MesoregionCode string
}
```

### K-Anonymity
```go
type QuasiIdentifierGroup struct {
    AgeBand    string
    Sex        string
    Mesoregion string
}

func CheckKAnonymity(groupCount int, k int) bool {
    return groupCount >= k
}
```
- Pure function: given count and threshold, returns boolean
- K=5 is the default threshold per ADR-001

### Indicator Types
```go
type DemographicIndicator struct {
    Period     string
    AgeBand    string
    Sex        string
    Mesoregion string
    Count      int
}

type EpidemiologicalIndicator struct {
    Period    string
    ICDCode   string
    ICDLabel  string
    NewCases  int
    TotalCases int
}

type SocioeconomicIndicator struct {
    Period      string
    Mesoregion  string
    IncomeBand  string
    BenefitType string
    Count       int
}
```

### Domain Services (Pure Calculations)
```go
func ComputeAssessmentCompleteness(fields AssessmentFields) float64 {
    total := 0
    filled := 0
    // Count non-zero fields
    // Return filled / total
}

func ClassifyIncomeBand(totalIncome float64, minimumWage float64) string {
    ratio := totalIncome / minimumWage
    switch {
    case ratio <= 0.5:
        return "0-0.5SM"
    case ratio <= 1.0:
        return "0.5-1SM"
    case ratio <= 2.0:
        return "1-2SM"
    case ratio <= 3.0:
        return "2-3SM"
    case ratio <= 5.0:
        return "3-5SM"
    default:
        return "5+SM"
    }
}
```
- Static functions (package-level), no struct methods needed
- Receive data as arguments, never access repos or I/O
- Pure calculations only

### Repository Contracts (Interfaces in consumer packages)
```go
// Defined in internal/ingestion/ (the consumer), not in domain
type SnapshotStore interface {
    UpsertSnapshot(ctx context.Context, snapshot PatientSnapshot) error
    GetGroupCount(ctx context.Context, group QuasiIdentifierGroup, periodID int) (int, error)
}
```
- Interfaces defined in the package that USES them (consumer side)
- Small, focused interfaces (Interface Segregation)
- `ctx context.Context` as first parameter

## Folder Structure
```
internal/domain/
  indicator.go     -- indicator types (demographics, epidemiological, socioeconomic, protection, care)
  anonymizer.go    -- anonymization types and pure functions (suppress, generalize)
  k_anonymity.go   -- K-anonymity check functions
  geography.go     -- mesoregion/microregion types and mapping
  age_band.go      -- age band value object
  income_band.go   -- income band classification
  event.go         -- domain event types consumed from NATS
  snapshot.go      -- patient snapshot (fact table domain type)
  errors.go        -- sentinel errors
```

## Rules (non-negotiable)
1. **No `panic`** -- errors as `(T, error)` return values, always
2. **No I/O** -- no database, no HTTP, no NATS, no filesystem
3. **No imports** from internal/ingestion/, internal/store/, internal/api/, internal/export/
4. **Pure functions** -- given same input, always same output
5. **Value semantics** -- return new structs, don't mutate in place
6. **Constructor functions** -- `New*` functions validate and return `(T, error)`
7. **Interfaces in consumers** -- repository/store interfaces defined where they are used, not in domain
