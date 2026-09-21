# Review verdict — TASK-260920-1sbj7o revision 2 (CR-TASK-260920-1sbj7o-2)

**Verdict: CHANGES_REQUESTED → `to-dev`.** Reviewer: claude-opus-5 (RUN-260920-fe1ba2), 2026-09-20.
Evidence bundle: `TASK-260920-1sbj7o_review-rev2-evidence.tar.gz` (drivers, logs, probe test, mutant outputs).

## Tree and gate identity (verified)

- Story worktree temp-index `write-tree` over `internal/ cmd/` = `5c8dd687254f9888795a416a482ee547d5ade806` = CR candidate tree; `HEAD` = base `6a6e2a14`; `git status --short` = exactly the 6 changed paths.
- Hosted gate run 35514565231 (rev2 validation log): `headSha a39a5c40…` → `git rev-parse a39a5c40^{tree}` = `5c8dd687…`, parent = base. Conclusion `success` on every lane (rose-air/candidate suite skipped by design).
- Review ran on disposable clones checked out at the gate commit (`/tmp/1sbj7o-rev2-review/{cand,probe}`, both `HEAD^{tree}` = candidate) and at the base (`base`); the Story worktree was never written.

## What the candidate does (read)

- `internal/buildrepo/pipeline.go:211-217`: `RunPipeline` returns an `Acquire` error whose `ErrorCode` is `CodeIdentityInvalid` unchanged, before the (production-unused) offline-snapshot substitution; every other acquire failure still collapses to `build_repository_source_unavailable`.
- `cmd/curator/draft_diagnostics.go:99-101`: new `draftRemediations` row keyed on the bare class `build_repository_identity_invalid` with the port/alias remediation text.
- Tests: pipeline unit row; CLI `install --dry-run` row (port-planned policy, both fixture identities); the two crossconformance rows assert the exact class through `install.Project`; `TestWithDraftRemediationTable` gains the positive row and **drops** the stbg4d legacy pin `build_repository_identity_invalid: wrong identity` (replaced by a `source_unavailable` pin).

## Independent reruns (bash driver, `set -o pipefail`, real exit codes; clone `cand`, load 6–10)

| check | rc |
|---|---|
| `go build ./...` | 0 |
| `go vet` buildrepo/install/cmd/curator/crossconformance | 0 |
| `gofmt -l` on the 6 files (empty) | 0 |
| `go test -count=1 -p 1 ./internal/buildrepo/` (full package, 112 s) | 0 |
| cmd/curator `-run` TestWithDraftRemediationTable\|TestDraftAttemptTransportTable\|TestDraftAttemptClauseShape\|TestEndpointExhaustionCarriesRemediation\|TestDraftTransportIdentityInvalidSurfacesThroughCLI\|TestDraftRemediationThroughCLI\|TestDraftTransportResolvedMatrix\|TestDraftTransportProviderAdmission\|TestDraftExternalStatusFailsClosedWithoutSource (8 tests, 64 s) | 0 |
| internal/install `-run 'TestRevision2\|TestDraftTransport\|TestUserConfigIgnored\|TestResolvedLane'` — note: matches only 5 tests (`TestDraftTransportEnabled/PlanConversion/Auth/PlanV2Conversion`, `TestUserConfigIgnoredByResolvedLane`); the producer's `TestRevision2*`/`TestResolvedLane*` masks match nothing in this package | 0 |
| crossconformance `-run 'TestDraftSourcesSemanticCases$/^v2-external-build-(port\|alias)-refused$'` | subtests 2/2 PASS; parent rc=1 by design (`executed(2) != total(94)`), ratio line `2 driven, 0 known-gap, 0 bound, 0 skipped` |

Hosted per-lane ledgers (`test-evidence-<os>`, run 35514565231): macOS `93 driven / 0 known-gap / 1 bound / 0 skipped`, ubuntu `92/0/1/1`, windows `42/0/1/51`; both refusal rows and `TestDraftTransportIdentityInvalidSurfacesThroughCLI` pass on macOS+ubuntu and skip on Windows ("test transport wrapper is POSIX-only"); `TestPipelinePreservesIdentityInvalidFromAcquisition` and `TestWithDraftRemediationTable` pass on Windows. Windows evidence for the install-level class is therefore the pure unit row only (production check is platform-neutral).

