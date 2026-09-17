# TASK-260910-1952mz — revision 2 review verdict

Verdict: **changes_requested**; route **to-dev**. No acceptance, commit_ack, or done transition.
Reviewer run: RUN-260917-227001. Candidate: b0e46950a9976ecb2476d23df6f5202904a768b7; base: 1de6f8e12f33212843820be756923b1179063d2c.
Published patch SHA-256 independently verified: b189d64f533f59f886877412205ebf91f1b8497ccf496f9dd1a400e0defca9ed.
All 11 changed paths matched candidate bytes; tracked worktree diff against candidate was empty. Candidate source was never edited. Adversarial probes/mutants used an archived disposable copy under the worktree.

## Required corrections

1. **P1 — malformed approval records authorize execution.** `internal/shell/shell.go:268` (awk lookup at :275) and :448–462 (PowerShell lookup) accept a matching path/digest without validating the closed four-field record. Independent production-hook probes under B-enforcing with (a) only path/digest, (b) approved_by=project, and (c) invalid timestamp all sourced silently. The Go List reader rejected all three, so its validation is bypassed by the actual consumer. Manager §8.1–8.2 admits only manager/operator records with the required shape; malformed/partial reads are not absence. Validate the closed record set in both emitted consumers, distinguish malformed/read failures from absence, and add production-hook negative tests covering both profiles. PowerShell shares the defective predicate by inspection; it was not executed locally.

2. **P1 — revision-1 portability failure was hidden, not repaired.** `internal/shell/shell.go:198–208` still emits Bash array syntax. `/bin/dash -n <generated-hook>` returns 2 at generated line 183, exactly the reported failure; macOS sh syntax check returns 0. `internal/shell/shell_hook_trust_test.go:297–309` deliberately removes sh and selects only bash/zsh. The round-2 task explicitly requires a POSIX/dash repair and forbids weakening gates. Preserve Bash prompt integration while making the required POSIX activation path parse/run under sh/dash, and run a real regression there. Also remove the blanket Windows POSIX skips at :228–230 and :343–345 when Git Bash is available: path spelling is an implementation/harness interoperability issue, not unavailable interpreter capability. Normalize Go/Windows/MSYS candidate identities and assertions and execute the Git Bash path. Current green hosted evidence does not cover these omitted paths.

3. **P2 — path identity contradicts the normative one-record requirement.** `internal/hookapproval/hookapproval.go:60–74` canonicalizes lexically only; emitted hook lookup compares the raw candidate spelling (`internal/shell/shell.go:275`, :458). A manager-approved real project opened through a symlink alias yields Lookup(found=false) and B refuses with shell_hook_env_unapproved. Manager §8.2 says two spellings of the same file MUST resolve to one record. The producer explicitly reports both symlink and Windows/MSYS differences as bounds, but these are implementation divergences, not a spec gap. Use consistent identity canonicalization across writer, reader, and both hooks; exercise real/alias paths and native/MSYS spellings without duplicating approvals. Preserve error reporting for unresolved/unreadable paths.

4. **P2 — approval replacement fallback is not atomic.** `internal/hookapproval/hookapproval.go:279–285` removes the published state after any rename failure, then retries. Readers can observe absence; a failed second rename destroys the prior record set. This contradicts the explicit atomic-write deliverable, particularly on fallback platforms. Use a platform-appropriate atomic replace or preserve old state and fail safely; add a failure-path test proving the last valid state survives publication failure. This is a direct code finding, not a locally injected OS failure.

## Per-item assessment

