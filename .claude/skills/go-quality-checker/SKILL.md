---
name: go-quality-checker
description: >
  Go code quality checker. Audits error handling, goroutine safety, naming conventions,
  idiomatic patterns, race conditions, resource management, and code organization.
  Use when checking Go code quality, concurrency safety, or API design guidelines.
user_invocable: true
---

# Go Quality Checker

## Error Handling
- Every error return is checked: `if err != nil { ... }` -- no exceptions
- No swallowed errors (`_ = err` or empty error blocks)
- Error wrapping with context: `fmt.Errorf("operation: %w", err)`
- Sentinel errors for expected conditions: `var ErrNotFound = errors.New(...)`
- Custom error types implement `Error() string` and optionally `Unwrap() error`
- No `panic` in production code (allowed in tests, init for truly unrecoverable)
- No `log.Fatal` or `os.Exit` outside `main()`
- Use `errors.Is` and `errors.As` for error matching (not string comparison)

## Goroutine Safety
- All goroutines respect `context.Context` for cancellation
- No goroutine leaks: every `go func()` has a termination path
- Shared mutable state protected by `sync.Mutex` or channels
- `defer mu.Unlock()` immediately after `mu.Lock()`
- `go test -race` must pass clean
- WaitGroup: `Add` before goroutine launch, not inside
- Channel direction in function signatures: `chan<-` (send), `<-chan` (receive)
- `select` with `case <-ctx.Done()` for cancellation

## Naming Conventions (Effective Go)
- Exported: PascalCase (`NewAgeBand`, `PatientSnapshot`, `Encoder`)
- Unexported: camelCase (`hashPatient`, `validatePeriod`, `parseQuery`)
- Interfaces: single-method use -er suffix (`Reader`, `Writer`, `Encoder`, `Resolver`)
- Multi-method interfaces: descriptive noun (`EventStore`, `IndicatorQuerier`)
- Packages: lowercase, singular, short (`domain`, `store`, `export`, `api`)
- Acronyms: all caps (`ID`, `HTTP`, `URL`, `SQL`, `JWT`, `NATS`, `FHIR`, `CSV`, `XML`)
- Boolean: `Is*`, `Has*`, `Can*` (`IsValid`, `HasChildren`)
- Constructors: `New*` (`NewAgeBand`, `NewConsumer`, `NewPostgresStore`)
- No stuttering: `export.Encoder` not `export.ExportEncoder`
- Getters: no `Get` prefix (`band.Label` not `band.GetLabel()`)
- Error variables: `Err*` prefix (`ErrNotFound`, `ErrInvalidAge`)

## Resource Management
- `defer` for cleanup immediately after acquisition
- HTTP response bodies: `defer resp.Body.Close()`
- Database rows: `defer rows.Close()`
- Context cancel functions: `defer cancel()`
- File handles: `defer f.Close()`
- Mutex unlock: `defer mu.Unlock()`
- `defer` not used in loops (use explicit cleanup or extract to function)

## Performance
- Preallocate slices: `make([]T, 0, knownCap)` when size is known or estimable
- `strings.Builder` for string concatenation in loops
- No blocking operations without context timeout
- `sync.Pool` for frequently allocated/freed objects (only if profiling shows benefit)
- Avoid unnecessary interface indirection in hot paths

## Code Organization
- Functions < 50 lines ideally, < 80 max
- `internal/` for encapsulation (Go visibility rules)
- No circular imports
- Test files co-located: `*_test.go` in same package
- Table-driven tests as default pattern
- Test helpers use `t.Helper()` for clean error reporting
- Interfaces defined in consumer package, not provider

## Idioms
- Accept interfaces, return structs
- Make the zero value useful
- Don't communicate by sharing memory; share memory by communicating
- A little copying is better than a little dependency
- `io.Reader` / `io.Writer` for streaming data
- `context.Context` as first parameter, always
- Return early for errors (guard clauses)
- Avoid `else` after `return`

## Reference
- handbook/principles/ (architecture decisions)
- Effective Go: https://go.dev/doc/effective_go
- Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
