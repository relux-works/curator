# Revision 7 review — accepted

Task TASK-260916-1l44nd; CR-TASK-260916-1l44nd-7.
Candidate tree f86d61c80c2dd2e27e793f04360154e1dca4f030; base 07453544e3defe726f5951af9023a0343c4402b7.

The single revision-6 finding is corrected. REMOVE_DIR, REMOVE_FILE and all seven MAKE_* rights are handled starting at ABI 1. Only REFER starts at ABI 2, TRUNCATE at ABI 3, IOCTL_DEV at ABI 5. Write masks are 0x1ff2 / 0x3ff2 / 0x7ff2 (ABI 3–4) / 0xfff2 (ABI 5+ for this build's known rights); execute adds 0x1. File grants still strip directory-only rights. Writable roots remain PrivateBase plus WritePaths, unchanged.

Revision 6 to 7 changes exactly five paths: landlock.go, landlock_test.go, preflight_test.go, reviewer_abi1_test.go and the platform ledger. The production mutation test no longer permits outside-grant mutation on ABI 1. The original reviewer contract is included and registered. Source comments and the attached results.md Revision 7 table are corrected. Other previously reviewed surfaces are unchanged and are not reopened, per the binding review note.

## Independent checks

Exact candidate extracted with git archive into a disposable directory inside the assigned worktree; product sources left unchanged. Shell zsh, Darwin, Go 1.26.0.

- Focused tests with -count=1: TestReviewerABI1MutationRights (9/9 rights), TestLandlockHandledMaskFollowsABI (11/11 cases), TestLandlockRuleRightsFollowObjectType (9/9 cases): exit 0, 0.662s. Repeated after mutant restoration: exit 0, 0.460s.
- Compared all 14 Linux constant aliases and off-Linux mirrors against the installed golang.org/x/sys/unix/zerrors_linux.go constants: all match. The mask tests independently pin their numeric results.
- Built actual CLI with go build -o ../curator ./cmd/curator after restoring all mutants: exit 0.
- Narrowing mutants in disposable extraction: restore ABI>=2 mutation gate; omit MAKE_REG; request REFER on ABI 1. All 3/3 killed by the focused mask/reviewer tests, each exit 1 with -count=1 -timeout=90s. Original landlock.go restored and cmp against the candidate Git object returned 0.
- No full landing-suite replay. One earlier build overlapped temporary mutant work and is deliberately excluded; the clean build above was rerun after restoration.

These are mask/helper mutant kills, not Linux worker-boundary mutant executions. ABI-1 kernel enforcement remains unverified because no available lane supplies that kernel. The corrected production call sites use these helpers at landlock_linux.go:137,237,257; no detached implementation is being tested.

## Hosted evidence

Gate https://github.com/relux-works/curator/actions/runs/35693783982 is successful. Head cbd7d0b8419c7ff76a6893dbe95bfb6e0fa23564 resolves via both local Git and the GitHub commit API to the exact candidate tree above.

Downloaded test-evidence-ubuntu-latest and race-evidence-ubuntu-latest and parsed their aggregate go-test.json files (no duplicate counting from served files).

| Row | Ubuntu Test | Ubuntu Race |
|---|---|---|
| TestLinuxLandlockConfinementMatchesProbe | PASS 0.07s | PASS 7.27s |
| TestLinuxWriteConfinementTruncateMatchesProbe | PASS 0.05s | PASS 7.10s |
| TestLinuxWriteConfinementDirectoryMutationMatchesProbe | PASS 0.06s | PASS 7.19s |
| TestReviewerABI1MutationRights | PASS 0s | PASS 0s |
| TestLandlockHandledMaskFollowsABI | PASS 0s | PASS 0s |
| TestLandlockRuleRightsFollowObjectType | PASS 0s | PASS 0s |

6/6 selected parent rows pass per Linux lane. Race JSON has one non-JSON module-download preamble, explicitly inspected; no malformed test event was hidden. Hosted executions are accepted evidence, not local Linux reruns. Overall gate includes green Windows/macOS, lint and other configured checks; rose-air and candidate-suite jobs are skipped, not passing.

## Bounds and lifecycle

No new finding. Prior review conclusions on unchanged revision-5/6 surfaces are retained. R-e pass-through streaming stays deferred to R5; newer-UAPI rights remain the documented bound. Prior production-boundary mutant coverage gaps are not upgraded to executed proof by this review. This acceptance resolves the expressly scoped revision-7 correction, without claiming ABI-1 runtime, fresh Windows runtime, ARM64, or exhaustive filesystem proof.

Verdict: accepted via accept_cr revision=7; route integrating, not done. Run RUN-260922-906323 is not goal-bound; no directives present. No logbook executable is available and campaign rules prohibit LOGBOOK.md edits; this task-scoped verdict preserves the review record. Evidence is attached before the verdict mutation. Integration remains producer-owned.
