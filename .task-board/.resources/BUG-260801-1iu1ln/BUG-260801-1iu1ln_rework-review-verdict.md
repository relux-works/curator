# BUG-260801-1iu1ln rework review verdict

Verdict: changes requested. Route: to-dev.

## Scope and provenance

Reviewed CocoaSkills branch `task/BUG-260801-1iu1ln-lifecycle-observed-traces` at signed commit `afc385f6cb24e12b9ff3ac83bc6d1036f3ea3eef`, preserving signed commit `9362cc8c076a85a49c04c82e76026d6f7473a311` over exact signed merge-base `ba250bfc4dfe104a160eadd5b5f4e340693bf892`. Base, prior commit, and HEAD pass `git verify-commit` with the expected ECDSA signer. The dedicated worktree is clean. No pin, schema-v7, tag, release, claim, CI, pyproject, or changelog surface changed.

The four previously requested sabotage cases now do change their normative cases: omitted skill validation, transient private-build home locking, omitted repair audit gating, and omitted generation-current checking are repaired. Acceptance still fails because three additional scalar fields remain insensitive to their CocoaSkills seams.

## Required findings

### 1. Successful private-build artifact execution is a literal false negative

`tests/protocol_lifecycle_observations.py:1967` assigns `artifacts_executed: False` in `all-misses-stage-and-verify-before-home-lock` without an execution trace.

Adversarial evidence: a wrapper around `installer._build_private_misses` executed both returned operation-private artifacts (`artifact-golden-tool` and `artifact-second-tool`, return code 0) before publication. The observer still reported `artifacts_executed=false`, and the complete observed case remained equal to the normative vector.

This violates the requirement that every normative leaf be backed by observed CocoaSkills behavior.

### 2. GC lock evidence is polluted by a fixture-owned transaction lock

The GC observer accumulates one shared `lock_events` list, explicitly acquires `ManagerHomeLock` for journal preparation at lines 1371-1380, then derives `only_lock` only from non-emptiness and the set of labels at lines 1459-1462. It does not prove that each guardless `gc.collect_runtime` call acquired its own manager-home lock.

Adversarial evidence: four guardless collection calls were redirected directly to `gc._collect_locked` with a no-op lock witness, simulating removal of the public GC lock acquisition. The unrelated fixture transaction still contributed one manager-home event. The observer reported `only_lock=manager-home-mutation-lock`, and the complete GC case remained equal to the normative vector.

### 3. Recovery scan scope is inferred from one journal

The interrupted-recovery fixture creates only `transaction-global-17`. At line 2373 it reports `all-incomplete-journals` whenever any recovery event exists.

Adversarial evidence: `TransactionEngine.recover` was replaced by an implementation that processes only `transaction_ids[:1]`. Its only non-empty inventory was `transaction-global-17`; the observer still reported `scan_scope=all-incomplete-journals`, and the complete recovery case remained equal to the normative vector.

## Required rework

- Instrument artifact execution across the full successful private-build operation and derive `artifacts_executed` from that trace; add a sabotage test that executes a staged artifact and requires case inequality.
- Scope GC lock tracing to each collection invocation. Prove a guardless public call acquires a valid manager-home witness, a guarded call uses the supplied witness, and no project/build lock participates. Exclude fixture-only transaction locks; add an omitted-acquisition sabotage test.
- Construct at least two incomplete journals with distinct transaction/project identities, recover them through one public scan, and verify both restore/removal outcomes. A first-only recovery implementation must change the case.
- Correct the LOGBOOK statement that every named GC/recovery/private-build field is individually backed, then rerun the authenticated conformance, product suites, strict mypy, package, diff, and signature gates in a new signed rework commit.

No external blocker or human-only decision exists. This is ordinary implementation rework.

## Reviewer validation

- Authenticated exact-root `tests/test_protocol_conformance.py`: 835 passed in 265.39s.
- Isolated original four sabotage tests: 4 passed, 831 deselected in 182.16s.
- Focused product regressions: 3 passed in 4.71s.
- Installer/global/currentness/transaction suite: 131 passed in 145.19s.
- Transaction/GC/status suite: 111 passed, 1 platform skip in 17.42s.
- Strict mypy: no issues in 68 source files.
- `compileall`, exact-base `git diff --check`, release-surface guard, and candidate pin guard: exit 0.
- Isolated PEP 517 build produced sdist and wheel `0.12.6.dev39+gafc385f6c`; twine passed both; the lifecycle helper is present in the sdist.
- Final worktree status is clean and HEAD signature verification passes.
