# BUG-260801-1iu1ln reviewer verdict

Verdict: changes requested. Route: to-dev.

## Scope reviewed

Reviewed signed CocoaSkills commit 9362cc8c076a85a49c04c82e76026d6f7473a311 on branch task/BUG-260801-1iu1ln-lifecycle-observed-traces. The worktree was clean. HEAD parent and merge-base both resolve to the required signed base ba250bfc4dfe104a160eadd5b5f4e340693bf892. Both commits pass git verify-commit with the expected ECDSA signer. No pin, schema-v7, tag, release, claim, CI, or pyproject file is in the committed diff.

## Baseline validation

The submitted baseline is green but does not meet the product-seam sensitivity acceptance criterion.

- Authenticated exact-root lifecycle plus scalar mutation subsets: 410 passed in 57.74s. This includes all 32 canonical cases and all 378 scalar-vector mutations.
- Focused product regressions: 3 passed in 7.53s.
- Strict mypy: no issues in 68 source files.
- compileall and git diff --check: exit 0.
- Signed branch stayed clean after review.

## Required findings

### 1. Planning gate inventory is manufactured from labels and survives removal of a real gate

tests/protocol_lifecycle_observations.py lines 1172-1314 define the normative gate names, append several names after broad helper calls, replace audit and registry gates with successful stubs, and emit the entire required list when the aggregate install succeeds. It does not bind each required gate or each failure_at_any_gate leaf to a distinct observed seam and failure probe.

Adversarial evidence: replacing installer._validate_skills with a no-op returning an empty issue list was invoked once. The observer still returned build-eligible, emitted all 11 required gates, and the complete observed planning case remained byte-for-byte equal to the normative vector.

### 2. Private-build failure effects and lock acquisition have a confirmed false negative

tests/protocol_lifecycle_observations.py lines 1487-1526 trap only cache publication, target swap, and GC, then manufacture eight forbidden-effect labels. The manager_home_lock_acquired field samples locking._STATE.home only after installer.install has returned, so a transient acquisition is invisible.

Adversarial evidence: a sabotaged private-build helper deliberately acquired and released ManagerHomeLock once on the second-build failure path before continuing. The observer reported manager_home_lock_acquired=false, emitted the full forbidden-effect list, reproduced all four events, and the complete observed case still equaled the normative vector.

This directly violates the acceptance requirement for private-build events and effects.

### 3. Repair pipeline labels survive removal of the audit gate

tests/protocol_lifecycle_observations.py lines 1706-1717 and 1837-1892 define the expected repair pipeline and append six stages whenever one high-level gate function returns, plus three stages whenever publication is entered. This does not prove the individual required stages or forbidden shortcuts.

Adversarial evidence: replacing installer.audit_pipeline.gate_plans with a no-op successful GateResult was invoked 25 times. The observer still emitted the complete 10-stage required pipeline, reported rebuilt-and-journaled, and the complete repair case remained equal to the normative vector.

### 4. Recovery restart guard is a literal and survives removal of the guard

tests/protocol_lifecycle_observations.py line 1667 assigns restart_if_plan_assumption_changed=true without creating a changed-generation scenario. The interrupted-journal case likewise assigns expected_action at line 1602 without using the collected restore-order evidence to establish the full action.

Adversarial evidence: replacing installer._assert_generation_current with a no-op was invoked once. The observer still reported restart_if_plan_assumption_changed=true and the complete recovery case remained equal to the normative vector.

### 5. The same pattern remains in other normative inventories

The dry-run effect taxonomies at lines 637-689 are returned wholesale after one aggregate read-only predicate at lines 818-845; many named effects have no corresponding trap. Status validated fields are emitted from _STATUS_VALIDATED whenever one current result is observed at lines 1693-1749. GC reports artifact_executed, entry_adopted, only_lock, and sweep-requirement labels partly as constants at lines 1012-1048. Cross-project shared_transactions_serialized is inferred from two sequential transactions rather than concurrent serialization at lines 582-603.

These are not merely stylistic concerns. They allow product behavior to violate normative scalar leaves while the new 378 vector-mutation test remains green. That test mutates only the expectation object and therefore proves complete equality sensitivity, not that every observed value is independently sourced from CocoaSkills behavior.

## Rework required

- Replace copied protocol taxonomies with a fail-closed field-to-probe classification in which each scalar leaf is backed by a specific CocoaSkills trace, state transition, negative probe, or reusable existing test helper.
- Add product-seam sabotage tests, not only vector-side mutation tests. At minimum the four adversarial cases above must fail: omitted skill validation, transient manager-home lock acquisition, omitted repair audit gate, and omitted generation-current guard.
- Instrument every private-build forbidden effect and lock acquisition as events over the whole operation, not final thread-local state.
- Exercise recovery generation change and restart, restore ordering and preimage guard; exercise cross-project serialization concurrently; bind dry-run, GC, currentness validation, and repair shortcut inventories to explicit probes.
- Preserve the three focused product fixes if they remain valid, rerun the authenticated exact-root and strict validation gates, update the inaccurate LOGBOOK claim that seam sabotage prevents self-consistency fallbacks, and produce a new signed rework commit.

No external blocker or human-only decision exists. This is ordinary implementation rework.