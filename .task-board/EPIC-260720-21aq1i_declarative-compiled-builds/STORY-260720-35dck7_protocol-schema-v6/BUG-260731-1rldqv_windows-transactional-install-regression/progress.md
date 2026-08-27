## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- BUG-260731-2rhy74

## Checklist
- [x] Reproduce and isolate each Windows failure family on ssh win or a faithful Windows harness.
- [x] Fix command-shim digest, replacement, and provenance behavior without weakening corruption guards.
- [x] Add focused Windows regression tests and keep Linux/macOS behavior green.
- [x] Publish a signed commit to the CocoaSkills PR 16 branch only on ivanopcode/cocoaskills.
- [x] Attach evidence, hand off to independent Opus review, and require the full PR 16 matrix green.
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-2dec82, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-2dec82)
Root cause analysis (in progress).

Failure classification of run 30594273278 windows py3.11 (45 failed):
- 28 digest-mode: TransactionCorruptionError transaction target changed while digesting <...>.cmd
- 12 winerror5: PermissionError WinError 5 on project/.claude|.codex/skills/<name> (8 direct + 4 test_e2e CLI exit 1)
- 4 publication-owner: cache_publication_invalid manager home owner does not match the current manager principal
- 1 test_audit_cli exit 1, same digest-mode family (skill ships scripts/tool.cmd)

CONFIRMED CAUSE 1 (digest-mode). CPython Modules/posixmodule.c update_st_mode_from_path() ORs 0o111 into st_mode for .bat/.cmd/.com/.exe, and it is only reachable from path-based stat (win32_xstat_impl). Python/fileutils.c _Py_fstat_noraise has no path so it cannot apply it. transactions._digest_file compares stat.S_IMODE(os.fstat(fd)) against stat.S_IMODE(path.lstat()) and therefore always sees 0o666 vs 0o777 for a .cmd shim. Deterministic, fires for every command shim, never for .csk-install.json. Same defect in _entry_content_digest and _staging_prefix_digest. The digest payload also embedded the synthesized mode, so a staged sidecar name and its live .cmd name digest differently on Windows.

CONFIRMED CAUSE 2 (winerror5). adapters.stage_project_adapter_targets stages the adapter entry as a relative symlink whose destination is computed for the live location, so it is dangling while staged. transactions._staging_tree_entry recorded link_is_directory=path.is_dir(), which resolves the dangling destination and yields False. _create_staging_entry then rebuilt it with target_is_directory=False, producing a Windows file-type symlink onto a directory, which cannot be traversed: os.stat raises WinError 5. POSIX ignores the flag, which is why only Windows breaks.

CAUSE 3 (publication-owner) still under diagnosis on a windows-latest probe job. The affected tests reach cache publication only under PR16 because c4131bd added the private_base build operation root and stubbed go_v1.build.

Fixes applied so far on the PR16 branch content: permission-identity based mode handling in transactions digest/guards, link type read from the reparse point instead of the resolved destination, adapters symlink probe now creates a directory link. New regression tests in tests/test_transactions.py.
Fix published to PR 16 as signed commit 7a66c73 on task/TASK-260720-3t8nr3-transactional-project-hybrid (parent 8a02e17). Four independent Windows platform causes, none of them transaction-engine logic.

CAUSE 1 (28 failures) - synthesized execute bits. CPython update_st_mode_from_path ORs 0o111 into st_mode for .bat/.cmd/.com/.exe and is reachable only from path-based stat; os.fstat has no path. Probe on windows-latest: tool.cmd lstat=0o777 fstat=0o666 same inode; tool.txt 0o666/0o666. transactions._digest_file compared the two directly. Same defect in _entry_content_digest and _staging_prefix_digest, plus a latent one in the digest payload (a staged sidecar name and a live .cmd name digested differently). Fix: _permission_identity, reusing the existing _permission_mode_identity; no-op on POSIX.

CAUSE 2 (12 failures) - Windows symlink type. A staged adapter link is dangling by construction, and link_is_directory was derived by resolving that destination, so the link was rebuilt as a file link onto a directory, which Windows cannot traverse (WinError 5). Probe: dangling dir link has st_file_attributes=0x410 while is_dir() is False. Fix: read the type from the reparse point; also compared during staging validation, and the adapters probe now creates the link kind it will actually create. The 3.14 cell differing by exactly these 12 (assert False instead of WinError 5, because Path.exists() swallows it on 3.14) is independent confirmation.

