# TASK-260910-2vnjej results

## Delivery

Implemented schema-2 signed bootstrap checkpoint handling, durable first-use TOFU posture, optional mirror-group root comparison, and read-only status reporting. Registry divergence remains report-only during resolution: accepted registries and their audit attestations are unchanged. Advisory policy reports a warning; strict policy reports an error and makes status non-current.

The applied-checkpoint identity is persisted with registry state. This keeps an unchanged bootstrap checkpoint current after later accepted network views advance its high-water, while a changed checkpoint is evaluated as a rebootstrap. The rc.13 first-use vector caught this status edge case.

## Normative clauses and production rows

Spec pin: curator-spec `v1.0.0-rc.13`, tag commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

| Rule and clause | Production consumer and row | Pinned rc.13 cases |
|---|---|---|
| Schema-2 registry members; manager profile §1 | `internal/config`: `TestManagerConfigV2SchemaCases`, `TestManagerConfigV2Vectors` | `valid-registry-bootstrap-members.json`; `schema2-registry-bootstrap-members` |
| Signed first-use checkpoint, first-use TOFU, rebootstrap persistence/regression; registry protocol §5 | `internal/install`: `TestRegistryBootstrapVectorsDriveInstall` drives `install.Project -> resolveRegistries` | `checkpoint-first-use-accepted`; `checkpoint-first-network-below-tampered`; `checkpoint-first-network-equal-different-tampered`; `no-checkpoint-first-use-tofu`; `checkpoint-signature-invalid-first-use`; `rebootstrap-advance-accepted`; `rebootstrap-equal-consistent-noop`; `rebootstrap-regression-refused`; `rebootstrap-equal-inconsistent-refused`; `rebootstrap-signature-invalid-ignored` |
| Cross-registry equivocation and same-log-size root comparison; registry protocol §5.1 | `internal/install`: `TestRegistryMirrorComparisonVectorsDriveInstall` drives `install.Project -> resolveRegistries` | `divergence-detected-advisory`; `divergence-detected-strict`; `divergence-views-agree`; `divergence-different-sizes-skipped`; `divergence-single-registry-skipped` |
| Per-registry bootstrap and mirror status; manager profile §10 | `registry.ReadBoundaryPostureWithPolicy`; CLI `curator status` / `curator env status` via `TestEnvStatusRegistryBoundaryPostureAndCheck` | Bootstrap and mirror vector consumers assert current status, severity, and persisted comparison outcome; `TestRegistryPostureReportsBootstrapAndMirrorSeverityWithoutWrites` checks read-only status |
| Protected checkpoint path; manager profile §1 and registry protocol §5 | `internal/registry/readBootstrapCheckpoint`; `TestOpenReadNoFollowRejectsFinalSymlink` | First-use checkpoint vectors run through install; final-component symlink refusal is separately tested |

The mirror comparison observes only accepted signed snapshots and accepted page-chain boundaries from the current operation. It compares enabled registries only when their configured `mirror_group` and `log_size` match. Status reports agree, diverged, or not-compared without writing during reads.

## Conformance gap ratchet

| Count | Before | After |
|---|---:|---:|
| All `conformance-gaps.tsv` data rows | 12 | 10 |
| Rows owned by STORY-260910-6bo7ej or TASK-260910-2vnjej | 2 | 0 |

Removed after their production config consumers passed: `manager-config-v2/schema-cases/valid-registry-bootstrap-members.json` and `manager-config-v2/vectors/schema2-registry-bootstrap-members`. Added pinned count rows for 10 bootstrap cases and 5 mirror cases, plus production-root/platform coverage rows. CI ledger consistency reported 478 rows checked across linux, darwin, and windows.

## Mutant evidence

Each narrowing/removal mutant produced the listed real non-zero exit; each was restored before the final passing suites.

