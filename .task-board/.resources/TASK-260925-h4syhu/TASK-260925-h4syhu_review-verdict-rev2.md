# TASK-260925-h4syhu review verdict — CR rev2: ACCEPTED

Reviewer: claude-opus-5-5 (low). Shell: zsh/bash with `set -o pipefail`, darwin. The worktree tree (temp index) = candidate tree a07dad19. Base is ab34556e.

## F1/M1 (draftsources.go:228-230)
- The candidate test `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` passes on darwin. It has 3 separate subtests (absent → fallback, lstat failure → typed stateread KindUnreadable/manager_state_unreadable plus os.PathError on the exact path, present-but-unusable → UnusableError). Exit code 0.
- With M1 (`continue` as the first statement of the Lstat error branch) applied in a disposable archive copy, `lstat_failure_stops_fallback` FAILS (`error = <nil>`). Exit code 1. **M1 is KILLED.**
- Survived before: base ab34556e has no test that references lockedNetworkRepository (`git grep` over *_test.go finds 0 files), so nothing on base could kill M1.
- Bound: the test calls the helper directly and does not go through the install entry. This is acceptable because the surface is the helper's classification, and the production call site draftsources.go:211 simply propagates the error.
- Windows: the row runs for real, with no ledger skip. `invalid?name` returns ERROR_INVALID_NAME, and Op=CreateFile. The ERROR_PATH_NOT_FOUND statement in the ledger row and the results is correct: under a regular-file parent this is not semantic absence, Go maps it to IsNotExist, and it is stated as a bound of the stateread.Lstat seam.

## Rev9 identity (per-file patch-id comparison of the rev9 patch vs base..candidate)
- 74 → 73 paths. The one DROPPED path is CHANGELOG.md, removed per the 2026-09-24 policy (the entry text goes into results).
- DIFF with identical +/- lines (context only, from the trunk refresh): cmd/curator/main.go, internal/envprofile/envprofile.go, internal/install/draftsources.go, internal/marker/marker.go.
- draftsources_test.go and platform-cases.tsv: only the owned additions (the new test and its 3 ledger rows).
- state_read_guard_test.go has three changes, all judged OK:
  - (a) The allowlist entry `manifest.go:LoadWithOptions` → `manifest.go:Load`. Trunk ab34556e only has `func Load`, so this is a correct rename migration.
  - (b) A new allowlist entry for gitops.FetchCommitFromURLIsolated. This is a trunk-added Lstat site (gitops.go:405, which treats IsNotExist as absent and every other error as a stop, matching its justification), caught by the deny-by-default guard.
  - (c) The coverage log now reports covered/scanned as a ratio. The equality assertion is unchanged, so this is an honest update.
- No revert of trunk. No CHANGELOG or LOGBOOK edit. No stray files.

## Hosted gate
Run 36198387119, headSha d4175d39, whose tree = a07dad19 (the exact candidate). Every lane succeeded, including Test windows/ubuntu/macos, Race, Lint, Gate self-test x3, Interop, and Naming. Rev1c's Windows fix (a CreateFile vs lstat Op assertion) is correct: it keeps the exact-path check and the typed-error check.

Verdict: ACCEPT.