CAUSE 3 (4 failures) - Windows never grants manager ownership. New objects belong to the token owner; on the runner _current_user_sid() is RID 500 while every created object is owned by S-1-5-32-544 BUILTIN\Administrators. The manager owned neither the home it created nor the artifact its compiler produced, so the protected build cache rejected both. POSIX gets that state free from st_uid plus mkdir(0o700). Not reachable on main, whose installer only plans builds; c4131bd is the first commit where install compiles and publishes. Fix on the producer side, guards untouched: locking.provision_new_manager_home provisions a home it creates (never repairs an existing one, so the drift guard and its test still hold), and builds.cache.make_publication_source_private stamps a fresh artifact before publication.

CAUSE 4 - _stub_trusted_toolchain hard-coded goos=linux and bin/<cmd>, so it could only ever publish on a non-Windows host. Now derives target and artifact path from the host.

EVIDENCE. Local macOS: mypy strict exit 0; full suite 1140 passed 100 skipped exit 0 (baseline before fixes 1136/99, delta is exactly the five new tests). windows-latest staged probes each removed one signature and exposed the next, no signature reappearing; final probe run 30620126160 = 253 passed, 55 skipped, success across test_install, test_status, test_gc, test_build_cache_windows, test_build_cache_posix, test_transactions, test_locking, test_adapters, test_config. Probe branches deleted. PR 16 full matrix run 30624304158 in flight on 7a66c73; mypy strict already green.

OPEN PRODUCT DECISION (not blocking this fix): a Windows home created by an elevated shell before this change stays Administrators-owned and fails closed. Adopting it would blunt the drift guard, so no automatic repair was added. Explicit opt-in repair vs documented manual re-provisioning is unresolved.
Host migration resumed on macbook-iv.local. CocoaSkills PR 16 head 7a66c73 currently has Ubuntu/macOS/mypy and Windows Python 3.11, 3.13, 3.14 green; Windows Python 3.12 remains in progress. Monitor that exact job, collect evidence, make focused rework only if it fails, then hand off to independent Opus review. Push only to ivanopcode/cocoaskills; no tags/releases.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-d6e3c5, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-d6e3c5)
RUN-260731-d6e3c5 (developer, resumed after migration).

FIFTH CAUSE FOUND AND FIXED: super-linear namespace validation.

State on arrival: commit 7a66c73 from the cancelled RUN-260731-2dec82 had already fixed the four platform causes and was published to PR 16. Windows 3.11/3.13/3.14 on run 30624304158 report SUCCESS, so the four original failure families are gone. The blocker was that Windows 3.12 had run 2h54m without finishing. The GitHub API completedAt field for the three green cells is stale/wrong: their own logs show Python 3.11 finished 1206 passed 152 skipped in 8323.00s = 2h18m43s, against about 14 minutes for the same suite on main run 30556125542. Individual install tests took 76-247 seconds each. So the whole Windows lane was pathologically slow, not just 3.12, and the AC requirement of a green matrix was not reliably reachable.

DIAGNOSIS. ssh win is offline (tailscale last seen 1d ago), so reproduction used a disposable probe branch running the suite on windows-latest with pytest-timeout and faulthandler. Two independent dumps, Python 3.11 and 3.12, landed on the identical frame: _canonical_target_path <- _namespace_parts <- _namespaces_overlap <- _validate_namespace_independence <- _validate_journal <- _save_journal. Probe branch deleted.

ROOT CAUSE. _validate_namespace_independence compares every declared namespace with every other one, and each comparison canonicalised BOTH operands via Path.resolve(). Filesystem work therefore grew with the square of the namespace count, and the pass runs on every _save_journal - including twice per 32 KiB staging chunk. Measured on macOS, where realpath is cheap: ONE install test = 750,620 _canonical_target_path calls, 182.5s of its 210s runtime, over 135 validation passes of about 74 namespaces each. Windows opens a handle per path component for the same question, which is why only Windows became unusable while POSIX merely looked slow. This is c4131bd-reachable: main only plans builds.

FIX (signed commit 98ab7a2, parent 7a66c73). A namespace is now a _NamespaceProbe that resolves its path, and reads its physical identity, at most once per pass; every comparison reads that one answer. GUARD UNCHANGED: parts equality, prefix containment and the samestat physical-alias check still run for every pair, identity is still lstat for entry targets and stat otherwise, and a real OSError is still never cached so it cannot be hidden from a later reader. Memoising per pass is if anything more internally consistent than re-asking the filesystem per comparison. Also removed the redundant re-derivation of manager-home parts once per added namespace.

MEASURED. Same install test: 750,620 -> 12,575 canonicalisations (59.7x fewer), 182.5s -> 10.4s canonical time, 210s -> 27.9s wall (7.5x). Windows should gain more than macOS because resolve is the part it amplifies.

