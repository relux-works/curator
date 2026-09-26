# TASK-260915-s4qkmr — Inventory of task-board*/curator*/tb-sessiond entries (BEFORE)

Captured 2026-09-22 ~21:00 +0400 (UTC+4) from zsh, `set -o pipefail`, before any change.
HOME=/Users/administrator. Hashes: `shasum -a 256`. mtimes in +0400.

Deviation from brief §"What is already true": the brief lists 7 legacy standalone
binaries; the host actually holds **10** (`task-board-main-e8205e78*` x3 extra).
The brief's `~/.local/bin` grep pattern also missed `~/.local/bin/tb-sessiond`
(it contains neither "task-board" nor "curator"); it is inventoried here.

## 1. ~/.local/bin — in-scope entries

| path | type | size | mtime (+0400) | sha256 |
|---|---|---|---|---|
| task-board | shim (POSIX sh) | 72 | 2026-09-17T17:48:03 | 5092fb50d34cdd27540531abdb71b3e68111866842becc8e117101fabf1a8cdb |
| task-board-tui | shim (POSIX sh) | 78 | 2026-09-16T14:06:35 | 44414a90b72916fab25eb8c177c5b3a43110149e1e4e89a78114b7df3e7aef6d |
| tb-sessiond | shim (POSIX sh) | 75 | 2026-09-16T14:06:35 | c972a22ae54bd43d3865831af9b269944969d86b13f4314c8df7c5b8e7b571c6 |
| task-board-main-154956a6 | standalone Mach-O x86_64 | 28066672 | 2026-09-14T07:57:39 | 7943d2c7e6b5ee6e2662ec6f72dcb9de1f0b30dadc55165717cca32fd734b846 |
| task-board-main-260914 | standalone Mach-O x86_64 | 28057648 | 2026-09-14T06:50:26 | 532cd9f16d308261ba487004c93c3aaa5abe2e28f69eb97e59069b37d7694383 |
| task-board-main-6cb09a23-curatorlike | standalone Mach-O x86_64 | 26839600 | 2026-09-14T11:02:16 | 18a0236197ac1f9c3e65e1d82f3812047ec83fd63a7b67fa68a520120d39da5d |
| task-board-main-91959c1a-curatorlike | standalone Mach-O x86_64 | 26742880 | 2026-09-14T09:53:27 | 6301aca5ee50d9678523ee5ce3cf67b1fe39b52a20d0472eeb81ff6000854e8d |
| task-board-main-91959c1a-nocgo | standalone Mach-O x86_64 | 27090416 | 2026-09-14T09:29:56 | ac96126a4db778313220cccea6a4eec49b7fdc10dc6eca8e1c0f11e09aec97de |
| task-board-main-91959c1a-nocgo-nop | standalone Mach-O x86_64 | 27090416 | 2026-09-14T09:40:21 | ac96126a4db778313220cccea6a4eec49b7fdc10dc6eca8e1c0f11e09aec97de |
| task-board-main-ac2ad9c0-curatorlike | standalone Mach-O x86_64 | 27411152 | 2026-09-16T11:17:53 | 9bfc3e68808076ed0fc16ea52b15310f405f970cdccb31344d09250d6513d05a |
| task-board-main-e8205e78 | standalone Mach-O x86_64 | 28304256 | 2026-09-14T08:19:18 | 9937fec400652ff1cc8b058bab4a02f199b469e1de412eae9a3762312a308ef5 |
| task-board-main-e8205e78-nocgo | standalone Mach-O x86_64 | 27077776 | 2026-09-14T09:19:20 | e582927b5e87caa79ebd5b0042820322b801b6a5e2320175b72815b24047ee2b |
| task-board-main-e8205e78-nop | standalone Mach-O x86_64 | 28066672 | 2026-09-14T09:06:37 | a49a4a3a458b2888cd2339433932ccf5044bce6cc5e157b5bba883e2fabc03ad |
| curator | standalone Mach-O x86_64 | 17562240 | 2026-09-08T15:44:01 | b0f16d52dbc5d1e929c434ccf4a194af1a4005fbf5ef13fbdb61a6cf93e652a2 |
| .curator-managed.json | marker JSON | 86 | 2026-09-16T14:06:36 | 01a049df99ea024a977478ec25a3bf06559947a71545e140f12731887ce626f9 |

Notes:
- `91959c1a-nocgo` and `91959c1a-nocgo-nop` are byte-identical (same sha256).
- No symlinks among in-scope entries (all regular files; `file(1)` confirms).
- `curator --version` => `curator main-04550e2`.

### Shim contents (~/.local/bin)

task-board (hand-written style: unquoted path, mtime Sep 17 — one day NEWER than the Curator-generated pair):
```
#!/bin/sh
exec /Users/administrator/.curator/global/bin/task-board "$@"
```
task-board-tui (Curator style: single-quoted path):
```
#!/bin/sh
exec '/Users/administrator/.curator/global/bin/task-board-tui' "$@"
```
tb-sessiond (Curator style):
```
#!/bin/sh
exec '/Users/administrator/.curator/global/bin/tb-sessiond' "$@"
```

### Marker (~/.local/bin/.curator-managed.json)
```json
{
  "schema_version": 1,
  "entries": [
    "task-board-tui",
    "tb-sessiond"
  ]
}
```
`task-board` is shimmed but NOT listed as managed (matches brief).

## 2. ~/.curator/global/bin — all entries (3, all Curator-generated shims)