## Gate attack (clone `probe`, candidate committed; restore via `git checkout --`; tree `5c8dd687` before and after)

| mutant | unit (`TestPipelinePreservesIdentityInvalidFromAcquisition`) | `TestWithDraftRemediationTable` | xconf rows (subtests pass) | CLI row |
|---|---|---|---|---|
| M1 re-collapse (delete the block; producer's mutant) | rc=1 | 0 | 0/2 | rc=1 |
| M2 narrow to `Operation == OperationInstall` (plan phase is dry-run) | rc=0 (helper bound) | 0 | **0/2 killed** | **rc=1 killed** |
| M3 narrow to `Declared.Tag == ""` | rc=0 | 0 | 2/2 | rc=0 → **survives** (bound: every driven row is an untagged declaration; not blocking) |
| M4 offline snapshot substitution before the identity return | **rc=1 killed** | 0 | 2/2 | rc=0 (production never sets `OfflineSnapshotKey`) |
| M5 remediation row removed | 0 | **rc=1** | 2/2 | **rc=1** |

The re-mask is killed at the production entry (install.Project rows and the CLI row), as the AC requires.

## Findings

### F1 (blocking) — the remediation row misfires on every other `build_repository_identity_invalid` message, including frozen-lane output
`cmd/curator/draft_diagnostics.go:99-101` keys the row on the bare class; `withDraftRemediation` (`:107-122`) matches by `strings.Contains`, so **any** message carrying the frozen v1 class gets the port/alias text. The class has ~30 producers besides `parseTransportPlan` (`internal/buildrepo/admission.go:189-217, 285-313, 386`, `credentials.go`, `sshbroker.go`, `httpsbroker.go`, `pipeline.go:424-440`), and the candidate's own `RunPipeline` change now lets the acquire-phase ones reach the CLI — on the legacy lane too.

Reproduction through the production entry (`cli.run`, probe `TestReviewProbeLegacyLaneIdentityInvalid`, candidate tree; `CURATOR_DRAFT_TRANSPORT_RESOLUTION` unset, no `source-policy.json`, `GIT_SSH` unset, SSH identity selected — i.e. the frozen lane with no manager SSH wrapper):

```
$ curator install app --dry-run      (and: curator status app — identical line)
error: skill-a.ssh-cmd: build_repository_identity_invalid: SSH requires the exact manager wrapper; fix the endpoint entry in machine source-policy.json: the strict external-build lane admits no explicit port and no host alias, then retry the explicit attempt
```
Same fixture on the base tree: `error: skill-a.ssh-cmd: build_repository_source_unavailable: exact external source is unavailable` (probe-base.out). The appended guidance is wrong for this message: the frozen lane never reads `source-policy.json`, and no port or alias is involved. `withDraftRemediation` also decorates (helper-level demonstration) `build_repository_identity_invalid: git path must be absolute` (`productionGitTool` zero value when git is absent, `admission.go:189`), `…: trusted Git version probe failed`, `…: invalid immutable lock`, `…: unsubstituted declared/effective source mismatch` and the stbg4d pin text `…: wrong identity` the same way (probe `TestReviewProbeRemediationKeyedOnBareClass`).

This violates the accepted stbg4d decision the file header states ("output without a stable class — every frozen v1 message — is unchanged") — the candidate removed that decision's pin (`draft_diagnostics_test.go:90`, `build_repository_identity_invalid: wrong identity` → replaced) instead of scoping the row — and the brief's "keep it closed and sanitized / legacy v1 unchanged".

Required: key the row on the refusal's own static diagnostic so only the §7 refusal is guided (e.g. `class: "build_repository_identity_invalid: transport plan endpoint"` — `strings.Contains` makes this a one-line change with no mechanism change; the install-level message is `skill.cmd: build_repository_identity_invalid: transport plan endpoint N carries an explicit port or host alias outside the strict external-build lane grammar`, unaltered by `RedactDiagnostic`); restore the legacy pin with a real lane message (e.g. `build_repository_identity_invalid: SSH requires the exact manager wrapper`) alongside the `source_unavailable` pin; keep the positive row and the CLI row.

### F2 (must be scoped or declared) — undeclared legacy-lane class change
The preservation in `RunPipeline` is class-based, not lane-based: on the frozen lane (`acquireDraftNetwork` → `buildrepo.AcquireNetwork`, `internal/install/drafttransport.go:120-121,131-132`) every `admitNetworkRequest`/`ValidateGitTool` identity refusal that was reported as `build_repository_source_unavailable: exact external source is unavailable` now surfaces as `build_repository_identity_invalid: <lane text>` at `install.Project`/CLI (before/after shown under F1; production-reachable, demonstrated through `cli.run` with a missing `GIT_SSH` wrapper for an SSH build repository). results.md's "Legacy v1: no frozen schema/lane file touched" is a file-level claim; the behaviour changed. No frozen vector or committed test pins the old collapse (gate green), and the manager profile's normative table maps `build_repository_identity_invalid` → state `unsupported` and `build_repository_source_unavailable` → `blocked` (profiles/manager.md:2246-2247), so surfacing the lane's own class is arguably the more conformant reading. Recommendation: keep the class-based fix but **declare** it (results.md + a committed legacy-lane row: switch off, no policy, `GIT_SSH` unset → `build_repository_identity_invalid: SSH requires the exact manager wrapper`, no `source_unavailable`, **no remediation suffix**). If the orchestrator wants the frozen lane byte-identical instead, scope the preservation to the resolved lane; either way the choice must be recorded, not silent. No functional change hides behind it: `OfflineSnapshotKey` is set by no production caller (`grep` — only `pipeline_test.go`), so the skipped offline substitution is test-only.

### F3 — docs and docs-pin not extended
Every other remediation row has a `### <class>` section in `docs/troubleshooting.md` and an entry in `TestDraftDocsPinExamples` (`draft_diagnostics_test.go:1160-1199`, "every stable class remedy"); the new row has neither, and the section intro ("These stable classes appear only on the draft Skillfile lane") is false for this frozen class. Add a section describing the §7 port/alias refusal remedy (scoped to the resolved lane) and the pin entry.

### F4 (coverage gap, not blocking) — no committed `status` row
The AC names `curator install/status`; the candidate commits only `install --dry-run`. The path is shared (`printFailures`, `main.go:658`), and my probe `TestReviewProbeStatusPortPlanned` confirms `status app` on the port-planned policy prints `error: skill-a.https-cmd: build_repository_identity_invalid: transport plan endpoint 1 carries … ; fix the endpoint entry in machine source-policy.json …`, no `source_unavailable`, zero fetches. One `capture(t, configPath, "status", "app")` assertion in the CLI row closes it.

### Residuals (record only)
- R1: `internal/install/drafttransport.go:274-281` builds `build_repository_identity_invalid: SSH transport resolution requires …` with `fmt.Errorf`, not `admissionError`; `ErrorCode` does not see it, so those still collapse to `source_unavailable` at install level. Pre-existing, outside this AC.
- R2: `pipeline.go:208-210` returns `unverified-offline` for `OperationSyntax` before the identity check; no production caller uses that operation.
- R3: M3 bound — a tag-conditioned re-mask survives; every driven row declares an untagged lock.
- R4: LOGBOOK.md not edited (campaign rules forbid worktree/control-root logbook writes); this resource is the record.

## Checklist mapping
- Implementation matches AC — **no** (F1: the remediation row is not closed to the class's refusal it describes; F2 undeclared).
- Solution fits project architecture — partially (root-cause placement in `RunPipeline` is right; the presentation-layer keying breaks the stbg4d contract).
- Tests green — yes (narrow reruns above; hosted gate green on tree `5c8dd687`).
- Verdict evidence added and routed — `set_status(to-dev)` with this resource.
