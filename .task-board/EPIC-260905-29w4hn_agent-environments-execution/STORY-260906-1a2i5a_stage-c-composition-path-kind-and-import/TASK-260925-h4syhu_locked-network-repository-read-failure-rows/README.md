# TASK-260925-h4syhu: locked-network-repository-read-failure-rows

## Description
Split from TASK-260907-2as5sx by the rev9 loop-detector response (one leaf per surface). SURFACE: internal/install/draftsources.go lockedNetworkRepository checkout classification — absent vs Lstat read failure (:228-230) vs present-but-unusable (:234-236) — and its platform-case ledger rows. Surface-table rows: (1) absent → fallback; (2) Lstat failure → typed manager_state_unreadable, no fallback (POSIX: regular file on the path, blocked/child → ENOTDIR); (3) present-but-unusable → stateread.UnusableError; (4) Windows: a real Lstat failure input if one exists, else a platform-cases/skip-classes ledger row with the exact reason; state whether ERROR_PATH_NOT_FOUND under a regular file counting as absent is correct. Coverage-map obligation: reviewer mutant M1 (continue as the first statement of the Lstat error branch) killed; each branch has its own subtest. Revision budget: 2. This leaf also carries the Story candidate: every other path of the Story delta must stay identical to TASK-260907-2as5sx revision 9 (accepted content of rev7 + rev8 refresh), verified by the reviewer.

## Scope
(define task scope)

## Acceptance Criteria
M1 killed with real exit codes; rows 1-3 on all OSes, row 4 real or ledgered; all other Story paths identical to 2as5sx rev9; hosted gate green on all lanes; no CHANGELOG edit
