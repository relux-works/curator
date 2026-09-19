# TASK-260910-3eu4cy — review verdict, revision 1: CHANGES_REQUESTED

Reviewer run RUN-260919-3f7b72 (claude-opus-5), 2026-09-19. Change Request
CR-TASK-260910-3eu4cy-1, base 6f173778, candidate tree 2d3a5d87.
Shell for every command: `#!/bin/bash` drivers with `set -o pipefail`, one
log line per command with the binary's own exit code; nothing piped through
`tail` for an exit code. Evidence bundle:
`TASK-260910-3eu4cy_review-rev1-evidence.tar.gz` (probe tests, raw outputs,
driver logs, per-run summaries).

## Verdict

**CHANGES_REQUESTED → `to-dev`.** Two production-entry defects in this leaf's
own delta (F1, F2) plus one production-wiring coverage gap (F3). Everything
else in the delta — the transactional publication sweep, the write-time
boundary recheck integration, repair revalidation, the refresh rollback
hardening, and `AttestRoot` v5 — verified and holds; the rework is narrow.

## Candidate identity (exact tree, three ways)

| proof | result |
|---|---|
| Story worktree working tree (`GIT_INDEX_FILE=$(mktemp) git read-tree HEAD && git add -A . && git write-tree`) | `2d3a5d8746cc517aedcae68ed5e424fd9c619378` = CR candidate ✔ |
| Disposable clone `/tmp/3eu4cy-rev1`: `git checkout --detach 6f173778 && git apply --index <rev1 patch> && git commit` → `HEAD^{tree}` | `2d3a5d87…` ✔ (patch sha256 `310f647c…` matches the resource) |
| Hosted gate run 35413340404: `headSha` `1b7287f1…` ("remote gate snapshot of STORY-260910-1s75e1") → `git rev-parse 1b7287f1^{tree}` | `2d3a5d87…` ✔ — gate `success`: Test/Race macos+ubuntu, Test windows, Lint, Naming, Interop, Gate self-tests ×3; rose-air skipped |

Leaf delta measured from the checkpoint tree `dda173a0` (1xs0pj rev3): 13 files,
+2263/−10 — production: `cmd/curator/main.go` (+50/−10), new
`cmd/curator/draft_status.go` (166), `internal/install/install.go` (+33),
`internal/install/global.go` (+1), `internal/registry/attest.go` (+46),
`internal/sourcelock/sourcelock.go` (+11), `internal/closure/resolve.go` (+14/−4);
the rest are tests. No change in `internal/transaction`, `internal/staging`,
`internal/adapters`, `internal/runtimestore` — the producer's "extend, don't
re-implement" claim is accurate: the engine's journaled commit, per-write
`BoundaryCheck`/durable `BoundaryProof` (`internal/install/commit.go:941-942`,
`internal/transaction/engine.go:491-513`) and unmanaged-destination refusal are
trunk code, and this leaf proves their draft-lane integration at `install.Project`.

## Independent reruns (clone at the candidate tree, prebuilt binaries, `-count=1 -v`)

Build/vet/fmt on the clone: `go build ./...` exit 0; `go vet` over the six
touched packages exit 0; `gofmt -l cmd internal` empty (exit 0); test binaries
compiled exit 0 ×6 (`build.log`).

| run | mask | result |
|---|---|---|
| `internal/sourcelock` | whole package | exit 0 — 31 pass (incl. `TestRestoreReturnsExactPriorBytes`) |
| `internal/registry` | `TestAttestRoot\|TestV5AttestIdentity` | exit 0 — 5 pass (+4 subrows) |
| `internal/closure` | `TestRefreshDraft\|TestOpenDraftFrozen` | exit 0 — 9 pass, 0 skip (the read-only-directory vector ran: not root) |
| `internal/install` publish rows | the 9 `draftpublish_test.go` tests | exit 0 — 9 pass (76 s) |
| `internal/install` sweep | `TestDraftFailureAtEveryTargetClassRestoresPriorState` | exit 0 — 7/7 classes (10-context … 90-consumer) pass, 68 s |
| `internal/install` legacy | `TestLegacyInstallUntouchedWhenDraftOff\|TestMovedTagWarningAndStrict\|TestGlobalStrictTagsDetectMovedTag` | exit 0 — 3 pass |
| `cmd/curator` draft | `TestDraftStatus\|TestClassifyDraftMember` | exit 0 — 11 pass (12 classifier + 4 presence subrows, `unreadable` ran, 0 skips) |
| `cmd/curator` legacy | `TestStatus\|TestProjectResolve` | exit 0 — 34 pass (47 incl. subtests), 0 fail |

