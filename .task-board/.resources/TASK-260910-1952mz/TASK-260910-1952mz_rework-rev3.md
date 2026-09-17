# Rework brief — TASK-260910-1952mz, revision 3 (answers review verdict rev2)

Read `TASK-260910-1952mz_review-verdict-rev2.md` first; it is the authority
for this round. Everything it marks PASS stays. Same Story worktree; rules
in `remediation-manager-producer-rules.md`. Four required corrections:

## R2 (P1) — repair the POSIX hook; never hide a gate failure
Revision 2 kept the Bash array syntax in the emitted hook
(`internal/shell/shell.go:198–208`; `/bin/dash -n` fails at generated line
183) and instead REMOVED `sh` from the test's shell list
(`shell_hook_trust_test.go:297–309`). That is a weakened gate, forbidden by
the campaign rules. Rewrite the emitted POSIX hook without Bash-only
constructs (no arrays, `[[ ]]`, `function`, `local -a`, `${x[@]}`, process
substitution; use `set --`/positional parameters or newline-separated
strings and `case`), keep the Bash/zsh prompt integration, and make the
test run the hook under `sh` AND `dash` when present (`command -v dash`) in
addition to bash/zsh; add `dash -n` / `sh -n` syntax checks of the generated
hook as a unit test that does not depend on the host shell.

## R1 (P1) — the hook must validate the closed record, not just path+digest
`shell.go:268/275` (awk lookup) and `:448–462` (PowerShell) accept a matching
path/digest even when the record is malformed (missing members,
`approved_by` outside `manager|operator`, invalid RFC 3339 timestamp) — three
probes under `B-enforcing` sourced silently. Make the emitted hooks accept a
record only when all four members are present and valid (closed set), and
add vector-style negative cases (malformed record ⇒ unapproved) to the unit
tests; keep the Go reader's validation as the single source of truth for the
record grammar (share the rule, e.g. by generating the hook's checks from the
same constants).

## R3 (P2) — one record for two spellings of one file
`hookapproval.go:60–74` canonicalizes lexically only, while the hook compares
the raw candidate spelling. Resolve symlinks/realpath on both sides (Go:
`filepath.EvalSymlinks` after `Abs`; hooks: a portable realpath — `cd -P` +
`pwd -P` on POSIX, `Resolve-Path`/`GetFullPath` on PowerShell) so a project
opened through an alias matches its record; add a test with a symlinked
project directory under both profiles.

## R4 (P2) — atomic replacement must preserve the last valid state
`hookapproval.go:279–285` removes the published state after a rename failure
and retries; a second failure destroys the prior record set. Use a
platform-appropriate atomic replace (rename onto the target; on Windows
`os.Rename` semantics or `MoveFileEx`-class replace) and, if replacement
cannot be atomic, keep the old state and fail the write; add a failure-path
test proving the last valid state survives a failed publication.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l internal/ cmd/`,
`go test -count=1` for `./internal/shell/... ./internal/hookapproval/...
./internal/envfiles/... ./internal/install/... ./cmd/curator/...` with the
conformance root set; `golangci-lint run` on the touched packages), update
`TASK-260910-1952mz_results.md` (R1–R4 closure with file:line, transcripts),
tick the checklist, then `task-board handoff TASK-260910-1952mz --role developer`.
