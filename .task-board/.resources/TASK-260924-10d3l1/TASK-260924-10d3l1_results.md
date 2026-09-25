# TASK-260924-10d3l1 results

## Changes

- Extended `TestDraftEvidenceExactMatch` with signed bad records that differ
  only in `source_identity` or only in `commit`. The positive exact-evidence
  install and existing wrong-name/context refusals remain covered.
- Each refusal now runs after a successful install and checks that the source
  lock, source bindings, marker, installed `SKILL.md`, and installed reference
  remain byte-identical. It also checks that the failed `install.Project`
  result carries no attestation and exposes neither the registry URL nor its
  pinned key in errors or messages.
- The test opts into the lane with `Options.DraftSourcesV1`, matching existing
  draft tests. Production behavior and `registry.ResolveExact` are unchanged.
- Updated `CHANGELOG.md`.

The production path exercised is `install.Project` -> `projectAttempt` ->
`resolveRegistries` -> `registry.ResolveExact`. This closes the matrix's B1
gap: the isolated repository and commit mismatches were previously helper-test
and reviewer-probe evidence only.

## Narrowing mutants

Each mutation was applied in a disposable copy under
`/tmp/TASK-260924-10d3l1-mutants.U5wVP2` and tested with
`go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1`.

| Comparison narrowed to | Row that exposed it | Outcome |
| --- | --- | --- |
| `record.SourceIdentity != ""` instead of exact equality | `wrong-repository-only` installed successfully under the mutant | Killed; expected-red exit code 1 |
| `record.Commit != ""` instead of exact equality | `wrong-commit-only` installed successfully under the mutant | Killed; expected-red exit code 1 |

Both of 2 comparison mutants were killed by the production-entry test.

## Validation

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/install -run 'TestDraftEvidenceExactMatch\|Registry'` (initial run) | 1 | Compile error: unused `project` return; fixed. |
| Same bounded command (second run) | 1 | Snapshot expected `agent-skill.json` in the installed tree, but the fixture does not install it; removed that invalid snapshot path. |
| `go test ./internal/install -run 'TestDraftEvidenceExactMatch\|Registry'` (final run) | 0 | Passed; package reported `ok` in 138.409s. |
| Repository comparison mutant: `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1` | 1 | Expected red: `wrong-repository-only` installed. |
| Commit comparison mutant: `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1` | 1 | Expected red: `wrong-commit-only` installed. |
| `golangci-lint run ./internal/install` | 0 | 0 issues. |
| `gofmt -d internal/install/draftevidence_test.go` | 0 | No formatting diff. |
| `git diff --check` | 0 | Clean. |

The task's landing gate is hosted CI at handoff; the full landing suite was not
run locally.

## Revision 2 (carry-forward republish)

Trunk moved to `a48f584c`; the accepted revision 1 delta was carried
uncommitted by `worktree converge` and republished here with no content
change except the orchestrator-directed CHANGELOG revert (see below).

### Per-path verification against TASK-260924-10d3l1_change-request_rev1.patch

The rev1 patch carries 2 paths:

- `internal/install/draftevidence_test.go` — trunk did NOT touch this path
  (rev1 base blob `8b9f4cb0` is still the HEAD blob). `git hash-object`
  of the worktree file is `494aae02dfb9c66a6688e6e4d46fe062875943aa`,
  equal to the rev1 `+++` blob `494aae02`: byte-identical to revision 1.
  No conflict markers. Changed nothing.
- `CHANGELOG.md` — intersecting path (trunk added the R5 script-worker-v1
  entry; rev1 base `17773d62` vs trunk base `171ec1f8`). Converge kept both
  sides with no duplication and no conflict markers: trunk R5 block, then
  this task's 3-line entry, then E4 (verified via `git diff CHANGELOG.md`
  before applying the CHANGELOG POLICY below).

### CHANGELOG entry (for release prep)

Per orchestrator policy 2026-09-24 the task's CHANGELOG hunk was reverted;
`git diff --quiet HEAD -- CHANGELOG.md` confirms the file equals trunk's.
The entry text, copied verbatim for the release-prep leaf:

- Production-entry tests for draft registry evidence that differs only in
  repository identity or commit, including prior-state preservation and
  diagnostic redaction at `install.Project`.

No stray root `TASK-*/BUG-*.md`, `test/` or `ledger/` paths exist
(`git status --porcelain --untracked-files=all` shows only
`M internal/install/draftevidence_test.go`).

### Validation (revision 2, real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/install -run 'Draft\|Evidence' -count=1` | 1 | FAIL: `panic: test timed out after 10m0s` in unrelated `TestDraftGitBuildsPublishReceipt3` (stuck in transaction staging file copy); the broad mask pulls in every `Draft*` test and exceeds the single-call bound. Not caused by this task: the delta touches only `draftevidence_test.go`, production code unchanged. |
| `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1 -v -timeout 8m` | 0 | PASS: all 5 subtests (`exact-admits`, `wrong-name-refuses`, `wrong-context-refuses`, `wrong-repository-only-refuses`, `wrong-commit-only-refuses`) in 43.759s. |
| `go vet ./internal/install` | 0 | Clean. |
| `gofmt -l internal/install/draftevidence_test.go` | 0 | No output (formatted). |
| `git diff --check` | 0 | Clean. |