TESTS. Five new tests in tests/test_transactions.py pin the cost contract and the preserved reach: one canonicalisation per namespace plus the manager home; growth linear in namespace count not quadratic; a probe answers repeated comparisons from a single stat; an absent namespace has no identity and does not raise; and a hardlink alias between two targets is still rejected as a namespace overlap. All 13 pre-existing namespace/overlap/lock-alias guard tests still pass.

LOCAL EVIDENCE (macOS, python 3.11.6). mypy strict exit 0, no issues in 67 source files. pytest tests/test_transactions.py tests/test_installer_transactions.py tests/test_locking.py tests/test_adapters.py exit 0, 156 passed 4 skipped in 38.63s. Full local suite running.

CI. PR 16 head is now 98ab7a2, run 30636027978 in flight. The previously hung 3.12 cell of run 30624304158 was superseded by this push.
FINAL (RUN-260731-d6e3c5). PR 16 matrix FULLY GREEN.

Run 30637483316, head 32737a8: completed/success, all 14 jobs - 12 test cells (ubuntu/macos/windows x 3.11/3.12/3.13/3.14), mypy strict, Build artifacts. Prior run 30636027978 on 98ab7a2 was also fully green, so both published commits are independently verified.

WINDOWS SUITE DURATION (1211 passed, 152 skipped in every cell)
  main b3a5031 baseline .......... 5m49s
  PR16 7a66c73 before this fix ... 2h18m43s; 3.12 never finished (3h+)
  PR16 98ab7a2 ................... 17m49s / 11m43s / 17m37s / 13m47s
  PR16 32737a8 ................... 10m27s / 7m55s / 15m20s / 16m17s

The four original failure families from 7a66c73 stayed fixed throughout; nothing regressed.

WHAT WAS ACTUALLY WRONG. Not a hang and not 3.12-specific. _validate_namespace_independence compares every declared namespace with every other one, and each comparison canonicalised BOTH operands via Path.resolve(); the pass re-runs on every _save_journal, including twice per 32 KiB staging chunk. Cost therefore grew with the square of the namespace count. One install test measured 750,620 _canonical_target_path calls = 182s of its 210s, on macOS where realpath is cheap. Windows opens a handle per path component for the same question, so only Windows became unusable. Reachable only under c4131bd, which is the first commit where install materialises transactionally.

Note for the record: gh run view --json jobs reported completedAt values for the three green Windows cells that were plainly wrong (implying ~34min). The job logs are authoritative and said 8323.00s. Do not trust that field.

TWO SIGNED COMMITS, both on ivanopcode/cocoaskills only.
  98ab7a2 - a namespace is a _NamespaceProbe that resolves its path, and reads its physical identity, at most once per pass. 750,620 -> 12,575 canonicalisations, test 210s -> 27.9s.
  32737a8 - the pairwise scan itself was then the remaining square term (364,635 comparisons per install). Naming, containment and physical aliasing are now three index lookups: a dict keyed by normalized parts, a lookup of each namespace proper prefixes over that dict (path depth is bounded), and a dict keyed by st_dev+st_ino, which is exactly what samestat compares. Test 27.9s -> 5.49s, 38x overall.

GUARD NOT WEAKENED. Same predicate, same rejections, same message shape, and the reported pair is still a genuinely colliding pair. Identity is still lstat for entry targets and stat otherwise. A real OSError is still never cached. Every validation pass that ran before still runs - both commits only make each pass cheaper. I deliberately did NOT reduce how often _validate_journal runs, because skipping passes changes WHEN corruption is detected, which is a guard question rather than a cost question.

TESTS. Five new tests in tests/test_transactions.py pin the cost contract and the preserved reach. Red/green verified: with per-access re-resolution restored, 60 namespaces produce 3597 canonicalisations instead of 61 and both cost tests fail. All 13 pre-existing namespace/overlap/lock-alias guard tests pass unchanged.

LOCAL EVIDENCE (macOS, CPython 3.11.6), real exit codes.
  python -m mypy .............................................. exit 0, no issues in 67 source files
  pytest test_transactions/installer_transactions/locking/adapters  exit 0, 156 passed 4 skipped
  pytest broad set incl. gc/status/activation_modes ........... exit 0, 174 passed 4 skipped
  full suite on 98ab7a2 ....................................... exit 0, 1145 passed 100 skipped, 626.80s
  full suite on 32737a8 ....................................... exit 0, 1145 passed 100 skipped, 574.42s
Recorded pre-fix baseline was 1140/100, so the delta is exactly the five new tests. No dedicated linter is configured in pyproject.toml; mypy strict is the project static gate and is green locally and in CI.

HARNESS. ssh win was offline all run (tailscale: last seen 1d ago), so reproduction and verification used windows-latest runners. A disposable probe branch ran the suite under pytest-timeout and faulthandler; two independent dumps, 3.11 and 3.12, landed on the identical frame, which is what proved a hot path rather than a hang. Probe branch deleted.

