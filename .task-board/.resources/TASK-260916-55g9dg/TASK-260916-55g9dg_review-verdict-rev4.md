# TASK-260916-55g9dg review verdict — CR rev4 — ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate: base bd3c0f43, tree 1664d96f (recomputed from worktree via temp index: identical). 8 paths.
Spec root: curator-spec clone at SPEC_PIN 23435129 (v1.0.0-rc.13) in $TMPDIR.

## Rev2 findings closure
- F1 knob pruning workaround: gone; `TestManagerConfigV2Vectors` compares full canonical JSON (environments_conformance_test.go:277-305), extra-key regression test `TestManagerEffectiveJSONComparisonRejectsExtraKnobs`.
- F2 null waivers: rejected at `Load` with "must be a list" (environments_test.go:635-653).
- F3 admission subset: driven by `TestSystemModuleAdmissionVectors` (contextmaterialize) at rc.13; killed by M2 below.

## Rev4 delta
- environments.go:182-190 sorted knob iteration → deterministic first unsupported field (test TestUnsupportedEnvironmentFieldFailureIsDeterministic); gap-ledger rows reattributed to the true first blocker (environments.permissions → STORY-260922-1cenbr; system-config signer lock → ioemse).
- status.go:570-662 unreadable transitive manifest no longer = "no drops": nil dropped set + DiagManifestInvalid diagnostic, status non-current (§8.4 unreadable≠absence). Policy error propagates instead of silently defaulting to drop.
- cmd/curator TestEnvStatusReportsDroppedSystemModuleThroughCLI: production CLI entry prints `transitive_system_modules=drop` and `dropped system module sysleaf 90-system.md`.

## Ledger
69 rows (70 lines incl. header), 0 rows owned by STORY-260916-2d9coh / this task. Histogram: 1cenbr 45, ioemse 10, BUG-2afgyq 5, 2qmrb8 3, 6bo7ej 2, 1i1gfo 2, 1v7pvn 2. E2 schema cases whose first blocker is external remain honestly attributed (permissions / source_signers); no E2∩E4 rows (provider_directories accepted).

## Independent validation (zsh, pipefail, rc.13 root)
- go build ./... ; go vet (config, envprofile, cmd/curator) ; gofmt -l internal cmd → rc=0, empty.
- go test -count=1 ./internal/config/... ./internal/contextmaterialize/... ./internal/interop/environments/... → ok, rc=0.
- go test ./cmd/curator → ok (480.9s).
- go test ./internal/envprofile/... full → host 600s timeout under load (host limit, not a failure signal); scoped `-run 'Status|Admission|SystemModule|Policy'` → ok 104s, rc=0. Hosted gate on this revision green (arbiter).

## Mutants (disposable copy)
- M1 waiver ignores scope (`Waived` = any waiver present): KILLED (TestSystemPromptWaiverAdmits).
- M2 transitive admitted (DirectSet admits every member): KILLED (TestDirectSet, TestSystemPromptDrop*/Error*, TestSystemModuleAdmissionVectors, …).
- M3 unreadable manifest → continue (absence): KILLED (TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops).
No survivors found.

## Bounds / notes
- Windows/Linux not rerun locally; hosted lanes are the evidence.
- Some E2 manager/system schema cases' E2 diagnostic is not reached at rc.13 until 1cenbr (permissions) / ioemse (signers) land; ledger records it.
- CHANGELOG untouched per 2026-09-24 policy.