| Mutant | Production test | Exit |
|---|---|---:|
| Require three roots before comparing instead of two | `TestRegistryMirrorComparisonVectorsDriveInstall` | 1 |
| Bypass signed-checkpoint verification | `TestRegistryInvalidFirstUseCheckpointFailsClosed` | 1 |
| Suppress first-use TOFU warning | `TestRegistryFirstUseReportsAndPersistsTOFUPosture` | 1 |
| Disable rebootstrap regression refusal | `TestRegistryRebootstrapRegressionKeepsExistingHighWater` | 1 |

## Local verification

| Command | Exit/result |
|---|---|
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -run '^(TestManagerConfigV2SchemaCases|TestManagerConfigV2Vectors)$' -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/install -run '^(TestRegistryBootstrapVectorsDriveInstall|TestRegistryMirrorComparisonVectorsDriveInstall)$' -count=1` | 0 |
| `go test ./internal/registry -count=1` | 0 |
| `go test ./cmd/curator -run '^TestEnvStatusRegistryBoundaryPostureAndCheck$' -count=1` | 0 |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` | 0 |
| `go test ./internal/pathboundary -count=1` | 0 |
| `golangci-lint run ./internal/config ./internal/install ./internal/registry ./internal/pathboundary ./cmd/curator` | 0, 0 issues |
| `go vet ./internal/config ./internal/install ./internal/registry ./internal/pathboundary ./cmd/curator` | 0 |
| `go build -o "$TMPDIR/curator-TASK-260910-2vnjej" ./cmd/curator` | 0 |
| `GOOS=windows go test -exec /usr/bin/true ./internal/pathboundary ./internal/registry ./internal/config ./internal/install ./cmd/curator -run '^$'` | 0; cross-compiles test binaries only, does not execute Windows tests |
| `bash .github/ci/ledger-consistency.sh "$TMPDIR/TASK-260910-2vnjej-ledger-final"` | 0, 478 rows checked across linux/darwin/windows |
| `git diff --check` | 0; `gofmt -l` on changed Go files returned no paths |

Earlier package-wide attempts `go test ./internal/install -count=1` and `go test ./cmd/curator -count=1` each exited 1 at the local ten-minute test ceiling. The install attempt also exposed an old endpoint-redaction assertion that rejected the registry URL even in the spec-required TOFU warning. The test now permits the public URL in warning messages while continuing to reject it in refusal errors; the focused `TestDraftEvidenceExactMatch` run passed. The bounded final production-entry tests above passed after the implementation changes. The hosted gate has not run in this session; the handoff runner runs it after the turn ends.

Development runs also recorded and resolved these failures: one vector invocation used a nonexistent, non-nested conformance path and exited 1; the corrected pinned-root invocation passed. The first bootstrap-vector run exited 1 because status treated the unchanged, already-applied checkpoint as a regression after the network high-water advanced; persisting its semantic identity fixed that case. A subsequent vector run exited 1 because the equal-version conflict fixture accidentally used equal roots; the fixture was corrected and the final 10-case and 5-case vector suites passed.

No CHANGELOG.md or LOGBOOK.md edit was made. The results resource records the relevant findings per campaign rules.

## CHANGELOG entry (for release prep)

Support signed registry bootstrap checkpoints and report first-use trust posture. Compare enabled mirror-group Merkle roots at matching log sizes, warn on advisory divergence, report strict divergence as an error, and keep installation resolution report-only. Expose bootstrap source and mirror comparison posture in read-only status.

## Windows gate rework (2026-09-28)

### Failure evidence and correction

The prior hosted run `36446388699` failed only the Windows test lane for registry checkpoint cases. I downloaded its evidence to `$TMPDIR` and read `test-evidence-windows-latest/test/go-test.json`. The failing install cases report `pathboundary.Validate` refusing the checkpoint parent because its DACL “grants mutation rights to another identity.” The checkpoint fixtures were created directly under `t.TempDir()`, whose inherited Windows ACL is broader than the manager-protected checkpoint contract. Their paths already used `filepath.Join`, and the JSON was generated by `json.Marshal`, so neither slash spelling nor source-file CRLF conversion caused this failure.

