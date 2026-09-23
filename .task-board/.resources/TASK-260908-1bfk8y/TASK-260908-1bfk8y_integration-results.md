# TASK-260908-1bfk8y integration preconditions (developer run)

Board status at check: `integrating` (via `task-board q 'get(TASK-260908-1bfk8y) { status title }'`).
Worktree: `.temp/STORY-260907-2bddfc/worktree`, branch `task-board/story/STORY-260907-2bddfc`.
Change present, uncommitted: `M .github/ci/platform-exclusions.tsv`, 7 insertions / 4 deletions.
No converge/integrate executed in this run per binding (runner performs the bound
landing synchronously). No board status writes made. No worktree files changed by
this run (this evidence lives in /tmp and is attached as a board resource).

## Accepted revision shape (rev 1, comment-only)
One-file comment-only change to `.github/ci/platform-exclusions.tsv`. No exclusion row touched.

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
`grep -v '^#' .github/ci/platform-exclusions.tsv`: `diff` exit 0.
Single data row unchanged:
`internal/godriver  TestProbeRejectsAnUncoveredPlatformBeforeTheWorker  linux  ...`
Command exit code: 0.

## Factual claim (exit 0)
- `SPEC_PIN` at `.github/workflows/ci.yml:52` = `87a0d0060bad64ab883d007dcdf35df7485368bf`.
- `git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec ls-tree -r --name-only <pin> | grep qualification` -> `conformance/v1/vectors/conformance-claim-v3-qualification.json`, exit 0.
- Vector content at pin: `platforms: linux=excluded (until_task TASK-260728-1skseh), macos=pending-downstream-native-evidence, windows=pending-downstream-native-evidence`. Matches tsv prose that linux is excluded under any root and the vector states the same normatively.

## Consumer analysis (column kept, live fallback)
- `.github/ci/excluded-packages.sh:65-73`: when `$ROOT/$QUALIFICATION_VECTOR` exists the vector decides (`source_of_truth="the root's own ..."`); only when no QV exists does it fall back to `default_excluded_on` (`source_of_truth="default_excluded_on (this root publishes no ...)"`). Exit-0 probe: `bash .github/ci/excluded-packages.sh linux /tmp` prints the `default_excluded_on` fallback line, exit 0.
- `.github/ci/suite-plan.sh:96-100`: reports `platform qualification read from the root's own ...` when QV present, else `the root publishes no ...; the recorded default_excluded_on applies`.
- `.github/ci/gate-selftest.sh:577-581`: pre-vector-root test deletes the vector from a partial root and asserts the fallback is recorded. This is the live path the new prose describes.
Conclusion: the new prose is accurate (fallback for a SUPPLIED pre-vector root, dormant on the default lane). Keeping the column is correct.

## Gates with real exit codes (shell: sh, each standalone, `set -o pipefail`)
- Narrow fallback suite `/tmp/narrow-1bfk8y.sh` (vector root excludes godriver on linux w/ vector source; same root excludes nothing on darwin; pre-vector root excludes godriver on linux w/ `default_excluded_on`; `excluded-packages.sh` direct both sources): exit 0 (`NARROW_SUITE_EXIT=0`, `ALL_NARROW_OK`).
- `env CI_EXCLUDED_PKGS=internal/godriver CI_EXCLUDED_GOOS=linux bash .github/ci/ledger-consistency.sh /tmp/lc-shipped-1bfk8y2`: exit 0 (`LEDGER_EXIT=0`, `241 rows checked`, `ledger-consistency: ok`).
- Full `bash .github/ci/gate-selftest.sh`: NOT run to green in this run. Started once, observed partial `ok` lines, then terminated by the run itself to respect the headless single-call time bound (reported as cancelled, not passing). Reuse note: the review instruction allows reusing the hosted gate evidence for the exact candidate tree; the narrow suite above plus ledger-consistency are this run's fresh evidence. No full-suite green is claimed here.
- No CHANGELOG (comment-only, no behaviour change).

## Finding (out of scope, no file changed)
`.github/ci/excluded-packages.sh:12-14` header still says "the committed released pin is such a root" — the same stale claim fixed in the tsv. Left untouched per "Change no file"; orchestrator may file a follow-up leaf.
