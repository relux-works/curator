# BUG-260920-2eg8nv results — marker reader admits external record without `substituted`

## Outcome

Fixed. `marker.Read` now refuses a v5 external `go-repository-v1` build record
without the `substituted` field, via a closed raw-shape check of the external
arm. The crossconformance schema row
`invalid-external-missing-substituted` flipped from bound to driven-pass
(111 driven / 4 bounds / 115 total, was 110/5). The `valid.json` quirk was
isolated to a single cause and is a corpus bug, not a reader bug: the fixture
carries top-level `build_source` with empty `builds`, violating the normative
"required exactly for active local go-v1 commands and absent otherwise" rule.
It stays a documented bound; the curator-spec discrepancy text is attached as
`BUG-260920-2eg8nv_curator-spec-discrepancy.md`. Narrowing mutant killed at
both the unit and schema levels.

## Production change

`internal/marker/marker.go` (only production file touched):

- New closed-shape tables for the v5 external arm:
  `v5ExternalBuildTypes` (18 members to JSON types), `v5ExternalBuildRequired`
  (16 required members incl. `substituted`), nested shapes for declared /
  effective identity, build-source, and both substitution variants (locked
  commit reuses `v5CommitShape`).
- New `validV5ExternalBuildShape`: required members present and typed, no
  foreign members even as null, `substituted`/`substitution` conditional
  (`true` requires / `false` forbids `substitution`), nested objects closed.
  Presence refusal for `substituted` lives in exactly one place (the required
  loop) so a mutant dropping it admits the record instead of surviving behind
  a redundant refusal.
- `rawJSONType` extended with `boolean` and `number` (package checks only use
  `string`/`object`, so they are unaffected).
- `validBuildState` parses raw `builds` for v5 and enforces the raw-shape
  check on records whose decoded driver is `go-repository-v1`, before the
  existing decoded `validV5Build`. Local-arm and v1–v4 behavior unchanged.

`internal/crossconformance/draftsources_schema_test.go`:

- Removed the `invalid-external-missing-substituted` bound and its special
  case; the row now drives through the default `(got != nil) != entry.Valid`
  assertion (entry `valid:false`, reader must refuse).
- Kept the `valid.json` bound with an updated comment pointing at the
  discrepancy text. Count assertion stays honest (`driven+bound == total`).

## Tests

New `TestMarkerV5ExternalBuildClosedShape` in
`internal/marker/marker_v5_builds_test.go`: 4 positive controls
(unsubstituted, substituted local-path, substituted network-git, declared-tag)
and 31 negative rows (missing / null / mistyped `substituted`, null
`substitution`, both conditional violations, foreign members incl. null,
nulls in required members, missing required members, nested extra members
and nulls, local-path-with-ref, network-git missing/null ref, unknown and
missing substitution type). Every row drives `marker.Read` via a
`rewriteV5Build` raw splice, the production entry.

## valid.json quirk — exact reason

Fixture `schema-cases/install-marker-v5/valid.json` (index `valid:true`) has
`"builds": {}` plus a top-level `build_source`. `validBuildState`
(`internal/marker/marker.go`) returns `!sourcePresent && m.BuildSource == nil`
for empty builds, so the document is refused. This matches the normative rule
in `protocol/skillfile-sources.md`: "Top-level `build_source` remains required
exactly for active local `go-v1` commands and absent otherwise". With zero
builds there are no active local commands, so `build_source` must be absent;
the fixture violates its own contract. JSON Schema cannot express the
cross-field rule, hence schema-valid but normatively inconsistent. The reader
is right; no workaround was added. Probe evidence below; discrepancy text
attached separately for the orchestrator to raise on curator-spec.

## Evidence (narrow, real exit codes)

Shell is `sh` via the task runner; `set -o pipefail` everywhere. Note: the
host stalls ~80s executing brand-new binaries and `go test`/`go run` wrappers
were killed under that stall, so tests were compiled once with
`go test -c -o /tmp/....test` (fast) and the binaries run directly.

- `go vet ./internal/marker/` → exit 0
- `gofmt -l internal/marker/ internal/crossconformance/` → empty, exit 0
- `golangci-lint run internal/marker/...` → `0 issues`, exit 0
- `golangci-lint run internal/crossconformance/...` → `0 issues`, exit 0
- marker full package: `/tmp/marker.test` from `internal/marker` → `PASS`,
  exit 0 (includes new `TestMarkerV5ExternalBuildClosedShape`, 35/35
  subtests PASS)
- schema suite: `/tmp/cc.test -test.run TestDraftSourcesSchemaCases` from
  `internal/crossconformance` → `PASS`, exit 0;
  `schema cases: 111 driven, 4 bounds, 115 total`;
  `invalid-external-missing-substituted.json` PASS (driven);
  `install-marker-v5/valid.json` PASS (bound, refusal still correct)
- pin + counts:
  `-test.run 'TestDraftSourcesSchemaCases|TestDraftSourcesPin|TestDraftSourcesCorpusCounts'`
  → `PASS`, exit 0
- valid.json probe (temporary in-module main, built and removed; only the
  three intended files remain modified):
  `valid.json as-is: readable=false`,
  `valid.json minus top-level build_source: readable=true`, exit 0 —
  isolates the refusal to the single `build_source` member.
- Narrowing mutant (removed `"substituted"` from `v5ExternalBuildRequired`,
  rebuilt, then restored):
  - unit: `TestMarkerV5ExternalBuildClosedShape/missing-substituted` → FAIL
    (`Read = &{...}`, `want readable=false`), exit 1 — killed.
  - schema: full `TestDraftSourcesSchemaCases` →
    `valid=false: marker.Read nil=false` on the
    `invalid-external-missing-substituted.json` row, exit 1 — killed.
  - An earlier variant with a redundant strict decode survived (documented
    here, not hidden); the shipped code makes the required loop the single
    refusal point so the one-line mutant is killed.

## Scope discipline

- Only `internal/marker/marker.go`,
  `internal/marker/marker_v5_builds_test.go`, and
  `internal/crossconformance/draftsources_schema_test.go` modified
  (`git status --short` confirms; probe directory removed).
- Sibling checkpointed leaves (3ukdk4, 2wyzde, 3v7x6j) untouched; legacy v1,
  v2–v4, and the v5 local arm unchanged (raw-shape gate applies to decoded
  `go-repository-v1` records on v5 only).
- Full local suite not run per wave note (remote gate runs at handoff); the
  one full-crossconformance attempt was terminated as out of scope after the
  narrow gates passed.

## Checklist

- [x] marker.Read refuses external record without substituted; schema row
      invalid-external-missing-substituted flips bound → driven-pass
- [x] valid.json quirk investigated: reader correct, exact reason above, bound
      kept, discrepancy text attached; narrowing mutant killed
- [x] narrow evidence with exit codes here; handoff via task-board handoff
- [x] code per task description and AC
- [x] tests for new/changed behavior written and passing
- [x] lint clean
- [x] relevant build/validation commands run; build not broken
- [x] outcome artifacts attached with task-scoped names
- [x] findings recorded here (no LOGBOOK.md edit per campaign rules)
