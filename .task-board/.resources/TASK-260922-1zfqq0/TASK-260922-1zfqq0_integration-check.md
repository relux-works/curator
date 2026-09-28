# TASK-260922-1zfqq0 — integration preconditions check (accepted CR rev 2)

Run: RUN-260923-b18d2c (role developer, archetype implementer)
Board status observed: `integrating` (via `task-board q 'get(TASK-260922-1zfqq0) { id status }'`, exit 0)
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher/.temp/STORY-260922-39hxog/worktree`
Branch: `task-board/story/STORY-260922-39hxog`
HEAD: `8c5b049 TASK-260906-2t2t6w: fold the four 0.2.1-review minors into SPEC 0.4.1-draft`
Directives: none (`task-board spawn directives` exit 0, "No directives recorded")

## 1. Candidate tree state (no commit by this run)

`git status --porcelain` (exit 0), exactly 8 modified, 0 untracked from this run:

```text
M CHANGELOG.md
M README.md
M SPEC.md
M cmd/curator-run/main.go
M cmd/curator-run/main_test.go
M cmd/curator-run/testdata/help.golden
M internal/diagnostics/diagnostics.go
M internal/diagnostics/gate_conformance_test.go
```

`git diff --stat` (exit 0): 8 files changed, 281 insertions, 103 deletions. No
commit, no checkout/clean/stash, no checkpoint/integrate executed by this run,
per the integration assignment.

## 2. Doc preconditions (re-verified read-only in this run)

- `specVersion` pins: SPEC.md line 3 `0.5.0-draft`; `cmd/curator-run/main.go`
  `specVersion = "0.5.0-draft"`; README `0.5.0-draft`; CHANGELOG `0.5.0-draft`;
  `help.golden` `specification 0.5.0-draft`; `main_test.go` want `0.5.0-draft`.
- F-S2 names cited in SPEC: `launch-env-fragment-v2` (§4.1 transport
  precondition), `permissions { mode, locked, source }` lattice.
- F-M1 v0.5.18 names cited: `LaunchRequest.PermissionMode`,
  `LaunchRequest.ToolRelease`, `LaunchRequest.NativeArgs`,
  `permission-grammar-v1`, `ErrPermissionModeUnverifiedRelease`. No provider
  flag spelled (D5).
- Headless detector: closed marker set {`CI`, `GITHUB_ACTIONS`}, stated as
  closed and versioned in SPEC §4.6 (`0.5.0-draft`), mirror of environments §10.1.
- Choice 4: exact stderr line
  `curator-run: effective-native-policy: relaxation=<selector[,selector...]> source=<settings-source>`
  and launch-record extension key
  `works.relux.curator.effective-native-policy` (tracked = ax launch document
  key; untracked = stderr provenance only).
- Diagnostics table (§6): `permission_policy_unsupported`,
  `permission_mode_tracked_unsupported`, `permission_mode_unsupported` with
  exit 1, no untracked fallback.

## 3. D5 grep row (real exit codes, standalone processes, pipefail set)

```text
grep -c -- "--dangerously" SPEC.md -> 0 matches (grep exit 1 = absent, clean)
grep -c -- "--dangerously" README.md -> 0 matches (grep exit 1 = absent, clean)
grep -rn -- "--dangerously" SPEC.md README.md CHANGELOG.md -> no output (exit 1 = absent, clean)
grep -rn -- "--full-auto|bypassPermissions|--approval|--sandbox" SPEC.md README.md -> no output (exit 1 = absent, clean)
```

The launcher's own `--yolo` alias and `--permissions native|yolo` remain present
(allowed; they are the launcher interface, not provider spelling).

## 4. Registry parity (rev-2 declaration-only scope)

`internal/diagnostics/diagnostics.go` registers exactly three codes:
`permission_policy_unsupported`, `permission_mode_tracked_unsupported`,
`permission_mode_unsupported` (call-site selected, exit 1). Gate
`TestGateCoverageCounts` pins 21 normative codes (was 18). No resolution,
refusal, or transport behaviour added in this leaf (F-L1b owns behaviour).

## 5. Gate evidence (this run, standalone, no tee/pipe chain)

Command: `make check` run directly in the worktree.
Exit code: 0.

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  github.com/relux-works/curator-agent-launcher/cmd/curator-run  57.484s
ok  github.com/relux-works/curator-agent-launcher/internal/axconfig  2.178s
ok  github.com/relux-works/curator-agent-launcher/internal/cli  2.591s
ok  github.com/relux-works/curator-agent-launcher/internal/composition  3.413s
ok  github.com/relux-works/curator-agent-launcher/internal/defaults  4.641s
ok  github.com/relux-works/curator-agent-launcher/internal/diagnostics  5.813s
ok  github.com/relux-works/curator-agent-launcher/internal/execution  27.735s
ok  github.com/relux-works/curator-agent-launcher/internal/fragment  18.739s
ok  github.com/relux-works/curator-agent-launcher/internal/mapping  6.959s
ok  github.com/relux-works/curator-agent-launcher/internal/plan  6.304s
ok  github.com/relux-works/curator-agent-launcher/internal/systemprompt  6.638s
go test ./... -count=1 -race
ok  github.com/relux-works/curator-agent-launcher/cmd/curator-run  83.818s
ok  github.com/relux-works/curator-agent-launcher/internal/axconfig  3.009s
ok  github.com/relux-works/curator-agent-launcher/internal/cli  3.393s
ok  github.com/relux-works/curator-agent-launcher/internal/composition  4.987s
ok  github.com/relux-works/curator-agent-launcher/internal/defaults  4.472s
ok  github.com/relux-works/curator-agent-launcher/internal/diagnostics  5.617s
ok  github.com/relux-works/curator-agent-launcher/internal/execution  23.240s
ok  github.com/relux-works/curator-agent-launcher/internal/fragment  13.525s
ok  github.com/relux-works/curator-agent-launcher/internal/mapping  5.432s
ok  github.com/relux-works/curator-agent-launcher/internal/plan  6.612s
ok  github.com/relux-works/curator-agent-launcher/internal/systemprompt  5.894s
```

(fmt-check is part of `make check` and passed silently inside the exit-0 run.)

## 6. Integration disposition

- Landing preconditions confirmed: `integrating` status held, candidate tree is
  the accepted 8-file uncommitted delta on the Story branch, docs verified,
  `make check` exit 0 in this run.
- This run changed no file, created no commit, ran no `worktree checkpoint` /
  `worktree integrate`, called no generic `handoff`, and set no status beyond
  the FIRST `integrating` confirmation (already `integrating`).
- Left for the runner: synchronous bound landing of CR-TASK-260922-1zfqq0-2 rev 2.
