# TASK-261001-3bsyvh — integration preconditions (fresh, RUN-261001-b0aaab)

Bound integration run for accepted CR-TASK-261001-3bsyvh-1 rev1 (role developer).
Board left at `integrating`; no status change, no handoff, no `worktree integrate`
invoked — the runner lands synchronously. No repository file changed in this run.

## Verdict

- Carrier identity vs accepted TASK-260728-rjxrgs rev1: EXACT (re-proven fresh below).
- Review: ACCEPTED (TASK-261001-3bsyvh_review-verdict-rev1.md already attached).
- Worktree: clean except the 17-path staged carrier; index tree 1e1c5d83 matches the accepted candidate tree.
- Trunk freshness: STALE — origin/main moved bab2433b -> bd126a9a; 3 of 17 paths
  overlap upstream; merge probe conflicts in exactly 1 file
  (.github/ci/conformance-case-counts.tsv, both-sides-keepable TSV rows).
  The 2 overlapping Go files auto-merge cleanly. Landing needs a converge/refresh
  decision by the runner/orchestrator; this run resolved nothing per "change no file".

## Fresh identity proofs (all rerun in this run, exit 0)

Base facts: HEAD = bab2433b (carrier base); accepted snapshot
refs/campaign/rjxrgs-rev1-20261001 = fe2b5f61, tree ac008990; candidate tree 1e1c5d83.

1. Staged paths: 17 (`git diff --cached --name-only`), zero unstaged, zero untracked.
2. Path lists accepted-vs-carrier: byte-equal (`diff`, 17/17).
3. `git diff --numstat` per-path add/del: identical (978 insertions, 43 deletions total).
4. +/- line multisets (1021 vs 1021, `sort | uniq -c | sort`, `cmp`): byte-identical.
5. Per-path blobs snapshot-vs-index: 17/17 EQUAL.
6. Base movement 5432c85f -> bab2433b: 16/16 tracked paths blob-equal; the 17th
   (internal/install/external_lifecycle_conformance_test.go) is absent from both
   bases (pure add). Base movement untouched the carrier surface.
7. `git merge-tree --write-tree --merge-base=5432c85f bab2433b <snapshot>` = 1e1c5d83
   (clean, no conflicts) — reproduces the candidate tree exactly.
8. `git write-tree` (index) = 1e1c5d83; `git diff 1e1c5d83 --name-only` = 0 paths.
9. Stray checks: TASK-260728-rjxrgs_results.md not in tree; no CHANGELOG/LOGBOOK or
   .task-board paths in the diff; added lines byte-identical to the accepted ones,
   so nothing the accepted delta lacked was introduced (incl. the employer-name rule).

Note: the carrier brief's `origin/main` 17-count no longer holds because trunk moved
ahead (39 paths vs origin/main); the recorded CR base bab2433b is the correct anchor,
and all counts above use it. `git diff --name-only HEAD` = 17.

## Trunk-overlap detail (fresh probe, read-only)

Upstream bab2433b..origin/main touches 25 non-board paths; overlap with the carrier 17:

- .github/ci/conformance-case-counts.tsv — CONFLICT (same insertion point after the
  status-repair-gc row; upstream adds 4 acquisition/* rows, carrier adds 4
  lifecycle/* rows; resolution keeps all 8).
- internal/buildrepo/protected.go — auto-merges (upstream: CacheArtifactName +
  artifact-path logic; carrier: one error-string word).
- internal/install/external.go — auto-merges (upstream: 2 artifactPath call sites;
  carrier: new const+func block at a different hunk).

Probe: temp commit via `git commit-tree` (no ref updated) + `git merge-tree
--merge-base=bab2433b origin/main <temp>` reports exactly the one TSV conflict.

## Fresh validation (standalone processes, real exit codes)

- `go build ./...`: exit 0.
- `go vet` on the 5 changed packages (buildrepo, install, marker, scopes,
  skillspec): exit 0.
- `gofmt -l` over the 14 changed Go files: exit 0, no output.
- `git diff --check HEAD -- . ':!.task-board'`: exit 0.
- `go test -count=1 -timeout 3m ./internal/skillspec/`: exit 0, ok 19.424s.
- `go test -count=1 -timeout 3m ./internal/buildrepo/ -run
  TestSkillBuildDescriptorRejectsPackageOutputAndPathDestinations -v`: exit 0,
  3/3 subtests pass (output, path_entry, path_entries) — negative gate for the
  carrier's descriptor-rejection hunk.

## Accepted from already-attached evidence (not rerun)

Per TASK-261001-3bsyvh_results.md and the CR rev1 green gate: golangci-lint clean;
root-enabled buildrepo/install/marker/scopes focused suites; split install
conformance streams (mixed-build 6/6, path/shim 3/3, signing 3/4 + 1 bound,
transaction 4/4); marker v2/v3/v4 and status/repair/GC rows with stated bounds.
Full-repo, race, and hosted-platform runs were not rerun. The known prior
expected-red items (broad package timeouts, pin-verifier rc skew on files outside
the delta) remain as reported there; nothing in this run re-opens them.

## Directives / stalls

- `task-board spawn directives` at checkpoint: none recorded.
- One `task-board worktree status` call stalled (host exec stall) and was
  terminated without conclusions; no finding depends on it.
