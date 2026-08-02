## Status
to-dev

## Review
none

## Task Class
research

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Verify exact signed head and review complete PR diff
- [x] Audit security portability lock and release guard behavior
- [x] Publish explicit provisional verdict as an outcome resource
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Verify exact signed head and review complete PR diff
- [ ] Audit security portability lock and release guard behavior
- [ ] Publish explicit provisional verdict as an outcome resource
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green at exact head (windows-latest pending)
- [ ] Verify exact signed head and review complete PR diff

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260802-23afc2, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260802-23afc2)
agent completed: [reviewer] reviewer (claude) (exit=1)
spawn run completed: claude (run=RUN-260802-23afc2, pid=70602, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260802-e328bf, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260802-e328bf)
Independent read-only review of CocoaSkills PR #19 at exact head 6e7742f0d28ad95ddd7d8e92364b84062571ad0b (verified via refs/pull/19/head, not branch tip). All 22 commits verify %G?=G, signer oparin@me.com.

PROVISIONAL ACCEPT. No correctness, security, lock, signature or release-guard defect found. Acceptance is contingent on the four windows-latest jobs at exact head reaching terminal success (run 30743353816: ubuntu x4, macos x4, mypy strict all SUCCESS; windows x4 still IN_PROGRESS at review time).

Key audited points:
1. Transaction-recovery relocation is safe. Dry-run stays mutation-free (removed if-not-dry_run guard is subsumed by the pre-existing dry-run early return before the ManagerHomeLock/recover block). Recovery-induced drift raises concurrent_state_change and is retried (attempts=3). Recovery still holds the home lock and runs before publish/commit.
2. except locking.LockError: raise restores the pre-existing EXIT_LOCK contract that the home-lock relocation would otherwise have swallowed into a generic per-project failure. LockOrderError also propagates. Covered by a CLI-level test.
3. New status._runtime_exposes_build_roots fails closed and its path (csk_home/runtime/<name>/<commit>) matches the installer/gc materialization layout exactly, so it cannot silently fail open. build_roots are already validated relative/strict-posix/link-free/disjoint, so no traversal is reachable.
4. Release guards hold. ci.yml is untouched, so the committed curator-spec pin is not advanced. The PR-body ref 432eb2e vs CI-pinned 0c81c1f discrepancy is cosmetic: both carry an identical conformance/v1 tree 36287c9 and the same pinned manifest digest 12e58b82. claims_emitted == [] and rc6_claim_schema is None are asserted.
5. Suite integrity is sound: manifest digest pin, per-file digest verification, absolute/.. rejection, protocol_version assertion, and exhaustive inventory counts (102 in-scope schema cases) that prevent silent case-set shrinkage.
6. The new test mass is not vacuous: every monkeypatched seam is a spy delegating to the real implementation, and meta-tests police the harness (AST scan requiring every hardcoded literal answer to be registered; a forbidden-pattern test banning known lossy proxies).

Non-blocking follow-ups: new security guard is POSIX_BUILD_VECTOR-gated so unexercised on Windows; a few suite reads bypass digest verification; the whole conformance suite is skipif(not CURATOR_CONFORMANCE_ROOT) so dropping that env var from ci.yml would silently skip ~1700 tests; commit 9362cc8 is typed test: but carries three production behaviour changes.

Full evidence: outcome resource BUG-260802-1s021p_pr19-independent-review.md
REVIEW VERDICT: CHANGES REQUESTED — CocoaSkills PR #19 at exact head 6e7742f0d28ad95ddd7d8e92364b84062571ad0b.

All 22 commits verify %G?=G (signer oparin@me.com). Head verified via refs/pull/19/head, not branch tip.

BLOCKING DEFECT — Windows portability fix is incomplete (1 of 5 call sites).

The head commit added _utime_portably() and rewrote exactly one call site (old line 2967 -> new 2989) in tests/protocol_lifecycle_observations.py. Four identical unguarded call sites survive in the same function _observe_gc (def at line 2751), all AFTER the repaired one:
  line 3055  os.utime(rejected_entry, (1, 1), follow_symlinks=False)
  line 3063  os.utime(entry, (young_mtime, young_mtime), follow_symlinks=False)
  line 3071  os.utime(entry, (2, 2), follow_symlinks=False)
  line 3112  os.utime(other_entry, (1, 1), follow_symlinks=False)

They were previously unreachable only because execution died at 2967 first. Reachability verified, not assumed:
- No os.name / sys.platform / pytest.skip gate anywhere in lines 2780-3120; lines 2999-3055 are straight-line code with no early return and no try/except covering the span.
- _observe_gc demonstrably runs on windows-latest: the previous run traceback is inside this function, and gc_cases:locked-mark-and-sweep-compiled-cache is in that run failure list.
- The active os.utime monkeypatch (line 638 -> observed_os_utime) forwards follow_symlinks verbatim to real_os_utime (lines 582-605), so NotImplementedError propagates.

