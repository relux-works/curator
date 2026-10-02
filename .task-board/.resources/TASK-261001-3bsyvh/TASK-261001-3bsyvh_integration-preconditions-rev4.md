# TASK-261001-3bsyvh — integration preconditions (fresh, RUN-261002-bbf5a2, rev4)

Bound integration run for accepted CR-TASK-261001-3bsyvh-4 rev4 (role developer, archetype implementer).
Board left at `integrating`; no status change, no handoff, no `worktree checkpoint` / `worktree integrate`
invoked — the runner lands synchronously. No repository file changed in this run.

## Verdict

- Carrier identity vs accepted TASK-260728-rjxrgs rev1: EXACT (re-proven fresh below).
- Review: ACCEPTED (TASK-261001-3bsyvh_review-verdict-rev4.md already attached; gate green per rev4-validation.log).
- Worktree: clean except the 17-path staged carrier; index tree 289196d5 matches the accepted candidate tree.
- Trunk freshness: LOCAL CACHE shows origin/main == base 67d83539 (0 ahead/behind, zero non-board diff),
  so no cached movement since the rev4 base. HOWEVER the protected authority is unreachable
  (SSH permission denied), so the true remote HEAD is indeterminate and no row can be classified:
  landing needs the authorized remote restored, then the runner's own freshness check.

## Fresh identity proofs (all rerun in this run, exit 0)

Base facts: HEAD = 67d83539 (rev4 base); accepted snapshot refs/campaign/rjxrgs-rev1-20261001 = fe2b5f61
(tree ac008990, base 5432c85f); rev3 snapshot refs/campaign/3bsyvh-rev3-20261001 = d540fbb1
(base e87d488b); candidate tree 289196d5.

1. Staged paths: 17 (`git diff --cached --name-only`); unstaged 0; untracked 0.
2. Path lists accepted-vs-carrier (`git diff --name-only 5432c85f fe2b5f61` vs staged, non-board):
   byte-equal (`diff` exit 0, 17/17).
3. +/- line multisets (`git diff -U0`, lines matching added/removed content, `sort | uniq -c | sort`, `cmp`):
   BYTE-IDENTICAL (660 unique lines, `cmp` exit 0).
4. `git diff --cached --numstat` total: 978 insertions, 43 deletions.
5. `git write-tree` (index) = 289196d5 = the accepted rev4 candidate tree from the verdict and identity-rev4.json.
6. Blobs vs rev3: 14 of 17 byte-identical; the 3 that differ are conformance-case-counts.tsv,
   conformance-gaps.tsv, internal/install/install.go — exactly the verdict's set.
7. `git merge-tree --write-tree --merge-base=e87d488b 67d83539 d540fbb1` = 289196d5, exit 0,
   clean: all base movement and trunk-only context proven, no manual resolution content.
8. Stray checks: rjxrgs results file not in HEAD or candidate tree; no CHANGELOG/LOGBOOK,
   .task-board, or results paths in the staged diff (grep exit 1, no match);
   `git diff --cached --check` exit 0, no output. Added lines are multiset-identical to the
   accepted delta, so nothing the accepted delta lacked was introduced.

## TSV union (fresh)

- case-counts: 0 removed, 4 added rows on digest be11bb1e (mixed_build_cases 6, path_shim_cases 3,
  signing_cases 4, transaction_cases 4); no duplicate (digest,family) key in the staged file.
- gaps: +5 marker/install-marker-v3 rows, 0 removed. root-artifacts: +1 internal/install row;
  internal/marker row extended with schema-cases/install-marker-v3. All equal to the rjxrgs additions
  (covered by the whole-delta multiset identity above).
- Independent recompute from fixture vectors/external-repository-lifecycle.json: mixed 6, path_shim 3,
  signing 4, transaction 4 = TSV rows (python3 exit 0).

## Fresh validation (standalone processes, real exit codes)

- `go build ./...`: exit 0.
- `go test ./internal/conformancecoverage -count=1`: exit 0, ok 0.481s.
- With CURATOR_CONFORMANCE_ROOT set to the fixture conformance/v1 tree:
  `go test ./internal/install -run 'TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal' -count=1`:
  exit 0, ok 43.842s (all 4 PASS).
- `go vet` on the 5 changed packages (buildrepo, install, marker, scopes, skillspec): exit 0.
- `gofmt -l` over the 14 changed Go files: exit 0, no output.
- `git diff --cached --check -- . ':!.task-board'`: exit 0, no output.

## Trunk-authority detail (fresh probe, read-only)

`task-board worktree integrating` (exit 0) classifies our row as REV 4 / indeterminate / DELTA present,
with every row unclassifiable because the protected authority is unavailable (canonical remote over
ssh github, SSH permission denied; remedy: restore the unique authorized remote and retry; do not
substitute local or cached authority). The local cached origin/main equals our base 67d83539
(`git rev-list --left-right --count HEAD...origin/main` = 0/0; non-board diff empty), so the cache
shows no intervening trunk change — but that cache is unverifiable while the remote is unreachable.
The runner must revalidate against real authority at landing time.

## Accepted from already-attached evidence (not rerun)

Per the rev4 review verdict, identity-rev4.json, and the CR rev4 green gate (remote gate success,
all lanes): rev3-to-rev4 per-path line-delta equality with the trunk e87d488b-to-67d83539 delta;
trunk (Muse and other) TSV rows untouched; scoped receipt/marker/GC/skillspec suites, ledger
consistency, narrowing-mutant exit 1 with byte-identical restoration, and the full gate matrix.
Full-repo, race, lint, and hosted-platform runs were not rerun.

## Directives / stalls

- `task-board spawn directives` at checkpoint: none recorded.
- No stalled or terminated calls in this run; every command above completed with its stated exit.
