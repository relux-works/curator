# TASK-260917-16l2md — review verdict, Change Request revision 3

Verdict: **changes_requested**, route to **to-dev**. Do not accept CR-TASK-260917-16l2md-3.

Reviewed base `3c45d4bbed81348dfc814bac98b87a58aadb02af`, candidate tree `17d166e7664eeaafa15c01078788f2ca12d75cbf`. Downloaded patch SHA256 is `0dfd13b04421fb4547643ada2c9dafa144a6be62fc6f8fed47c12af968ed27e8`, matching the assignment. Tests use a disposable `git archive` copy of that exact tree; no candidate code changed. Tracked worktree files match the candidate tree, and the untracked Windows test's blob equals `23e521d331b7e532dd9612785d12498af7b7253c` in the tree. Spec worktree is exactly `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

## Required corrections

### 1. High — emit S4 declaration rows before publication and materialization

The pinned environments §2.3 (lines 478–482) requires the manager to **print** the rows after audit succeeds and before publishing the lock or materializing any surface. Computing strings before publication does not satisfy that obligation.

`internal/envprofile/envprofile.go:744` computes rows, but `:758` publishes the lock and `:775` activates before returning `Info.Surfacing`. `cmd/curator/profile.go:117–129` only prints the rows after `Install` returns. The same ordering exists for changed path reinstalls (`envprofile.go:878–904`) and updates (`:1099–1123`; CLI `profile.go:308–320`). A publication/resync failure can also discard the computed rows through an empty `Info` error return.

Independent production-entry probe `TestReviewerSurfacingBeforePublication` reuses the committed CLI install fixture and invokes real `run()` with a stdout writer that stats the profile lock when it receives the first `mcp-declaration` row. It fails on unmodified production code:

```
lock.json already exists when first mcp-declaration row is written: .../profiles/withmcp/lock.json
--- FAIL: TestReviewerSurfacingBeforePublication
EXIT:1
```

Fix the operation-to-CLI emission seam so the actual output happens at the required point, including install/reinstall/update and before activation/resync. Keep informative surfacing non-fatal and avoid printing twice. Add committed tests observing filesystem/publication state at emission and old-lock identity on update, including a failure after the emission point.

The current `runSurfacingOrderCase` at `internal/envprofile/envpassthrough_conformance_test.go:609–659` checks that the vector states the expected order, audit-blocked operations return no rows, and a helper formats the candidate lock. It never observes the runtime order. Both order vectors pass while the production probe fails. Replace this proxy coverage with observation of the required event order; do not weaken vectors or claim 2/2 order cases behaviorally proved by the existing tests.

### 2. Medium — reject explicit null provider_directories

`internal/config/environments.go:315` guards parsing with `present && rawPD != nil`, silently accepting `provider_directories:null` as `[]`. The pinned manager-config-v2 schema at lines 531–543 requires an array, with default `[]`; the system schema references the same definition. Unlike `passable_env_names`, this knob has no null meaning.

Independent production `Load` probe reports:

```
--- PASS: TestReviewerProviderDirectoriesNullRejected/omitted
--- PASS: TestReviewerProviderDirectoriesNullRejected/empty
Load accepted provider_directories:null as []string{}
--- FAIL: TestReviewerProviderDirectoriesNullRejected/null
EXIT:1
```

Parse every present value and reject null with the existing wrong-type configuration error naming the knob. Add committed Load regressions for user and system configuration, retaining omitted/empty-list acceptance. This is the same absence-versus-null error class already fixed for E2 waivers; the E2 correction itself remains closed.

### 3. Low — supply the requested operator documentation

The candidate has no `docs/` changes. Outside the historical security-audit report there is no documentation of the new knobs in `docs/` (search for `provider_directories`, `transitive_system_modules`, `passable_env_names`). The task's DoD explicitly requires docs updated. Add concise operator documentation covering the three knobs/policies, waiver example, warning defaults, migration steps, and explicit null versus absent semantics. Preserve the historical audit as a historical report rather than treating it as current usage documentation. CHANGELOG already carries the required entries.

