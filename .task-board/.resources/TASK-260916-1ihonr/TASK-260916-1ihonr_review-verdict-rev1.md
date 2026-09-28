# TASK-260916-1ihonr review verdict — CR-TASK-260916-1ihonr-1 rev1: ACCEPTED

Reviewer: claude-opus-5-5 (low). Shell: zsh, `set -o pipefail`.
Candidate: base 27cc242d, tree 0b302f60. I recomputed the worktree tree from a temp index and got exactly 0b302f603cc6…, 18 paths, no stray files.

## Brief items
1. **Load sites:** cmd/curator-run has two load sites, main.go:109 `axconfig.Load` and main.go:181 `defaults.Load`. Both now go through `configfile.Read`, which does the ancestor walk, Lstat symlink refusal, O_NOFOLLOW open (ELOOP → symlink), fstat regular-file check, uid == config-dir uid, `mode&0o022` refusal, and then reads from the same fd. There is no other os.ReadFile/Open of these files. On Windows it uses FILE_FLAG_OPEN_REPARSE_POINT + reparse refusal, requires the file owner to equal the directory owner, refuses a null/absent DACL, and refuses an allow-ACE write/delete/WRITE_DAC/WRITE_OWNER grant to anyone except owner/OWNER RIGHTS/CREATOR OWNER/SYSTEM/Administrators. Absent file/dir → absent. An open/stat/read failure → present+error → `defaults_config_invalid`, never absent.
2. **SPEC:** §4.7 has the new contract paragraph. The exit table row cites §4.3/§4.6/§4.7 and is consistent with them.
3. **Rows:** The goldens (symlink, group-writable) and 16 refusal rows (machine/operator × defaults/ax × symlink/group/world/unreadable) are driven through `run()`. There is a foreign-owner row for machine and operator defaults, a happy path through the production resolver, and a Windows DACL row at `configfile.Read` level plus a pure policy table. Windows mode rows are skipped with a reason.
4. **CHANGELOG:** The repo convention allows leaf entries (5d0d0a3 and 6c387f6 both edit CHANGELOG.md), so the entry is acceptable.

## Independent reruns (real exit codes)
- `go vet ./...` → 0
- `go test ./cmd/curator-run ./internal/configfile ./internal/defaults ./internal/axconfig` → 0 (all ok)
- `GOOS=windows go vet ./internal/configfile` → 0. (`GOOS=windows go vet ./cmd/curator-run` fails in the existing internal/execution/process.go, which this CR does not touch.)

## Mutants (on a git-archive copy of 0b302f60, `go test ./cmd/curator-run -run TestRunConfigSecurity`)
| mutant | exit | failing tests |
|---|---|---|
| symlink check removed (Lstat branch + O_NOFOLLOW) | 1 | 7 |
| permission narrowed 0o022→0o002 (world only) | 1 | 7 |
| permission check removed | 1 | 11 |
| owner check disabled (`&& false`) | 1 | DifferentDirectoryOwner/machine + /operator |
| unreadable→absent (ReadAll error → `nil,false,nil`) | 1 | 4 unreadable rows |

## Bounds / minors (not blocking)
- SPEC names the owner reason `configuration file owner`, but the emitted text is "configuration file is owned by a different identity than its configuration directory". The literal differs; it is still clearly the same reason.
- The foreign-owner row covers defaults.json only. ax.json shares the same `configfile.Read` path, so this is a bound, not a gap.
- The Windows DACL refusal is proven at `configfile.Read` level, not through `run()`. Windows evidence comes from the hosted lane (the orchestrator note says the gate is green); I did not verify it locally.
