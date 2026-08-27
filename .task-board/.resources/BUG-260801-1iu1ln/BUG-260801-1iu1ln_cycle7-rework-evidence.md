# BUG-260801-1iu1ln cycle-7 developer rework evidence

## Provenance and signed handoff commit

- Dedicated CocoaSkills worktree: `/Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`
- Branch: `task/BUG-260801-1iu1ln-lifecycle-observed-traces`
- Required signed base and exact merge-base: `ba250bfc4dfe104a160eadd5b5f4e340693bf892`
- Preserved signed cycle-6 parent: `77b75dd7d20af2d0e901737f9900363b8384c675`
- Signed cycle-7 commit: `a0046fdfbd37ecce4c5d6d0e21152628c2d2432f`
- `git verify-commit HEAD`: exit 0; good ECDSA signature for `oparin@me.com`.
- Parent, exact merge-base, committed diff, restricted-surface diff, clean branch, and no-tag checks: exit 0.

## Cycle-7 reviewer findings and causal repairs

1. The shared persistent-mutation observer now resolves descriptor-relative paths on Darwin with `fcntl.F_GETPATH`, in addition to `/dev/fd` and `/proc/self/fd`. Cache publication observation is installed over the exact live entry after the atomic rename. It records and rejects descendant open/write/truncate/fsync restoration and permission changes while distinguishing the legitimate cache-root seal.
2. Project and global upgrade dry-runs, every planning-gate probe, and the private-build failure case now install the same high- and low-level persistent observer over their exact CocoaSkills effect surfaces. Any observed mutation changes the lifecycle answer even when final bytes, modes, and timestamps are restored.
3. All-project upgrade now reports deduplication only when the observed fetch multiset is exactly the nonempty direct-plus-transitive repository closure, once each, with the unrelated repository absent. Dedicated zero-fetch and duplicate-fetch probes both change the normative case.
4. The stronger Darwin descriptor trace exposed a legitimate permission change while repair quarantines and replaces an invalid candidate. Rebuild evidence remains bound to the complete repair pipeline, resulting currentness, and absence of candidate execution. The separate chmod-and-adopt shortcut remains fail-closed on permission mutation.

Cycle 7 changes only `tests/protocol_lifecycle_observations.py`, `tests/test_protocol_conformance.py`, and `LOGBOOK.md`; it preserves the accepted earlier PR19 repairs and does not add a product-code workaround.

## Expected-red and refinement evidence

- Initial exact six-case cycle-7 gate: exit 1; 5 failed and 1 passed in 256.86s. The already-protected duplicate-fetch variant passed, while live-child mutation, both upgrade dry-runs, planning mutation, private-failure mutation, and zero-fetch survived before repair.
- First observer revision of the same gate: exit 1; 4 failed and 2 passed. macOS returned `EINVAL` for descriptor `readlink`, demonstrating that Darwin needed native `F_GETPATH` attribution.
- Initial canonical/scalar/classification run after Darwin attribution: exit 1; 416 passed and 1 failed. The single failure was the legitimate repair quarantine/replacement `fchmod`, not a normative survivor. Refining rebuild versus chmod-and-adopt evidence removed that false positive.
- The unsabotaged affected-case diagnostic also exited 1 with 5 passed and 1 failed until the legitimate cache-root publication seal was distinguished from forbidden descendant mutations.

These are recorded as expected-red or diagnostic failures, not passing gates. Every corresponding final gate below ran separately and exited 0.

## Final direct gates

- Exact six new cycle-7 sabotage cases: exit 0; 6 passed in 288.91s.
- All 22 inherited lifecycle sabotage probes: exit 0; 22 passed in 1,078.88s.
- Canonical/scalar/classification/helper gate: exit 0; 417 passed in 54.42s. This includes all 32 lifecycle cases, every scalar-leaf mutation, fail-closed literal/lossy-proxy classification, all argv elements, exact project identity, rollback guards, and unknown-field rejection.
- Full authenticated exact-root protocol conformance: exit 0; 863 passed in 1,372.22s.
- Installer, global install, build currentness, and installer transaction suites: exit 0; 131 passed in 142.24s.
- Transaction, GC, and status suites: exit 0; 111 passed and 1 expected platform skip in 16.45s.
- Strict configured mypy: exit 0; no issues in 68 source files.
- `python -m compileall -q src tests`: exit 0.
- Unstaged, staged, committed, and exact-base diff checks: exit 0. No standalone formatter or linter is configured.
- Exact-base release/version/CI restricted-surface diff (`pyproject.toml`, `CHANGELOG.md`, `.github`, and `src/csk/_version.py`): exit 0.

## Signed-tree package validation

- Isolated PEP 517 build from detached signed `a0046fd`: exit 0; built `cocoaskills-0.12.6.dev44+ga0046fdfb` sdist and wheel.
- Twine check of both distributions: exit 0; both passed.
- Sdist membership check for `tests/protocol_lifecycle_observations.py` and `tests/test_protocol_conformance.py`: exit 0.
- Detached build source and primary branch remained clean: exit 0.
- Temporary build artifacts were moved to Trash after validation and can be recovered there if needed.

No PR, main landing, tag, release, claim, pin, schema-v7, CI, changelog, pyproject, or generated-version action or change was made. The signed commit is ready for later PR19 integration and reviewer acceptance.