## Per-item assessment

| Item | Assessment |
|---|---|
| Pin and scope | One SPEC_PIN at ci.yml:53 equals dced9b8; all five suite checkouts reference env.SPEC_PIN. The distinct inputs.candidate_ref is the pre-existing opt-in candidate checkout. release.yml unchanged. No vendored spec or E1/R1/S6 implementation in the 37-path delta. |
| Union fidelity | Independently compared added non-comment lines of all three supplied patches against the candidate and inspected parser, lock maps, status, and vector drivers. Missing lines are documented union-arm/formatting rewrites, root-content skip/ledger deletions, E4 selected-shim-directory hardening, or excluded board-results files. E2 admission and S4 runtime additions remain. This line check supplements code inspection; it is not semantic completeness proof. See union.log. |
| Config/locks | Both E2 and E4 lockable keys retained, waivers not lockable; transitive policy locks only toward error. S4 renders absent as [] and explicit null as null, carrying presence to runtime. E4 null bug is correction 2. |
| E2 behavior | Direct-set root/overlay/dependency admission, waiver admission, drop warning and admitted bytes, error refusal, pre-publish install/reinstall/update gate, and posture retained. All 5 admission vectors and 7 schema-subset cases pass. Independent error-threshold mutant killed. |
| E2 prior review closure | Exact manager-config comparison retained; no postPinKnobs/prunePostRevisionKnobs. Null-waiver committed regression passes. Dedicated subset drivers remain and fail for absent/partial publication, replacing obsolete root-content skips. |
| E4 behavior | Active constant is revision A; spec expressly retains PATH selection with outside-root warning. B searches roots and refuses PATH-only candidates. Managed/published-directory refusal, unreadable-root handling, same-directory/ancestor identity checks and status currency inspected. All 14 vectors run under both revisions (28/28 expected outcomes). Local case-variant identity test passes. Windows 8.3 test read; Windows execution accepted only from exact rev3 hosted evidence. |
| S4 behavior | Active s4-warn and selectable s4-enforce retained; default/explicit-null/list semantics and allowlist-empty install/update/status warnings exercised. Fragment bound calls ResolvePassthrough from production buildFragment. Real Resolve empty-list regression kills mutant. Row formatting/posture pass, but output order fails correction 1. |
| Conformance coverage | 48/48 manager-config cases, 7/7 E2 schema cases, 5/5 admission cases, 14/14 umbrella cases with both revisions, and all 25 S4 vector subtests execute with zero skips in targeted logs. S4 ordering has 2 syntactically executed cases but 0/2 actual emission-order observations in the committed driver; independent install observation fails. |
| Skip accounting | Four candidate root-content ledger rows and six absence skip branches removed; ledger remains base-identical. Identity tests use recognized host-capability reasons. No new ledger rows for identity tests despite rev2 note requesting them; hosted gate permits these classed host-capability skips. No vector family is hidden by these changes. |
| Shell-hook-trust bound | No shell-hook-trust vector driver exists in candidate; producer explicitly reports it un-driven. Shell package passes, which is not proof that this vector family executes. Do not mark the brief's broader vector-family assertion fulfilled. S6 implementation is expressly outside this review's scope. |
| Release/docs | E2 default-drop/error-opt-in, E4 A-before-B, S4 warn-before-enforce, and rc.12 pin notes present in CHANGELOG. Operator docs missing: correction 3. |
| Architecture | Admission and passthrough helpers reused at production boundaries; E4 lookup stays in CLI. S4 requires an output seam across the existing operation boundary, not a spec change or human-only architecture decision. |

## Independent checks and evidence reuse

Shell bash, `set -o pipefail`, `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12-review/conformance/v1`, Go tests use `-count=1`. Full commands and stdout/stderr are in the task-scoped evidence archive.

