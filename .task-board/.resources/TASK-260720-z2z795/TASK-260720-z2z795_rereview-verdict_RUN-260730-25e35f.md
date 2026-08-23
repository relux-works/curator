# TASK-260720-z2z795 independent rereview verdict

- Run: `RUN-260730-25e35f`
- Role: reviewer
- Date: 2026-07-30
- Verdict branch: **changes requested -> `to-dev`**
- Goal-bound check: none; the run is not goal-bound
- Commit acknowledgement: not supplied (reviewer-archetype run)
- Source modification by reviewer: none

## Reviewed bytes and contract

The review covered the exact unstaged handoff in:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

The branch and remote task ref both point to
`5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`. The accepted base
`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor. The
reviewed source hashes are:

```text
dd8b60fa544e0663e55d6da06c97d578175be8b19f97769a4b5502807953fdad  src/csk/locking.py
7fd2256a78c84096ce76b60987d3ba80135adb57741961ccc77d2c6ec8f16dd2  src/csk/transactions.py
2149953a8d2fb1ad781b0ed27e084ffab7cdc334f0256cb4eb7ede3e9bc284d7  tests/test_locking.py
3f5a158361ed10f0538089e613185d7407101e06fd958afeec59568dac956c7c  tests/test_transactions.py
```

The review used the task acceptance criteria, Curator manager lock/transaction
specification, the accepted Go behavior under
`curator/.temp/TASK-260720-1pvfj5/rework/composite`, all prior task verdicts,
and the current producer handoff.

## Blocking findings

### 1. A lock for one manager home authorizes mutation in a different home

`HomeLockWitness` exposes only `assert_held()` (`transactions.py:62-63`).
`TransactionEngine` independently resolves its configured home
(`transactions.py:157-162`), while `prepare`, `commit`, `recover`, and journal
reference enumeration only ask whether the supplied witness is held. They never
bind that witness to the engine's canonical manager-home identity.

A deterministic temporary-directory probe held `ManagerHomeLock(home-a)` and
passed it to `TransactionEngine(home-b)`. Prepare and commit succeeded, changed
the target to the desired bytes, and wrote/recovered the home-b journal namespace
without a home-b lock:

```text
wrong_home_commit_succeeded=True
wrong_home_lock_for_engine_exists=False
```

This violates the single manager-home mutation-lock requirement. A genuine
home-b owner can concurrently hold `home-b/.lock`, so the witness check does not
establish mutual exclusion for the state being mutated.

Required rework:

- Make the witness carry an immutable canonical manager-home identity and reject
  a mismatch before any journal or target mutation/read that assumes exclusive
  mutation ownership.
- Apply the check consistently to prepare, commit, recovery, journal reference
  enumeration, and every public home-locked operation.
- Add a regression using two real `ManagerHomeLock` instances that proves a
  home-a witness cannot create a home-b journal, sidecar, target mutation, or
  lock-path residue.

### 2. A managed symlink target is silently redirected into its referent

`MutableTarget` has no target-kind discriminator (`transactions.py:66-74`).
Journal construction resolves the live final component
(`transactions.py:673`) and the desired final component
(`transactions.py:680`) before `digest_path` can reject a link. The resulting
journal owns the referent path, not the caller-supplied directory entry.

The deterministic probe supplied a live managed symlink to a regular-file
referent:

```text
journal_live_is_referent=True
managed_entry_still_symlink=True
referent_was_replaced=True
```

This can overwrite storage outside the manager-owned entry while reporting a
successful target commit. It also diverges from the accepted behavior reference:
`internal/transaction/types.go:66-81` defines `KindEntry` specifically for
manager-owned adapter symlinks and preserves the final path component. The
Python project currently creates adapter and shim symlinks
(`src/csk/adapters.py:153-166`, `src/csk/shims.py:173-191`), so this is not a
theoretical target shape.

Required rework:

- Add the stable target-kind interface from the accepted contract and implement
  entry-target digest, staging, backup, commit, rollback, cleanup, and namespace
  validation without resolving the final path component.
- If entry targets cannot be implemented in this slice, fail closed on a live or
  desired final-component symlink before creating a journal or sidecar. Do not
  silently reinterpret it as a bytes target.
- Add success, rollback, recovery, stale-preimage, and alias tests for symlink
  entry targets, including proof that the referent is never mutated.

### 3. Recovery deletes foreign bytes that reappear after cleanup was durably removed

Once a cleanup record is durably in `state="removed"`, both its canonical sidecar
and cleanup tomb were absent before that state was saved. A later reappearance is
therefore foreign state and must be preserved. Instead,
`_discard_sidecars` accepts both `"tombed"` and `"removed"` at
`transactions.py:1215`, and if the tomb exists it validates and deletes it at
`transactions.py:1220-1243`. Exact digest equality is treated as renewed
ownership even though ownership was already durably relinquished.

The probe crashed after `journal_tombed`, recreated the previously removed backup
tomb with bytes and mode matching the recorded digest, then invoked cross-project
recovery:

```text
reappeared_tomb_matches_recorded_digest=True
reappeared_tomb_exists_before_recovery=True
reappeared_tomb_exists_after_recovery=False
journal_tomb_exists_after_recovery=False
```

Recovery silently deleted the recreated path. This violates the stale-preimage
and concurrent-consumer safety rule: a completed removal record cannot reclaim a
path merely because another producer later created identical bytes there.

Required rework:

- For `state="removed"`, require both the canonical sidecar and its tomb to be
  absent. Any reappearance must raise corruption before mutation and preserve all
  witnesses.
- Keep the current resumable deletion behavior only for `state="tombed"`.
- Add a regression at the final journal-removal crash boundary for exact-digest
  and different-digest reappearance, for both file and directory sidecars.

### 4. Windows first-use identities are not recanonicalized after missing parents are created

For a multi-component missing suffix,
`_canonicalize_windows_identity` preserves only the first missing component
according to the existing parent's case-sensitivity and then unconditionally
sets later components to case-insensitive (`locking.py:182-187`). Lock objects
capture that provisional identity in their constructors
(`locking.py:405-415`, `531-538`, and `602-605`). Acquisition later creates the
missing parent at `locking.py:282`, but none of the lock classes recanonicalizes
the new physical path before selecting process-state ownership.

Pure deterministic routing probes show both the collision and the identity
change once case-sensitive children exist:

```text
windows_provisional_upper=C:\SENSITIVE\Parent\CHILD
windows_provisional_lower=C:\SENSITIVE\Parent\CHILD
windows_nested_missing_leaf_identities_equal=True
windows_provisional_identity=C:\SENSITIVE\Parent\CHILD
windows_post_creation_identity=C:\SENSITIVE\Parent\Child
windows_identity_changes_after_creation=True
```

This can split process-wide cross-class lock-order accounting between objects
constructed before and after first creation. The accepted reference explicitly
closes this gap by creating the configured home and then recanonicalizing before
choosing the lock root and process state
(`internal/managerlock/managerlock.go:70-91`).

Required rework:

- Recanonicalize the manager home after creating the missing home/lock parents
  and before the final process-state reservation and physical lock namespace are
  selected.
- Add a deterministic before/after-creation regression and a real Windows
  per-directory-case-sensitive concurrency regression covering home, project,
  and build lock classes.

## Validation result

All requested existing gates pass on the reviewed bytes:

- current post-migration regressions: 18 passed
- preserved migration regressions: 14 passed
- corrected prior lock-integrity regressions: 14 passed
- earlier reviewer regressions: 14 passed
- contract-targeted regressions: 13 passed
- focused locking/transaction pytest: 105 passed, 1 skipped
- full pytest: 675 passed, 20 skipped
- strict mypy: no issues in 57 source files
- Ruff lint and format checks: passed
- package sdist and wheel build: passed
- `git diff --check`: passed

The green suites do not exercise the four unsafe boundaries above. Detailed
commands, timings, build hashes, and probe output are attached in
`TASK-260720-z2z795_rereview-validation_RUN-260730-25e35f.log`.

## Verdict

Changes are requested and the task must return to `to-dev`. These are
deterministic, locally reproducible implementation defects within the task's
existing scope, not an external blocker or a human-only tradeoff. A fresh
independent reviewer cycle is required after the fixes and regressions are
attached.