Prior run 30737293076 (9228054) failed windows x4 with exactly: NotImplementedError: utime: follow_symlinks unavailable on this platform at protocol_lifecycle_observations.py:2967. Step-level inspection confirms genuine Run tests failures, not cancellations. Because all 32 lifecycle cases draw from one shared cached observation build, this single raise fails the whole lifecycle suite.

PREDICTION (falsifiable): run 30743353816 windows x4 at this head will fail with the same NotImplementedError at line 3055. At review time ubuntu x4, macos x4 and mypy strict are SUCCESS; windows x4 still IN_PROGRESS (~1h50m elapsed; prior Windows jobs took ~2h56m). The static evidence above does not depend on that confirmation.

REQUESTED CHANGES:
1. Route all four call sites through the existing _utime_portably() helper.
2. Add a source-level regression guard so this does not need a sixth CI round: extend the existing self-policing pattern (test_rc6_lifecycle_observer_rejects_known_lossy_proxy_forms) to fail if follow_symlinks=False appears in protocol_lifecycle_observations.py outside the body of _utime_portably. Fixing one call site per ~3h Windows CI round is the real cost driver.

AUDITED SOUND — no rework needed:
- Transaction-recovery relocation. Dry-run stays mutation-free (the removed if-not-dry_run guard is subsumed by the pre-existing dry-run early return before the ManagerHomeLock/recover block). Recovery-induced drift raises concurrent_state_change and is retried (attempts=3). Recovery still holds the home lock and precedes publish/commit. TransactionError still surfaces as a per-scope failure.
- Lock semantics. except locking.LockError: raise restores the pre-existing EXIT_LOCK contract that moving the home lock into the commit phase would otherwise have swallowed into a generic per-project failure. LockOrderError also propagates. Covered by a CLI-level test.
- Security. New status._runtime_exposes_build_roots fails closed and its path csk_home/runtime/<name>/<commit> matches the installer/gc materialization layout exactly, so it cannot silently fail open. build_roots are already validated relative/strict-posix/link-free/disjoint, so no traversal is reachable.
- Release guards. ci.yml untouched, so the committed curator-spec pin is not advanced. The PR-body ref 432eb2e vs CI-pinned 0c81c1f discrepancy is cosmetic: both carry an identical conformance/v1 tree 36287c9 and the same pinned manifest digest 12e58b82. claims_emitted == [] and rc6_claim_schema is None are asserted.
- Suite integrity. Manifest digest pin, per-file digest verification, absolute/.. rejection, protocol_version assertion, exhaustive inventory counts (102 in-scope schema cases) preventing silent case-set shrinkage.
- Anti-vacuity. Every monkeypatched seam is a spy delegating to the real implementation; meta-tests police the harness (AST scan requiring every hardcoded literal answer to be registered; forbidden-pattern test banning known lossy proxies).
- csk global upgrade fetch fix is correct and symmetric with project scope; covered by a transitive-closure test.

NON-BLOCKING FOLLOW-UPS: new security guard is POSIX_BUILD_VECTOR-gated so unexercised on Windows; a few suite reads bypass digest verification; the whole conformance suite is skipif(not CURATOR_CONFORMANCE_ROOT) so dropping that env var from ci.yml would silently skip ~1700 tests; commit 9362cc8 is typed test: but carries three production behaviour changes.

This verdict is head-bound. Re-review is required at the new exact head after the fix.

Full evidence: outcome resource BUG-260802-1s021p_pr19-independent-review.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260802-e328bf, pid=75803, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260802-1s021p_spawn-log_-reviewer--reviewer--claude-_RUN-260802-23afc2.log](file://BUG-260802-1s021p/BUG-260802-1s021p_spawn-log_-reviewer--reviewer--claude-_RUN-260802-23afc2.log) — System spawn log captured by task-board
- [BUG-260802-1s021p_spawn-log_-reviewer--reviewer--claude-_RUN-260802-e328bf.log](file://BUG-260802-1s021p/BUG-260802-1s021p_spawn-log_-reviewer--reviewer--claude-_RUN-260802-e328bf.log) — System spawn log captured by task-board
- [BUG-260802-1s021p_pr19-independent-review.md](file://BUG-260802-1s021p/BUG-260802-1s021p_pr19-independent-review.md) — Independent read-only review verdict for CocoaSkills PR #19 at exact head 6e7742f: CHANGES REQUESTED. Windows portability fix repairs 1 of 5 identical os.utime(follow_symlinks=False) call sites; the other 4 are on the same unguarded path in _observe_gc. Correctness, lock/EXIT_LOCK semantics, runtime build-root security guard, signatures, release guards and suite integrity all audited sound.

## Created
2026-08-02T10:22:56Z

## Last Update
2026-08-02T12:06:32Z

## Assigned To
[reviewer] reviewer (claude)
