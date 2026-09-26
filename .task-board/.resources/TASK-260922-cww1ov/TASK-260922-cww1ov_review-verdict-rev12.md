# TASK-260922-cww1ov — review verdict, revision 12: ACCEPTED

Candidate tree 0e01a0fe on base 60498052 (refresh of accepted rev11).

## 1. rev11 content preserved, trunk kept
Reconstructed rev11 onto 60498052 (temp index, `git apply --3way` of TASK-260922-cww1ov_change-request_rev11.patch) = tree 0706afdb.
`git diff 0706afdb 0e01a0fe` touches only 9 paths: 1 new doc section, 4 ledger rows, and the new stateread migrations and rows below. All
other 29 rev11 paths are byte-identical to rev11-on-trunk. The trunk-touched files (platform-cases.tsv, profile_test.go, envprofile.go,
envprofile_f10f11f12_test.go, draftsources.go) only gain lines on top of trunk, so no trunk content is dropped. No LOGBOOK.md or CHANGELOG.md in the delta.

## 2. New paths = trunk state readers migrated onto stateread
- envprofile.go loadMachinePolicy (1f2ng0): `os.Lstat`+IsNotExist replaced by stateread.Lstat. Absent → default; any failed read → DiagSourceInvalid wrapping the typed error. Row: TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/blocked_parent (PASS locally).
- draftsources.go replay checkout probe (11burj): same migration. EnsureRepo failure is now typed unusable. Rows: TestDraftReplayRefusesUnreadableGitSnapshotCache and TestDraftInstallRefusesUnreadableLocalSnapshotCache. Both drive the production `install.Project` entry and assert that the lock and the blocker are unchanged. PASS locally.
- snapshot OpenLocal/authenticateLocalSnapshot/AuthenticateGit (11jgkt): only an absent path gives source_snapshot_unavailable; a failed read gives typed stateread.Error. Row: TestFrozenSnapshotReadersRefuseBlockedParents (local + git). PASS locally.
- troubleshooting.md: manager_state_unreadable section. Ledger: 4 rows, three lanes, no skip.

Narrowing mutant (disposable archive of 0e01a0fe): in AuthenticateGit, a stateread.Lstat read failure was mapped to the absent
"not in the store" message (read-failure → absence fallback). KILLED: TestFrozenSnapshotReadersRefuseBlockedParents/git_snapshot, capture_test.go:41.

## 3. Validation
- Local: internal/snapshot ok, internal/stateread ok, and internal/envprofile policy rows ok. The full `./internal/install` package and `-run TestDraft` and the full ./internal/envprofile package timed out locally at 600 s and 500 s. This matches the host's known exec-stall/load behaviour; targeted draft rows pass in 2 s.
- Hosted gate run 36219425103 (commit 8266566a, whose non-board tree equals 0e01a0fe exactly): Test ubuntu/macos/windows, Race, Lint, Gate self-test x3, Interop, and Naming are all success. This is the arbiter for the full install suite.

Residual (minor, no action): AuthenticateGit now treats a present-but-non-repo `repo` as unusable rather than unavailable. This is intended by the absence/failure rule and is green on all lanes.
- go vet (darwin) and GOOS=windows go vet over snapshot/envprofile/install: OK.
