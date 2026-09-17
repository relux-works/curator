# TASK-260910-1952mz — revision 4 review verdict

Verdict: **changes_requested**. Route: **to-dev**. No acceptance or done transition.
Reviewer: RUN-260917-43f2a8. Candidate tree: f78eacc6e7b74430c8b343fe93f5be21f31b3490.
Base: 1de6f8e12f33212843820be756923b1179063d2c.
Published patch SHA-256 verified: 9571d46211bcfd8a5862dcba341b0ca948607268050d0bf3e18f4d23646de245.
All 11 changed paths independently matched the candidate bytes. Candidate code was not modified; mutants ran in an archived disposable copy below this worktree.

## Required correction — remaining R2/R3 Windows interoperability gap (P1)

The prior verdict explicitly required removing blanket Windows POSIX skips when Git Bash is available and reconciling native/MSYS identities. Revision 4 still unconditionally skips on runtime.GOOS == windows at internal/shell/shell_hook_trust_test.go:236 (all POSIX vectors), :386 (hostile checkout), :717 (malformed records), and :822 (symlink identity). This does not probe interpreter availability. The ledger at .github/ci/platform-cases.tsv:150–157 reclassifies the known path-spelling problem as host-capability. The original revision-1 failure already demonstrated that the Windows runner could execute the hook and produce its warning; an assertion/path interoperability failure is not missing capability.

Production still compares the POSIX candidate literally against stored paths (internal/shell/shell.go:259–299, :380). The Go writer uses native filepath canonicalization (internal/hookapproval/hookapproval.go:82–123), while the POSIX resolver produces shell filesystem spelling; there is no native/MSYS conversion before the equality check. The producer results explicitly acknowledge this retained difference. This violates the outstanding review correction and leaves manager §8.2's same-file identity obligation unverified for Git Bash. Under B, a native manager approval can be missed; under A, it produces an erroneous unapproved warning. This is a code/path-format finding, not a claim of an independent Windows execution on this macOS host.

Required next change: use one identity across native Go and Git Bash lookup; run the vector, hostile, malformed, and alias cases on Windows when Git Bash is actually available; normalize CRLF and assertion path spellings without weakening the expected sourced/diagnostic/warning outcome. Skip only a genuinely absent interpreter/capability, not GOOS alone. Demonstrate that one native manager record authorizes the same file reached through MSYS spelling under both profiles. Preserve the now-passing sh/dash and PowerShell coverage. This is ordinary implementation rework, not a human-only decision or external blocker.

The rev3 rework brief focuses on sh/dash but says the rev2 verdict is its authority. The rev4 brief asks to verify all R1–R4 closures. The producer's request to carry the Windows bound does not close the explicit prior correction.

## Per-item assessment

| Requirement / prior finding | Evidence and result |
|---|---|
| R1 closed-record validation | PASS for the requested negative shapes: shell.go:329–393 and :655–677 validate four fields, lowercase digest, closed approver and timestamp. Committed TestShellHookRejectsMalformedRecords executes 14 malformed shapes under sh/dash/bash/zsh and pwsh, both A/B. All pass independently. Bound: not an exhaustive parser equivalence proof; Go and shell timestamp implementations differ on some non-manager-emitted offsets. |
| R2 POSIX/dash syntax and execution | PASS for Unix: shell.go:204 guards Bash integration in eval; TestGeneratedHookParsesUnderPOSIXShells passes all 32 syntax checks (8 generated variants × 4 interpreters), plus real sh/dash activation. Windows POSIX portion remains OPEN as above. |
| R3 canonical identity | PASS for local real/alias identity: hookapproval.go:82, shell.go:259 and :570; TestShellHookTrustResolvesSymlinkedProject passes both profiles under all four POSIX interpreters and real pwsh. Native/MSYS identity remains OPEN. |
| R4 atomic replacement | PASS for requested failure preservation: hookapproval.go:299–342 stages and makes one rename, never removes published state. TestFailedPublicationPreservesLastValidState (:232) injects rename failure into Upsert and Revoke, verifies prior bytes/records survive and temp files are removed. Local test passes. Windows OS atomicity was not independently stress-tested. |
| Manager-home API / errors | List/Lookup/Upsert/ApproveFile/Revoke at hookapproval.go:170–297; malformed/unreadable Go state returns errors. Hook read-error branches refuse. Manager-home TSV format documented. No project/profile/package lookup in ordinary configured-home path. |
| Manager records writes | envfiles.go:78–126 and install/commit.go:670–694. Independent envfiles tests and real install.Project approval/dry-run tests PASS. |
| Default A / selectable B / diagnostics | shell.go:48–85, :418–457, :708–736. Default A-warning, closed generation-time profile option and exact diagnostics. PASS in executed cases. |
| Actual gate / warnings | Calls before source at shell.go:504 and :773. HookWithProfile-generated scripts executed twice in one process; sourced markers, per-activation diagnostic and warning counts checked. Migration command and path named. PASS. |
| Vectors driven | shell_hook_trust_test.go:70–178 consumes external vector family. **14/14 cases pass locally, zero skips** with pwsh on PATH: 8 POSIX cases × 4 interpreters plus 6 PowerShell cases. Two forged-project-record cases included. Root-content missing-family skip independently exercised; ledger validates. Windows POSIX exclusion remains the coverage defect above. |
| Hostile checkout | TestShellHookRefusesHostileCheckout passes both profiles under four POSIX shells and pwsh. |
| CLI / posture | Explicitly delegated to sibling TASK-260910-3ungjy, not required here. |
| Release / scope | CHANGELOG.md:28 names S6, A-warning, diagnostics, approval migration, later B. No SPEC_PIN change or vendored spec bytes, ax/proposal changes, commits or branch changes. Install wiring is relevant scope. |
| Spec gaps | No specification gap excuses Windows identity/skips. Local spec checkout differs from 0da4020 elsewhere, but §8 and shell-hook-trust.json are unchanged against that pin (verified diff). |

