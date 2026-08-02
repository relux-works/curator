# BUG-260801-1iu1ln cycle-6 reviewer verdict

## Verdict

Changes requested. Route to implementation rework (`to-dev`). This is ordinary recoverable harness work, not a Stop-The-Line blocker.

Reviewed signed CocoaSkills commit `77b75dd7d20af2d0e901737f9900363b8384c675` on `task/BUG-260801-1iu1ln-lifecycle-observed-traces` in the dedicated worktree. The worktree is clean; HEAD and required base `ba250bfc4dfe104a160eadd5b5f4e340693bf892` have valid ECDSA signatures; the exact merge-base is the required base; no tag points at HEAD and no restricted release/pin/schema-v7/CI/version surface changed.

## Acceptance-blocking findings

### 1. Live published child mutations remain invisible

`tests/protocol_lifecycle_observations.py:826` defines `destination_matches_live` as exact live-directory inode equality. The mutation hooks at lines 855-968 therefore record namespace/open operations only when their direct target is the live directory itself, not a receipt or artifact below that directory. The final predicate at lines 1036-1045 then accepts restored final bytes as `atomic-complete-directory`.

Adversarial evidence: a POSIX `_rename_noreplace` wrapper performed the real atomic rename, then chmoded the now-live artifact, wrote corrupt bytes, fsynced them, restored the original bytes/mode, and returned. The probe recorded 42 live-child corruptions across the full lifecycle observation. Nevertheless the complete `publish-complete-immutable-entry-under-home-lock` object remained byte-for-byte equal to the normative case and reported `publication=atomic-complete-directory`, `result=published`, and `manager_home_lock=true`.

This is a direct false negative for the publication/immutability AC. Rework must observe mutating operations anywhere below the selected live entry after exposure (including chmod and descriptor-relative writes), or provide an equivalent fail-closed exposure witness on every supported backend. Add a retained product-seam sabotage test for transient post-rename child corruption and preserve the existing alternate-rename/partial-entry probes.

### 2. Negative-path persistent-effect assertions still use lossy final snapshots

Three normative areas remain vulnerable to write-and-restore behavior:

- Project/global upgrade dry-run: `_observe_upgrade_dry_run` classifies effects from call markers plus before/after tree equality. A wrapper around the real project/global install seam wrote extra bytes to the existing config and restored the original bytes before continuing. Both complete `project-upgrade` and `global-upgrade` cases remained exactly equal to the vectors; mutations were confirmed for both.
- Planning gates: `_observe_planning_gate_failures` records a persistent mutation only when the final tree differs (lines 2915-2965). A wrapper at project selection wrote/restored `Skillfile.json` before each gate. Twelve mutations occurred, while the complete planning case still reported `persistent_mutations=[]`, all required gates, and `build-eligible`.
- Private-build failure: lines 3440-3444 and 3589-3623 compare final trees and selected effect calls. A wrapper around the real `_build_private_misses` seam wrote/restored `persistent-generation` once during the two-build failure. The complete case remained exactly normative, including all eight `forbidden_effects`, identical before/after values, and `build-failed`.

These survivors violate the required-gates, private-build-effects, and dry-run-effects portions of the AC. Rework should apply the reusable high/low-level persistent mutation observer (or an equivalent fail-closed trace) to every persistent surface in these three paths, not only currentness and compiled dry-run. Add retained sabotage tests that require each complete case to differ under transient write/restore.

### 3. All-project upgrade deduplication passes vacuously with zero fetches

`_observe_upgrade_fetch` computes `deduplicated` as `len(fetched_paths) == len(set(fetched_paths))` at line 5615. It does not require any expected closure member to be fetched for the `all-projects-deduplicate` case.

Adversarial evidence: a wrapper forced `fetch_existing=False` for the two real projects in the all-project scenario. The omission was exercised twice for projects `one` and `two`; no all-project fetch phase ran. The complete observed case still equaled the vector: `{"deduplicate": true, "name": "all-projects-deduplicate", "scope": "project", "selection": "all"}`.

Rework must make `deduplicate=true` depend on the exact non-empty direct/transitive closure being fetched once each across both projects, while unrelated repositories remain excluded. Add a retained zero-fetch sabotage and a duplicate-fetch sabotage.

## Independent positive verification

- Exact cycle-6 five-probe gate: 5 passed in 212.49s.
- Canonical 32 cases, 378 scalar mutations, classification and helper gate: 417 passed in 43.80s.
- Strict configured source-only mypy: success, 68 source files.
- `compileall -q src tests`: exit 0.
- Signed-base/HEAD verification, exact merge-base, clean-tree, no-tag, restricted-surface and diff checks: exit 0.

The disclosed whole-repository legacy CLI lock-fixture failure is not causally introduced by this branch: `tests/test_cli.py` and `src/csk/locking.py` are unchanged from the exact base, and the only `src/csk/cli.py` diff is the global-upgrade fetch option. It is not the reason for this verdict.

## Required next handoff

1. Close all three complete-case false-negative classes above with focused product-seam sabotage tests.
2. Preserve the five cycle-6 probes and all prior 17 adversarial probes.
3. Re-run the canonical/scalar/classification gate, full authenticated exact-root protocol suite, relevant regressions, strict configured mypy, compileall, diff/release guards, build/Twine checks, signatures and clean-tree checks.
4. Produce a new signed commit directly on the dedicated task branch and update the task-scoped producer evidence.

Reviewer modified no CocoaSkills code, tags, releases, pins, claims, schemas, CI, changelog, or packaging metadata.