OPEN, NOT BLOCKING. The product decision from the previous run is unchanged: a Windows manager home created by an elevated shell before 7a66c73 stays Administrators-owned and fails closed; adopting it automatically would blunt the drift guard. Explicit opt-in repair vs documented manual re-provisioning is still unresolved.

Artifacts: BUG-260731-1rldqv_namespace-cost-fix.md and BUG-260731-1rldqv_probe-logs.tar.gz. Ready for independent Opus review.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-d6e3c5, pid=23836, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-2bd01b, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-2bd01b)
REVIEW VERDICT (RUN-260731-2bd01b, independent Opus, read-only): CHANGES REQUESTED -> to-dev. Evidence: BUG-260731-1rldqv_independent-review-verdict.md.

WHAT HOLDS. Every AC claim re-derived from primary sources, none carried over from the developer report, and all of it held. PR 16 run 30637483316 on head 32737a8 is 14/14 success (12 test cells ubuntu/macos/windows x 3.11-3.14, mypy strict, Build artifacts). Head confirmed 32737a8, MERGEABLE. All three commits 7a66c73/98ab7a2/32737a8 carry good GPG signatures (%G? = G) and origin is ivanopcode/cocoaskills only. Locally at 32737a8 on macOS CPython 3.11.6: mypy strict exit 0 (67 files); full suite exit 0, 1145 passed 100 skipped in 345.43s, matching the reported 1145/100 exactly; targeted transactions/locking/adapters/installer_transactions exit 0, 156 passed 4 skipped.

GUARD-NOT-WEAKENED CLAIM TESTED, NOT ACCEPTED. 32737a8 replaces the O(n^2) pairwise namespace scan with three index lookups, the riskiest edit in the change. I copied the OLD pairwise algorithm verbatim out of 8a02e17 and ran it against the SHIPPED indexed scan over identical namespace sets on a real filesystem, covering duplicates, containment at several depths, hardlink alias, symlink-to-dir, symlink-to-file, dangling symlink, absent paths, absent-under-absent-parent, and entry/bytes mixtures: 1771 exhaustive pair/triple sets + 4000 randomized sets = 5771 total, ZERO disagreements in the accept/reject decision. The indexed form is also a superset: the old code reached the physical-identity check only for pairs surviving the parts checks, the new pass indexes identity for every namespace. Containment coverage is complete because _namespace_parts resolves to an absolute path, so the only prefix range(1,len(parts)) cannot reach is the empty one, which no namespace can have.

DIGEST NO-OP-ON-POSIX CLAIM MEASURED. digest_path over a tree of .cmd/.exe/.bat/plain files across modes 0o755/0o700/0o644/0o600/0o444 under a 0o750 dir is BYTE-IDENTICAL at 8a02e17 and 32737a8 (sha256:346f63ad...5251 both). No POSIX receipt or install marker is invalidated.