- `go build ./...`: exit 0.
- `go vet ./...`: exit 0.
- `gofmt -l internal cmd`: empty, exit 0.
- Full tests for `internal/config`, `contextresolve`, `contextmaterialize`, `contextaudit`, `shell`, `interop/environments`, `envfragment`, `globalbins`: all pass, exit 0 (`narrow.log`).
- `cmd/curator -run 'Test(Umbrella|ProviderInputs|ActiveRevision|EnvStatus|ProfileInstallLists|ProfileUpdateLists)'`: exit 0, 39 PASS entries including subtests, zero skips (`cli.log`). Includes local identity logic and real install/update/status.
- `internal/envprofile -run 'Test(EnvironmentsEnvPassthroughVectors|InstallError|UpdateError|Drop|StatusReportsPolicy|StatusErrorReports|PolicyFromConfigCarries)'`: exit 0, 35 PASS entries including subtests, zero skips (`profiles.log`).
- Explicit config vector/subset/null-waiver rerun: exit 0, 58 PASS entries, zero skips (`config-vectors.log`). Admission-vector rerun: exit 0, 6 PASS entries, zero skips (`admission-vectors.log`).
- `bash .github/ci/gate-selftest.sh`: exit 0, 180 passed / 0 failed (`gate-selftest.log`).
- `golangci-lint run ./internal/config/... ./internal/contextmaterialize/... ./internal/envfragment/... ./internal/envprofile/... ./cmd/curator/... ./internal/globalbins/...`: exit 0, 0 issues (`lint.log`).
- Reviewer probes: publication-order and null-provider probes both exit 1 for the defects described above. These are additional reviewer tests through overlays, not committed tests.

Full envprofile and CLI suites were **not** independently replayed: this review used the stated targeted masks and found reproducible rework. Linux/Windows/race/full-package coverage is accepted only as existing hosted evidence, not claimed as a local rerun. Attached rev3 validation log reports run 35286627425, all required test/race/lint/interop/self-test lanes successful, exit 0; optional candidate and rose-air lanes skipped. Its own test_case_coverage is unknown. No remote gate rerun performed.

## Narrowing mutants: 3/3 killed

All use Go overlays against the disposable exact candidate, no candidate edits. This is bounded mutation evidence, not exhaustive gate coverage.

| Candidate | Narrowing | Committed tests that kill it | Exit |
|---|---|---|---|
| E2 | Refuse error-policy materialization only when more than one module is dropped (`len(dropped)>0` → `>1`) | TestSystemPromptErrorRefusesFirst; admission vector system-module-transitive-error through SystemPrompt | 1 |
| E4 | Refuse managed/published candidates only under revision B, bypassing the revision-A refusal | Three umbrella vectors; TestUmbrellaRefusedDirectoriesBothRevisions; TestUmbrellaSymlinkIntoManagedRefused; case-variant identity regression | 1 |
| S4 | Admit an unlisted name when effective list length is zero, retaining the nonempty-list bound | Enforce-absent vectors; TestResolvePassthroughProfiles; production Resolve test TestResolvePassthroughKnobs/empty-list-bounds-all | 1 |

Harness corrections: initial /tmp overlay keys did not match Go's /private/tmp canonical paths; those early no-op runs are not mutation evidence. Final overlay keys use resolved paths. The first null probe had a field-name compile typo, then an unrelated missing-system-file error; final probe clears the optional system override, checks rejection attribution, and runs omitted/empty controls. Only final substantive failures count.

## Lifecycle and logbook

`task-board spawn goal "$TASK_BOARD_RUN_ID"` reports no active goal binding. No directives pending at the checked checkpoint. Campaign instructions prohibit LOGBOOK.md edits; this verdict plus board notes is the persistent findings/logbook record. Attach verdict and evidence before routing to-dev. No accept_cr, commit_ack, code edit, commit, branch change or push. These are ordinary implementation/test/documentation corrections, not an external blocker or a human-only decision.
