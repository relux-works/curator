# TASK-260916-55g9dg results — E2 direct-only system modules

## Candidate and scope

- Story worktree: `STORY-260916-2d9coh`, branch `task-board/story/STORY-260916-2d9coh`.
- Candidate base: current trunk, `bd3c0f436226ec831c13b3a307c30acd035c73c2`; the task scope's E2 production implementation is present on current trunk. This revision adds the rc.13 first-blocker accounting, explicit E2 subset accounting, deterministic config diagnostics, and the null/status/CLI regression coverage described below.
- Curator `SPEC_PIN` remains rc.13 commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` (`.github/workflows/ci.yml:43`). No spec bytes were vendored and no pin was changed.
- Normative contract checked: curator-spec `protocol/environments.md` §3, §5.5, §5.7, §12.1, §12.2, §12 posture, §13; `profiles/manager.md` §1. The rc.13 source was checked out in a temporary clone at the exact pin. It is outside the worktree.

## Changes in this revision

- `internal/config/environments.go:182-190,290-305,940-946,1082-1093`: deterministic ordering for unsupported environment fields; explicit `null` for `system_module_waivers` is a wrong-type error while omission defaults to `[]`; config serialization and the system-file lock direction remain explicit.
- `internal/config/environments_test.go:635-669,671-722`: `Load`-entry regressions for explicit null, omitted/empty defaults, stable first unsupported-field diagnostics, and the `drop` lock-direction / waiver-not-lockable rules.
- `internal/config/system_module_schema_test.go:1-85`: explicit accounting for the seven E2 schema cases from the pinned schema index. The existing manager/system family consumers feed published exact case bytes through `Load`. Current rc.13 publishes all seven; tests fail loudly if the pinned family disappears. No `root-content` skip is used because this committed pin serves the family and the gate forbids a skip that can never legitimately occur.
- `internal/config/environments_conformance_test.go:233-304,306-350` and `.github/ci/conformance-gaps.tsv`: restore full canonical-JSON equality for manager vectors, with a regression rejecting extra normalized keys; establish each schema/vector row's first production blocker and attribute it to that surface. For manager cases the deterministic first blocker is `environments.permissions` (STORY-260922-1cenbr); for system cases it is locked `environments.source_signers` (STORY-260916-ioemse). E2 is not credited for those still-blocked cases.
- `internal/envprofile/status.go:296-307,551-663`: propagate admission-policy/materialization errors; an unreadable transitive manifest reports `context_manifest_invalid` and an unknown dropped set instead of treating the failed read as absence. Status still reports the effective policy and each dropped module's package/path.
- `internal/envprofile/admission_test.go:368-430`: portable unreadable-manifest regression; the test asserts unknown dropped modules, a path-bearing diagnostic, and non-current status.
- `cmd/curator/env_test.go:330-381`: production `env status` command regression for the effective `drop` policy and the `sysleaf 90-system.md` dropped-module entry.

## E2 behavior and acceptance mapping

| Requirement | Evidence |
|---|---|
| Defaults, grammar, writing, and lock direction | Defaults are `drop` and empty waivers (`internal/config/environments.go:150-168`); parsing/closed validation at `:290-305,825-850`; effective output at `:1082-1093`; lockable key at `internal/config/config.go:54-67`; `drop` is rejected in a system lock at `internal/config/environments.go:940-946`. New null and deterministic-diagnostic tests are listed above. |
| Direct, transitive, and waiver admission | `internal/contextmaterialize/admission.go:66-102,129-150` defines direct packages as the root, active overlays, and contexts-required packages; waived packages are admitted. |
| Drop warning/output and error refusal | `internal/contextmaterialize/admission.go:117-127,153-198` reports `context_system_module_transitive` with package and module under `error`; `internal/contextmaterialize/contextmaterialize.go:257-285` assembles admitted bytes only and returns dropped modules. Resolution checks before publish at `internal/envprofile/envprofile.go:764-766,877-879,1361-1398`; the warning/fragment admitted-set path is `internal/envprofile/managed.go:2270-2312,1905-1925`. Existing E2 integration tests cover install/update refusal without replacing the lock, drop warning, waiver and admitted fragment semantics (`internal/envprofile/admission_test.go:65-245`). |
| Always-warn finding preserved | Existing `context-system-module-present` audit behavior remains in `internal/contextaudit`; E2 admission tests do not replace that finding with the admission warning. |
| Status posture | Effective policy and all dropped package/path pairs are exposed at `internal/envprofile/status.go:581-591,604-662` and printed by `cmd/curator/envstatus.go:107-115`. Drop warnings do not themselves mark a provisioned row non-current (`internal/envprofile/admission_test.go:266-325`). Failed manifest reads are unknown plus diagnostic, never absence (`status.go:604-639`, regression above). |
| Vectors | `internal/contextmaterialize/system_module_admission_test.go:34-43,100-141,143-240` selects and byte-compares the five rc.13 system-module admission cases through `SystemPrompt`/`ClassifySystemModules`: direct, transitive-drop, transitive-error, waived, and overlay-direct. `internal/config/environments_conformance_test.go:268-304` compares the full manager effective JSON and rejects unexpected normalized keys. Both drivers require every named case from the pinned root. |
| Schema cases | `internal/config/system_module_schema_test.go:30-85` accounts for all seven E2 cases; the manager/system schema tests run their exact published bytes through production `Load`. The cases blocked earlier by other surfaces remain in the ledger under those owners. |

## Previous verdict closure (revision 2)

1. Removed `postPinKnobs` / `prunePostRevisionKnobs`; `TestManagerConfigV2Vectors` now compares canonical JSON for the complete normalized object (`internal/config/environments_conformance_test.go:268-304`). The rc.13 exact manager vector family is green (31 driven, 25 separately attributed known gaps). `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` names the regression and the matching narrowing mutant below proves expected-key-only comparison fails it. No SPEC_PIN change.
2. `TestSystemModuleWaiversNullRejected` calls production `Load`, rejects explicit `null`, and keeps absence/empty accepted (`internal/config/environments_test.go:635-651`).
3. The five-vector admission subset is explicitly selected from the current pinned root and fails if any case is missing (`internal/contextmaterialize/system_module_admission_test.go:34-43,100-141`); the seven schema cases have the same fail-closed presence accounting (`internal/config/system_module_schema_test.go:30-85`) and their family consumers run through `Load`. The rc.13 pin publishes these families. I did not add an impossible `root-content` skip/ledger row: current-pin gate policy rejects that skip when the pin serves the content. This follows the current rc.13 rework brief and keeps the pinned suite fail-closed.

## Conformance and gap ledger

Conformance root used for pinned runs:

`/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/curator-spec-rc13-checkout-9am5ezqk/spec/conformance/v1`

The repository's normal `curator-spec` checkout was newer than the curator pin. I used a temporary shared Git clone checked out at the exact rc.13 commit so its fixture bytes retain literal `$Format` placeholders. An earlier `git archive` extraction substituted those placeholders; the full interop test failed on the altered fixture (exit 1). I discarded that extracted root and reran against the exact clone, where the interop package passed (exit 0).

| Driver | Result on exact rc.13 root |
|---|---|
| `TestSystemModuleAdmissionVectors` | 5/5 E2 vectors driven through production `SystemPrompt`; byte-exact expectations pass; 19 other cases reported bounded; 0 skipped. |
| `TestManagerConfigV2Vectors` | 31 driven, 25 known gaps, 0 bound/skipped (56 total). |
| `TestManagerConfigV2SchemaCases` | 78 driven, 29 known gaps (107 total). |
| `TestSystemConfigV2SchemaCases` | 36 driven, 6 known gaps (42 total). |
| E2 schema subset | All 7 pinned E2 cases are explicitly selected/accounted, and the schema-family consumers feed their exact inputs through `Load`. They remain known gaps where another earlier blocker stops processing: five manager cases first-block at unsupported `permissions`; two system cases first-block at locked `source_signers`. |
| E2-owned materialization gap rows | 0 before and 0 after; all five E2 admission vectors pass. |

`.github/ci/conformance-gaps.tsv` remains at 69 rows. Owner histogram before → after:

| Owner | Before | After |
|---|---:|---:|
| BUG-260923-2afgyq | 5 | 5 |
| STORY-260910-2qmrb8 | 3 | 3 |
| STORY-260910-6bo7ej | 2 | 2 |
| STORY-260916-1i1gfo | 2 | 2 |
| STORY-260916-ioemse | 11 | 10 |
| STORY-260916-ioemse + STORY-260922-1cenbr | 28 | 0 |
| STORY-260922-1cenbr | 16 | 45 |
| STORY-260925-1v7pvn | 2 | 2 |

The 28 combined overlay rows and the manager `schema2-every-knob` first-blocker row now belong to STORY-260922-1cenbr because `permissions` is the first unsupported/differing field; the E1 count falls by one. This records the first actual blocker rather than claiming E2 or E1 code was reached. No remaining E2 row passed unnoticed or was removed from the ledger.

## Validation transcripts

Commands below were run directly as standalone processes in `zsh`; no pipeline was used. Exit codes are the real process exits.

Green:

```text
0  go build ./...
0  go vet ./...
0  gofmt -l internal/ cmd/                  (no output)
0  git diff --check
0  golangci-lint run ./internal/config/... ./internal/contextresolve/... ./internal/contextmaterialize/... ./internal/contextaudit/... ./internal/envfragment/... ./internal/envprofile/... ./cmd/curator/... ./internal/interop/environments/...  (0 issues)
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 ./internal/config/...
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 -run '^TestManagerConfigV2Vectors$' -v ./internal/config (31 driven, 25 known-gap, 0 bound/skipped)
0  go vet ./internal/config/...
0  golangci-lint run ./internal/config/... (0 issues)
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 -run '^TestSystemModuleAdmissionVectors$' -v ./internal/contextmaterialize
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 -run '^(TestManagerConfigV2SchemaCases|TestSystemConfigV2SchemaCases|TestSystemModuleSchemaSubset|TestManagerConfigV2Vectors|TestOverlayGapOwnersMatchFirstProductionBlocker)$' ./internal/config
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 ./internal/interop/environments/...
0  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 ./internal/envprofile/... -run '^Test(InstallError|UpdateError|DropRepair|DropAllDropped|StatusReportsPolicyAndDropped|StatusErrorReportsPolicyWithoutDropped|StatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|PolicyFromConfigCarriesAdmission)$'
0  go test -count=1 ./internal/contextresolve/...
0  go test -count=1 ./internal/contextmaterialize/...
0  go test -count=1 ./internal/contextaudit/...
0  env CURATOR_CONFORMANCE_ROOT=<configured conformance root> go test -count=1 ./internal/envfragment/...
0  go test -count=1 ./cmd/curator/... -run '^TestEnvStatusReportsDroppedSystemModuleThroughCLI$' -v
```

Expected/red run, reported as such:

```text
1  env CURATOR_CONFORMANCE_ROOT=<exact rc.13 root> go test -count=1 ./internal/envprofile/...
   Go's 10-minute test timeout fired in TestWeightRulesApplyInOrder while gitops.Clone/ensureRepo was running. The focused E2 envprofile selection above passed. A standalone retry of `go test -count=1 -timeout 45s ./internal/envprofile/... -run '^TestWeightRulesApplyInOrder$' -v` exited 0 in 12.6s (three subtests); the broad package run remains non-green and the cause was not established.