ALSO VERIFIED. make_publication_source_private is called immediately before the ONLY CachePublication( construction site in the tree, and both routes a manager home comes into existence go through provision_new_manager_home, so coverage is complete. The chmod lands after a content-only sha256 and the publication re-hash is also content-only, so it cannot desynchronise the receipt. _staging_tree_entry passes an lstat of the link itself to _link_is_directory, so st_file_attributes describes the reparse point; adding link_is_directory to _validate_staging_entry_modes STRENGTHENS the staging guard. Test changes are faithful to the platform rather than accommodations of it: conftest.csk_home now builds its home through the product own provision_new_manager_home, and _stub_trusted_toolchain derives target and artifact path from the host via the product own derived_artifact_path. Diagnosis independently corroborated by curator BUG-260731-33v6zz, which found the same 0x410 junction reparse point and the same unprotected-manager-home fact in the Go implementation.

REWORK 1 (the reason this is not accepted). src/csk/locking.py:126-132, introduced by 7a66c73. Path.mkdir(exist_ok=True) does not swallow FileExistsError, CPython does if not exist_ok or not self.is_dir(): raise. The rewrite to a bare except FileExistsError: return dropped that is_dir() condition, so a non-directory at the manager-home path is now accepted. Measured differentially on both commits: symlink-to-MISSING-dir went from LockError cannot create manager home to OK with the home materialised at the link destination at mode 0o755 and NEVER PROVISIONED (mode=0o700 create mode skipped, and on Windows the whole provision_manager_home ownership/DACL stamp skipped); regular file went from LockError to a raw FileExistsError that is not in the tuple cli.main catches, so it surfaces as a traceback instead of error: ... plus EXIT_LOCK. The absent, existing-dir and symlink-to-existing-dir rows are unchanged. This sits in the exact function the commit added to establish the manager home private state, and its docstring claims only a home this call creates is provisioned while a home it did not create is now accepted un-provisioned. SEVERITY STATED HONESTLY: latent, not live. I checked reachability rather than assuming it - every GlobalLock(cfg.path.parent) site in cli.py (573, 592, 639) is preceded by config.load_config(), which fails first on a non-directory home, and csk bootstrap already died with an unhandled FileExistsError on this input BEFORE the change, so that path is not a regression. It is still rework: this task scope names ownership and provenance checks explicitly and forbids weakening them, the defect is inside the new ownership-provisioning code, 7a66c73 newly made provision_new_manager_home a public two-caller module function, and no test covers the case, which is what let it through. FIX: restore the condition mkdir(exist_ok=True) applied - except FileExistsError: if not csk_home.is_dir(): raise LockError(f"cannot create manager home: {csk_home}"); return - which keeps the symlink-to-existing-dir row working. Plus a tests/test_locking.py regression test for the two changed rows with symlink-to-existing-dir as the positive control.

REWORK 2. 98ab7a2 cost contract is pinned by two tests; 32737a8 contract (that the pairwise comparison itself is gone) is not. Restoring a pairwise scan over memoised probes would leave both existing cost tests green while reinstating the 364,635 comparisons that commit removed, which was the dominant Windows term. Add a comparison/lookup counter in the same shape as the existing two tests.

REWORK 3. Curator LOGBOOK.md entry (line 59) stops at PR 16 matrix running on 7a66c73. The two most transferable findings are missing from the cross-project record: the quadratic _validate_namespace_independence (a POSIX-latent cost defect that only Windows made fatal) and the tooling anomaly that gh run view --json jobs reported completedAt values contradicted by the job logs (34 minutes against an actual 8323.00s). Both are in the CocoaSkills logbook and board notes already.

NOT BLOCKING. link_is_directory is persisted in StagingTreeEntry and now recomputed as False on POSIX where path.is_dir() gave True, so a pre-fix journal with a directory link would fail _validate_staging_entry_modes on resume - unreachable in practice since entry targets with links arrive with c4131bd, which is not on main. The config -> locking -> builds.cache two-level function-local import chain works and is commented but is worth watching. The open product decision on opt-in repair of a pre-existing Administrators-owned Windows home is correctly scoped out and must NOT be folded into the rework.

SCOPE OF REWORK: bounded and autonomous. Two lines plus a test, one test, one logbook paragraph. Nothing needs a human decision. Nothing in the Windows fix, its evidence or the green matrix needs redoing. On return the reviewer needs the new tests red-then-green, mypy strict, and a green PR 16 matrix on the new head. Reviewer archetype is read-only: no commit_ack supplied, no code modified, all verification ran in throwaway git worktrees of 8a02e17 and 32737a8 under /tmp, since removed.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-711b96, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-711b96)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-2bd01b, pid=72916, exit=0)
REWORK ROUND (run RUN-260731-711b96, developer). Reviewer RUN-260731-2bd01b requested three changes; none needed a human decision. All three are done and pushed as one signed commit f8b90a5 on ivanopcode/cocoaskills only, PR 16 head. Nothing in the original Windows fix or its green matrix was touched.

REWORK 1 - provision_new_manager_home fails closed again. Restated the condition Path.mkdir(exist_ok=True) carries (CPython: if not exist_ok or not self.is_dir(): raise) inside the except FileExistsError that had replaced it with a bare return. Verified differentially rather than by reading: the same five-shape probe (absent / existing dir / symlink to existing dir / symlink to MISSING dir / regular file) run against a throwaway worktree of 8a02e17 and against the fixed module now agrees on all five rows, and the two rows that had flipped to accept raise LockError again. Correction to the verdict on one detail: provision_new_manager_home itself returned cleanly on the regular-file input, it did not let a raw FileExistsError escape; the traceback the reviewer saw comes from a later mkdir once locking used the accepted non-home. Both are fail-open, so the fix is unchanged, but the new GlobalLock test pins the LockError-not-OSError boundary cli.main depends on.

REWORK 2 - the pairwise-scan-is-gone contract is pinned. New test counts every read of the state a namespace is compared by, plus every parts-tuple equality, at three equally spaced sizes; equal namespace steps must produce equal work steps. Indexed: 156 / 296 / 436 units for 32 / 60 / 88 namespaces. Restored pairwise scan over the same memoised probes: 3500 / 12446 / 26880. Red proof done by actually restoring that scan - it fails this test ALONE, both older canonicalisation tests staying green, exactly as predicted.

