# TASK-260918-11f9l1 results — combine rc.12 union with moved trunk

Story-final Change Request work: replay accepted checkpoint `73fc8a4`
(TASK-260917-16l2md rev 5, base `3c45d4b`) onto the moved trunk and hand off.

## 1. Replay transcript

> NOTE: `task-board worktree refresh-candidate` refuses
> deterministically — see §7. The replay below was computed as detached
> experiments in `/tmp/replay-exp` (trunk `6401d3c`) and
> `/tmp/replay-exp2` (trunk `1c464c5`) — cherry-pick of `73fc8a4` onto
> trunk, no branch touched — to prepare exact resolution bytes. The
> story worktree is untouched at `73fc8a4`.

- `task-board worktree obligations` — no row for STORY-260917-3w3lvj
  besides review/checkpoint bookkeeping (5 rows, all other stories).
- `git status --short` in the story worktree: clean at `73fc8a4`.
- `task-board worktree refresh-candidate TASK-260918-11f9l1` (7 attempts,
  trunk `6401d3c` → `b56089e` → `1c464c5`):
  `change_request_checkpoint_conflict: persisting recovery intent before
  the base refresh of refs/heads/task-board/story/STORY-260917-3w3lvj ...`,
  exit 1, `--json` reports `INTERNAL_ERROR`. No replay worktree retained.
- Experimental merge at both trunks: exactly 3 content conflicts —
  `CHANGELOG.md`, `cmd/curator/envstatus.go`,
  `internal/envprofile/status.go`; `cmd/curator/main.go` auto-merges.
  Trunk moves during the run (`b56089e`: board-only; `1c464c5`: rustup
  lane + board) touch none of the union's 40 files and none of the 4
  merge-surface files, so the conflict set and the resolution bytes
  are identical at every observed trunk.

## 2. Resolutions (union of both sides, nothing dropped)

### 2a. `cmd/curator/envstatus.go` — S6 block FIRST, then union §12 block

```go
for _, warning := range status.ShellHookTrustWarnings {
    _, _ = fmt.Fprintf(stdout, "shell-hook-trust: warning: %s\n", warning)
}
for _, row := range status.ShellHookTrust {
    _, _ = fmt.Fprintln(stdout, formatTrustRow(row))
}
// §12 posture: the active S4 profile with the effective
// passable_env_names, the machine-level warnings, and the §2.3
// surfacing rows for the current profile of each reported scope.
_, _ = fmt.Fprintf(stdout, "s4_profile: %s, passable_env_names: %s\n", status.S4Profile, formatPassable(status.PassableEnvNames))
for _, warning := range status.Warnings {
    _, _ = fmt.Fprintf(stdout, "warning: %s\n", warning)
}
for _, scope := range status.MCPDeclarations {
    _, _ = fmt.Fprintf(stdout, "scope %s profile %s mcp-declarations:\n", scope.Scope, scope.Profile)
    for _, row := range scope.Rows {
        _, _ = fmt.Fprintln(stdout, row)
    }
}
```

Both sides byte-verbatim (6 S6 lines + 13 union lines); one closing
brace restored where git shared it as common context. gofmt-clean.

### 2b. `internal/envprofile/status.go` — field order

`Homes, Scopes, Adapters, Targets, Profiles, Providers,
UnregisteredEnvironments, Orphans, Notes, NonCurrent` (union alignment,
tag col 41), then `ShellHookTrust, ShellHookTrustWarnings` (S6,
verbatim), then `S4Profile, PassableEnvNames, Warnings,
MCPDeclarations` (union, verbatim), then `RequireCurrentProfile, ...`
unchanged. Imports auto-merged (union 4 + S6 2, alphabetical).
gofmt-clean.

### 2c. `CHANGELOG.md`

S6 entry kept verbatim in place, union E2/S4/pin entries kept verbatim
after it; union E4-Added and E4-Fixed hunks auto-merged. No rewording.

## 3. Closed output order (spec §10)

`curator env status` prints, after the optional
`require_current_profile` line:

1. `shell-hook-trust: warning: ...` rows (S6)
2. shell-hook trust rows `shell-hook-trust: <path> ...` (S6)
3. `s4_profile: ..., passable_env_names: ...` (union S4)
4. `warning: ...` machine warnings (union)
5. `scope <s> profile <p> mcp-declarations:` + rows (union §2.3)
6. homes, ... (unchanged), `provider <name>: ...` (union E4), targets,
   profiles, ... (unchanged)

Order authority: manager profile §10 closed posture inventory
(hook-trust is gate 1 of 13, ahead of env-passthrough /
provider-trust-roots / mcp rows). That inventory is landed on spec
main; rc.12 (`dced9b8`) §10 states only the recompute-and-report
discipline and orders nothing against it, and rc.12 environments §12
row sequence is preserved inside the union block. `curator status`
carries the S6 hook rows only — the union added no posture rows there
(cmdStatus never calls the env printer; union main.go diff is the
umbrella hunk only), so adding any would be new behavior beyond the
combination.

