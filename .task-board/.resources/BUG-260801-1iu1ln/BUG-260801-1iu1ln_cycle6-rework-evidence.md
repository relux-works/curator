# BUG-260801-1iu1ln cycle-6 developer rework evidence

## Provenance and handoff commit

- Dedicated worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`
- Branch: `task/BUG-260801-1iu1ln-lifecycle-observed-traces`
- Required signed base and exact merge-base: `ba250bfc4dfe104a160eadd5b5f4e340693bf892`
- Preserved signed cycle-5 parent: `27ace8b6d1b7d6a1b54acac30e670098ca3b5110`
- Signed cycle-6 commit: `77b75dd7d20af2d0e901737f9900363b8384c675`
- Base, parent, and cycle-6 signature checks: exit 0; expected ECDSA signer `oparin@me.com`
- Post-build worktree status is clean and no tag points at the cycle-6 commit.

## Causal repairs for the five cycle-5 findings

1. The cross-project success case now runs two real installs through successful private builds, protected publication, materialization commit, and consumer registration in one scenario. It records the actual `ProjectConfig.path` at each commit, reads both registered consumers and the shared protected cache hit, proves distinct private cache-key builds overlap, handles the normal generation retry, and proves the real publish/commit handoffs serialize under the manager-home lock. The obsolete separate synthetic success transaction helper was removed.
2. Publication, cross-project success and rollback, dry-run, GC, private build, status, repair, and deterministic transaction order now derive normative cache/receipt identities from their own plans, publications, markers, protected inspections, or repaired state. The operation-side sabotage changes those identities while leaving the authenticated fixture untouched and makes every corresponding complete case differ.
3. Process observation resolves every path-like argv element against explicit or inherited subprocess cwd for both `run` and `Popen`. The retained regression executes `./bin/golden-tool` from an untrusted protected candidate and is detected.
4. Persistent mutation observation covers high-level `Path` calls and descriptor-relative low-level open/write/unlink/remove, mkdir/rmdir, rename/replace, link/symlink, truncate, timestamp, and permission families. The retained low-level write/fsync/unlink-and-timestamp-restore sabotage changes dry-run, clean currentness, and the full failure matrix.
5. Atomic-publication evidence observes all supported namespace operations targeting the exact live cache destination, including alternate `os.rename` on POSIX and move/rename parity on Windows, rather than trusting only the named no-replace helper.

Status/repair fixtures also use the normative `golden-tool` operation identity. Wrong-target and wrong-toolchain repair cases now place independently invalid identities in the protected candidate receipt while preserving the desired native operation, so the observed rebuild returns the normative key through the real repair pipeline.

The LOGBOOK was corrected to distinguish the earlier cycle-5 claims from these cycle-6 observations. No product source module changed in cycle 6.

## Direct gate evidence

- Initial exact five-probe expected-red gate: exit 1; 5 failed in 204.23s because all five reviewer survivors still preserved complete-case equality.
- Exact cycle-6 five-probe gate after repair: exit 0; 5 passed in 209.15s.
- Preserved prior adversarial gate: exit 0; 17 passed in 734.55s.
- Canonical/scalar/classification/helper gate: exit 0; 417 passed in 43.53s. This includes all 32 lifecycle cases, all 378 scalar leaves, fail-closed literal and lossy-proxy classification, unknown-field rejection, cwd/path and exact-project-identity helpers, and both rollback contract fields.
- Full authenticated exact-root protocol conformance: exit 0; 857 passed in 998.98s.
- Focused preserved product regressions: exit 0; 3 passed in 3.18s.
- Installer, global install, build currentness, and installer transaction suites: exit 0; 131 passed in 126.02s.
- Transaction, GC, and status suites: exit 0; 111 passed and 1 expected platform skip in 14.37s.
- Strict configured mypy: exit 0; no issues in 68 source files.
- `compileall src tests`: exit 0.
- Unstaged, staged, exact-base, and committed diff checks: exit 0. No standalone formatter or linter is configured.
- Signed-base and parent verification, exact merge-base, cycle-6 signature, clean-tree, no-tag, and restricted release/pin/CI/version surface checks: exit 0.
- Isolated PEP 517 build from detached signed `77b75dd`: exit 0; built `cocoaskills-0.12.6.dev43+g77b75dd7d` sdist and wheel.
- Twine check of both distributions: exit 0; both passed. Sdist lifecycle-observer membership check: exit 0.

## Non-green command record

- During implementation, focused status/repair diagnostics exited 1 while exposing operation-key projection and invalid wrong-target receipt construction; the repaired exact case subsequently passed, and all three status/repair canonical vectors passed together.
- During implementation, cross-project canonical diagnostics exited 1 while exposing shared snapshot retirement, optimistic retry handling, same-key build serialization, and commit-label projection. Each was repaired causally; the isolated canonical case, all 32 cases, the five-probe gate, the 17 inherited probes, and exact-root protocol suite subsequently exited 0.
- The first expanded identity sabotage exited 1 because its altered publication key leaked into adjacent publication fixtures; scoping the operation key fixed the harness, and the expanded sabotage subsequently passed.
- One related-suite command used the nonexistent historical path `tests/test_installer.py`: exit 4, no tests ran. The corrected command using `tests/test_install.py` exited 0 with 131 passed.
- An additional whole-repository diagnostic (broader than the required exact-root protocol gate) exited 1 with 2,125 passed and 54 skipped in 1,321.76s. Its sole failure was `tests/test_cli.py::test_cli_lock_contention_returns_lock_exit`: the unchanged legacy `.lock` file fixture expected exit 3, while signed HEAD intentionally rejected online legacy-lock migration with exit 1. The isolated test reproduced the same failure (exit 1), and `git diff --exit-code HEAD -- src/csk tests/test_cli.py` exited 0, proving cycle 6 did not touch that pre-existing surface.

No PR, main landing, tag, release, claim, pin, schema-v7, CI, changelog, pyproject, or generated-version action or change was made. The signed commit is for later PR19 integration and review.