REWORK 3 - curator LOGBOOK.md now carries both transferable findings: the quadratic _validate_namespace_independence as a POSIX-latent cost defect (750,620 canonicalisations, 182s of 210s, on macOS) that only a slower syscall made fatal, and the gh run view --json jobs completedAt anomaly (implied ~34min against an actual 8323.00s in the job logs). CocoaSkills LOGBOOK.md carries the two code reworks.

TESTS. Six new tests, all red-then-green verified. tests/test_locking.py: 5 covering every shape the home path can be in, with the two adopted rows as positive controls - without the fix 3 fail and those 2 pass. tests/test_transactions.py: 1 cost-contract test, red only against a restored pairwise scan.

LOCAL EVIDENCE (macOS, CPython 3.11.6), real exit codes, each command run standalone.
  python -m mypy ................................................. exit 0, no issues in 67 source files
  pytest test_locking -k new tests ............................... exit 0, 5 passed
  pytest test_locking -k new tests, fix reverted .................. exit 1, 3 failed 2 passed (red proof)
  pytest test_transactions -k namespace ........................... exit 0, 19 passed
  pytest test_transactions -k namespace, pairwise restored ........ exit 1, 1 failed 18 passed (red proof)
  pytest transactions/locking/adapters/installer_transactions ..... exit 0, 162 passed 4 skipped (156 baseline + 6)
No dedicated linter is configured in pyproject.toml; mypy strict is the project static gate. Full suite and PR 16 matrix on f8b90a5 in flight; results appended when they land.
REWORK COMPLETE (RUN-260731-711b96). PR 16 MATRIX FULLY GREEN ON THE NEW HEAD.

CI run 30641011440, head f8b90a5: conclusion=success, 14/14 jobs - 12 test cells (ubuntu/macos/windows x 3.11/3.12/3.13/3.14), Type check / mypy strict, Build artifacts. Windows cells read from the JOB LOGS, not the API timing field: 1217 passed 152 skipped in 672.01s / 851.24s / 493.38s / 660.22s for 3.11/3.12/3.13/3.14. 32737a8 reported 1211/152 on every Windows cell, so the delta is +6 passed and ZERO new skips - all six new tests actually executed on Windows. Job log confirms the four test_provisioning_* cases and test_locking_refuses_to_bind_* PASSED at 59% and test_namespace_independence_never_walks_the_pairs PASSED at 99%.

FULL LOCAL SUITE: exit 0, 1151 passed 100 skipped in 386.22s, against the 1145/100 baseline at 32737a8 - again exactly +6.

VERIFIED PLATFORM FACT, AND I HAD IT BACKWARDS. CreateDirectoryW on a dangling symlink returns ERROR_ALREADY_EXISTS; it does NOT reparse to the link destination and create a directory there. With ssh win offline (connect to 100.120.84.42 port 22: Operation timed out) I reasoned it would reparse, since a final path component is normally reparsed when FILE_OPEN_REPARSE_POINT is absent, and had a platform-tolerant version of that test staged instead of the strict one. The runner settled it: both symlink tests RAN on windows-latest rather than skipping, so the strict LockError-on-every-platform assertion is correct and is what shipped. No test was weakened to accommodate a platform behaviour that had only been reasoned about.

TOOLING NOTE (corollary to the earlier completedAt anomaly): gh run view --log and gh run view --job <id> --log returned empty or single-job output for this completed run, while gh api repos/.../actions/jobs/<id>/logs returned the full 233 KB. Try the API endpoint before concluding a job log is missing.

DELIVERED. One signed commit f8b90a5 (%G? = G), parent 32737a8, pushed to ivanopcode/cocoaskills only; origin has no other remote and no tags were created. src/csk/locking.py +8/-1, tests/test_locking.py +82, tests/test_transactions.py +107, CocoaSkills LOGBOOK.md +31. Curator LOGBOOK.md carries the cross-project entry with both transferable findings the earlier entry omitted, plus the Windows reparse fact and the gh-logs corollary; it is left uncommitted in the working tree the way previous rounds did, for the orchestrator to land.

Throwaway 8a02e17 worktree removed; stale local probe/BUG-260731-1rldqv-win312 branch deleted.

STILL OPEN, UNCHANGED, DELIBERATELY NOT FOLDED IN: a Windows manager home created by an elevated shell before 7a66c73 stays Administrators-owned and fails closed. Automatic adoption would blunt the drift guard. Explicit opt-in repair vs documented manual re-provisioning remains a product decision.

Artifact: BUG-260731-1rldqv_review-rework.md. Ready for review.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-711b96, pid=83494, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-d23d45, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-d23d45)
REVIEW VERDICT (RUN-260731-d23d45, independent Opus, read-only, round 2): ACCEPTED -> done. Evidence: BUG-260731-1rldqv_review-verdict-round2.md.

