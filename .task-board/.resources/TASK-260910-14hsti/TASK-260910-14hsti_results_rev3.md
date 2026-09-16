# TASK-260910-14hsti — rework-2 handoff (rev3)

Run: RUN-260916-2cadc5 (rework after reviewer verdict
TASK-260910-14hsti_review-verdict-rev2.md, CHANGES_REQUESTED).
Prior evidence: TASK-260910-14hsti_results.md (rev1),
TASK-260910-14hsti_results_rev2.md (rev2), rev1/rev2 CR patches and logs.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`.
- Uncommitted candidate; the handoff snapshot publishes the authoritative tree.
- Tree delta vs base: `M internal/adapters/adapters.go`
  (rev1 casing-alias hardening, unchanged this round),
  `M internal/adapters/stage.go`, `M internal/install/commit.go`,
  + 9 new files (`internal/{snapshot,staging,adapters}/boundaries{,_test}.go`,
  `internal/privatedir/staging{,_test}.go`,
  `internal/install/commit_boundaries_test.go`). No other files touched.
- Spec checkout: curator-spec main `871d11b` (protocol/skillfile-sources.md
  sections 2 and 5, revision-1 semantics only). Diagnostics keep spec section 5
  names. Frozen v1 wire schemas untouched.

## What rework-2 demanded, and what changed

1. HIGH — guards not wired into production paths. Fixed by wiring every
   guard into a production entry point (details and tests below):
   - `snapshot.PrepareLocalAcquisition` (NEW): validates
     overlap/root_inputs/boundaries via `ValidateLocalPackage` BEFORE any
     traversal or staging side effect, enumerates via `EnumerateInputs`,
     reserves staging via `privatedir.TempStaging`. The capture leaf
     (TASK-260910-16k7xy, backlog, same story) consumes it; byte capture,
     hashing, race detection and store publication stay its scope.
   - `adapters.Group.Admitted` (NEW field) + gate in `stage()`:
     `StageProject`/`StageGlobal` now enforce `ValidateDestinations` on
     the planned mirrors when groups pin admitted inputs. Empty (all
     legacy callers) skips the gate with identical behavior; no signature
     changed.
   - `install.runCommit` recheck seam (NEW, 10 lines + field): when
     `scopeTargets.boundaries` is set, the real serialized publisher
     calls `plan.Recheck` after planning/validation and before the first
     publication write (`journalPlan`, which creates directories).
     Refusal returns before journaling, so there is nothing to roll
     back; legacy scopes (nil) behave byte-identically. Draft planning
     that sets the field lands with the schema-2 install integration
     leaf — this seam is the publisher side of that contract.
2. HIGH — spelling-only Snapshot admitted same-path replacement. Fixed:
   `staging.Snapshot` is now a struct recording resolved spellings PLUS
   `os.SameFile` identity for every existing destination ancestor and
   every admitted input; `Recheck` refuses changed/vanished identities.
   The reviewer's counterexample (plan parent/out, Snapshot, rename
   parent away, mkdir parent, Recheck) is committed as
   `TestPlanRecheckRefusesSameSpellingParentReplacement` and fails
   without the fix (proven by mutant 2 below).
3. Survived mutant must bite. Fixed: `TestValidateManagedSourceRejected`
   is now joined by `TestValidateManagedMissingRefusedAsOverlapBeforeTraversal`,
   which isolates the FIRST output gate — a missing selector under a
   managed root refuses as `source_output_overlap` before any
   existence-dependent traversal. Narrowing that gate to
   `directory == "."` now fails (mutant 1, killed). A companion positive
   test pins resolve-then-check semantics (managed spelling resolving to
   authored bytes admits those bytes, per spec section 1).
   Linux skip-class handling from rev2 is unchanged (skip audit below).

## Deliberate scope extension (review note)

`internal/install/commit.go` is outside the four named scope packages.
Rework-2 item 1 explicitly orders the Recheck to "run at publication in
the real transaction publisher", and the real publisher of staging plans
is `runCommit` (no in-scope publisher exists for them; inventing one
would be a bypassable side channel, and touching the frozen transaction
engine core would be worse). The diff is minimal (one nil-default field
+ one nil-guarded gate + comments), v1 behavior is bit-identical when
unset (proven by the full `internal/install` suite, exit 0), and the
gate is proven through the real publisher by committed tests. The
schema-2 planning side that populates the field is explicitly
coordinated as the install-integration leaf's work, not silently
assumed.

## Acceptance trace — every row reaches a production entry point

| Contract row | Production entry point | Committed test(s) driving it |
|---|---|---|
| Canonicalize paths/symlinks/case | `PrepareLocalAcquisition`, `StageProject`, `runCommit` | `TestValidateSymlinkManagedAndEscape`, `TestValidateCaseAliasUsesFilesystemIdentity`, `TestValidateManagedSpellingResolvingOutsideAdmitted` (acquisition); `TestStageProjectRefusesMirrorOverAdmittedInput` (planner); `TestRunCommitRecheckRefusesSwappedParent` (publisher) |
| Authored vs managed outputs | `StageProject` | `TestStageProjectRefusesMirrorOverAdmittedInput`, `TestStageProjectAdmitsDisjointAdmittedInputs` |
| root_inputs validation | `PrepareLocalAcquisition` | `TestPrepareLocalAcquisitionAdmitsRootInputs`, `TestPrepareLocalAcquisitionRefusals/root-without-inputs` |
| Reject unsafe overlap before traversal | `PrepareLocalAcquisition` | `TestPrepareLocalAcquisitionValidatesBeforeStaging` (order proof: invalid package + bogus staging parent yields the validation error), `TestPrepareLocalAcquisitionRefusals`, `TestValidateManagedMissingRefusedAsOverlapBeforeTraversal` |
| Recheck at publication | `runCommit` (real serialized publisher) | `TestRunCommitRecheckRefusesDestinationInsideAdmitted`, `TestRunCommitRecheckRefusesSwappedParent` (journal untouched: prepared=commits=ids=0, live bytes intact), `TestRunCommitRecheckPassesUnchangedBoundaries`, `TestRunCommitWithoutBoundariesSkipsRecheck` |
| Protect unmanaged files | `StageProject` | `TestUnmanagedTakeoverRefusedBesideAdmittedCheck` (rev1, still green) |
| Path `.` + safe subdirectory valid | `PrepareLocalAcquisition` | `TestPrepareLocalAcquisitionAdmitsSafeSubdirectory` |
| Same-spelling identity replacement | `Plan.Recheck` via `runCommit` + direct | `TestPlanRecheckRefusesSameSpellingParentReplacement`, `TestPlanRecheckRefusesVanishedAncestor`, `TestPlanRecheckRefusesChangedAdmittedIdentity`, `TestPlanRecheckRefusesUnrecordedAdmitted`, `TestPlanRecheckWithoutSnapshotStillEnforcesSeparation`, `TestPlanRecheckCoversTargetsAddedAfterSnapshot`, `TestPlanSnapshotRefusesMissingAdmitted` |

Production-caller coverage for the five integration surfaces is now 5/5
at a production entry point: ValidateLocalPackage + EnumerateInputs +
TempStaging via PrepareLocalAcquisition; ValidateDestinations via
StageProject/StageGlobal; Snapshot/Recheck via runCommit.

## Direct validation (bash/zsh, standalone commands, real exit codes)

All commands run by this producer on the final candidate.

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 ./internal/staging/ ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | all four green (0.46s / 2.22s / 0.61s / 1.46s), post-restore |
| `go test -count=1 ./internal/install/ -run TestRunCommit` | 0 | 4/4 new publisher tests pass |
| `go test -count=1 -timeout 9m ./internal/install/` (FULL package, unmodified) | 0 | ok, 338.8s — nil-boundaries legacy path bit-identical |
| `go vet` on staging/snapshot/adapters/privatedir/install | 0 | clean |
| `gofmt -l` on the five packages | 0 | empty output |
| `golangci-lint run` on the five packages | 0 | `0 issues.` (one G602 raised during the run on new code, fixed, relinted) |
| `git diff --check` | 0 | no whitespace errors |
| `go build ./cmd/curator` (binary removed) | 0 | CLI compiles |
| skip-reason audit (`grep t.Skip` over the five test files) | 0 | only `symlinks unavailable: %v` (x10) and the rev2 `test filesystem is case-sensitive; ...` (x2) — both `host-capability`/`allow` in skip-classes.tsv; no new skip text |

The full landing suite was not run manually; the handoff owns the single
remote-gate execution. Other platforms are unverified locally. No
installs, daemon restarts, tags, releases, LOGBOOK.md edits,
runtime-home changes, live credential export, or ax calls.

## Measured negative evidence (this run): 3/3 narrowing mutants killed

Bytes snapshotted with sha256 before mutation; each mutant reverted and
`sha256sum -c` verified OK after its kill. Post-restore suite rerun
green (row 1 above ran after restoration; the lint fix after it was
retested: lint exit 0, staging exit 0).

| Narrowing mutation | Killer command | Real exit/result |
|---|---|---|
| snapshot `ValidateLocalPackage`: first `else if pruned` narrowed to `else if pruned && directory == "."` (the reviewer's rev2 survivor) | `go test -count=1 ./internal/snapshot/ -run TestValidateManagedMissingRefusedAsOverlapBeforeTraversal` | 1, expected failure: missing-under-managed now reports `source_member_missing` instead of `source_output_overlap`. `TestValidateManagedSourceRejected` still survives via the post-resolution gate, as designed — the new test isolates the first gate. |
| staging `recheckIdentity`: `!os.SameFile(...)` narrowed with `&& recorded.IsDir() != current.IsDir()` (admits same-kind replacement) | `go test -count=1 ./internal/staging/ -run 'TestPlanRecheckRefusesSameSpellingParentReplacement\|TestPlanRecheckRefusesChangedAdmittedIdentity'` | 1, both FAIL with `err = <nil>` — the committed reviewer's-counterexample regression bites |
| adapters `stage()`: `len(admitted) > 0` narrowed to `> 1` (single-pin bypass) | `go test -count=1 ./internal/adapters/ -run TestStageProjectRefusesMirrorOverAdmittedInput` | 1, expected failure: overlap admitted |

0 survivors.

## Bounds and review needs

- Byte capture, inventory hashing, capture-race detection
  (`source_snapshot_changed`), store publication and lock issuance
  belong to TASK-260910-16k7xy (capture) and later leaves; this leaf
  proves validation-before-traversal, private staging, planning
  enforcement and publication recheck only.
- The schema-2 install flow that populates `scopeTargets.boundaries`
  (draft planning inside stageTargets) is future integration-leaf work;
  this leaf proves the publisher invokes the gate (nil = legacy skip).
- Mid-commit violations (after Prepare) remain the engine's
  preimage/rollback domain, covered by existing transaction tests, not
  re-proven here. Ancestor TOCTOU between Recheck and the swaps is a
  residual race narrowed by the serialized transaction, as in the
  upstream design.
- Same stated bounds as rev1/rev2: Unicode-normalization aliases of
  missing paths are a blind spot (documented on `staging.Within`);
  root-input "required context input" coverage is bounded to SKILL.md,
  the effective manifest file, and declared runtime/build roots;
  revision-2 policy shapes are hooks only.
- Independent review must verify the exact published CR tree (base
  `12f1287`, uncommitted delta as listed above), rerun the narrow
  suites plus the remote gate, and re-attack the gates.
