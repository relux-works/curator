# TASK-260910-hwxr26 results rev8 — duplicate-key gate (rework 7)

Base: `9b185d2503d7bfb4fe66a33ad591a50abe9f8c90` (Story branch `task-board/story/STORY-260910-20sx61`).
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-20sx61/worktree` (left uncommitted for handoff snapshot).
Scope: `internal/audit`, `registry`, `artifactpolicy`, `scriptpolicy` — only audit files touched this rev; unrelated dirty `install.go`/`registry.go` modifications left untouched.

## Fix

`internal/audit/sourceaudit.go`:
- `ParseSourceAudit` (+7 lines): `protocoljson.Validate(payload)` on the original object bytes before any map/struct decoding; refuses `source_audit_rejected: malformed` on duplicates/trailing data. Covers top level and nested package/commit recursively.
- `validateEvidenceReportRaw` (+6 lines): same shared validator on the original report bytes before lossy decoding; refuses `report is malformed`, wrapped as `source_audit_rejected: evidence` by callers. No new decoder written; vendored schemas, closed shapes, typed renewal ordering unchanged.

`internal/audit/sourceaudit_test.go` (+148 lines): `TestParseSourceAuditDuplicateKeys` — object duplicates (top schema_version, hidden foreign-arm package, nested kind/snapshot, git commit hex/object_format, hidden outer+inner commit) plus report duplicates per arm (top revoked, nested kind) and git commit-hex rows with recomputed digests; valid local/git/report controls.

`internal/install/draftaudit_test.go` (+260 lines): `TestDraftAuditDuplicateKeysRefuse` at production entry `install.Project` — one `strictLocalProject` fixture restored per row; 12 rows (object/report × top-duplicate, nested kind, nested snapshot, hidden foreign-arm commit:null, hidden commit-inner-duplicate) assert `source_audit` refusal on mutating AND read-only paths with no overwrite of either record; intact controls pass both paths after restore.

## Evidence (real exit codes, zsh, `set -o pipefail`)

- `go test -p 1 ./internal/audit -run '^TestParseSourceAuditDuplicateKeys$' -count=1 -timeout=60s -v` → exit 0, 0.626s, 18/18 subtests pass.
- `go test -p 1 ./internal/install -run '^TestDraftAuditDuplicateKeysRefuse$' -count=1 -timeout=100s -v` → exit 0, 12.832s, 12/12 rows pass (each refuses both paths, no overwrite).
- `go test -p 1 ./internal/audit -run 'Test(Source|ParseSource|ValidateSource|CheckSource|PolicyDigest)' -count=1 -timeout=60s` → exit 0, 1.949s.
- `go test -p 1 ./internal/install -run '^TestDraftAudit(ClosedPackageShapeRefuses|MalformedShapeRefuses|RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable)$' -count=1 -timeout=110s` → exit 0, 85.404s (prior shape/renewal controls preserved).
- `go test -p 1 ./internal/install ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'TestLegacyInstallUntouchedWhenDraftOff|TestCheckLocalPackage|TestEffectiveLabels|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditPinAdmits|TestDraftAuditRevocationBlocks' -count=1 -timeout=100s` → exit 0 all four packages (install 8.307s, registry 0.703s, artifactpolicy 0.475s, scriptpolicy 0.596s).
- `go vet ./internal/audit ./internal/install` → exit 0. `gofmt -l` on touched files → clean. `git diff --check` → exit 0.
- Narrowing mutant (report gate disabled, `_ = evidence`): `go test -p 1 ./internal/install -run '^TestDraftAuditDuplicateKeysRefuse/report-top-duplicate-revoked$'` → exit 1, `report-top-duplicate-revoked was NOT refused on the mutating path` (audit gate passed, fell through to `install marker is invalid for schema 2`). Restored immediately; rerun same row → exit 0. Proves the report-side Validate is load-bearing; object-only validation does not suffice. Unit table additionally kills nested-only omissions (commit-hex rows fail without recursive Validate).
- Reviewer rev7 probe rows covered 1:1 by production rows `object-top-duplicate-schema_version`, `report-top-duplicate-revoked`, `object-hidden-foreign-arm`, `report-hidden-foreign-arm` (same prefix forgeries + digest recompute for reports).

## Checklist

- [x] Implement the scoped production behavior with traceability to the accepted draft contracts.
- [x] Run task-specific positive, negative and legacy regression checks; record exact revision and evidence for independent review.
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant (finding: Go last-wins decoding hid foreign-arm members; shared validator reused, no new decoder)

## Handoff

Ready for review (non-final leaf; rev8). Full suite not run locally per wave note (host stalls on fresh binaries; remote gate owns landing).