All three requested changes delivered in signed commit f8b90a5 (parent 32737a8). Every figure I could measure was re-derived from primary sources rather than carried over, and all of it reproduced exactly.

PROVENANCE. PR 16 head is f8b90a5, the exact commit reviewed; OPEN, base main, MERGEABLE. All six branch commits carry good GPG signatures (%G? = G, one key). origin is ivanopcode/cocoaskills only. NO TAG contains any of the four fix commits and remote heads are just main, the PR branch and a pre-existing feat branch - the probe branches are gone. f8b90a5 diffstat is exactly LOGBOOK.md +31, locking.py +8/-1, test_locking.py +82, test_transactions.py +107.

REWORK 1 CONFIRMED. I re-ran the five-shape probe myself through _prepare_manager_home against throwaway worktrees of 8a02e17, 32737a8 and f8b90a5: the fixed module agrees with the pre-fix baseline on ALL FIVE rows, and the two rows that had flipped to accept raise LockError again. The two adoption rows are untouched, so the condition did not over-tighten. RED PROOF REPRODUCED: reverting only the except FileExistsError body gives exit 1, 3 failed 2 passed, the two controls passing either way. The developer correction to my predecessor is RIGHT and I verified it directly - provision_new_manager_home returned cleanly on the regular-file input at 32737a8; the raw FileExistsError comes from a downstream mkdir once locking used the accepted non-home. Also checked for over-reach: is_dir() swallowing OSError fails closed (safe direction), the mkdir->is_dir window is the same TOCTOU CPython mkdir(exist_ok=True) already carries, and conftest.csk_home calls the function on an ABSENT path so it still takes the create branch and loses no fidelity.

REWORK 2 CONFIRMED. The test measures at 4/8/12 targets = 32/60/88 namespaces; the reworks table is labelled by namespaces. My own measurement: 156/296/436, exact. I also measured with FRESH monkeypatch fixtures per size (the shipped test reuses one fixture across three nested patches) and got byte-identical counts, ruling out the nesting distorting the contract. RED PROOF REPRODUCED INDEPENDENTLY: I restored 98ab7a2 pairwise _namespaces_overlap scan over the same memoised probes - exit 1, 1 failed 18 passed, assert (7364-2128) == (15736-7364). It fails this test ALONE; both canonicalisation cost tests stay green.

GUARD EQUIVALENCE RE-DERIVED, NOT CARRIED OVER. I did not accept the round-1 differential; I ran my own against the pairwise scan I had just restored, over 18 filesystem shapes (plain, dir, child-under-dir, absent, absent-under-absent-parent, hardlink alias, symlink-to-dir, symlink-to-file, dangling symlink, three depths, five unused): all pairs and triples, every same-path-twice duplicate, plus 3000 randomized sets of size 2-6 = 3867 sets, compared on (raised?, exception type NAME). ZERO DISAGREEMENTS. Different harness, different shape pool, same answer as round 1.

NO TEST WEAKENED. git diff 8a02e17 f8b90a5 -- tests/ adds exactly three skip constructs: two host-capability guards inside the NEW symlink tests (the Windows logs prove they did not fire) and one skipif(os.name != posix) on a NEW POSIX-only test. No pre-existing test acquired a skip. Windows skip set-diff main b3a5031 vs f8b90a5: six tests skipped at head and not at base, ALL SIX new-file or new-test - five in test_installer_transactions.py, which does not exist on main, gated by POSIX_BUILD_VECTOR that c4131bd introduced (the PR own baseline, not the fix), plus one new POSIX-only privacy test. Nothing that ran before is skipped now.

AC RE-PROVED AT THE PR HEAD. Run 30641011440 on f8b90a5: success, 14/14. Counts read from raw job logs via gh api .../jobs/<id>/logs: windows 1217 passed 152 skipped on every cell (672.01/851.24/493.38/660.22s), ubuntu 1295/74, macOS 1319/50 - collection totals 1369 on every platform, so no lane collects less. Delta vs 32737a8 is +6 passed and 0 new skips; all six new tests confirmed PASSED BY NAME in the Windows logs (four test_provisioning_* and test_locking_refuses_to_bind_* at 59%, test_namespace_independence_never_walks_the_pairs at 99%). Original signatures grepped to zero in the Windows 3.11 log: transaction target changed while digesting 0, WinError 5 0, cache_publication_invalid 0, FAILED 0, ERROR 0. Every originally-failing file passes on Windows case by case, none skipped wholesale: activation_modes 6, audit_cli 12, closure_install 12, dev_substitution 7, gc 8, global_install 33, hybrid_scope 9, install 52, mcp_dependencies 24, status 2 - all FAILED=0.