Windows/Ubuntu rows: hosted gate only (the job log carries the ledger and the
platform-case gate's "180 skips recorded", not per-test lines). The two new
POSIX vectors use reasons that are declared in `.github/ci/skip-classes.tsv`
(`host-capability`: "this process can write through a read-only directory";
"this environment can (inspect|read) a mode-000 (directory|file)"), and the
platform-case gate — which is fatal on an undeclared reason — passed on
windows-latest. No new skip class.

## Findings

### F1 (blocking) — `detectMovedTagsIn` v5 arm reports a *declared* tag bump as a moved tag; `--strict-tags` refuses it

`internal/install/install.go:1235-1250` (+ `recordedPackageCommit`, :1262). The
new arm warns whenever the marker's locked commit differs from the frozen
node's commit for any `tag`-kind declaration. A schema-5 marker carries no
`ref` (the v5 schema forbids it), so the arm cannot tell "the same tag now
points elsewhere" from "the operator changed the declaration from v1 to v2".
The legacy predicate (`:1252`) requires `recorded.Ref == node.Resolved.Ref`; the
flag's contract is `cmd/curator/main.go:557` "fail if an installed tag moved to
another commit".

Reproduction (probe `zz_review_tagbump_test.go`, production entry
`install.Project`, run on the candidate clone, `probeA.out`):

1. `setupGitInstall` → real draft install at tag `v1` (commit C1); marker v5 binds C1.
2. New commit C2, **new tag `v2`** (v1 untouched — the probe asserts v1 still resolves to C1).
3. Manifest `"tag":"v1"` → `"tag":"v2"`; `closure.RefreshDraft` (as the CLI does).
4. `install.Project(StrictTags: true)` →
   `Status:failed  Errors:[moved tag for review: v2 9026c34c… -> c59d4c08…]`
   — v2 never pointed at 9026c34c; nothing moved. `TestReviewDraftDeclaredTagBumpIsNotAMovedTag` FAIL, exit 1.
5. Legacy control on the frozen v1 lane (`TestReviewLegacyDeclaredTagBumpIsNotAMovedTag`,
   same bump under a v4 marker): `status=ok`, no moved-tag message — PASS.

Impact: on the draft lane every routine pinned-tag bump prints a false "moved
tag" advisory, and any `curator install --strict-tags` pipeline refuses the
bump. This is the "inferred from a proxy signal" shape: the commit difference is
not evidence that a tag moved. The candidate's own row
(`TestDraftMovedTagWarnsThroughTheLock`) covers only the same-tag move, so
nothing in the suite would have failed.

Viable fixes (producer's choice; both keep 1xs0pj's accepted decisions):

- **A (bound, recommended):** drop the v5 arm — a schema-5 marker never matches
  the legacy moved-tag gate; state at the call site that on the draft lane
  install never re-resolves refs, so a tag move is observable only at explicit
  refresh (this is exactly the bound the 1xs0pj rev3 verdict accepted). Turn the
  producer's test into the negative rows: declared bump → no warning, StrictTags
  ok; same-tag move → new lock, no install-time warning, marker binds the new commit.
- **B (provable detection at refresh):** in `closure.RefreshDraft` /
  `cmd/curator project_resolve.go`, when a prior lock exists and its
  `manifest_sha256` equals the new plan's (declaration unchanged), a Git member
  whose commit changed under a `tag` declaration IS a moved tag → report it from
  `project refresh`. Whether `--strict-tags` should exist there is a product
  call; do not couple it to install. (Recording the declared ref in the marker is
  not an option — the v5 closed shape forbids it.)

### F2 (blocking) — `curator status --check` exits 0 on a stale lock (zero-member lock + newly declared skill); stale-lock rows describe the wrong skill set

`cmd/curator/main.go:769-776` diverts a failed dry run with `!BuildsComplete`
into the row path when `draftLockIsStale` (`draft_status.go:62-72`) is true, and
`draftStatusDrift` (`draft_status.go:99-107`) then synthesises `needs-install`
rows **from the stale lock's member names**. Consequences at the CLI:

- Reproduction B (`TestReviewDraftStatusStaleEmptyLockIsNotCurrent`, `probeBC.out`):
  `project resolve app` on `"skills":[]` succeeds ("resolve app: 0 skills",
  lock with 0 members); edit the manifest to `include:["review"]` (never
  resolved, never installed); `status app` → **no rows, no diagnostic, exit 0**;
  `status --check app` → **exit 0**. `checkFailed({}, nil)` is false. A stale
  lock with a declared, uninstalled skill is reported current — the
  "absent evidence treated as satisfied" shape, at the production entry the
  AC names ("Status … enforce exact currentness and frozen inputs";
  spec §3 "changed manifest hashes fail with `source_lock_stale`", §4
  "checking status returns nonzero").
- Reproduction C (`TestReviewDraftStatusStaleLockRowsAndDiagnostics`): lock =
  {review}, manifest swapped to `include:["other"]` → rows say
  `review needs-install` (review is no longer declared) and never mention
  `other` (declared, not installed); stderr is empty — the dry run's
  `source_lock_stale: … explicit refresh required` refusal is swallowed
  (the `printStatusRefusal` call is skipped on this branch). `--check` is
  nonzero here only by accident of the old member list.

The row set cannot be derived from the new manifest without a collection
rescan (`include` globs), which status must not do — so a stale lock is not a
per-member verdict at all. Recommended fix: remove the diversion and let a
stale lock take the existing failed-dry-run path (`printResult`, exit
non-zero) exactly like every other `!BuildsComplete` refusal — one deletion
in `main.go`, `draftLockIsStale` and the stale branch of `draftStatusDrift`
become dead (the drift function should then treat `CheckStale != nil` as an
error, never as rows). Re-point `TestDraftStatusStaleManifestNeedsInstall` to
expect the refusal + nonzero exit, and add the zero-member row. If the
orchestrator prefers rows, the minimum is: always print the refusal, and force
`exitFail` under `--check` whenever the lock is stale regardless of member count.

### F3 (coverage, fix with the rework) — `Result.Attestations` → status wiring is unverified by any test

`internal/install/install.go:532` sets `Result.Attestations`; the only consumer is
`cmd/curator/main.go:790`. Mutant **MR7** (`result.Attestations = nil`) SURVIVES
the whole candidate suite (`cmd/curator` `TestDraftStatus|TestClassifyDraftMember`
11 pass; `internal/install` attestation rows 2 pass): no test reads the field
and the CLI fixtures have no registry, so the "changed registry/status/key →
non-current" clause is driven only at the `classifyDraftMember` seam with
hand-built inputs. Needed: an `install.Project` dry-run row asserting
`Result.Attestations` equals the injected `ResolveAttest` map (and is nil when
resolution fails), and preferably one `curator status` row through a signed
`httptest` registry (`internal/install/registry_e2e_test.go` has the fixture
shape) proving attested → up-to-date and changed status → needs-install.

### F4 (coverage, small) — two classifier gates are killed only at the helper seam

- **MR1** (drop `recorded.LockSHA256 != lockSHA256`, `draft_status.go:152`):
  KILLED by `TestClassifyDraftMemberComparesExactCurrentness/lock-mismatch`,
  SURVIVES all 9 CLI rows (`TestDraftStatusNeedsInstallAfterRefreshWithoutInstall`
  changes the package too). Probe D (`TestReviewDraftStatusLockOnlyChangeNeedsInstall`:
  widen `include` to `["review","other"]`, refresh, no install → `review
  needs-install`, `--check` nonzero) PASSES on the candidate (`probeD.out`) and
  KILLS MR1 at the CLI (`probeD-mr1.out`, exit 1). Please adopt it.
- **MR8** (legacy marker in a draft project treated as current,
  `draft_status.go:139`): KILLED by `TestClassifyDraftMemberPresenceRows/legacy-marker`,
  SURVIVES the CLI rows. Optional CLI row: write a v4 marker into
  `.agents/skills/review` of a resolved draft project → `needs-install`.

### Minor / notes (non-blocking)

- `restorePriorFile` → `sourcelock.Restore` is a robustness improvement with no
  distinguishing test (byte-equal outcome either way); fine as is.
- Draft status rows are keyed by lock members (closure), the legacy surface by
  declared roots — a deliberate, spec-consistent difference; worth one sentence
  in the results so the JSON consumers know.
- `registry.Matches` accepts a content-hash-only match and checks no `name`
  (spec §4 "exact name, canonical repository, commit and context hash"); trunk
  behaviour shared with install since v1 — bound, not this leaf.

## Narrowing mutants (mutant clone `/tmp/3eu4cy-rev1-mut`, candidate committed, `git checkout --` restores; end state `git diff --quiet HEAD` exit 0, only the two probe files untracked)

| id | mutant | killer | result |
|---|---|---|---|
| MR1 | status drops the `lock_sha256` comparison | classifier `lock-mismatch` row | KILLED (helper) exit 1 / **SURVIVED at the CLI** exit 0 → F4; probe D kills it exit 1 |
| MR2 | stale-lock branch reports `up-to-date` | `TestDraftStatusStaleManifestNeedsInstall` (CLI) | KILLED exit 1 |
| MR8 | legacy marker in a draft project reported current | presence row `legacy-marker` | KILLED (helper) exit 1 / SURVIVED at the CLI → F4 |
| MR7 | `Result.Attestations` never exposed to status | — | **SURVIVED** (cmd 11 pass, install 2 pass) → F3 |
| MR5 | v5 moved-tag arm never warns | `TestDraftMovedTagWarnsThroughTheLock` | KILLED exit 1 (the row pins the F1 behaviour; it must change with the fix) |
| MR4 | configured-git attested through its source path | `TestAttestRootV5ArmsWithoutRegistryIdentityAreUnattestable` | KILLED exit 1 |
| MR3 | engine rollback skips every `20-runtime` target | sweep: 30-shim-canonical, 50-env-file, 60-adapter-ledger, 80-removal, 90-consumer rows ("rollback did not restore the prior state: home/runtime …") | KILLED exit 1 |

Producer's M3a/b analysis (per-write `BoundaryCheck` removed, durable proof
kept → equivalent) checked against `engine.checkBoundary`
(`engine.go:491-513`): with no in-memory guard the engine re-verifies the
durable proof for the same target before the same write — a redundant layer,
so the survivor is equivalent and M3c/d's kill is the evidence. Agreed.