## 4. Identity proof (verified at trunks `6401d3c` and `1c464c5`)

- 36 of 40 union files byte-identical to `73fc8a4`; the 4 diffs are
  exactly the 3 resolved files + `main.go` (auto-merge).
- `main.go`: exp == checkpoint + S6 hunks; exp == trunk + union
  umbrella hunk (both directions verified with diff).
- Changed set vs trunk == union 40-file set + 1 test fix-up file.

## 5. Fix-up delta (combination-only, test-only, +33 lines, 0 production)

- `cmd/curator/hook_posture_test.go` (+17): `provisionedEnvMatrix`
  plants stub `curator-run`/`curator-session` providers on PATH
  (mirrors the union's own `--check` test setup in `env_test.go`).
- `cmd/curator/hook_test.go` (+16): same block at the top of
  `TestEnvStatusReportsShellHookTrustPosture` (own matrix setup).

Cause: the union attaches §12 provider rows on every `env status`;
a missing provider is non-current, so S6's `env status
--check`-expects-OK assertions failed on the missing `session`
provider. Stubs warn outside-trust-roots under revision A but stay
current, keeping `--check` evidence of the trust posture alone. No
production behavior touched.

## 6. Prepared artifacts (attached to the board)

Successor/orchestrator use after the replay unblocks:

| Artifact | SHA-256 | Use |
|---|---|---|
| `TASK-260918-11f9l1_resolved-CHANGELOG.md` | `38b42f0d…c1ca9` | `--replay-resolutions` replacement bytes for `CHANGELOG.md` |
| `TASK-260918-11f9l1_resolved-envstatus.go` | `ceeccf45…304cf` | same for `cmd/curator/envstatus.go` |
| `TASK-260918-11f9l1_resolved-status.go` | `d696023d…62090` | same for `internal/envprofile/status.go` |
| `TASK-260918-11f9l1_fixup.patch` | `70d56174…f8d6` | test-only fix-up (`git apply` on the replayed tree) |

(Full digests: see §2 files; `38b42f0d6434b1b3d01d7611ee7217a6ca3beff1e00e125678078ef0e85b1ca9`,
`ceeccf459dece03b4be4799e0b453c6df290b65bbc1d19fe09174193afc304cf`,
`d696023d369e3afe37f02203678f2fe782bce5438258bff84265d859a8a62090`.)

## 7. Blocker

`refresh-candidate` fails before retaining any replay (see §1).
Eliminated: lease (held by this run), dirty tree (clean), authority
observation (fresh trunk picked up across all 3 trunk OIDs), signing
configured, CR records parse, no stale locks, network OK, control-root
state normal. Prime suspect: the second checkpoint on this story —
TASK-260917-2ecpjv rev 1 (empty delta, base `73fc8a4`) — checkpointed
11:51:38Z, after this task was created (11:34:57Z); the replay path
was designed against a single-checkpoint story. No sanctioned bypass
exists: `converge` would demote the accepted union to stale;
hand-rebasing the story branch is forbidden. REQUIRED:
orchestrator/tool intervention to complete the replay (or a fixed
tool), after which the prepared resolutions (§2, artifacts in §6)
and fix-up (§5) transfer directly.

## 7. Gate transcripts (all in /tmp/replay-exp, CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1 @ dced9b8)

| Gate | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` (sources; only pre-existing `.task-board` resource files listed) | 0 |
| `golangci-lint run ./cmd/curator/ ./internal/envprofile/` (0 issues) | 0 |
| `go test ./internal/hookapproval/` (4s) | 0 |
| `go test ./internal/envregistry/ ./internal/globalbins/` | 0 |
| `go test ./internal/envfragment/` | 0 |
| `go test ./internal/config/` (31s) | 0 |
| `go test ./internal/contextmaterialize/` | 0 |
| `go test ./internal/shell/` (60s) | 0 |
| `go test ./internal/envfiles/` | 0 |
| `go test ./internal/interop/environments/` | 0 |
| `go test ./internal/envprofile/` — 174/174 pass; suite wall >15m under load, last test verified solo (8s) | 0 (split) |
| `go test ./internal/install/ -timeout 25m` (791s suite) | 0 |
| `go test ./cmd/curator/ -run TestEnv\|TestHook` (434s; 1 fail before §5 fix) | 0 |
| `go test ./cmd/curator/ -run TestStatus\|TestGlobalStatus` (215s) | 0 |
| `go test ./cmd/curator/ -run TestProfile\|TestUmbrella\|TestProvider` (329s) | 0 |
| `go test ./cmd/curator/ -skip <A,B,C masks>` (1308s, 223 pass) | 0 |
| `/tmp/replay-exp2` @ trunk `1c464c5`: resolutions apply, `go build ./...` | 0 |
| `/tmp/replay-exp2`: `git apply` fixup patch + `go vet ./cmd/curator/` | 0 |

Full-tree `go test ./...` was not run (wall-clock); every package
either side touched is covered above. The runtime's `remote-gate.sh`
at handoff is the final arbiter (per campaign rules).