| path | size | mtime (+0400) | sha256 | resolves to (exec target) |
|---|---|---|---|---|
| task-board | 306 | 2026-09-16T14:06:35 | be8fbf92bec559c862b92e8a7e0a2956be802577168bffadcd0b00fc92cd9be7 | cache/go-v1/e77b151704d2a922f9afba8207265c32be4da0ddfe6cc8c3a37371d1ee1b6c67/bin/task-board |
| task-board-tui | 310 | 2026-09-16T14:06:35 | 37c5a7d222be3f8a1935638c0a642ef9ad2b296ea7c88051aaa405b9e2dabdda | cache/go-v1/9eaec67d27938228ed3f5dd8aa97afeb966584e42a18e49b2fbcfafeb7a4859b/bin/task-board-tui |
| tb-sessiond | 307 | 2026-09-16T14:06:35 | 04804e481bef92ad8896583e0c4984f9a30ad40aa7e96983243fb8b877d03623 | cache/go-v1/533829077f05293bbfbc5970d009b7a64947ac41b18be03cdb7d27b8282e33c4/bin/tb-sessiond |

Each global shim prepends `~/.curator/global/bin` to PATH then execs the absolute
cache-build binary (full bodies captured in run transcript, exit 0).

### Cache-build targets (idle on disk; NO live process resolves these builds)

| build hash (short) | binary | size | sha256 |
|---|---|---|---|
| e77b1517… | bin/task-board | 27239264 | 0cda4580850ce49ac1e3fa1f8caedf6549679cc4655ecbbddce20c3658bd1e4b |
| 9eaec67d… | bin/task-board-tui | 16257152 | 3f5ff90dc86b2c6928c00d328e7e8353ffa97ae7d3e0552b05c8ed2abb8ea6b3 |
| 53382907… | bin/tb-sessiond | 16846608 | 4adb3bfbfd6bbcbe2f37315e0c5c8caf671755679b5da54c9550b82a00ed34e6 |

## 3. Live-process references (BEFORE, 2026-09-22 ~21:00 +0400)

`ps -ax -o pid,ppid,etime,lstart,command | grep -E "task-board|tb-sessiond|curator"`:

| PID | PPID | started | command |
|---|---|---|---|
| 35709 | 1 | Thu Sep 10 18:18:44 2026 | cache/go-v1/3b57ed0a…/bin/tb-sessiond --board-dir …/skill-project-management/.task-board |
| 46192 | 1 | Tue Sep 8 15:56:43 2026 | cache/go-v1/e2029bed…/bin/tb-sessiond --board-dir /Users/administrator/.task-board |
| 54120 | 6046 | Thu Sep 10 19:32:21 2026 | cache/go-v1/10f46bb6…/bin/task-board codex --model gpt-6-astra … |
| 39451 | 1 | Tue Sep 22 20:53:10 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike --no-update-check spawn-runner --manifest …/RUN-260922-839c83/manifest.json |
| 39477 | 1 | Tue Sep 22 20:53:11 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn wait RUN-260922-839c83 |
| 41409 | 1 | Tue Sep 22 20:56:46 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike --no-update-check spawn-runner --manifest …/RUN-260922-17ade7/manifest.json |
| 41430 | 1 | Tue Sep 22 20:56:46 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn wait RUN-260922-17ade7 |

(RUN-260922-17ade7 is this task's own run; RUN-260922-839c83 is the sibling BUG run.)

`lsof -Fn -c task-board -c tb-sessiond` txt mappings additionally show the three
cache-build binaries are mapped via Curator sweep-renamed paths
(`cache/build/go-v1/.sweep-<hash>-<id>/bin/…`) — i.e. Curator already GC-renamed
those build dirs while the processes kept running. The `ps` command lines above
show the pre-sweep absolute paths. Either way: live daemons resolve OLD cache
builds, not the current global-shim targets, and must not be disturbed.

Per-entry reference summary:
- `task-board-main-6cb09a23-curatorlike`: LIVE (4 processes, incl. this run's own spawn chain) — OFF LIMITS, residual.
- All other `task-board-main-*` standalone binaries: no live process, no marker, no launchd, no config reference found.
- `curator` binary: no live process resolves it (curator runs are short-lived CLI calls, none active at snapshot).
- Shims + marker: resolved by ordinary shell PATH lookup only (no daemon argv references them).

## 4. launchd references

None. All 10 plists in ~/Library/LaunchAgents grepped for
`task-board|tb-sessiond|curator|local/bin`: the only hits are `/usr/local/bin`
PATH strings (colima/caddy/memori plists) — false positives. `launchctl list`
shows no task-board/sessiond/curator job (only `com.apple.nsurlsessiond`,
matched on substring "sessiond", unrelated).

## 5. Campaign/board config references

- Board top-level config (`curator/.task-board/*.json|*.yaml`): no `task-board-main*` or `.curator/(global|cache)` references (grep exit 1, empty).
- Spawn manifests RUN-260922-839c83 / RUN-260922-17ade7: no binary-path references.
- The campaign's binding to `task-board-main-6cb09a23-curatorlike` is by absolute
  path in live process argv (table §3) — that is the reference of record.
- `task-board m 'project_config()'` is not a mutation on this binary (unknown-operation error); not pursued — not needed for the inventory.

## 6. Resolution baseline (BEFORE)

- `command -v task-board` (my shell) => `/Users/administrator/.local/bin/task-board`; `task-board --version` => `task-board version dev` (exit 0).
- Fresh login PATH (`zsh -lc 'echo $PATH'`) = `~/.local/bin:/usr/local/bin:…` — does NOT include `~/.curator/global/bin`, so all three managed commands resolve via the `~/.local/bin` shims.