Narrowing-mutant evidence from revision 1 stands (both repository and commit
comparison mutants killed, recorded above); revision 2 makes no production
or test-content change, so no new mutant run was required.

## Revision 2 re-verification (RUN-260924-5de9ad, 2026-09-24)

The prior implementer run published Revision 2 and handed off, but its runner
heartbeat expired, returning the task to `development`. This run re-verified
the carried delta without changing any content (`git status` shows only
`M internal/install/draftevidence_test.go`, left uncommitted).

- `internal/install/draftevidence_test.go`: trunk untouched (HEAD blob
  `8b9f4cb0` = rev1 `---` blob); worktree `git hash-object`
  `494aae02dfb9c66a6688e6e4d46fe062875943aa` = rev1 `+++` blob —
  byte-identical to revision 1. No conflict markers (`grep` exit 1).
- `CHANGELOG.md`: `git diff HEAD -- CHANGELOG.md` empty — file equals trunk
  `a48f584c` (task hunk reverted per policy; entry text preserved verbatim in
  the "CHANGELOG entry (for release prep)" section above). No stray root
  `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths.
- Reran myself, real exit codes: `go test ./internal/install -run
  '^TestDraftEvidenceExactMatch$' -count=1 -v -timeout 8m` → exit 0, all 5
  subtests pass (exact-admits, wrong-name/wrong-context/
  wrong-repository-only/wrong-commit-only-refuses) in 104.6s;
  `go vet ./internal/install` → 0; `gofmt -l` → clean; `git diff --check` → 0.
- Broad mask `go test ./internal/install -run 'Draft|Evidence' -count=1` was
  NOT rerun (needs >10m, exceeds the single-call bound); its exit-1 result
  (unrelated `TestDraftGitBuildsPublishReceipt3` timeout) is accepted from the
  already-attached Revision 2 evidence above. Narrowing-mutant evidence from
  revision 1 stands; revision 2 has no content change, so no new mutant run.

## Revision 3 (carry-forward republish)

Trunk moved `a48f584c` → `ab34556e`; `worktree refresh-candidate` advanced
(`TrunkOID`/`BranchOID` `ab34556e`, `refresh_advanced`). The rev2 patch
(`TASK-260924-10d3l1_change-request_rev2.patch`) carries 1 path. No content
change was made in this revision; the worktree file was never edited.

### Per-path verification against the rev2 patch

- `internal/install/draftevidence_test.go` — INTERSECTING path. Trunk touched
  it (`8b9f4cb0`→`f2986c47`, commit `5dddbb57`
  STORY-260924-3eywt2 skillfile-schema-2-enabled-by-default: removed
  `DraftSourcesV1: true` at the 2 pre-existing call sites). Worktree
  `git hash-object` = `494aae02dfb9c66a6688e6e4d46fe062875943aa`, equal to
  the rev2 `+++` blob: byte-identical to revision 2. No conflict markers
  (`grep` exit 1). The two sides contradict on the flag lines (trunk deletes
  what rev2 sets), so a both-sides union is impossible there; the rev2 side
  is kept, for the evidenced reason below. All rev2 additions (both
  single-field mismatch rows, prior-state preservation, attestation/redaction
  checks) are present; nothing dropped or duplicated.

### Why trunk's 2-line hunk is not adopted in this worktree

A disposable-copy probe (`/tmp`, worktree production unchanged) dropped the
flag at both sites and ran the task-scope test: exit 1, all 5 subtests FAIL
— worktree production still gates schema-2 fixtures on the opt-in
(`Errors:[schema_version: unsupported Skillfile schema_version 2 ...]`).
Every sibling draft test and the worktree production still use
`DraftSourcesV1`. Adopting trunk's hunk in this file alone breaks the test
in-worktree and orphans it from the worktree it lands with. The flag removal
must ride with the production convergence (sibling/orchestrator scope
following `5dddbb57`); story integrator note: drop `DraftSourcesV1` across
the draft tests when the worktree production converges with trunk's
schema-2-default rewrite.

### CHANGELOG policy

Rev2 has no CHANGELOG hunk (1-path patch). `git diff HEAD -- CHANGELOG.md`
is empty: the file equals trunk `ab34556e`; nothing to revert. The entry
text stays verbatim under "CHANGELOG entry (for release prep)" above.

No stray root `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths; no markers.