| Requirement | Evidence / assessment |
|---|---|
| Closed manager-home approval state / APIs | Package docs and List/Upsert/ApproveFile/Revoke in hookapproval.go; Go shape validation works. Hook validation fails R1; canonical identity fails R3; atomic fallback fails R4. |
| Manager writes record manager digests | envfiles.go:80–129; install/commit.go:670–695 calls recording after publication under lock. Writer and install entry-point tests pass. |
| Default A, selectable B, exact diagnostics | shell.go:48–85, :306–347, :484–510. PASS for exercised ordinary fixtures; default stays A-warning. |
| Trust gate called before source | shell.go:394–395 and :542–543; actual emitted hook exercised via HookWithProfile. No project-local lookup in ordinary configured manager-home path. |
| Once/session warning + migration hint | shell.go:278–303, :465–483; two-activation vector assertions pass for 8/8 executable POSIX cases. |
| Forged project record ignored | 2/2 provided forged-record vectors executed and passed. Supplied fixture is JSON in project data; this is not proof against every possible state redirection or malformed-state class. |
| §8.7 vectors / root-content | shell_hook_trust_test.go:69–92 consumes external family and runs all declared cases through generated hooks. 8/14 executed PASS; 6/14 PowerShell SKIP because pwsh absent. Root-content absent-file branch and platform-cases.tsv rows present. Windows blanket POSIX skips remain R2. |
| Hostile checkout | Existing POSIX hostile test passes A/B; PowerShell skipped locally. |
| Status posture / approval CLI | Explicitly deferred to sibling TASK-260910-3ungjy by the task brief; not a defect in this revision. |
| Release note | CHANGELOG.md:28 names S6, A-warning, both diagnostics, migration command, later B release. PASS. |
| Scope | 11 paths; no SPEC_PIN change, vendored spec, ax/proposals, branches/commits. install/commit.go is relevant production wiring. |
| Spec gap | No gap identified: §8.2 directly covers R1/R3. Current spec checkout HEAD is 23dafa7, not advertised 0da4020; compared pinned diff: §8 and shell-hook-trust vectors unchanged, so this did not alter the review contract. |

## Independent validation

Shell: bash with pipefail for checks. CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1. Commands used -count=1 for executed tests.

- `go build ./... && go vet ./... && gofmt -l .`: exit 0. gofmt listed pre-existing board-resource Go snippets and disposable review-copy snippets; it does not fail on formatting differences. Separate `gofmt -l internal/ cmd/`: no output. Production/test source formatting clean.
- `go test -count=1 -v ./internal/shell/... ./internal/hookapproval/... ./internal/envfiles/...`: exit 0. Vectors: 8 PASS / 6 host-capability SKIP, no pwsh. Existing hostile POSIX PASS, PowerShell SKIP.
- `go test -count=1 ./internal/install -run 'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals'`: exit 0 (8.848s).
- `go test -count=1 ./cmd/curator -run TestShellInit`: exit 0 (0.657s).
- `golangci-lint run ./internal/shell/... ./internal/envfiles/... ./internal/hookapproval/... ./internal/install/...`: exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh .temp/review-1952mz/ledger`: exit 0, 241 rows checked.
- Additional disposable-copy adversarial tests: exit 1, expected defect reproductions: 3/3 malformed rows admitted, 1/1 approved alias refused, dash syntax failed; sh syntax passed. No claim of local PowerShell or Windows execution.

Full cmd/curator suite and remote gate were NOT rerun. Reused attached revision-2 validation log: hosted run 35178718756 reports success, exit 0, including Test/Race Linux/macOS and Test Windows. Its own footer says test_case_coverage=unknown. The committed SPEC_PIN predates this vector family, so hosted green is not evidence that all 14 new vectors ran. Independent local coverage is exactly 8/14; 6 PowerShell cases remain unexecuted here.

## Narrowing mutants

Mutated only the disposable candidate copy, restoring exact shell.go bytes afterward. Both ran committed `TestShellHookTrustVectors` with -count=1 and external root.

| Mutant | Narrowed refusal | Result |
|---|---|---|
| M1_only_unapproved_refused | Changed-file B branch additionally requires empty recorded digest, so changed records bypass refusal | KILLED, exit 1: changed-env-sh-B-enforcing-not-sourced observed sourced1=1 |
| M2_only_changed_refused | Unapproved B branch additionally requires nonempty recorded digest, so absent records bypass refusal | KILLED, exit 1: unapproved-env-sh-B-enforcing-not-sourced and forged-project-record-env-sh-B-enforcing-not-sourced |

Killed 2/2 selected mutants, 0 survivors. Bound: these are POSIX refusal-class mutants, not exhaustive guard, PowerShell, corruption, identity, or filesystem-failure mutation coverage. The real R1–R3 defects survive the existing committed green suite.

## Lifecycle and handoff

`task-board spawn goal "$TASK_BOARD_RUN_ID"`: run not goal-bound. No directives. Findings persisted as this verdict and a task-scoped review logbook outcome; campaign forbids LOGBOOK.md edits. Attach evidence before setting to-dev. Producer must correct R1–R4 and obtain another reviewer cycle. No human-only decision or external blocker is required.
