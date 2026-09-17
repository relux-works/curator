# Brief — TASK-260916-3oh0u8: provider lookup from trust roots (E4, manager)

Story `STORY-260916-2otjbn` (umbrella-provider-trust-roots), wave 1; this
story blocks proposal 0016 / `path_prepend`. Rules:
`remediation-manager-producer-rules.md` (attached). Role: developer. The
sibling `TASK-260916-16ys92` (launcher prints the resolved provider path)
lives in curator-agent-launcher and is independent.

## Spec (read first; it is the contract)
curator-spec `0da4020` (landed by PR #62): `protocol/environments.md` §11
(trust roots: the manager's install directory, then the machine knob
`provider_directories` in listed order; revision A = ambient `PATH` still
selects, warn `subcommand_provider_outside_trust_roots` with resolved path,
roots consulted and the migration hint; revision B = install dir → configured
list, never `PATH`, refuse PATH-only / manager-published / managed providers
with `subcommand_provider_untrusted`; disjoint outcomes incl.
`subcommand_provider_missing` (roots consulted) and
`subcommand_provider_root_unreadable` — a read failure is never absence and
never a fallback; diagnostic-only `PATH` probe under B names the untrusted
path), §11.1 diagnostics table, §12.1 `provider_directories` (list of
absolute paths, default `[]`), §12.2 (lockable), §12 posture (resolved
absolute provider path and trust verdict per discovered `curator-<name>`,
always for `curator-run` and `curator-session`, missing/unreadable
reported); `profiles/manager.md` §1 lock set; `cli/curator.md`; vectors
`conformance/v1/vectors/umbrella-provider-resolution.json` (both revisions; the
S6-planted PATH case; PATH-vs-trusted; unreadable root) and the
`provider_directories` schema cases (manager-config-v2 / system-config-v2).
Audit: `docs/security-audit-2026-09.md` E4; `verify-e-findings-rev4.md` on
`TASK-260916-dv7xv5` (`cmd/curator/umbrella.go:30-63` uses `exec.LookPath`
on the ambient PATH; only manager-published dirs refused).

## Deliverable
1. **Config knob** (`internal/config`): `provider_directories` (closed list
   of absolute paths, POSIX-absolute or Windows drive-absolute as the spec
   says, no duplicates, default `[]`), lockable in the system file; schema
   cases consumed from the root (root-content skip when absent).
2. **Resolution** (`cmd/curator/umbrella.go`): implement both revisions
   behind ONE option/constant; ship **revision A** as the default: keep the
   `PATH` selection, compute the trust verdict against the roots (install
   directory of the running executable resolved through symlinks, then
   `provider_directories`), warn on an outside-roots selection with the
   resolved path, the roots consulted and the `provider_directories` hint;
   revision B: search install dir then the list in order, first executable
   regular file wins, never descend, refuse manager-published/managed
   candidates and PATH-only matches (`subcommand_provider_untrusted` naming
   path and roots), `subcommand_provider_missing` naming the roots consulted
   when nothing matches and every root was readable, and
   `subcommand_provider_root_unreadable` naming the first unreadable root
   (a read failure never activates a fallback) — the five outcomes are
   disjoint exactly as §11 states. Keep every existing §11 rule (identifier
   grammar, implemented-subcommand-wins, no provider registry, no implicit
   install, profile/marker/fragment data never influence dispatch).
3. **Posture** (`cmd/curator/envstatus.go` / `curator status`): the resolved
   absolute provider path and trust verdict for every discovered
   `curator-<name>`, always `curator-run` and `curator-session`, reported
   missing when absent and unreadable with the directory when a root cannot
   be read; a refused/failed provider row is non-current.
4. **Vectors**: a Go test executes the umbrella-provider vector cases from
   the root for BOTH revisions (materialize the case's directories and
   executables in a temp tree, set PATH, run the resolver, compare
   resolved/diagnostic and the named roots) with the root-content skip and
   ledger row; the AC's hostile case (a `curator-run` planted through a PATH
   entry outside the trust roots) is asserted refused under B and warned
   under A.
5. `CHANGELOG.md` Unreleased: "E4: … warning release (revision A) …",
   noting revision B follows in a later release and the 0016 dependency.

## Out of scope
The launcher side (`TASK-260916-16ys92`), proposal 0016 itself, SPEC_PIN,
spec edits.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l`, `go test`
for `./internal/config/... ./cmd/curator/...` with the conformance root set).
Attach `TASK-260916-3oh0u8_results.md`, tick the checklist, then
`task-board handoff TASK-260916-3oh0u8 --role developer`.