The production check was correct to refuse that directory. The tests now create a dedicated checkpoint subtree and call `pathboundary.ProtectTree` before and after writing fixtures, so Windows tests exercise the accepted protected-tree path. Registry protocol §5 requires a missing, unreadable, or malformed checkpoint to be a configuration error and leaves rebootstrap state unchanged; environments §8.4.1 supplies the absence-versus-read-failure rule. `reconcileBootstrapCheckpoint` no longer suppresses present unreadable or malformed checkpoint errors when established state exists: the caller fails closed while the previous high-water remains unchanged. A structurally valid but cryptographically invalid rebootstrap candidate remains rejected without replacing established state, as covered by the pinned `rebootstrap-signature-invalid-ignored` vector.

Added `TestReadBootstrapCheckpointAcceptsCRLFJSON` and `TestReconcileBootstrapCheckpointFailsClosedForMalformedRebootstrap`, with platform-ledger rows so all three hosted OS lanes must execute them. The malformed case verifies the established state value remains unchanged while the present checkpoint error is returned.

### Mutants

| Mutant | Test | Real exit |
|---|---|---:|
| Suppress malformed checkpoint errors whenever prior state exists | `go test ./internal/registry -run '^TestReconcileBootstrapCheckpointFailsClosedForMalformedRebootstrap$' -count=1` | 1 |
| Require at least three same-boundary views before comparing | `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/install -run '^TestRegistryMirrorComparisonVectorsDriveInstall$' -count=1` | 1 |

Both mutants were restored. The final registry and install runs below passed.

### Final local verification

| Command | Real exit/result |
|---|---|
| `go test ./internal/registry -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/install -run '^(TestRegistryBootstrapVectorsDriveInstall|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater)$' -count=1` | 0 |
| `go test ./internal/pathboundary -count=1` | 0 |
| `GOOS=windows go vet ./internal/registry ./internal/install` | 0 |
| `GOOS=windows go test -exec /usr/bin/true ./internal/pathboundary ./internal/registry ./internal/config ./internal/install -run '^$'` | 0; compiles Windows test binaries, does not execute them |
| `go vet ./internal/config ./internal/install ./internal/registry ./internal/pathboundary ./cmd/curator` | 0 |
| `golangci-lint run ./internal/config ./internal/install ./internal/registry ./internal/pathboundary ./cmd/curator` | 0, 0 issues |
| `go build -o "$TMPDIR/curator-TASK-260910-2vnjej" ./cmd/curator` | 0 |
| `bash .github/ci/ledger-consistency.sh "$TMPDIR/TASK-260910-2vnjej-ledger-final"` | 0; 480 rows checked across linux, darwin, windows (the two new platform rows increase the prior 478) |
| `git diff --check` | 0 |
| `gofmt -l internal/registry/checkpoint.go internal/registry/checkpoint_test.go internal/install/registry_e2e_test.go` | 0; no paths returned |
| `git diff --stat ac31ab6d` | 0; non-empty diff observed |

The first focused checkpoint test invocation exited 1 because its invalid-signature fixture used malformed base64 and therefore tested parse rejection instead of a well-formed bad signature. The fixture now contains a validly encoded 64-byte invalid signature; the focused checkpoint tests and full registry package reran green. One earlier 30-second install invocation returned without a captured exit code and was discarded as evidence; the install command above was rerun to completion and exited 0.

The hosted gate has not rerun after these changes. The handoff runner owns that gate and will report its result after this turn ends. Windows test execution therefore remains pending hosted evidence; only Windows vet and test-binary compilation were run locally.

## Verification update (RUN-260928-a7ec76, 2026-09-28)

The pin's full `conformance/v1` tree was extracted from curator-spec commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` under `$TMPDIR/TASK-260910-2vnjej.AKju6k/conformance/v1`.

