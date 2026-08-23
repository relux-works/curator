# BUG-260801-1iu1ln cycle-4 reviewer verdict

Verdict: changes requested. Route: to-dev.

## Scope and provenance

Reviewed the clean dedicated worktree at `/Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`, branch `task/BUG-260801-1iu1ln-lifecycle-observed-traces`, signed tip `963d224c9bb3d6fc274b9dbfbac4bdcafd243c2b` directly atop signed `bb2e5801d3f4c31e48018028097b525238126b33`. Merge-base is the exact required signed base `ba250bfc4dfe104a160eadd5b5f4e340693bf892`. Signature verification, clean status, no-tag check, committed diff check, and forbidden release-surface guard pass. This run is not goal-bound.

The three submitted cycle-4 probes pass and the prior argv-element, same-basename identity, and in-place GC mutation holes are fixed. Acceptance still fails because seven independent product-seam sabotages preserve complete normative case equality.

## Acceptance-blocking findings

### 1. Atomic publication is inferred from final state only

At `tests/protocol_lifecycle_observations.py:319-338`, `publication=atomic-complete-directory` is derived only from a final cache hit and final bytes. A sabotage around the real POSIX `_rename_noreplace` exposed an incomplete live cache-key directory and file, removed it, then allowed the real rename. It produced 38 transient partial live exposures across the lifecycle observation. The complete `publish-complete-immutable-entry-under-home-lock` object remained equal. Atomicity therefore has no event or concurrent-reader evidence.

### 2. Cross-project private-build overlap never reaches the build seam

The cross-project helper overlaps project locks and consumer commits but never runs concurrent `_build_private_misses` calls. A process-wide mutex was placed around every private-build call, making overlap impossible. The observer made 37 private-build calls with maximum concurrency 1, yet the complete `two-project-success-preserves-both-consumers` case still matched with `private_builds_may_overlap=true`. The scalar is a lock-fixture proxy, not observed private-build behavior.

### 3. Registered-consumer GC liveness is hidden by overlapping roots

At lines 1438 and 1622-1630, `registered-consumer` is emitted from ledger membership while the same entry is simultaneously retained by a configured marker root. After preserving the initial registration, a sabotage made every later GC mark phase ignore the consumer registry. All 37 subsequent consumer loads were empty, yet the complete `locked-mark-and-sweep-compiled-cache` case remained equal and still claimed `registered-consumer` as a mark root. Each normative root needs an isolated causal probe or exact reference-set trace.

### 4. Recovery backup retention uses `any` instead of exact per-target evidence

At lines 2564 and 2660, `backups_retained_until_recovery_succeeds` is backed by `any(path.exists() for path in backup_paths)`. A sabotage deleted the already-restored consumer backup while another backup remained. Recovery completed, both journals were removed, and the complete recovery case remained equal. The plural retention guarantee needs all expected backup identities and digests through successful recovery cleanup.

### 5. Repair can execute an untrusted artifact and still claim every forbidden shortcut

The repair observer records broad pipeline calls but has no process-execution trap. Its `forbidden_shortcuts` values at lines 3463-3504 are inferred from a later successful rebuild, not from absence of adoption, chmod, trust, or execution events. A sabotage executed the untrusted protected `tool` artifact successfully before the real rebuild. The complete `repair-rebuilds-invalid-compiled-entry` object remained equal with every shortcut still reported as forbidden.

### 6. Currentness read-only fields miss transient persistent mutations

The clean-status path at lines 3015-3027 and every matrix path at lines 3252-3258 compare only before and after trees. Cache and build calls are trapped, but permission mutations are not. A sabotage performed chmod-and-restore on the protected cache entry during status collection. The clean case tolerated 2 permission mutations and the 14-condition matrix tolerated 28; both complete objects remained equal with `mutations=[]`, `adopt=false`, `quarantine=false`, and `repair=false`.

### 7. Rollback success and reverse order do not verify restored target state

At lines 3942-3998, restore events and lock state are recorded, but the six live transaction targets are never compared with their exact preimages after rollback. A sabotage called the real rollback for every target and then corrupted each restored live target. All six corruptions survived while the complete `reverse-rollback-under-home-lock` object remained equal with `result=rolled-back` and exact reverse order.

## Required rework

- Bind atomic publication to a trace or concurrent-reader negative probe that rejects any partial live destination exposure.
- Run actual concurrent private builds for distinct projects and prove overlap before the shared manager-home handoff; verify the exact shared cache identity from both build plans.
- Isolate every GC mark root or compare the exact reference set produced by the GC mark phase; preserve the ignored-consumer-registry sabotage.
- Require every expected recovery backup, exact target association, and digest until the corresponding target is safely restored and recovery cleanup succeeds.
- Add repair process and persistent-mutation traps. Derive forbidden shortcuts from observed absence of candidate reuse, chmod/adoption, self-consistent receipt trust, and artifact execution rather than from eventual rebuild success.
- Trap permission and other transient persistent mutations in clean currentness and every failure-matrix probe.
- Compare every rollback live target byte-for-byte and mode-for-mode to its exact preimage after rollback, and preserve the post-restore corruption sabotage.
- Preserve all ten submitted sabotage probes and add these seven regressions. Correct the LOGBOOK claim after the causal audit is genuinely complete, rerun exact-root, strict mypy, diff, build, signature, and release-surface gates, and produce a new signed commit.

## Reviewer validation

- Submitted cycle-4 sabotage probes: 3 passed in 120.81s.
- Authenticated exact-root canonical 32, exhaustive 378 scalar mutations, literal classification, lossy-proxy guard, and process-path helper: 413 passed in 40.96s.
- Seven independent reviewer sabotage runs all returned complete case equality; concrete counts are recorded above.
- Strict configured mypy: no issues in 68 source files.
- Exact-base diff check, forbidden release-surface guard, signature verification, exact merge-base, clean worktree, and no-tag check: exit 0.
- The signed producer evidence for isolated PEP 517 build, Twine, and sdist membership was inspected and remains green; a repeated package build is not material to this logical rejection.

The 378 vector-leaf mutation test proves only that changing an expected literal differs from the fixed observed object. These surviving product-seam sabotages prove it does not establish causal normative coverage. No external blocker or human-only decision exists; this is ordinary implementation rework.