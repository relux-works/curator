# Brief — TASK-260910-1952mz: shell-hook trust gate and manager-recorded env-file digests (S6, manager)

Story `STORY-260910-2awkzu` (shell-hook-project-env-approval-gate), wave 1.
Rules: `remediation-manager-producer-rules.md` (attached). Role: developer.
Sibling `TASK-260910-3ungjy` (approval commands + status posture) runs AFTER
this task in the same Story worktree; do not implement the CLI commands or
status rows here, but design the approval state so that task can add them.

## Spec (read first; it is the contract)
curator-spec `0da4020`, `profiles/manager.md` §8.1–§8.7 (landed by PR #60):
trust gate, approval record shape, diagnostics, warn-first rollout, posture,
conformance surface and execution binding. Vector:
`conformance/v1/vectors/shell-hook-trust.json` (14 cases: approved /
unapproved / changed × `.agents/env.sh` / `.agents/env.ps1` × `A-warning` /
`B-enforcing`, plus two forged-project-record cases). Audit context:
`docs/security-audit-2026-09.md` S6/I1 in this repository.

## Current code
`internal/shell/shell.go` (`Hook`, `InstallHook`, `posixHook`,
`powershellHook`: the cached hook walks up from `$PWD` and sources
`.agents/env.sh` / `.agents/env.ps1`); `internal/envfiles/envfiles.go`
(`WriteProject`, `WriteGlobal`: the manager writes those files);
`cmd/curator` shell/hook commands. Read them before designing.

## Deliverable
1. **Approval state** (§8.2): manager-home state below the manager home used
   by the rest of the manager (follow how other manager-home state is
   located; outside every profile/package/project surface). One closed
   record per absolute, canonicalized env-file path:
   `{ path, sha256, approved_by: "manager" | "operator", approved_at }`
   (lowercase hex SHA-256 over the exact bytes, RFC 3339). Provide a small Go
   package (e.g. `internal/hookapproval`) with read / upsert / revoke / list,
   atomic writes, two spellings of one file resolving to one record, and a
   format the emitted shell hook can read cheaply (document it). Never read
   a record from package, project or profile data.
2. **Manager-recorded digests** (§8.1 `manager` trust source): when the
   manager itself writes `.agents/env.sh` / `.agents/env.ps1`
   (`envfiles.WriteProject`, and the global files if the hook sources them),
   record the digest as `approved_by: manager` at that time.
3. **Hook trust gate** (§8.1/§8.4/§8.5): regenerate `posixHook` and
   `powershellHook` so that, before sourcing a candidate, the hook computes
   the candidate's SHA-256 (`shasum -a 256` / `sha256sum` fallback on POSIX,
   `Get-FileHash` on PowerShell), looks up the record for the absolute path,
   and:
   - trusted → source silently;
   - no record → diagnostic `shell_hook_env_unapproved`; changed digest →
     `shell_hook_env_changed`; in both cases warn ONCE per shell session
     (session marker variable), naming the absolute path and
     `curator hook approve <path>`, and continue without failing activation;
   - rollout profile: implement both `A-warning` (still sources) and
     `B-enforcing` (does not source) selectable by ONE generation-time
     option/constant; the shipped default is `A-warning`. A forged approval
     record placed in project data MUST be ignored.
   Keep the existing upward-search/progress guarantees and Git-Bash rules of
   the profile intact.
4. **Vector execution test** (§8.7 recipe): a Go test reads
   `shell-hook-trust.json` from `CURATOR_CONFORMANCE_ROOT` (root-content skip
   + ledger row when absent), and for every case: materializes the fixture
   bytes at a temp candidate path, seeds a temp manager home with exactly the
   case's approval record (or none), places the forged record only in
   project data when present, generates the hook for the case's profile, runs
   activation TWICE in a real shell (`sh`/`bash` for POSIX; `pwsh` for the
   PowerShell cases with a legitimate host-capability skip when `pwsh` is not
   installed), and asserts sourced / diagnostic / warning count exactly as the
   case expects. Also a hostile-checkout unit test (AC): foreign
   `.agents/env.sh` bytes are refused under `B-enforcing` and warned under
   `A-warning`.
5. `CHANGELOG.md` Unreleased: "S6: shell-hook trust gate — warning release
   (`A-warning`) …" with the migration hint; note that `B-enforcing` follows
   in a later release.

## Out of scope
`curator hook approve|approvals|revoke` and the status posture rows
(`TASK-260910-3ungjy`); the E2/E4/S4 findings; SPEC_PIN; PowerShell hook
execution on hosts without `pwsh` (skip legitimately, do not stub).

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l`, `go test`
for `./internal/shell/... ./internal/envfiles/... ./internal/hookapproval/...
./cmd/curator/...` with the conformance root set). Attach
`TASK-260910-1952mz_results.md`, tick the checklist, then
`task-board handoff TASK-260910-1952mz --role developer`.