| Rule and rc.13 clause | Production entry and coverage row | Pinned outcome |
|---|---|---|
| Checkpoint bootstrap and first-use TOFU posture; registry protocol §5 | `install.Project -> resolveRegistries`, `TestRegistryBootstrapVectorsDriveInstall`, `registry-client/bootstrap-cases` | 10/10 driven; includes signed checkpoint, tampered first response, TOFU warning, and rebootstrap regression cases |
| Optional mirror-group equivocation detection; registry protocol §5.1 | `install.Project -> resolveRegistries -> NewMirrorViewObserverWithState`, `TestRegistryMirrorComparisonVectorsDriveInstall`, `registry-client/mirror-comparison-cases` | 5/5 driven; advisory divergence warns, strict divergence is an error diagnostic, equal roots and incomparable boundaries do not diverge |
| Schema-2 registry configuration members; manager profile §1 | `TestManagerConfigV2SchemaCases` and `TestManagerConfigV2Vectors` | The Story-owned `valid-registry-bootstrap-members.json` and `schema2-registry-bootstrap-members` cases pass. The schema family reports 105 driven plus 2 gaps owned by STORY-260910-2qmrb8; the vector family reports 56/56 driven. |

The gap ledger has 12 data rows at the worktree checkpoint and 10 now; its two STORY-260910-6bo7ej rows are both removed. There were no rows owned directly by this task. The five mirror-comparison cases are newly registered in the expected-count, root-artifact, and platform-consumer tables.

### Mutant evidence

Mutant: raise the distinct-root threshold from two to three, suppressing a two-registry same-boundary divergence.

| Gate | Result |
|---|---|
| `go test ./internal/registry -count=1` with the mutant active | Exit 0; the helper/package-only suite lets the mutant survive |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/TASK-260910-2vnjej.AKju6k/conformance/v1 go test ./internal/install -run '^TestRegistryMirrorComparisonVectorsDriveInstall$/^divergence-detected-advisory$' -count=1 -v` with the same mutant | Exit 1; the production-entry vector rejects the missing `registry_view_divergence` warning |
| Restore from the saved pre-mutation file, then `cmp` against it | Both commands exit 0; production source restored byte-for-byte |
| Full `TestRegistryMirrorComparisonVectorsDriveInstall` after restoration | Exit 0; 5/5 cases driven |

### Commands run in this recovery

| Command | Real exit/result |
|---|---|
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/TASK-260910-2vnjej.AKju6k/conformance/v1 go test ./internal/install -run '^TestRegistry(MirrorComparison|Bootstrap)VectorsDriveInstall$' -count=1 -v` | 0; bootstrap 10/10 and mirror comparison 5/5 |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/TASK-260910-2vnjej.AKju6k/conformance/v1 go test ./internal/config -run '^TestManagerConfigV2(SchemaCases|Vectors)$' -count=1 -v` | 0 |
| `go test ./internal/registry -count=1` after restoration | 0 |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` | 0 |
| `make lint` | 0; golangci-lint reported 0 issues |
| `go vet ./internal/registry ./internal/install ./internal/config` | 0 |
| `go build -o $TMPDIR/TASK-260910-2vnjej.AKju6k/curator ./cmd/curator` after restoration | 0 |
| `git diff --check` | 0 |

The earlier sections retain prior bounded platform and recovery evidence. No CHANGELOG.md or LOGBOOK.md edit was made; the release-prep entry above remains the proposed text.

## Rev3 — Windows gate fix 2 (DACL instead of mode bits)

Root cause (run 36460771718): `readBootstrapCheckpoint` checked `info.Mode().Perm()&0o022`; Go reports 0666 for ordinary Windows files, so every checkpoint was refused as "permits group or other mutation".

