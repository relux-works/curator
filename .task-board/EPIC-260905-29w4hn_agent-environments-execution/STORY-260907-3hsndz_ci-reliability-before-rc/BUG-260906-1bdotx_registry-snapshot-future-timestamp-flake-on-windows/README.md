# BUG-260906-1bdotx: registry-snapshot-future-timestamp-flake-on-windows

## Description
internal/install TestRegistryAttestationLandsInMarker fails on the windows-latest lane alone with registry test-reg snapshot timestamp is too far in the future, then every trusted audit registry served a tampered snapshot. The same test on the same head passes elsewhere, and the stage that surfaced it does not touch internal/install. The fixture writes created_at with RFC3339, which truncates downward and cannot produce a value ahead of the clock that wrote it, while the gate compares against time.Now() at the call site with a 300-second default skew — so a monotonic reading cannot fire it and the real mechanism is unknown. Establish it from evidence before changing anything, and do not weaken a gate whose whole purpose is catching a snapshot dated in the future, which is what a rollback or equivocation attack looks like.

## Scope
curator repository, branch fix/snapshot-timestamp-flake in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-snapshot-flake, base 919e2e9c74eb4878a648dae77de85a845fe32e1a. Files: internal/install, internal/registry and their tests. The gate itself is internal/registry/snapshot.go:158; the fixture is internal/install/registry_e2e_test.go:44; the policy default is internal/config/config.go:41.

## Acceptance Criteria
The mechanism behind the future-dated snapshot is established from evidence and stated, or its absence is stated plainly with the smallest deterministic fix proposed instead. The gate still refuses a genuinely future-dated snapshot: a narrowing mutant that admits exactly one such snapshot fails a named test, and that test exists after this change whether or not it existed before. The clock-skew tolerance is not widened and the assertion is not removed; if a tolerance is part of the fix it follows from the established mechanism and is stated as a bound. No Windows result is claimed that was not measured on a hosted runner.
