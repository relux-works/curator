# TASK-260720-4bd0it review cycle 2

Verdict: changes requested; route to to-dev.

## Required rework

1. Strict v2 locale validation is incomplete. internal/marker/marker.go:279-285 accepts any non-empty string up to 64 runes, so a schema-2 marker with locale "!" is accepted. The authoritative install-marker-v2 schema references common.locale, whose pattern is ^[A-Za-z0-9](?:[A-Za-z0-9-]{0,62}[A-Za-z0-9])?$. The project already exposes identifiers.ValidLocale. Make non-null marker locale validation match the schema and add a focused rejection test.

2. Historical v1 compatibility and v2 string-set compatibility are narrowed at internal/marker/marker.go:335-342. validStringSet rejects the empty string, but common.stringSet permits any JSON string and requires only array type and uniqueness; both marker-v1 and marker-v2 use it for requirers. The pre-task v1 reader accepted an empty requirer. Preserve non-nil, uniqueness, and schema-2 sorting without adding a non-empty item constraint, and add regression coverage for a valid v1 marker plus the v2 reader/writer behavior.

These are task-owned strict-reader defects, not external blockers. No product code was modified during review.

## Passing independent evidence

- Authoritative rc.4 marker race suite: pass, 81.5% coverage.
- Legacy TestGoldenMarkerObject: pass.
- Scoped golangci-lint through go run: 0 issues.
- make check: pass.
- go test -race ./... -count=1: pass.
- Native build and Linux/Windows compile graphs: pass.
- git diff --check and focused gofmt: pass.
- Currentness architecture and source/context/cache/receipt/artifact drift behavior match the accepted protocol boundary.
