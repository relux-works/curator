# TASK-260729-35tb37 review verdict — cycle 1

**Verdict:** CHANGES REQUESTED. Route to `analysis`; the remaining work is research-artifact correction, not implementation and not an external blocker.

## Independently supported

- CocoaSkills is clean and unstaged on local `main` at `edce8816dda44bb121d661b7c4dea942558ce408`. Read-only `git ls-remote origin refs/heads/main` and local `refs/remotes/origin/main` both resolve `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`; divergence is `0 2`. The two commits are `deb971f` and `6fc2fd9`.
- The package and CI map is supported at upstream `6fc2fd97`: `setuptools.build_meta`, `src` discovery, Python 3.11 through 3.14, console entry `csk = csk.cli:main`, strict mypy, pytest, 12-platform test matrix, build and twine gates, and rc.3 conformance ref `00b1688a...`. CLI routing, coarse `GlobalLock`, dry-run lock conflict, config/env/PATH boundaries, and reusable atomic-write/strict-JSON patterns match the cited source.
- The root ownership fits the accepted architecture: schema-v6 work belongs in `src/csk/skillspec.py`, validation-only `src/csk/skillcheck.py`, a build-domain initializer, and focused parser/check tests; transaction infrastructure belongs in `src/csk/locking.py`, new `src/csk/transactions.py`, and focused lock/journal tests. Live installer, planner, global, status, repair, and GC routing remains downstream.
- The rc.5 candidate at `.temp/TASK-260729-3nx97g/worktree` remains detached at `57c1f568`, unstaged but uncommitted, with manifest SHA-256 `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`, 11 `expected/build-driver` files, 12 manifest entries, build input SHA-256 `529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`, and receipt SHA-256 `919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`. The old accepted rc.5 snapshot still lacks both golden paths and all manifest references.
- The schema case inventory is 24 files each for `agent-skill-v6` and `csk-skill-v6`. Accepted protocol core and manager sections support the static root/module rules, lock ordering, journal ordering, recovery timing, reverse rollback, and consumer-last plan. The transaction diagram really is stale because it performs recovery before planning.
- No CocoaSkills product, pin, dependency, checkout, or protocol candidate was changed by this review. No broad suite was run.

## Required corrections

1. Add the explicitly required rc.2/rc.5 regression record and distinguish the stale local checkout from the implementation base. Accepted logs under `TASK-260729-1b9tc3` prove that local `edce881` with its committed rc.2 ref `cbe912d0...` runs `tests/test_protocol_conformance.py` as `98 passed`, while the same local code against immutable rc.5 produces `1 failed, 97 passed`; the failure adds `scripts/golden-tool` because local `load_skill_spec` ignores `agent-skill.json`. The manifests are semantically identical. Upstream commit `deb971f` already fixes canonical/legacy resolution and adds `tests/test_protocol_conformance.py::test_skill_manifest_resolution_vectors`, and `6fc2fd9` advances upstream CI to rc.3. The revised outcome must state that this is historical local-baseline evidence and a regression gate for work based on upstream, not a new product fix or authorization to alter a pin.

2. Update board drift. `TASK-260729-v5hqnv` is now `analysis`, not `reviewing`: cycle-1 review requested correction of an overbroad claim about removing stale hashes from the whole board and restoration of an out-of-scope `TASK-260720-12r55p.notes` mutation. The seven brief-field retargets and two dependency edges were verified, but the retarget task is not accepted. Revise sections 1, 6.2, references, and recommendation with this current state and its effect.

3. Correct the upstream delta count. `git diff --name-only HEAD..origin/main` contains 19 distinct paths, not 20. There are 20 commit-level touch events only because `.github/workflows/ci.yml` changes in both commits.

4. Make the local/upstream CI boundary explicit in the integration map: local `edce881` pins rc.2 `cbe912d0...`; upstream `6fc2fd97` pins rc.3 `00b1688a...`. Add the existing focused manifest-resolution conformance test to the schema-root regression gate or explain precisely why direct `test_skillspec.py` coverage is sufficient while preserving the upstream protocol test.

After these scoped artifact corrections, repeat current task projections and hand the revision to a fresh reviewer. No source edit, pin movement, dependency install, pull, checkout, or broad test is needed.