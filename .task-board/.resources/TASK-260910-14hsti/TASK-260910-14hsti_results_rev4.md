# TASK-260910-14hsti — Windows gate fix handoff (rev4)

Run: rework after rev3 remote-gate failure (run 35086213047).
Prior evidence: TASK-260910-14hsti_results.md (rev1),
TASK-260910-14hsti_results_rev2.md (rev2),
TASK-260910-14hsti_results_rev3.md (rev3, rework-2 wiring + identity),
rev1/rev2/rev3 CR patches and logs, reviewer verdict rev2.

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2`.
- Uncommitted candidate; the handoff snapshot publishes the authoritative tree.
- Delta vs rev3: `internal/staging/boundaries.go` (+pinIdentity, 2 call
  sites, doc), `internal/staging/boundaries_test.go` (two identity tests
  restructured to single-Recheck). No other files touched this round.
- Spec checkout: curator-spec main `871d11b`
  (protocol/skillfile-sources.md sections 2 and 5, revision-1 semantics).
  Diagnostics keep spec section 5 names. Frozen v1 wire schemas untouched.

## Root cause of the rev3 Windows failure

Remote gate run 35086213047: every lane green except Test
(windows-latest), exactly one failure —
`internal/install TestRunCommitRecheckRefusesSwappedParent`:
`commit_boundaries_test.go:106: runCommit err = <nil>, want
source_output_overlap` (from the `test-evidence-windows-latest`
artifact, `go-test-served.json`).

Cause: Go's `os.SameFile` on Windows resolves the file index LAZILY
(`os/types_windows.go`: `loadFileId` opens the stored path on first
comparison and caches vol/idx). An `os.FileInfo` stored at Snapshot time
carries no identity until first compared — so a single Recheck after a
same-spelling parent swap compares the NEW object against itself and
passes. On unix the stat buffer pins dev+ino at Stat time, which is why
only Windows failed. The staging-level counterpart
(`TestPlanRecheckRefusesSameSpellingParentReplacement`) passed on
Windows only incidentally: it Rechecked once BEFORE the swap, warming
the cache with the old IDs. Same latent mask in
`TestPlanRecheckRefusesChangedAdmittedIdentity`. The production path
(plan once, Recheck once at publication) never warms — hence the gate
failure.

Fix: `Snapshot` now pins identity eagerly via `pinIdentity`
(`os.SameFile(info, info)` at record time, fail-closed on false) for
every admitted input and every recorded destination ancestor. No-op on
unix; on Windows it caches the planning-time file index so a later
`SameFile(recorded, current)` compares old vs new. `Within` and the
snapshot-package identity comparison are single-time (both stats fresh)
and unaffected by laziness — verified by inspection, unchanged.

Tests: both staging identity tests now take TWO snapshots — one for the
unchanged positive check, one FRESH record compared exactly once after
the swap, mirroring the production publisher. Without the fix these fail
on Windows (lazy self-comparison passes); with it they refuse via the
identity gate. The install `TestRunCommitRecheckRefusesSwappedParent`
(production entry point, single Recheck) is unchanged and is the third
regression.

## Direct validation (bash, standalone commands, real exit codes)

All commands run by this producer on the final candidate.

| Command | Exit | Evidence |
|---|---|---|
| `go test -count=1 ./internal/staging/` | 0 | ok, 0.44s post-restore |
| `go test -count=1 ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | all three ok (2.62s / 0.52s / 1.28s) |
| `go test -count=1 ./internal/install/ -run TestRunCommit` | 0 | 4/4 new publisher tests pass |
| `go vet` on staging/snapshot/adapters/privatedir/install | 0 | clean |
| `gofmt -l` on the five packages | 0 | empty output |
| `GOOS=windows go build` on the five packages | 0 | cross-compiles |
| `GOOS=windows go vet` on staging/install | 0 | clean |
| `GOOS=windows go test -c` staging + install | 0 | both Windows test binaries compile (removed after) |
| `golangci-lint run ./internal/staging/ ./internal/install/` | 0 | `0 issues.` |

No new skips added (`grep t.Skip` unchanged from rev3 audit). The full
landing suite was not run manually; the handoff owns the single
remote-gate execution. No installs, daemon restarts, tags, releases,
LOGBOOK.md edits, runtime-home changes, live credential export, or ax
calls.

## Measured negative evidence (this run)

Bytes snapshotted with sha256 before mutation; each mutant reverted and
`sha256sum -c` verified OK after. Post-restore suite rerun green.

| Narrowing mutation | Killer command | Real exit/result |
|---|---|---|
| staging `recheckIdentity`: `!os.SameFile(...)` narrowed with `&& recorded.IsDir() != current.IsDir()` | `go test -count=1 ./internal/staging/ -run 'TestPlanRecheckRefusesSameSpellingParentReplacement\|TestPlanRecheckRefusesChangedAdmittedIdentity'` | 1, KILLED — both FAIL with `err = <nil>`; restructured single-Recheck tests bite |
| staging unrecorded-admitted gate narrowed with `&& len(inputs) > 1` | `go test -count=1 ./internal/staging/ -run TestPlanRecheckRefusesUnrecordedAdmitted` | 1, KILLED — FAIL as designed |
| `pinIdentity` neutered to `return nil` | `go test -count=1 ./internal/staging/` | 0, SURVIVED locally — warming is unobservable on unix. Bound: this mutant kills on the Windows lane only; the rev3 gate log is the pre-fix kill evidence (unpinned code failed `TestRunCommitRecheckRefusesSwappedParent` on windows-latest). Post-fix Windows green is the kill. |

2/3 killed locally, 1 platform-bound survivor with stated killer.

## Bounds and review needs

- The Windows laziness fix is proven locally by construction (Go runtime
  source: `loadFileId` caches on first `SameFile`) and on the hosted
  Windows lane by the gate; no local lane executes Windows file-ID
  semantics. Independent review must confirm the windows-latest Test lane
  is green on the rev4 gate run.
- Residual parity bound (unchanged, both platforms): file-index/inode
  reuse after delete+recreate could self-compare equal; narrowed by the
  serialized transaction and the short plan-to-publish window, same as
  the upstream design. The enforced threat is same-spelling replacement
  of a live parent, which is covered.
- All rev3 bounds carry over: byte capture/hashing/race detection/store
  publication belong to TASK-260910-16k7xy; schema-2 planning populating
  `scopeTargets.boundaries` is future install-integration-leaf work;
  mid-commit violations stay the engine preimage/rollback domain;
  NFD-alias blind spot on `staging.Within` documented in code.
- Independent review must verify the exact published CR tree (base
  `12f1287`, uncommitted delta as listed in rev3 results plus this
  round's two staging files), rerun the narrow suites plus the remote
  gate, and re-attack the gates.
