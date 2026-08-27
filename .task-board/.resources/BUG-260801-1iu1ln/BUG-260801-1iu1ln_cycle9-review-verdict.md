# BUG-260801-1iu1ln cycle-9 reviewer verdict

## Verdict

Changes requested; route to implementation rework. This is not an external blocker.

## Passing evidence

- Dedicated worktree remains clean on branch task/BUG-260801-1iu1ln-lifecycle-observed-traces.
- HEAD dbd23cbf24602ccf5125ca78510d9eeb002d0bae and exact merge base ba250bfc4dfe104a160eadd5b5f4e340693bf892 both verify with the expected ECDSA signature; merge-base is exact and HEAD has no tag.
- Independent exact-root canonical/scalar/classification gate: exit 0, 417 passed and 452 deselected; all 378 scalar mutations remain rejected.
- Independent strict mypy: exit 0, no issues in 68 source files.
- Exact-base git diff --check: exit 0; source worktree remained clean.
- Submitted cycle-9 regression, full conformance, related-suite, build, Twine and release-surface evidence is internally consistent.

## Acceptance-blocking survivor

The new raw atomic destination witness observes only after the callable supplied by cache_posix.ctypes.CDLL returns. An independent POSIX probe replaced that loader with a delegating wrapper whose renameatx_np or renameat2 callable performed the real native no-replace rename, opened the resulting 64-hex live root through captured descriptor APIs, toggled one permission bit with captured fchmod, restored the original mode, and then returned. The operation executed 42 such mutations. Despite the transient post-handoff mutation, observe_manager_lifecycle_case returned the complete normative cache_publication_cases case unchanged: {mutations: 42, survived: true}.

This is the same normative requirement targeted by the cycle-9 repair: the published entry must remain immutable after the actual atomic handoff. The current observer wraps the replaceable loader result, so code inside that result can mutate and restore state before raw_atomic_destinations is sampled. The new root-fchmod regression passes because its sabotage is one wrapper farther out at _rename_noreplace; it does not close this adjacent loader-level gap.

## Required rework

1. Add a retained regression reproducing a mutating/restoring rename function supplied through cache_posix.ctypes.CDLL, with the Windows-equivalent native API boundary covered or explicitly fail-closed where that platform cannot be exercised.
2. Make the complete publication case differ when mutation occurs after the actual no-replace handoff but before the replaceable native callable returns. The evidence must anchor at a defensible trusted primitive or persistent operation witness rather than merely moving another return-time wrapper inward.
3. Rerun the cycle-9 barrier, all inherited sabotage probes, the 417 canonical/scalar/classification gate, full exact-root conformance, strict mypy, diff/provenance and relevant platform gates.
4. Preserve the exact signed base and all release-surface exclusions; provide a new signed commit and task-scoped outcome artifact.
