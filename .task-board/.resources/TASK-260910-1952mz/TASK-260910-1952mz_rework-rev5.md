# Rework brief — TASK-260910-1952mz, revision 5 (answers review verdict rev4)

Read `TASK-260910-1952mz_review-verdict-rev4.md` first; it is the authority
for this round. R1, R2 (Unix), R3 (local alias), R4 are PASS — keep them
byte-for-byte unless the correction below forces a touch. Same Story
worktree; rules in `remediation-manager-producer-rules.md`. One required
correction (P1):

## Windows / Git Bash identity and coverage
1. **One identity for native Go and Git Bash lookup.** The Go writer stores
   native canonical paths (`hookapproval.go:82–123`, e.g.
   `C:\Users\me\proj\.agents\env.sh`), while the POSIX hook under Git Bash
   resolves the candidate to MSYS spelling (`/c/Users/me/proj/.agents/env.sh`)
   and compares literally (`shell.go:259–299`, `:380`). Convert to ONE
   canonical identity before comparison: in the POSIX hook, when running
   under MSYS/Git Bash (`uname -s` matches `MINGW*|MSYS*|CYGWIN*`), map the
   resolved candidate to native spelling with `cygpath -w` (fall back to a
   documented refusal-with-warning when `cygpath` is absent — never a silent
   source), normalize the drive letter case and separators the same way the
   Go canonicalizer does, and compare case-insensitively on Windows exactly
   as Go does. Document the identity rule next to the canonicalizer so both
   sides cite one definition.
2. **Run the Windows POSIX coverage instead of skipping on GOOS.** Replace the
   unconditional `runtime.GOOS == "windows"` skips at
   `shell_hook_trust_test.go:236, :386, :717, :822` with interpreter probes:
   when Git Bash (`bash.exe` from the Git for Windows install, or `sh` on
   PATH under MSYS) is present, run the vector, hostile-checkout, malformed
   record and symlink-alias cases through it; skip only for a genuinely
   absent interpreter/capability (host-capability class, naming what is
   absent). Normalize CRLF in captured output and compare warning paths by
   the spelling the hook itself prints on that platform, without weakening
   the expected sourced/diagnostic/warning outcomes. The ledger rows at
   `.github/ci/platform-cases.tsv:150–157` must then describe the real
   condition (Git Bash present → must run; absent → host-capability).
3. **Prove the cross-spelling case**: a committed test on Windows where one
   native manager record authorizes the same file reached through MSYS
   spelling under both profiles (`A-warning` silent, `B-enforcing` sourced),
   and a changed file is refused under B.

Keep the now-passing sh/dash and PowerShell coverage intact.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l internal/ cmd/`,
`go test -count=1 ./internal/shell/... ./internal/hookapproval/...`, plus
`golangci-lint run` on the touched packages) — the Windows behaviour is
proven by the hosted gate, which the runtime runs at handoff. Update
`TASK-260910-1952mz_results.md` (closure with file:line, the identity rule,
transcripts), tick the checklist, then
`task-board handoff TASK-260910-1952mz --role developer`.
