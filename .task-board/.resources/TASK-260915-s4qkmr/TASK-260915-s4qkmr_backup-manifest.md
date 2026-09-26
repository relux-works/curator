# TASK-260915-s4qkmr — Backup manifest (copy-only, sha256-verified)

Method: `cp -p` only. No `mv`, no `rm`, no overwrite of any live path.
Rollback (written BEFORE running, per brief rule 5):
`rm -rf /Users/administrator/.curator/backups/20260922T165929Z-legacy-task-board`
— removes only files created by this step.

- Backup dir: `/Users/administrator/.curator/backups/20260922T165929Z-legacy-task-board/`
- Created: 2026-09-22T16:59:29Z (20:59 +0400). Shell: zsh, `set -o pipefail`.
- Total: 262M, 17 files + MANIFEST.txt (on-disk manifest at backup-root).
- Verification: every copy's sha256 equals its source (17 OK, FAIL=0), re-checked
  after a manifest path fix (checked=17 missing_or_drift=0). Exit codes 0 throughout.

## Contents (backup-relpath -> source, sha256, bytes)

| backup relpath | source | sha256 | bytes |
|---|---|---|---|
| local-bin/.curator-managed.json | ~/.local/bin/.curator-managed.json | 01a049df99ea024a977478ec25a3bf06559947a71545e140f12731887ce626f9 | 86 |
| local-bin/task-board | ~/.local/bin/task-board | 5092fb50d34cdd27540531abdb71b3e68111866842becc8e117101fabf1a8cdb | 72 |
| local-bin/task-board-tui | ~/.local/bin/task-board-tui | 44414a90b72916fab25eb8c177c5b3a43110149e1e4e89a78114b7df3e7aef6d | 78 |
| local-bin/tb-sessiond | ~/.local/bin/tb-sessiond | c972a22ae54bd43d3865831af9b269944969d86b13f4314c8df7c5b8e7b571c6 | 75 |
| local-bin/task-board-main-154956a6 | ~/.local/bin/task-board-main-154956a6 | 7943d2c7e6b5ee6e2662ec6f72dcb9de1f0b30dadc55165717cca32fd734b846 | 28066672 |
| local-bin/task-board-main-260914 | ~/.local/bin/task-board-main-260914 | 532cd9f16d308261ba487004c93c3aaa5abe2e28f69eb97e59069b37d7694383 | 28057648 |
| local-bin/task-board-main-6cb09a23-curatorlike | ~/.local/bin/task-board-main-6cb09a23-curatorlike | 18a0236197ac1f9c3e65e1d82f3812047ec83fd63a7b67fa68a520120d39da5d | 26839600 |
| local-bin/task-board-main-91959c1a-curatorlike | ~/.local/bin/task-board-main-91959c1a-curatorlike | 6301aca5ee50d9678523ee5ce3cf67b1fe39b52a20d0472eeb81ff6000854e8d | 26742880 |
| local-bin/task-board-main-91959c1a-nocgo | ~/.local/bin/task-board-main-91959c1a-nocgo | ac96126a4db778313220cccea6a4eec49b7fdc10dc6eca8e1c0f11e09aec97de | 27090416 |
| local-bin/task-board-main-91959c1a-nocgo-nop | ~/.local/bin/task-board-main-91959c1a-nocgo-nop | ac96126a4db778313220cccea6a4eec49b7fdc10dc6eca8e1c0f11e09aec97de | 27090416 |
| local-bin/task-board-main-ac2ad9c0-curatorlike | ~/.local/bin/task-board-main-ac2ad9c0-curatorlike | 9bfc3e68808076ed0fc16ea52b15310f405f970cdccb31344d09250d6513d05a | 27411152 |
| local-bin/task-board-main-e8205e78 | ~/.local/bin/task-board-main-e8205e78 | 9937fec400652ff1cc8b058bab4a02f199b469e1de412eae9a3762312a308ef5 | 28304256 |
| local-bin/task-board-main-e8205e78-nocgo | ~/.local/bin/task-board-main-e8205e78-nocgo | e582927b5e87caa79ebd5b0042820322b801b6a5e2320175b72815b24047ee2b | 27077776 |
| local-bin/task-board-main-e8205e78-nop | ~/.local/bin/task-board-main-e8205e78-nop | a49a4a3a458b2888cd2339433932ccf5044bce6cc5e157b5bba883e2fabc03ad | 28066672 |
| global-bin/task-board | ~/.curator/global/bin/task-board | be8fbf92bec559c862b92e8a7e0a2956be802577168bffadcd0b00fc92cd9be7 | 306 |
| global-bin/task-board-tui | ~/.curator/global/bin/task-board-tui | 37c5a7d222be3f8a1935638c0a642ef9ad2b296ea7c88051aaa405b9e2dabdda | 310 |
| global-bin/tb-sessiond | ~/.curator/global/bin/tb-sessiond | 04804e481bef92ad8896583e0c4984f9a30ad40aa7e96983243fb8b877d03623 | 307 |

Coverage vs brief: all 10 legacy standalone binaries (brief named 7; the 3
`e8205e78*` extras are included), all 6 current shims (3 in ~/.local/bin incl.
`tb-sessiond`, 3 in ~/.curator/global/bin), plus the marker JSON (needed as a
pre-image for the managed-publication step).

Deliberately NOT copied: `~/.local/bin/curator` (live product binary, version
main-04550e2, neither legacy nor shim; hashed in the inventory instead).

Nothing was moved or removed: sources re-hashed after the copy match the BEFORE
hashes in TASK-260915-s4qkmr_inventory.md (see the re-verification loop:
manifest sha == source sha == backup sha for all 17).