```

The hosted `scripts/remote-gate.sh` was not run manually; handoff runtime owns that one full gate. This host is macOS, so local validation does not independently establish Linux/Windows behavior.

## Narrowing mutants

Each mutant was installed using Go's `-overlay` from `/tmp/e2-mutants`, outside the worktree. Every mutant test command exited 1 because its committed regression caught the narrowed behavior; survivors: 0/8.

| Mutant | Catching regression | Exit |
|---|---|---:|
| Treat explicit null waivers as absence | `TestSystemModuleWaiversNullRejected` (`Load`) | 1 |
| Exclude active overlays from `DirectSet` | overlay-direct admission vector / direct-set unit coverage | 1 |
| Apply error refusal only when a transitive package has at least two modules | `TestInstallErrorRefusesTransitiveSystemModule` | 1 |
| Suppress drop warning for a single dropped module | `TestDropRepairWarnsAndFragmentFollowsAdmitted` | 1 |
| Treat unreadable transitive manifest as an empty dropped set | `TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops` | 1 |
| Ignore waiver map during classification | `TestSystemPromptWaiverAdmits` | 1 |
| Permit system-file lock toward `drop` | `TestSystemTransitiveDirection` | 1 |
| Compare only expected manager-vector keys, allowing extra normalized knobs | `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` | 1 |

Bounds: mutants establish that these committed tests reject the listed narrowing changes. They do not prove all possible malformed manifests, filesystem races, or platform-specific path behavior. The exact rc.13 vectors and cross-platform hosted gate cover additional cases; hosted CI remains pending handoff.

## Checklist mapping and remaining bounds

- Config grammar/defaults/writing and lock rule: implemented and tested. The older rc.11 `root-content` requirement was superseded by the current rc.13 pin contract; current-pin tests require the cases and fail if absent.
- Admission, drop/error diagnostics, unchanged lock on resolution refusal, waiver, admitted-set fragment, status policy/path posture, and five vectors: implemented and covered as cited.
- Lint and build/vet/format validation: green. The config package, five admission vectors, focused E2 envprofile regressions, CLI posture test, and standalone weight-rule retry are green. Full `internal/envprofile/...` is the one validation command that timed out red, recorded above; no claim is made that this broad package command passed.
- CHANGELOG and LOGBOOK files were not edited per current campaign rules. Findings and release text are in this task outcome resource.
- No SPEC_PIN update, no vendored conformance files, and no ax integration or proposal work.

## CHANGELOG entry (for release prep)

E2: restrict system-class context modules to direct packages and explicitly waived packages. The default `transitive_system_modules=drop` is non-breaking and warns with `context_system_module_dropped`; `error` is opt-in and refuses resolution with `context_system_module_transitive`. Add `system_module_waivers` for reviewed package exceptions.

## Revision 6 — re-apply on 97ca3370

This revision re-applies the accepted E2 delta from `refs/campaign/55g9dg-rev4-full-20260927` onto clean trunk `97ca3370`. It remains uncommitted on `task-board/story/STORY-260916-2d9coh`.

### Conflict resolutions and boundary

- `.github/ci/conformance-gaps.tsv`: the rev4 ledger described `permissions` as an unsupported blocker, but trunk now parses and writes that knob. Kept trunk's permissions and fragment-v2 ledger state, then updated the five affected E2 manager-schema rows to the actual first remaining blocker, E1's `require_source_signers`. The current gap rows are at lines 7–10 and 33.
- `internal/config/environments_conformance_test.go`: kept the accepted rev4 exact full-object JSON comparison and extra-key regression. For the overlapping gap attribution, kept the current first-blocker logic for signer fields and expected `require_source_signers` as the deterministic first unsupported field after trunk's supported `permissions` knob.
- The other six paths applied cleanly: `cmd/curator/env_test.go`, `internal/config/environments.go`, `internal/config/environments_test.go`, `internal/config/system_module_schema_test.go`, `internal/envprofile/admission_test.go`, and `internal/envprofile/status.go`.
- Trunk features remain present: the permissions parser/serializer is still in `internal/config/environments.go:28-31,242-247,1106-1120`; `cmd/curator/env_test.go:61` still asserts `launch-env-fragment-v2`. `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the eight paths above, `git diff --cached --name-only` is empty, and `git diff --check` is clean. No commit was created.

