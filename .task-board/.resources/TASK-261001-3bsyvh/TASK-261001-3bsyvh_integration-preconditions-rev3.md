# TASK-261001-3bsyvh — integration preconditions (fresh, RUN-261001-564a79, rev3)

Bound integration run for accepted CR-TASK-261001-3bsyvh-3 rev3 (role developer).
Board left at `integrating`; no status change, no handoff, no `worktree integrate`
invoked — the runner lands synchronously. No repository file changed in this run.

## Verdict

- Carrier identity vs accepted TASK-260728-rjxrgs rev1: EXACT (re-proven fresh below).
- Review: ACCEPTED (TASK-261001-3bsyvh_review-verdict-rev3.md already attached).
- Worktree: clean except the 17-path staged carrier; index tree 9cf8a21b matches
  the accepted candidate tree.
- Trunk freshness: LOCAL CACHE shows origin/main moved e87d488b -> b0a640b6 with
  3/17 paths overlapping, but the read-only merge probe is CLEAN (zero conflicts;
  merged tree dc9782f8; TSV union verified, no duplicate keys). HOWEVER the
  protected authority is unreachable (SSH permission denied), so the true remote
  HEAD is indeterminate and no row can be classified: landing needs the authorized
  remote restored, then the runner's own freshness check. This run resolved
  nothing per "change no file".

## Fresh identity proofs (all rerun in this run, exit 0)

Base facts: HEAD = e87d488b (rev3 base); accepted snapshot
refs/campaign/rjxrgs-rev1-20261001 = fe2b5f61, tree ac008990; candidate tree 9cf8a21b.

1. Staged paths: 17 (`git diff --cached --name-only`); unstaged 0; untracked 0.
2. Path lists accepted-vs-carrier (`git diff --name-only 5432c85f fe2b5f61` vs
   `git diff --name-only e87d488b 9cf8a21b`, non-board): byte-equal (`diff`, 17/17).
3. +/- line multisets (`git diff -U0`, `grep ^[+-]`, `sort | uniq -c | sort`,
   `cmp`): BYTE-IDENTICAL (660 unique lines).
4. `git diff --numstat` total: 978 insertions, 43 deletions.
5. `git write-tree` (index) = 9cf8a21b = the accepted rev3 candidate tree.
6. Stray checks: TASK-260728-rjxrgs_results.md not in tree; no CHANGELOG/LOGBOOK
   or .task-board paths in the diff; added lines multiset-identical to the
   accepted ones, so nothing the accepted delta lacked was introduced.

## Trunk-overlap detail (fresh probe, read-only, LOCAL-CACHE-BOUND)

`task-board worktree integrating` (exit 0) classifies our row as REV 3 /
indeterminate / DELTA present, with every row unclassifiable because the
protected authority is unavailable (canonical_remote_url ssh github, SSH
permission denied). The local cached origin/main (b0a640b6) sits ahead of our
base; that cache is unverifiable while the remote is unreachable.

Upstream e87d488b..b0a640b6 (local cache) touches ~50 non-board paths; overlap
with the carrier 17 (3 paths):

- .github/ci/conformance-case-counts.tsv — auto-merges (carrier: 4 lifecycle
  rows after status-repair-gc; upstream: new bd03456b digest block at file end).
- .github/ci/conformance-gaps.tsv — auto-merges (carrier: 5 marker/install-
  marker-v3 rows; upstream: bd03456b agent-environment-marker-v3 rows).
- internal/install/install.go — auto-merges (distinct hunks).

Probe: temp commit via `git commit-tree` (no ref updated) + `git merge-tree`
old-style: 0 `<<<<<<<` markers; `git merge-tree --write-tree
--merge-base=e87d488b b0a640b6 <probe>` = dc9782f8, exit 0, no conflicts.
Merged case-counts holds both sides' added rows (carrier mixed_build 6,
path_shim 3, signing 4, transaction 4 present exactly once) with zero duplicate
(digest,family) data keys. The runner must still revalidate against real
authority; this probe only shows the cached trunk needs no manual resolution.

## Fresh validation (standalone processes, real exit codes)

- `go build ./...`: exit 0.
- `go test ./internal/conformancecoverage -count=1`: exit 0, ok 0.348s.
- `go test ./internal/install -run
  'TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal'
  -count=1`: exit 0, ok 13.727s.
- `go vet` on the 5 changed packages (buildrepo, install, marker, scopes,
  skillspec): exit 0.
- `gofmt -l` over the 14 changed Go files: exit 0, no output.
- `git diff --check HEAD -- . ':!.task-board'`: exit 0, no output.

## Accepted from already-attached evidence (not rerun)

Per the rev3 review verdict and the CR rev3 green gate: CURATOR_CONFORMANCE_ROOT
suites (mixed-build/path-shim/signing/transaction entry tests, all PASS with
stated timings); independent TSV recompute against the rc.13 clone; tsv union
row-level verification; merge-tree identity of 16/17 paths vs the accepted
snapshot. Full-repo, race, lint, and hosted-platform runs were not rerun.

## Directives / stalls

- `task-board spawn directives` at checkpoint: none recorded.
- One `task-board worktree status` call stalled (host exec stall) and was
  terminated without conclusions; no finding depends on it.
- `task-board worktree integrating` completed (exit 0) and supplies the
  indeterminate-classification finding above.