## Independent validation

Shell: /bin/bash, set -o pipefail. External root: /Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1. Tests use -count=1. Full transcripts attached as TASK-260910-1952mz_review-evidence-rev4.txt.

- go build ./...: exit 0.
- go vet ./...: exit 0.
- gofmt -l .: exit 0; output names pre-existing board-resource and disposable-copy snippets. Separate gofmt -l internal/ cmd/: no output, exit 0. Candidate source formatting clean.
- go test -count=1 -v ./internal/shell/... ./internal/hookapproval/... ./internal/envfiles/...: exit 0. Initial PATH had no pwsh, so PowerShell cases skipped in this run.
- PATH=/tmp/pwsh/app:$PATH go test -count=1 -v ./internal/shell -run 'TestShellHookTrustVectors|TestShellHookRejectsMalformedRecords|TestShellHookTrustResolvesSymlinkedProject|TestShellHookRefusesHostileCheckout': exit 0, 70.192s, zero skips. Reused the existing temporary real pwsh binary, did not install software.
- go test -count=1 ./internal/install -run 'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals': exit 0, 5.754s.
- go test -count=1 ./cmd/curator -run TestShellInit: exit 0, 0.494s.
- golangci-lint run ./internal/shell/... ./internal/envfiles/... ./internal/hookapproval/... ./internal/install/...: exit 0, 0 issues.
- bash .github/ci/ledger-consistency.sh .temp/review-1952mz/ledger: exit 0, 241 rows across linux/darwin/windows.
- TestShellHookTrustVectors with an empty conformance root: expected root-content SKIP and exit 0.

Full install/CLI suites and hosted gate were not replayed. Accepted attached exact revision-4 hosted log as broader evidence: run 35194739427, exit 0, Test Linux/macOS/Windows, Race Linux/macOS and Lint successful. Its own footer says test_case_coverage=unknown. Hosted green does not prove execution of vectors absent at the released SPEC_PIN, nor cover explicitly skipped Windows POSIX tests.

## Narrowing mutants

Only the archived disposable tree was mutated; original shell.go bytes restored and cmp-verified. Each mutant ran committed TestShellHookTrustVectors with external root and -count=1.

| Mutant | Narrowing | Result |
|---|---|---|
| M1 | Changed-file B refusal additionally requires empty recorded digest | KILLED, exit 1: changed-env-sh-B-enforcing-not-sourced sources instead of refusing. |
| M2 | Unapproved-file B refusal additionally requires nonempty recorded digest | KILLED, exit 1: unapproved and forged-project-record B cases source instead of refusing. |

Killed 2/2 selected mutants, 0 survivors. Bound: POSIX refusal-class mutants only; not exhaustive PowerShell/parser/filesystem/Windows mutation coverage. All 14 vectors passing on this host does not establish native/MSYS correctness.

## Lifecycle

Run goal queried: not goal-bound. No directives. Verdict, transcripts and task-scoped logbook attached before routing to-dev. No LOGBOOK.md edits (campaign prohibition), commit_ack, handoff, acceptance, or done transition. Producer should address the remaining Windows correction and obtain another independent review.