## Acceptance clauses — status

| clause | evidence | status |
|---|---|---|
| one recoverable transaction; faults roll back consistently | sweep 7/7 classes at `install.Project` (whole-state digest incl. lock+bindings, reverse-order rollback, no journal); refresh first-write/second-write/crash-window rows; MR3 killed | holds |
| recheck boundaries at write time | retarget + admitted-replacement rows fail `source_output_overlap` from a `PointPrepared` mutation with full rollback; producer M3c/d | holds (trunk engine, integration proven) |
| status enforces exact currentness and frozen inputs | up-to-date/needs-install/content-drift/not-installed/invalid-marker/live-mutation rows hold; **stale lock: F2**; attestation wiring: F3 | **not yet** |
| repair revalidates the locked source and evidence | tampered snapshot → `source_snapshot_changed`, lock+marker byte-identical; revoked evidence → refusal, lock+marker byte-identical; drift repaired without a lock write | holds |
| refresh replaces lock+marker only after success | refresh-then-failed-install keeps the old marker beside the new lock; stale bindings → `source_lock_stale`, no lock bytes written; retry recovers | holds |
| unmanaged files protected | foreign file at the mirror path → refusal before any write, bytes intact, no journal | holds |
| retarget / copy mutation failure paths | proven (above) | holds |
| identity consumers deferred by 1xs0pj | `AttestRoot` v5: network-git audited/revoked, local + configured-git `unattestable`, legacy untouched (MR4 killed); `scopeStatusDrift`: replaced by the draft lane for schema-2 + switch, legacy byte-identical (34 legacy rows green); `detectMovedTagsIn`: **F1** | **not yet** |
| legacy goldens with the switch off | `TestLegacyInstallUntouchedWhenDraftOff`, legacy moved-tag rows, 34 legacy status/resolve rows | holds |

## Bounds

- Whole-repository suite, race lanes and Windows/Ubuntu rows: hosted gate on
  the exact candidate tree only; rose-air skipped.
- `status --attest` rendering for v5 is covered at `AttestRoot`; no CLI attest row (as the producer states).
- Global scope sets `Result.Attestations` for symmetry; no consumer.

## Required for revision 2

1. F1: bound the v5 arm (option A) or move detection to refresh (option B); negative rows for the declared bump under `StrictTags` at `install.Project`.
2. F2: a stale lock must never yield exit 0 under `--check` and must surface its refusal; zero-member and swapped-selection CLI rows.
3. F3: `Result.Attestations` asserted at `install.Project`; a status row through a real registry fixture if feasible.
4. F4: adopt probe D (lock-only change through the CLI); optional legacy-marker CLI row.

Verdict recorded via `set_status(TASK-260910-3eu4cy, status=to-dev)`; not accepted.
