# TASK-260910-2n0233 results

## Implementation

Implemented the curator client half of R1/P1 against curator-spec `v1.0.0-rc.13` (`23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`), with `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1`. The rollout is direct: no legacy-accept mode or configuration knob. Optional `/v1/log` replay is out of scope, and the client has no `/v1/log` reader.

- `internal/registry/http.go:98-184,187-220,222-292` requires the exact records v2 envelope, verifies each page before its records contribute, and rejects a whole chain without cache fallback or cache replacement. A previous cache entry remains on disk after rejection but is not served for that operation; an outage fallback remains eligible only when that cache boundary matches persisted high-water.
- `internal/registry/boundary.go:10-27,55-157` defines the three closed diagnostics and applies the §9.3 order: missing/unverifiable boundary first, then signature-verified later-page byte comparison, then first-page §5 high-water comparison. Higher first-page state is atomically persisted through the existing snapshot-state writer before records contribute. Later pages do not check or advance high-water. Read-only policy verifies without persistence.
- `internal/registry/snapshot.go:32-38,186-192` stores `boundary_verified` with the existing per-registry high-water and preserves the bit across `/v1/snapshot` state updates. `internal/registry/boundary.go:159-213` restricts cache acceptance to an already-persisted matching high-water and reports read-only posture for trusted registries with pinned keys.
- `internal/registry/attest.go:62-90` carries boundary rejection detail into attest results. `internal/install/install.go` wires state and registry pins into the fetcher. `cmd/curator/envstatus.go:188-224`, `cmd/curator/main.go:832-904`, and `internal/envprofile/status.go` expose `registry_posture` in both status surfaces; unreadable or missing-after-use state makes `--check` fail, while first use is informational. Status reads do not create or repair state.
- Unit tests in `internal/registry/boundary_test.go:139-512` cover first/equal/lower high-water cases, each equal-version body field, exact chain boundary bytes, missing and invalid signatures, precedence, closed v2 envelopes, read-only behavior, atomic persistence before contribution, cache rejection behavior, URL diagnostics, and attestation. CLI posture and `--check` coverage is in `cmd/curator/env_test.go:332-439` and `cmd/curator/main_test.go` (status end-to-end coverage).
- `internal/registry/boundary_conformance_test.go:27-145` drives all pinned `page_boundary_cases` through the production HTTP fetch entry point. The nine expected names are asserted, and rejected vectors assert no returned records, unchanged high-water, registry URL diagnostics, and no cache write.
- User guidance is in `docs/cli.md:250-265` and `docs/troubleshooting.md:711-752`.

The acceptance criterion is exercised by the production fetch path: the below-high-water page is rejected with no records and leaves the persisted high-water and record cache unchanged. This includes the `below-high-water-rejected` rc.13 vector.

## Conformance and ledger

- Pinned rc.13 `page_boundary_cases`: **9/9 passed** from `/tmp/spec-rc13/conformance/v1`.
- `.github/ci/conformance-gaps.tsv`: **72 total rows before / 72 after; 0 owned by STORY-260910-25yc0h or TASK-260910-2n0233 before / 0 after**. No page-boundary row needed re-attribution, and all cases were platform-independent, so no ledger row was added.

## Verification transcripts

Commands ran directly as standalone shell processes; listed exit codes are the observed results.

- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test -count=1 ./internal/registry` — **exit 0** (25.028s).
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test -count=1 -v -run '^TestPageBoundaryConformanceVectors$' ./internal/registry` — **exit 0**; all nine vector subtests passed.
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test -count=1 -run 'TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestStrictRegistryPolicyFailsUnknown' ./internal/install` — **exit 0** (12.499s).
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test -count=1 -run 'TestEnvStatusRegistryBoundaryPostureAndCheck|TestCLIEndToEndInstallStatusAndTamperCheck' ./cmd/curator` — **exit 0** (41.015s).
- `go build ./...` — **exit 0**.
- `go vet ./...` — **exit 0**.
- `golangci-lint run ./internal/registry/... ./internal/install/... ./internal/envprofile/... ./cmd/curator/...` — **exit 0**, 0 issues.
- `gofmt -l cmd internal` — **exit 0**, no output.
- `git diff --check` — **exit 0**.
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test -count=1 ./internal/install` — **exit 1**, interrupted with Ctrl-C after more than three minutes because the broad integration package run was still executing. The focused registry install tests above passed.
- `go test -count=1 ./internal/envprofile` — **exit 1**, interrupted with Ctrl-C after about 1:16 because the broad package run was still executing. The production status path is covered by the passing focused `cmd/curator` tests above.
- The full landing suite was not run locally; the configured handoff runtime runs it once.

## Narrowing mutant evidence

Mutations were restored after each run; current source was rerun by the green commands above. The values below are real `go test -count=1` process exit codes.

- Removing the isolated `log_size` comparison: the prior aggregate equal-version mismatch test **survived (exit 0)** because head/root also differed; the new `log_size`-only case **killed the mutant (exit 1)**.
- Skipping signature verification on later pages: the prior later-page mismatch case **survived (exit 0)**; the added bad-signature/later-mismatch precedence case **killed it (exit 1)**.
- Skipping first-page signature verification: the `bad_signature` case **killed the mutant (exit 1)**.
- Narrowing below-high-water rejection to miss stale versions: the boundary case suite **killed the mutant (exit 1)**.
- Narrowing later-page mismatch so a higher mismatched boundary could pass: the higher-and-mismatch vector/test **killed the mutant (exit 1)**.
- Skipping first-page high-water persistence: `TestRecordsPageBoundaryPersistsBeforeContribution` **killed the mutant (exit 1)**.
- Allowing read-only persistence: `TestRecordsPageBoundaryReadOnlyNeverPersists` **killed the mutant (exit 1)**.
- Allowing fallback for boundary errors: `TestRecordsPageBoundaryRejectsRefreshWithoutServingOrReplacingCache` **killed the mutant (exit 1)** and observed stale cached records returned.
- Returning a partial page chain after a later-page rejection: the later-mismatch test **killed the mutant (exit 1)** and observed the first-page record leak.
- Widening the v2 envelope to accept an extra member: `TestRecordsPageBoundaryRejectsNonV2EnvelopeBeforeStateAdvance` **killed the mutant (exit 1)** and observed a record leak.
- Omitting the pinned-key filter from posture rows: `TestReadBoundaryPostureOnlyListsRegistriesWithPinnedKeys` **killed the mutant (exit 1)**.
- Omitting the env status `NonCurrent` update on unreadable state: `TestEnvStatusRegistryBoundaryPostureAndCheck` **killed the mutant (exit 1)**.
- A mutant that removed only the direct `curator status` unreadable-state exit-code assignment **survived (exit 0)** because the preceding read-only install projection already returns a failure on the same unreadable rollback state. This is a subsumed direct branch; the production failure path and its status row are covered, and the result does not claim that specific assignment mutant was killed.

For new rules that had no corresponding test in the prior source, there is no pre-change executable test gate to call a survivor; the tests above demonstrate the newly added negative gates. This distinction is retained rather than claiming a baseline result that was not run.

## Spec wording to review

§9.3 accepts an equal-version/same-body boundary with “nothing persisted,” while its posture row asks whether the last page boundary received was verified. To honor the no-write rule, `boundary_verified` is persisted atomically with first-use/higher high-water updates; an equal-version accepted page leaves that bit unchanged. The current posture therefore describes the last persisted verification fact. Clarify whether the status fact should update on equality; doing so would require a persistence rule that is not explicit in the pinned contract.

## CHANGELOG entry for release prep

### R1/P1 — client registry page-boundary high-water check

Require a §2-verified signed `boundary` on every `/v1/records` page; reject missing, mismatched, and stale boundaries as `registry_page_boundary_missing`, `registry_page_boundary_mismatch`, and `registry_page_boundary_stale`. Apply §9.3 page ordering and §5 high-water rules before records contribute, and report persisted boundary posture in `curator status` and `curator env status`. Rejected chains do not authorize attestations or update the record cache. Rollout is direct, with no legacy-accept mode or configuration knob; log replay remains optional.

Per campaign release-prep policy, this leaf did not edit `CHANGELOG.md` or `LOGBOOK.md`; the release entry and spec interpretation finding are recorded here. Final `git status --short --untracked-files=all` showed source, tests, and docs only; no binaries, build outputs, `.review/`, coverage files, or `LOGBOOK.md` are in the candidate.

## Revision 3 — seam route for the high-water read

The rev2 hosted validation identified the deny-by-default manager-owned absence-read guard at `internal/envprofile/state_read_guard_test.go:221`: `internal/registry/boundary.go:ReadOnlyStateDir` inspected the protected rollback-state directory with direct `os.Stat`. The gatefix instruction required routing this check through `internal/stateread` and distinguishing proven absence from unreadable state.

`internal/registry/boundary.go:218-228` now calls `stateread.Stat` and selects the legacy directory only when the seam returns `KindAbsent` without an error. Present or unreadable state keeps the protected path selected; the caller's following state read reports the unreadable diagnostic and cannot use legacy state as a fallback. `internal/registry/boundary_test.go:124-177` adds absent, present, and unreadable rows. The unreadable row uses a blocked parent, asserts the seam reports `KindUnreadable`, and verifies read-only posture reports `manager_state_unreadable`.

### Revision 3 verification

Each check was a standalone process with its real exit code:

- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` — **exit 0** (20.721s).
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test ./internal/registry -count=1` — **exit 0** (24.780s), including the rc.13 client-boundary conformance test.
- Narrowing mutant: changed the fallback condition to treat any stat error as absence; `go test ./internal/registry -run '^TestReadOnlyStateDirFallsBackOnlyForProvenAbsence/unreadable_state_does_not_fall_back$' -count=1` — **exit 1**, expected; the test caught the protected-path-to-legacy fallback. The mutation was restored.
- Restored focused test `go test ./internal/registry -run '^TestReadOnlyStateDirFallsBackOnlyForProvenAbsence/unreadable_state_does_not_fall_back$' -count=1` — **exit 0** (1.052s).
- `gofmt -w internal/registry/boundary.go internal/registry/boundary_test.go` — **exit 0**; follow-up `gofmt -l` on both files — **exit 0**, no output.
- `git diff --check` — **exit 0**.

I fetched `origin main` successfully; it resolved to `55b94af251d72fe78637e1af9b51fbc807e7ed67`. I did not manually run the hosted/full landing gate; the handoff runtime runs it. The candidate status contained only source, test, and documentation paths, with no build outputs, `.review/`, coverage files, or `LOGBOOK.md`. No CHANGELOG or LOGBOOK edit was made; the existing release-prep text and findings remain in this outcome resource.

## Revision 4 — unreadable high-water through the fetch path

Revision 4 addresses review finding F1 from `TASK-260910-2n0233_review-verdict-rev3.md`. No production behavior changed: `openPageChain` already returns a `pageStateError` when the existing high-water file is unreadable. The missing evidence was a FetchFn-level row that made that exact read fail without letting the catalog's separate missing-after-use guard mask it.

- `internal/registry/boundary_test.go:124-168` builds the persistent `NewHTTPFetch`, persistent `NewHTTPFetchWithPolicy`, policy read-only, and `NewHTTPFetchWithPolicyReadOnly` fetchers and checks the typed fail-closed error/no-record result.
- `internal/registry/boundary_test.go:170-244` exercises each constructor against an httptest registry with a stale signed page and either a directory at the state-file path or corrupt JSON. The valid empty catalog models interruption after state-file publication but before catalog update, so this row independently proves the unreadable read is not collapsed to absence. It asserts no request reaches the registry, no record cache is written, the catalog is unchanged, and corrupt bytes or the directory marker remain unchanged.
- `internal/registry/boundary_test.go:246-281` separately covers a state path that is missing while the catalog still lists it, across all four fetch constructors.
- Production fail-closed branches remain at `internal/registry/boundary.go:60-72`.

### M2 narrowing mutant

The exact reviewer mutant changed `openPageChain` to discard `readSnapshotState` errors and continue with an empty state. Against the pre-existing boundary tests, the mutant survived (exit 0). The first new test setup also survived (exit 0) because the catalog-listed-missing guard independently rejected it; I corrected the test to use a valid empty catalog and kept a separate row for the catalog-listed branch. With that corrected setup, the same M2 mutation was killed (exit 1): the persistent directory cases detected network contact, the persistent corrupt-state cases detected an accepted fetch/state replacement, and the read-only cases detected the accepted stale page. I restored production source after the mutant run; the registry suite is green on the restored candidate.

The previous independent rev3 review also recorded M1 (stale high-water comparison narrowed away) killed by four tests (exit 1); that production check was unchanged in revision 4.

### Conformance ledger and trunk refresh

The rc.13 `page_boundary_cases` set remains 9/9, driven through `TestPageBoundaryConformanceVectors` at `internal/registry/boundary_conformance_test.go:27-68` from `/tmp/spec-rc13/conformance/v1` (pin `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`). The refreshed `d41da0fb` baseline has 70 non-comment rows in `.github/ci/conformance-gaps.tsv`; before/after the revision there are 0 rows owned by `STORY-260910-25yc0h` or `TASK-260910-2n0233` (0 -> 0). The ledger file matches trunk; no platform row was needed because the directory-in-place and corrupt-JSON cases run on all supported platforms.

The authoritative managed candidate refresh verified base `d41da0fbf210f6bc0ea1a5cef3d81a12af9eb439` (`refresh_already_current`). `git diff --name-only origin/main -- . ':!.task-board'` lists only this leaf's tracked paths; `git status --short --untracked-files=all` contains 17 modified tracked paths and the three boundary Go files, with no `.task-board`, binary, build output, `.review/`, `LOGBOOK.md`, or coverage file.

### Revision 4 verification

Each command ran as a standalone process; exit codes below are observed results on the refreshed candidate:

- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test ./internal/registry` — exit 0 (23.043s), including all nine pinned page-boundary vectors.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$'` — exit 0 (25.129s).
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1 go test ./cmd/curator -run '^(TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI)$'` — exit 0 (76.165s); both adjacent status tests survived the current-trunk union.
- `go build ./internal/registry ./internal/envprofile ./cmd/curator` — exit 0.
- `go vet ./internal/registry ./internal/envprofile ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/registry/... ./internal/envprofile/... ./cmd/curator/...` — exit 0, 0 issues.
- `gofmt -l cmd internal` — exit 0, no output.
- `git diff --name-only origin/main -- . ':!.task-board'` — exit 0; only leaf paths.

The full landing suite was not run locally; the handoff runtime runs it once.

### Board environment correction

The shell's inherited `TASK_BOARD_DIR` pointed at the worktree checkout copy. Once detected, I used the authoritative board explicitly with `--board-dir /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board` for subsequent reads and writes, reconfirmed the task at `development`, and verified the managed candidate refresh there. The checkout copy's incidental changes were restored from the candidate HEAD and are absent from the candidate status. No manual edits were made to the authoritative board.
