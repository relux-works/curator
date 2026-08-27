# BUG-260823-1vx45a: managerlock-windows-tempdir-contention

## Description
The managerlock subprocess helper reports blocked whenever its context deadline expires, so a lock that is merely slow to take is indistinguishable from a lock another process holds. Windows CI observed the false positive as: independent build key helper = "blocked", want acquired.

## Scope
internal/managerlock test harness: the helper deadline in TestManagerLockHelper and the parent expectations in TestSubprocessContentionAndIndependentProjects and TestSubprocessBuildKeyDeduplicationAcrossProjects. No production code is implicated.

## Acceptance Criteria
An expected-acquired outcome cannot be defeated by a slow host. A deterministic regression proves that a deliberately tiny helper deadline yields blocked on an uncontended lock, so the conflation cannot return unnoticed. Both named tests pass repeatedly on a native Windows host and on the Windows CI lane.