Fix (diff vs rev2 tree 2bc75612, limited to):
- `internal/pathboundary/pathboundary.go`: new exported `CheckPrivateFile(path, info)` = the existing per-node `checkNode` (owner == effective operator + build-tagged `checkMutationPermissions`: Unix 0o022 bits unchanged; Windows owner/DACL — no non-owner ACE with mutation rights).
- `internal/registry/checkpoint.go`: mode-bit check replaced by `pathboundary.CheckPrivateFile`; still `stateread.UnusableError` (fail-closed, never absent). CRLF row and present-but-unverifiable fail-closed behaviour from rev2 unchanged.
- Fixtures already create checkpoints via `pathboundary.ProtectTree` (the product's owner-only DACL helper) — no relaxation.
- Tests: `pathboundary TestCheckPrivateFileRefusesForeignMutation` (cross-platform: protected file accepted; Everyone:GENERIC_WRITE ACE on Windows / 0777 on Unix refused); `registry TestReadBootstrapCheckpointRefusesGroupWritableFile` (unix, 0620 → refused, not absent).

Commands (zsh, `set -o pipefail`, real exit codes):
- `GOOS=windows go vet ./internal/registry ./internal/install ./internal/pathboundary` → 0
- `go vet ./internal/registry ./internal/install ./internal/pathboundary` → 0
- `go test ./internal/pathboundary ./internal/registry` → 0
- `go test ./internal/install -run 'Checkpoint|Bootstrap|Registry'` → 0
- `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` → 0
- `golangci-lint run ./internal/registry/... ./internal/pathboundary/...` → 0 (0 issues)

Mutants (GroupWritable row):
- M-file (CheckPrivateFile call → nil): row passes (exit 0) — survivor is expected: the preceding `pathboundary.Validate(parent)` tree walk already refuses the file. Stated bound: the per-file check only distinguishes under a race between the tree walk and the Lstat; defence in depth.
- M-tree (Validate(parent) → nil): row passes (exit 0) — killed-by-design: CheckPrivateFile alone refuses.
- M-both: row FAILS (exit 1) → killed. Restored: exit 0.

Windows DACL path verified only by cross-compile vet here; the hosted windows-latest lane is the arbiter (unverified locally).

## Revision 5 — re-apply on `6a7deb11dc92229c0380d2ea64cf302b86a51383`

Reapplied accepted rev4 (`refs/campaign/6bo7ej-rev4-20260929`, based on `213a53e5`) to the current Story worktree at trunk `6a7deb11dc92229c0380d2ea64cf302b86a51383`. No commit was created. `git diff --name-only origin/main -- . ':!.task-board'` reports exactly 26 paths, matching the accepted rev4 path set; no additional repository path changed.

### Per-path delta comparison and conflict resolutions

Compared `git diff 213a53e5 refs/campaign/6bo7ej-rev4-20260929 -- P` with `git diff origin/main -- P` for every path. Added and removed lines match on 24 paths. The two intentional resolutions are described below.

| Path | Comparison |
|---|---|
| `.github/ci/conformance-case-counts.tsv` | Same added and removed lines |
| `.github/ci/conformance-gaps.tsv` | Same added and removed lines; the only current diff is the two rev4 removals shown below |
| `.github/ci/platform-cases.tsv` | Same added and removed lines |
| `.github/ci/root-artifacts.tsv` | Same added and removed lines |
| `cmd/curator/env_test.go` | Same added and removed lines |
| `cmd/curator/envstatus.go` | Same added and removed lines |
| `internal/config/config.go` | Same added and removed lines |
| `internal/config/config_test.go` | Same added and removed lines |
| `internal/install/draftevidence_test.go` | Same added and removed lines |
| `internal/install/install.go` | Conflict resolution: uses the combined detailed snapshot and mirror observer result; see below |
| `internal/install/registry_e2e_test.go` | Same added and removed lines |
| `internal/pathboundary/open_read_nofollow_unix.go` | Same added and removed lines |
| `internal/pathboundary/open_read_nofollow_unsupported.go` | Same added and removed lines |
| `internal/pathboundary/open_read_nofollow_windows.go` | Same added and removed lines |
| `internal/pathboundary/pathboundary.go` | Same added and removed lines |
| `internal/pathboundary/pathboundary_test.go` | Same added and removed lines |
| `internal/registry/boundary.go` | Same added and removed lines |
| `internal/registry/boundary_test.go` | Same added and removed lines |
| `internal/registry/checkpoint.go` | Same added and removed lines |
| `internal/registry/checkpoint_test.go` | Same added and removed lines |
| `internal/registry/checkpoint_unix_test.go` | Same added and removed lines |
| `internal/registry/http.go` | Same added and removed lines |
| `internal/registry/mirror.go` | Same added and removed lines |
| `internal/registry/registry.go` | Same added and removed lines |
| `internal/registry/registry_test.go` | Same added and removed lines |
| `internal/registry/snapshot.go` | Conflict resolution: retains trunk's structured `SnapshotCheck` and combines it with the accepted mirror observer; see below |

Conflict resolutions:

- `.github/ci/conformance-gaps.tsv`: retained trunk's current rows exactly, except the two bootstrap-member gap rows removed by accepted rev4. Did not restore the posture rows that trunk had already removed. `diff <(git show origin/main:.github/ci/conformance-gaps.tsv) .github/ci/conformance-gaps.tsv` output:
  ```text
  14,15d13
  < manager-config-v2/schema-cases	valid-registry-bootstrap-members.json	STORY-260910-6bo7ej	Load rejects the valid registry bootstrap members case published by S2.
  < manager-config-v2/vectors	schema2-registry-bootstrap-members	STORY-260910-6bo7ej	Parse rejects the S2 bootstrap_checkpoint/mirror_group registry members.
  ```
  There were 18 data rows before and 16 after; rows owned by STORY-260910-6bo7ej or TASK-260910-2vnjej went from 2 to 0.
- `internal/install/install.go`: calls the new detailed observer APIs and copies `Unreachable`, snapshot warnings, mirror diagnostics, and final observer findings into `registryResolution`. This preserves trunk's structured outage reporting together with rev4's mirror comparison.
- `internal/registry/snapshot.go`: the internal check continues returning trunk's `SnapshotCheck`; detailed observer wrappers preserve the mirror observer and its warnings without dropping `Unreachable`. Existing tuple-return observer wrappers remain available.

### Revision 5 verification

Commands were run as standalone zsh processes with captured real exit codes.

| Command | Result |
|---|---|
| `go test ./internal/registry` | Exit 0 (`21.422s`) |
| `go test ./internal/install -run 'Registry|Snapshot|Root|Mirror|Checkpoint|Bootstrap'` | Exit 0 (`218.758s`) |
| `go test ./cmd/curator -run 'Registry|Env'` | Exit 0 (`449.232s`) |
| `GOOS=windows go vet ./internal/registry ./internal/install` | Exit 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/install -run '^TestRegistry(Bootstrap|MirrorComparison)VectorsDriveInstall$'` | Exit 0 (`96.551s`); 10 bootstrap and 5 mirror vectors driven through install |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$'` | Exit 0 (`36.739s`) |
| `golangci-lint run ./internal/config ./internal/install ./internal/registry ./internal/pathboundary ./cmd/curator` | Exit 0; 0 issues |
| `go build -o "$TMPDIR/TASK-260910-2vnjej-curator" ./cmd/curator` | Exit 0; binary stayed in `$TMPDIR` |
| `git diff --check` and `git diff --cached --check` | Both exit 0 |

The prior accepted rev4 review evidence remains the mutant evidence for this reapply: the named regression test is `TestRegistryMirrorComparisonVectorsDriveInstall`; the narrowing mutant raises the minimum comparison group from two roots to three, and the `divergence-detected-advisory` case kills it with exit 1. This mutant was not rerun in rev5 because the mirror implementation was unchanged; the regression suite was rerun against the re-applied tree and passed above. The accepted review verdict also records the checkpoint high-water, unverifiable-checkpoint, and DACL mutants.

The local curator-spec checkout resolves `23435129` to `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`, matching the pin. `task-board q` confirmed the task checklist is fully checked. No CHANGELOG.md or LOGBOOK.md file was changed. Hosted gate results remain pending the handoff runner.

## Revision 6 — carry-forward onto cea992e2 (no content change; rev5 failed only the Naming gate on a trunk board resource)

Shell: zsh. No file changed in this run.
- `git rev-parse origin/main HEAD` → cea992e28d5fd8277cc79d40a7e3e8415cf5596d (both).
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` → 26 (exit 0).
- `git diff | grep -c '^+<<<<<<<'` → 0; `git diff --cached | grep -c '^+<<<<<<<'` → 0.
- No tests rerun in this carry (content identical to rev5, same +/- multiset verified by the orchestrator); hosted gate is the arbiter.

## Revision 7 — re-apply on f30c2b34 (pathboundary_test.go merged with uyak0e)

Shell: zsh. Patch: `git diff cea992e2 refs/campaign/6bo7ej-rev6-20260929 -- . ':!.task-board'` applied with `git apply --3way` (rc=1: one conflict, `internal/pathboundary/pathboundary_test.go`; pathboundary.go applied cleanly).
Resolution: both sides kept — trunk's five uyak0e tests (EntryRemovedBetweenReadDirAndLstat, EntryReplacedBySymlink…, FailsClosedForOtherEntryLstatErrors, SkipsEntryVanishedDuringOwnerProbe, FailsClosedForOtherOwnerProbeErrors) intact, followed by rev6's TestCheckPrivateFileRefusesForeignMutation. Nothing else added. New files are intent-to-add (`git add -N`), otherwise uncommitted.

Verification:
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` → 26.
- Sorted +/- line multiset `git diff cea992e2 rev6 -- P` vs `git diff origin/main -- P`: 0 differences on all 26 paths, including both pathboundary files (the uyak0e hunks are in the base, so rev6's delta is reproduced exactly). pathboundary.go keeps uyak0e's `vanished` skip (lines 99–153) and rev6's `OpenReadNoFollow` (line 208).
- `go test -v ./internal/pathboundary` runs all 14: TestValidateTrustedDirectoryTree, TestValidateRootIgnoresIndependentChildRoots, TestValidateRejectsGroupOrWorldWritableComponents, TestValidateRejectsInternalSymlink, TestValidateRejectsEscapingSymlinkAsContainmentFailure, TestOpenReadNoFollowRejectsFinalSymlink (rev6), TestValidateRejectsMissingDirectory, TestValidateRejectsInjectedForeignOwner, 5× uyak0e (above), TestCheckPrivateFileRefusesForeignMutation (rev6) — all PASS.

Focused runs (real exit codes):
- `go test -count=1 ./internal/pathboundary ./internal/registry` → rc=0
- `go test -count=1 ./internal/install -run 'Registry|Snapshot|Root|Mirror|Checkpoint|Bootstrap'` → rc=0 (150 s)
- `GOOS=windows go vet ./internal/pathboundary ./internal/registry` → rc=0
Hosted gate is the arbiter. No CHANGELOG/LOGBOOK edit.

## Revision 8 — rev7 re-applied unchanged on bdb77413 (origin/main)
- `git apply` of `git diff f30c2b34 refs/campaign/6bo7ej-rev7-20260929 -- . ':!.task-board'`: exit 0 (clean, no content change). New files marked intent-to-add (`git add -N`) so they appear in `git diff`.
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` = 26
- sorted +/- line diff rev7 vs worktree: empty (diff exit 0)
- No CHANGELOG/LOGBOOK edit. Tests not rerun (identity re-apply; hosted gate is the arbiter).