REVIEWER-LOCAL (macOS, CPython 3.11.6), real exit codes, throwaway worktree of f8b90a5: mypy exit 0, no issues in 67 source files; locking new tests exit 0, 5 passed; reverted exit 1, 3 failed 2 passed; namespace tests exit 0, 19 passed; pairwise restored exit 1, 1 failed 18 passed; FULL SUITE exit 0, 1151 passed 100 skipped in 371.89s - matching the reported figure exactly against the 1145/100 baseline.

REWORK 3 CONFIRMED. Curator LOGBOOK.md 1810 entry carries both transferable findings the earlier entry omitted (quadratic _validate_namespace_independence as a POSIX-latent cost defect, and the gh completedAt anomaly), plus the CreateDirectoryW reparse fact and the gh-logs corollary. It is uncommitted in the working tree for the orchestrator to land, the way previous rounds and the accepted sibling BUG-260731-27h1yc did.

NON-BLOCKING. (1) A symlink-to-existing-directory home is still adopted un-provisioned, deliberately matching the 8a02e17 baseline; on Windows it skips the DACL stamp and fails closed later at the build cache - same shape as the open product decision, not a new hole. (2) The cost test three nested monkeypatches are sound but subtle; I verified with fresh fixtures. Worth a comment if that helper is extended. (3) STILL OPEN AND CORRECTLY NOT FOLDED IN: a Windows manager home created by an elevated shell before 7a66c73 stays Administrators-owned and fails closed; automatic adoption would blunt the drift guard. Explicit opt-in repair vs documented manual re-provisioning is a product decision. (4) gh run view --log again returned nothing useful while gh api .../jobs/<id>/logs returned the full ~233KB.

FOR THE COMMIT-OWNING MOVER: land PR 16 at head f8b90a5 (base main b3a5031, MERGEABLE, all commits GPG-good); commit the curator LOGBOOK.md working-tree change; no tag or Release was created and none is implied.

Reviewer archetype: no commit_ack supplied. All verification ran in throwaway git worktrees of 8a02e17, 32737a8 and f8b90a5 under /tmp, since removed; no repository file was modified.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-d23d45, pid=97548, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-2dec82.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-2dec82.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_root-cause-and-fix.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_root-cause-and-fix.md) — Root cause analysis of all four Windows failure families and the fixes applied
- [BUG-260731-1rldqv_evidence.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_evidence.md) — Commands run, exit codes, local and windows-latest CI evidence
- [BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-d6e3c5.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-d6e3c5.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_namespace-cost-fix.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_namespace-cost-fix.md) — Root cause, fix, measurements and full green CI evidence for the Windows namespace-validation cost regression
- [BUG-260731-1rldqv_probe-logs.tar.gz](file://BUG-260731-1rldqv/BUG-260731-1rldqv_probe-logs.tar.gz) — windows-latest faulthandler probe logs (3.11/3.12) plus the PR16 green-cell and pre-fix CI logs
- [BUG-260731-1rldqv_spawn-log_-reviewer--reviewer--claude-_RUN-260731-2bd01b.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-reviewer--reviewer--claude-_RUN-260731-2bd01b.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_independent-review-verdict.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_independent-review-verdict.md) — Independent Opus review verdict: changes requested. All AC re-verified from primary sources; one fail-open regression in provision_new_manager_home requires rework.
- [BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-711b96.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-implementer--developer--claude-_RUN-260731-711b96.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_review-rework.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_review-rework.md) — Review rework: fail-closed manager-home provisioning restored, pairwise-scan cost contract pinned, cross-project logbook completed; red/green proofs and green PR 16 matrix on f8b90a5
- [BUG-260731-1rldqv_spawn-log_-reviewer--reviewer--claude-_RUN-260731-d23d45.log](file://BUG-260731-1rldqv/BUG-260731-1rldqv_spawn-log_-reviewer--reviewer--claude-_RUN-260731-d23d45.log) — System spawn log captured by task-board
- [BUG-260731-1rldqv_review-verdict-round2.md](file://BUG-260731-1rldqv/BUG-260731-1rldqv_review-verdict-round2.md) — Independent Opus review verdict, round 2 (RUN-260731-d23d45): ACCEPTED. Rework 1/2/3 re-verified from primary sources, both red proofs reproduced, 3867-set guard differential with zero disagreements, PR 16 run 30641011440 14/14 green on head f8b90a5.

## Created
2026-07-31T01:59:51Z

## Last Update
2026-07-31T15:35:54Z

## Assigned To
[reviewer] reviewer (claude)
