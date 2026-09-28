# TASK-260923-xq4pjj — backup manifest

Backup directory: /Users/administrator/.curator/backups/20260923T121543Z-shim-adoption

Both files were copied with cp -p before the shim swap. SHA-256 values match the source and backup; cmp exited 0 for each pair. stat showed source and backup modes, sizes, and mtimes identical.

| Original path | Backup path | SHA-256 | Mode | Size | Original mtime |
|---|---|---|---|---:|---|
| /Users/administrator/.local/bin/task-board | /Users/administrator/.curator/backups/20260923T121543Z-shim-adoption/task-board | 5092fb50d34cdd27540531abdb71b3e68111866842becc8e117101fabf1a8cdb | 711 | 72 | 2026-09-22T23:51:12Z |
| /Users/administrator/.local/bin/.curator-managed.json | /Users/administrator/.curator/backups/20260923T121543Z-shim-adoption/.curator-managed.json | 7a4efaacaf3d78c6a5dae6f838c67500b367ded96d1529593bb2e7e6ab55a3d5 | 600 | 64 | 2026-09-22T23:59:47Z |

Commands and exit codes:
- cp -p shim: 0
- cp -p marker: 0
- shasum -a 256 on source and copies: 0, source/copy hashes equal
- cmp shim source vs backup: 0
- cmp marker source vs backup: 0
