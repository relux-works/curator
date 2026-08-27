# TASK-260729-35tb37 review verdict — cycle 2

**Verdict:** CHANGES REQUESTED. Route to `analysis`; the remaining work is one
current-state research-artifact correction, not implementation and not an
external blocker.

## Independently supported

- CocoaSkills remains clean and unstaged on local `main` at
  `edce8816dda44bb121d661b7c4dea942558ce408`. Read-only
  `git ls-remote origin refs/heads/main` resolves
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`; local divergence is `0 2`.
  No pull, fetch into refs, checkout, install, edit, stage, or commit occurred.
- The upstream delta correction is exact: 19 distinct paths and 20
  commit-level touch events. `deb971f` touches 19 paths, `6fc2fd9` touches
  `.github/workflows/ci.yml`, and that workflow is touched by both commits.
- The packaging, CLI, environment, PATH, and CI map matches
  `origin/main@6fc2fd97`: setuptools `src/` discovery, Python 3.11–3.14,
  `csk = csk.cli:main`, strict mypy, the 12-cell pytest matrix, build/twine
  gates, rc.3 conformance ref `00b1688a...`, current coarse `GlobalLock`
  routing, dry-run boundary, atomic-write patterns, and strict protocol JSON.
- The two root file plans match their live briefs and current source seams:
  schema-v6 owns `src/csk/skillspec.py`, new
  `src/csk/builds/__init__.py`, validation-only `src/csk/skillcheck.py`,
  `tests/test_skillspec.py`, and `tests/test_skillcheck.py`; transaction
  infrastructure owns `src/csk/locking.py`, new `src/csk/transactions.py`,
  `tests/test_locking.py`, and new `tests/test_transactions.py`. Installer,
  planner, global, recovery/GC, build execution, cache, packaging, workflows,
  and pin movement remain downstream/out of scope. The focused pytest/mypy
  gates are appropriately narrow, and the schema-root plan now includes the
  owner-published manifest-resolution vector replay with an explicit rc.3 root.
- The historical regression evidence is exact. Board logs
  `TASK-260729-1b9tc3_baseline-pin-rc2.log` and
  `TASK-260729-1b9tc3_candidate-rc5.log` report respectively `98 passed`
  (exit 0) and `1 failed, 97 passed` (exit 1). The only failure adds
  `scripts/golden-tool`. The rc.2 `csk-skill.json` and rc.5
  `agent-skill.json` normalize to equal JSON, while `expected/marker.json` and
  `expected/context_files.json` are byte-identical. Local `load_skill_spec`
  ignores `agent-skill.json`; upstream `deb971f` supplies canonical/legacy
  resolution plus `test_skill_manifest_resolution_vectors`, and `6fc2fd9`
  advances CI to rc.3. Revision 2 correctly frames this as a stale-base
  regression gate, not root-task product work or pin authorization.
- The local/upstream CI boundary is supported: local rc.2
  `cbe912d0...` has manifest SHA-256 `728f7729...` and 81 files; upstream rc.3
  `00b1688a...` has SHA-256 `7951cda1...` and 93 files, including eight
  manifest-resolution cases.
- The accepted rc.5 golden candidate remains based at `57c1f568...`, with 130
  uncommitted entries and nothing staged. Its manifest SHA-256 is
  `b6f56aac...`; it has 11 `expected/build-driver` files and 12 manifest
  references. The build-input and receipt hashes remain `52937012...` and
  `919fbbad...`. The older accepted rc.5 snapshot still lacks the golden
  vector/tree and manifest references. No publication, pin movement, or
  product change is implied.
- Live dependency edges and all other board-state rows match revision 2:
  both roots are `backlog`, each directly blocked by `TASK-260720-1pvfj5`;
  `1pvfj5` remains `backlog` behind done `2qqq0w` and development
  `jrrgw9`; `3ag6pi` remains `blocked`; `3nx97g` and `1b9tc3` remain `done`.
  The two root diagrams are correctly identified as stale in protocol wording
  and recovery timing.

## Required correction

`TASK-260729-v5hqnv` completed its second independent review after revision 2
was written and is now `done`, with accepted evidence in
`TASK-260729-v5hqnv_review-verdict-cycle-2.md`. The baseline outcome is
therefore no longer current in all of these places:

- executive finding 3 says the task is `to-review` and unaccepted;
- section 6.2 reports current state `to-review`;
- section 6.2 says the seven rc.5-aligned brief texts are not
  reviewer-accepted and may still change in the pending second review;
- recommendation item 7 asks to complete the pending review cycle;
- references omit the accepted cycle-2 verdict.

Revise only this research outcome:

1. set the `TASK-260729-v5hqnv` current state to `done`;
2. state that the seven brief-field retargets and the two provenance dependency
   edges are reviewer-accepted, while preserving the correct no-product-change,
   no-test-run, no-root-unblock, and fail-closed `3ag6pi` boundaries;
3. remove the pending-review warning/recommendation and cite
   `TASK-260729-v5hqnv_review-verdict-cycle-2.md`;
4. re-query the tracked status/dependency projection once more before the next
   reviewer handoff.

No CocoaSkills, Curator, protocol, pin, dependency, checkout, task brief,
diagram, product file, or test needs to change. No broad test is needed. The
historical rc.5 red result must remain attributed as evidence against the stale
local base, not rewritten as a green current-product claim.

## Review execution evidence

Focused read-only checks covered live board projections, both historical pytest
logs, semantic/byte fixture comparisons, CocoaSkills git provenance and
upstream path counts, upstream source/workflow/package inspection, root-task
briefs, candidate/old-snapshot identities, schema-case counts, dependency
edges, and both diagrams. No test suite was run by this reviewer because the
task is research-only and broad tests are forbidden.