### Validation (revision 3, real exit codes, run in this turn)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/install -run 'Draft\|Evidence' -count=1 -timeout 4m` | 1 | Expected-red: 4m timeout panic across the unrelated long draft tests (240.5s FAIL); same shape as the rev2 broad-mask evidence. |
| `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1 -v -timeout 8m` | 0 | PASS: all 5 subtests (`exact-admits`, `wrong-name/wrong-context/wrong-repository-only/wrong-commit-only-refuses`) in 15.2s. |
| `go vet ./internal/install` | 0 | Clean. |
| `gofmt -l internal/install/draftevidence_test.go` | 0 | No output (formatted). |
| `git diff --check` | 0 | Clean. |

Narrowing-mutant evidence from revision 1 stands (both repository and commit
comparison mutants killed); revision 3 makes no production or test-content
change, so no new mutant run was required. The disposable-copy probe above
is a convergence check, not a gate-narrowing mutant.

## Revision 4 — stale-candidate repair (2026-09-25, RUN-260925-6d2ed2)

Answers review verdict rev3 F1 (CHANGES_REQUESTED): the rev3 candidate tree
was the old `a48f584c` snapshot plus the test change, reverting 60 trunk
files against base `ab34556e`. Repaired per `10d3l1-repair-4.md`:

1. Saved the accepted change: rev2 patch
   (`TASK-260924-10d3l1_change-request_rev2.patch`, 1 path) copied to
   `$TMPDIR/10d3l1-rev2.patch`. Working-file diff vs `a48f584c` patch-id
   `c7ce917d…` confirmed it is the accepted content (dirty state preserved
   under `refs/campaign/10d3l1-delta-20260925`; not deleted).
2. `git restore --source=HEAD --staged --worktree -- .` — tracked tree now
   equals trunk `ab34556e` (verified: `git diff --name-only HEAD` empty).
3. `git clean -n` listed exactly one untracked file,
   `internal/envprofile/draft_capture_test.go` (Sep-24 leftover, not on
   trunk, not added by the rev2 patch, belongs to no path in this task).
   Backed up to `$TMPDIR/draft_capture_test.go.bak`, then removed.
   `.task-board` untouched.
4. `git apply --3way` of the rev2 patch: applied with ONE conflict hunk
   (the `Options{…}` return line). Resolved keeping BOTH sides — trunk's
   Options shape plus the accepted test plumbing:
   - `installOpts := Options{Platform: installPlatform()}` (trunk `m28s6b`
     removed the `DraftSourcesV1` field from `Options` entirely, so the
     rev2 `DraftSourcesV1: true` literal cannot compile on trunk);
   - `return Project(cfg, project, "test", Options{Platform:
     installPlatform()}), stub, priorState` (trunk call shape + rev2
     stub/prior-state tuple).
   - Both single-field mismatch rows (`wrong-repository-only` on
     `source_identity`, `wrong-commit-only` on `commit`), the seed-install
     prior-state snapshot, and the attestation/redaction checks are intact.
5. Result: `git diff --name-only HEAD` ==
   `internal/install/draftevidence_test.go` only. Per-file patch-id
   `65e72d75…` vs accepted `c7ce917d…`: differs SOLELY by the trunk merge
   above (flag removal at 2 call sites); every accepted addition present,
   no conflict markers (`grep` exit 1), `gofmt` clean.
6. No CHANGELOG edit (`git diff HEAD -- CHANGELOG.md` empty).

### Revision 3's flag-keeping rationale is obsolete

Rev3 kept `DraftSourcesV1: true` because a disposable-copy probe showed the
flag-free test failing — but that probe ran against STALE-tree production
(`a48f584c`, which still gated schema-2 fixtures on the opt-in). On trunk
production the flag is gone and schema-2 is default: the merged flag-free
test passes the focused gate below, proving convergence. The "drop the flag
across draft tests when production converges" condition rev3 named is now
satisfied for this file.

### Validation (revision 4, real exit codes, run in this turn)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1 -v` | 0 | PASS: all 5 subtests (`exact-admits`, `wrong-name/wrong-context/wrong-repository-only/wrong-commit-only-refuses`) in 32.1s on trunk production. |
| `go vet ./internal/install` | 0 | Clean. |
| `gofmt -l internal/install/` | 0 | No output (formatted). |
| `git diff --check` | 0 | Clean. |
| `git status --short --untracked-files=all` | 0 | Only `M internal/install/draftevidence_test.go`; no stray paths. |

Narrowing-mutant evidence from revision 1 stands (both repository and
commit comparison mutants killed at the production entry
`install.Project`); revision 4 changes no comparison and no refusal
semantics — only the trunk-mandated Options shape — so no new mutant run
was required. The broad-mask `Draft|Evidence` suite was NOT rerun (exceeds
the single-call bound; prior rev2/rev3 expected-red evidence stands).