### Regression and narrowing mutant

- Named regression: `TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops`, `internal/envprofile/admission_test.go:366-433`. It drives `StatusOf` with an unreadable transitive package manifest and requires the dropped-module set to remain unknown, the row to be non-current, and a manifest diagnostic with the package/path context.
- Narrowing mutant: replaced the unreadable-manifest diagnostic return in `internal/envprofile/status.go:633-640` with `continue`, which treats the unreadable transitive manifest as absent. With the mutant kept buildable, the named regression failed at `admission_test.go:420`: it observed `DroppedSystemModules: []` where nil/unknown was required (mutant command exit 1). The candidate source was restored, and the focused status tests passed afterwards.
- A first, broader mutant attempt also exited 1 at compilation because deleting the diagnostic left `manifestPath` unused. That attempt was not counted as a killed mutant; the compile-safe mutant above is the evidence.

### Validation transcripts

All validation commands were run directly as standalone processes; exit codes below are the actual process exits.

```text
0  go test ./internal/contextmaterialize ./internal/contextresolve -count=1
0  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/contextmaterialize ./internal/config -run 'SystemModuleAdmissionVectors|ManagerConfigV2|SystemConfigV2|SystemModuleSchemaSubset|SystemModuleWaivers' -count=1
0  go test ./internal/envprofile -run '^TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops$' -count=1
1  go test ./internal/envprofile -run '^TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops$' -count=1  (compile failure on the first, overly broad mutant attempt; not counted as a kill)
1  go test ./internal/envprofile -run '^TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops$' -count=1  (compile-safe narrowing mutant; expected regression failure at the unknown-set assertion)
0  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -count=1
0  go test ./cmd/curator -run '^TestEnvStatusReportsDroppedSystemModuleThroughCLI$' -count=1
0  go test ./internal/envprofile -run 'TestStatus(UnreadableTransitiveManifestDoesNotBecomeNoDrops|ReportsPolicyAndDropped|ErrorReportsPolicyWithoutDropped)$' -count=1
0  go build ./...
0  go vet ./...
0  gofmt -l cmd/curator/env_test.go internal/config/environments.go internal/config/environments_conformance_test.go internal/config/environments_test.go internal/config/system_module_schema_test.go internal/envprofile/admission_test.go internal/envprofile/status.go  (no output)
0  git diff --check
```

The full hosted `scripts/remote-gate.sh` was not run in this bounded re-apply; handoff runtime owns that gate. The old rev4 verdict resource (`TASK-260916-55g9dg_review-verdict-rev4.md`) records acceptance, although the appended Review Round Brief calls rev4 a rejection. The re-apply followed `55g9dg-reapply-6.md` and also supplies the requested named regression plus a killed narrowing mutant; no new review verdict was issued in this developer run.
