# TASK-260907-187z6x: git-same-source-reinstall-drops-use-and-takeover

## Description
profile install <git-url> --use --takeover after the 9.5 stop accepts both flags, does nothing, and exits 0 saying updated profile. A trunk defect, not stage (c): reproduced identically on origin/main db444157.

## Scope
Stage (c) review cycle 6 (C6-m1, repeat-of C5-M1). The same-source reinstall of a git root delegates to updateLocked with no --use/--takeover activation handling, so the retry the 9.5 stop invites is a silent no-op that reports success. Rework 5 deliberately scoped its fix to path roots and the flags work correctly on a git FIRST install, which is why this is a separate trunk task. The reviewer corrected a false clause in the stage (c) rework-5 report claiming the behaviour is undrivable hermetically here: it drives in about thirty lines using the repository own fixture pattern, a local repo served under a fake canonical identity through a GIT_CONFIG_GLOBAL insteadOf rewrite, as internal/envprofile/network_fixture_test.go and cmd/curator/profile_test.go already do. Cycle 5 had tried file://, which canonicalGit correctly refuses.

## Acceptance Criteria
A same-source reinstall of a git root honours --use and --takeover exactly as a first install does, or refuses with a diagnostic that tells the operator what to run instead. Driven through run() for all four flag combinations on a git root that is current and one that is not, using the insteadOf fixture pattern. A narrowing mutant kills a named test. No path-root behaviour changes.
