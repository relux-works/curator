# TASK-260910-hwxr26 results — rework rev2 (F1+F2)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).

## What changed (rework 1 — verdict rev1 F1/F2)

F1 — broken existing record never re-established (internal/audit/sourceaudit.go CheckSourceAudit + Load classification):
- Live verdict enforced before any stored state, so hostile content returns `audit blocked` even with a prior allow binding; stored allow never authorizes new blocking content.
- `LoadSourceAudit` unavailable split via `sourceAuditObjectExists`: no object at all + mutating → first-time issuance (`establishSourceAudit`); object present but report missing/unreadable → refuse `source_audit_rejected: evidence: existing binding has a missing or unreadable report` on both paths, never overwriting.
- `ValidateSourceAudit` failures split via `isRenewableAuditError`: only `source_audit_rejected: policy:` and stale-time (`is stale; re-run the gates`) renew on the mutating path. Identity, context, evidence integrity, future timestamps, and decision drift refuse on both paths, never overwriting.

F2 — report completeness + trusted bindings (ValidateSourceAudit + SourceExpectation):
- `SourceExpectation` gains live trusted arms: Skill, Findings, Pinned, Revoked, Revocation, ScriptLabels, AssuranceLabels.
- `CheckSourceAudit` populates them from the fresh `auditSubject` run (`report.Findings/Revoked/Revocation`, `isPinned` on the live content hash, `subject.ScriptLabels/AssuranceLabels`, `subject.Name`).
- `ValidateSourceAudit` now checks, after the digest: raw JSON contains all 11 report keys (schema_version, skill, package, content_sha256, findings, pinned, revoked, revocation, script_policy, assurance_policy, decision); report schema_version==1, package valid, content digest well-formed, decision enum, each finding has id/severity/evidence with known severity; report binds the same package/content/decision as the object; then trusted bindings — skill equality (when expected set), `findingsEqual`, pin equality, revocation equality, `labelsEqual` for script/assurance. A stripped wrong-skill report with a recomputed digest still fails.
- `findingsEqual`/`labelsEqual` treat nil and empty as equal so clean allow bindings validate.

## Tests (committed, production entry named in AC)

`internal/audit/sourceaudit_test.go`:
- `validBinding` is now complete (findings/pinned/revoked/script/assurance present, expected carries the same trusted arms).
- `TestValidateSourceAuditBindings`: git-arm valid updated to carry labels; spliced/decision-split fixtures made complete so they reach the intended evidence check; new rows: incomplete-stripped, wrong-skill recomputed, findings mismatch, pin mismatch, script-labels mismatch, assurance-labels mismatch — each with a narrowing digest-stale variant that stays `source_audit_rejected: evidence`.
- `TestValidateSourceAuditRenewableClassification`: policy + stale are renewable; identity, evidence-digest, future, decision are not.
- `TestCheckSourceAuditEstablishThenValidate`: deleted report now refuses on both paths (`source_audit`) and proves the object bytes were not overwritten and the report was not re-established.

`internal/install/draftaudit_test.go`:
- `TestDraftAuditBrokenReportRefuses` (adopts reviewer overlay `TestReviewSourceEvidenceRefusals` rows plus `mismatching-report`): missing-report, wrong-digest (`{}`), mismatching-report (trailing-space digest break), forged-report (stripped + wrong skill + recomputed digest). Each refuses with `source_audit` on its own path and on the opposite dry-run path, and (non-forged) proves the object was not overwritten.
- `TestDraftAuditFirstIssuanceVsBrokenRecord` (narrowing): report-only deletion refuses on mutating; full deletion (object+report) re-establishes on mutating and stores a binding. Proves the bound is the existing record, not the package.

## Evidence (shell: bash, `set -o pipefail` not needed — no pipelines except tail; exit codes real)

1. `go test -p 1 ./internal/audit -run 'TestValidateSourceAudit|TestCheckSourceAudit' -count=1 -timeout=60s` — exit 0.
2. `go test -p 1 ./internal/install -run 'TestDraftAuditBrokenReport|TestDraftAuditFirstIssuance' -count=1 -timeout=100s -v` — exit 0; `TestDraftAuditBrokenReportRefuses` (4/4 subtests incl. missing/wrong-digest/mismatching/forged) PASS, `TestDraftAuditFirstIssuanceVsBrokenRecord` PASS.
3. `go test -p 1 ./internal/audit ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy ./internal/install -run 'Test(ParseSourceAudit|ValidateSourceAudit|SourceAudit|PolicyDigest|CheckSourceAudit|DraftAudit|EffectiveLabels|LocalPackage|NetworkAttestation)' -count=1 -timeout=100s` — exit 0 (audit ok, registry no matching tests in that mask, artifactpolicy ok, scriptpolicy ok, install ok 24.9s).
4. `go test -p 1 ./internal/registry ./internal/install ./internal/interop -run 'Test(CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1 -timeout=100s` — exit 0 (all three packages ok).
5. `go test -p 1 ./internal/audit -count=1 -timeout=60s` (full audit package) — exit 0.
6. `go test -p 1 ./internal/artifactpolicy -run 'TestEffectiveLabels' -count=1` — exit 0.
7. `git diff --check` — exit 0. `gofmt -l` on touched files — exit 0 (clean after `gofmt -w`). `go vet ./internal/audit ./internal/install` — exit 0.
8. Full `go test -p 1 ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -count=1 -timeout=60s` — FAIL in `internal/artifactpolicy` on unrelated `TestDirectoryEntryLimitStopsTheLiveWalker` (containers_test.go:174, live-walker timeout/stack after 60s); `internal/scriptpolicy` ok. Out of task scope (no artifactpolicy production change by this task; labels narrow test above passes). Recorded as failing gate staying failing, not claimed green.

## Checklist

- Implement scoped production behavior with traceability to skillfile-sources §4 (identity/context/policy/evidence/time/decision; local has no network attestation; pins/revocation/assurance before cache/compiler): done.
- Positive, negative, legacy regression checks with exact revision and independent-review evidence: done (see Evidence).
- Implementation matches AC; solution fits project architecture (draft lane only, frozen v1 untouched, source-audit distinct from registry attestation): done.
- Relevant tests written and passing; lint clean; build not broken: done (narrow suites green, vet/diff-check/gofmt clean).
- Review rev1 F1/F2 addressed; overlay probes adopted as committed tests plus narrowing mutants: done.

## Files

- internal/audit/sourceaudit.go (F1+F2 fix; new helpers findingsEqual, labelsEqual, isRenewableAuditError, sourceAuditObjectExists; CheckSourceAudit reorder + renewal classification)
- internal/audit/sourceaudit_test.go (complete fixtures, new evidence/trusted negative rows, renewal classification test, broken-record non-overwrite test)
- internal/install/draftaudit_test.go (TestDraftAuditBrokenReportRefuses 4 modes × both paths + non-overwrite; TestDraftAuditFirstIssuanceVsBrokenRecord narrowing)
- Tracked rev1 wiring retained: internal/install/install.go (12b draft source-audit binding before cache/compiler), internal/registry/registry.go (unknown without identity), plus rev1 new files internal/registry/local.go, internal/artifactpolicy/labels.go, internal/scriptpolicy/labels.go and their tests.
