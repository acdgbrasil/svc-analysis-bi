---
name: quick-path
description: >
  Fast-path for trivial tasks that don't need the full pipeline.
  Use for bugfixes, small refactors, single-file changes, documentation, or config tweaks.
  Bypasses architect/test-writer/review-loops.
  Trigger for: "quick fix", "hotfix", "simple change", "just add", "small tweak".
user_invocable: true
---

# Quick Path -- Fast Track for Trivial Tasks

## Decision Rule

Use quick-path if ALL of these are true:
- Touches <= 3 files
- Stays within 1 layer (domain OR ingestion OR store OR api OR export)
- Does NOT touch auth/middleware/security
- Does NOT change public API contracts or event schemas
- Is a bugfix, small refactor, config change, or doc update

If ANY of these is true, use full pipeline instead:
- Touches 3+ files across 2+ layers
- Touches auth, JWT, middleware, or anonymization pipeline
- Changes indicator types or event schemas
- Adds new endpoint or new export format
- Needs design discussion

## Execution Flow

1. **Classify** -- is this really trivial? Check the decision rule above.
2. **Discuss** (optional, `--discuss`) -- quick alignment with user
3. **Implement** -- make the change
4. **Test** (optional, `--test`) -- run `go test -run TestName ./internal/...`
5. **Verify** -- `go vet ./... && go build ./...` must pass
6. **Review** (optional, `--review`) -- quick code-reviewer pass
7. **Report** -- summary of what changed

## Flags
- `--discuss` -- ask clarifying questions before implementing
- `--test` -- write/run tests for the change
- `--review` -- run code-reviewer agent on the changed files
- `--security` -- run secure-code-reviewer on the changed files

## Commit
```
fix(<scope>): <description>
```
or
```
chore(<scope>): <description>
```
