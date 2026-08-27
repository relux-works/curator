# BUG-260801-1iu1ln cycle-10 reviewer verdict

## Verdict

Changes requested; route to implementation rework. This is not an external blocker.

## Passing evidence

- Dedicated worktree `/Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260801-1iu1ln/worktree` is clean on branch `task/BUG-260801-1iu1ln-lifecycle-observed-traces`.
- HEAD `80b5b1673db170e3db9be349c3649d9b4e03d520` and exact base `ba250bfc4dfe104a160eadd5b5f4e340693bf892` verify with the expected ECDSA signature; merge-base is exact.
- Independent exact-root canonical/scalar/classification gate passed: 417 passed, including all 32 lifecycle cases and all 378 scalar mutations rejected.
- Independent strict project mypy passed: no issues in 68 source files.
- Exact-base `git diff --check`, clean-tree, signature, and merge-base checks passed.
- The retained cycle-10 `fchmod` native-loader regression closes the exact cycle-9 hole.

## Acceptance-blocking survivor

The cycle-10 trusted primitive records only CPython `os.chmod` audit events (`tests/protocol_lifecycle_observations.py`, audit filter near line 1177), and publication completeness only reconciles that permission-event count with the ordinary root seal (near line 1429). Other real post-handoff metadata mutations are therefore invisible when performed inside the replaceable native callable and restored before it returns.

An independent POSIX probe supplied a delegating `renameat2`/`renameatx_np` callable through `cache_posix.ctypes.CDLL`. After the real no-replace rename succeeded, it used previously captured `os.stat` and `os.utime` callables with the destination dir-fd to advance the live 64-hex root mtime and then restore the original atime/mtime before returning. The probe exited 42 only when at least one two-call mutation sequence executed and the complete `publish-complete-immutable-entry-under-home-lock` observation still equaled the normative vector; it exited 42. Root ctime is intentionally normalized by `_same_tree_across_atomic_rename`, so final snapshots do not expose the restored sequence. This violates the same immutable-entry requirement and proves the observed complete case remains a lossy proxy.

## Required rework

1. Retain a regression reproducing captured `os.utime` mutate-and-restore inside the POSIX native-loader callable, plus the Windows-equivalent boundary or an explicit fail-closed classification where that platform cannot be exercised.
2. Make `publish-complete-immutable-entry-under-home-lock` differ for any post-handoff mutation of the live entry, not only chmod/fchmod. The primitive-level witness must fail closed across relevant content, metadata, ownership, timestamp, xattr/flags, and namespace mutation classes supported by the platform; do not merely add a one-off expected count for `os.utime`.
3. Audit the process-global audit-hook design for complete event classification and retain a fail-closed test so a newly relevant mutation event cannot silently become unobserved.
4. Rerun the native publication barrier, inherited sabotage probes, exact-root canonical/scalar/classification gate, full authenticated conformance, strict mypy, diff/provenance, packaging/release guards, and relevant Windows lane.
5. Preserve the exact signed base and release-surface exclusions; produce a new signed commit and task-scoped outcome artifact, and record this finding in the CocoaSkills logbook during producer rework.