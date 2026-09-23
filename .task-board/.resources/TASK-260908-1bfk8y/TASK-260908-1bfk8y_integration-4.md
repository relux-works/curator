# TASK-260908-1bfk8y integration preconditions — fresh evidence (RUN-260923-b89626)

Board status at check: `integrating`
(`task-board --no-update-check q 'get(TASK-260908-1bfk8y) { id status }'` → `{"id":"TASK-260908-1bfk8y","status":"integrating"}`, exit 0).
Worktree: `.temp/STORY-260907-2bddfc/worktree`, branch `task-board/story/STORY-260907-2bddfc`.
No converge/integrate executed in this run per binding (runner performs the bound
landing synchronously). No board status writes made. No worktree files changed by
this run (narrow suite and evidence live in /tmp; this file is attached as a board resource).

## Accepted revision shape (rev 1, comment-only)

One-file change, uncommitted: `M .github/ci/platform-exclusions.tsv`, 7 insertions / 4 deletions.
`git diff --name-only` → only `.github/ci/platform-exclusions.tsv`. No CHANGELOG (comment-only, no behaviour change).

Before (HEAD):

```
# `default_excluded_on` is the fallback for a root that predates the vector.
# The committed released pin is such a root: it publishes no qualification
# vector at all, so without a fallback the default lane would lose an exclusion
# that is a property of the implementation rather than of the root.
```

After (worktree):

```
# `default_excluded_on` is the fallback for a SUPPLIED root that predates the
# vector (e.g. an older CURATOR_CONFORMANCE_ROOT passed explicitly). The
# committed released pin is NOT such a root: it publishes
# `vectors/conformance-claim-v3-qualification.json`, so on the default lane the
# vector is authoritative and this fallback stays dormant. Without it, a
# pre-vector root would lose an exclusion that is a property of the
# implementation rather than of the root.
```

## Data integrity (exit 0)

`git show HEAD:.github/ci/platform-exclusions.tsv | grep -v '^#'` vs
`grep -v '^#' .github/ci/platform-exclusions.tsv` → `diff` exit 0,
`NON-COMMENT-IDENTICAL:yes`. Single data row unchanged:

```
internal/godriver	TestProbeRejectsAnUncoveredPlatformBeforeTheWorker	linux	an uncovered inventory platform is refused before the worker starts
```

## Factual claim, incl. moved base (exit 0)

- Worktree `SPEC_PIN` at `.github/workflows/ci.yml:52` = `87a0d0060bad64ab883d007dcdf35df7485368bf`.
- Trunk `6c19e5ee` promotes `SPEC_PIN` to `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
  (control-root `git show 6c19e5ee:.github/workflows/ci.yml`); trunk
  `.github/ci/platform-exclusions.tsv` head is byte-identical to the worktree base
  stale text, so the one-file comment patch still applies.
- Both pins publish the vector (exit 0 each):
  `git -C curator-spec ls-tree -r --name-only <pin> | grep qualification`
  → `conformance/v1/vectors/conformance-claim-v3-qualification.json` for both pins.
- `CURATOR_CONFORMANCE_ROOT` = `${{ github.workspace }}/protocol-spec/conformance/v1`
  (`.github/workflows/ci.yml:215,336,441,525`), so the TSV's relative
  `vectors/conformance-claim-v3-qualification.json` resolves inside the supplied root.
- Vector content at BOTH pins: `platforms: linux=excluded (until_task TASK-260728-1skseh),
  macos=pending-downstream-native-evidence, windows=pending-downstream-native-evidence`.
  Matches the TSV prose (linux excluded under any root; the vector states the same normatively).
  Acceptance holds on the moved base; no re-review trigger from the pin promotion.

## Consumer analysis (column kept, live fallback; shell `sh`, `set -o pipefail`)

- `.github/ci/excluded-packages.sh:41`: `QV` set only when `$ROOT/$QUALIFICATION_VECTOR` exists.
- `.github/ci/excluded-packages.sh:61`: reads column three of the TSV into `defaults`.
- `.github/ci/excluded-packages.sh:65-74`: when QV present the vector decides exclusively
  (`source_of_truth="the root's own …"`, skip unless vector excludes this GOOS); only when
  no QV exists does it fall back to `default_excluded_on`
  (`source_of_truth="default_excluded_on (this root publishes no …)"`).
- `.github/ci/suite-plan.sh:89`: resolves exclusions via `excluded-packages.sh` (shared helper).
- `.github/ci/suite-plan.sh:96-100`: reports `platform qualification read from the root's own …`
  when QV present, else `the root publishes no …; the recorded default_excluded_on applies`.
- `.github/ci/ledger-consistency.sh:90`: same helper drives the per-platform exclusion set.
- `.github/ci/gate-selftest.sh:577-581`: pre-vector-root test deletes the vector and asserts
  the fallback is recorded (`default_excluded_on`) — the live path the new prose describes.
- Go search (`grep -rn "platform-exclusions" --include='*.go' .`): no matches (exit 1, no Go consumer).
Conclusion: the new prose is accurate (fallback for a SUPPLIED pre-vector root, dormant on
the default lane). Keeping the column is correct.

## Gates with real exit codes (each standalone, `set -o pipefail`, shell `sh`)

- Narrow fallback suite `/tmp/narrow-1bfk8y-run2.sh` (drives production entry point
  `excluded-packages.sh` for vector/pre-vector × linux/darwin, plus `suite-plan.sh` report
  wording both ways): 8 × `ok`, `ALL_NARROW_OK`, `NARROW_SUITE_EXIT:0`.
- `env CI_EXCLUDED_PKGS=internal/godriver CI_EXCLUDED_GOOS=linux bash .github/ci/ledger-consistency.sh /tmp/lc-shipped-1bfk8y-run2`:
  `241 rows checked`, `ledger-consistency: ok`, `LEDGER_EXIT:0`.
- `go build ./...`: `GOBUILD-EXIT:0`.
- Full `bash .github/ci/gate-selftest.sh`: NOT green in this run. It was started once as a
  standalone process; partial `ok` lines were observed, then it exceeded the headless
  single-call time bound and was terminated at 300s (sections include rustup/network-dependent
  gates unrelated to this comment-only change). Reported as not-run/timeout, not passing.
  The brief allows the narrow equivalent; the narrow suite + ledger-consistency above are this
  run's fresh evidence. A prior full-suite green exists on the board
  (`TASK-260908-1bfk8y_review-selftest-rev1.log`: `gate-selftest: 185 passed, 0 failed`).

## Finding (out of scope, no file changed)

`.github/ci/excluded-packages.sh:12-14` header still says "the committed released pin is
such a root" — the same stale claim fixed in the TSV. Left untouched per scope
("no exclusion row change"; TSV prose only) and per "change no file". Orchestrator may file
a follow-up leaf.
