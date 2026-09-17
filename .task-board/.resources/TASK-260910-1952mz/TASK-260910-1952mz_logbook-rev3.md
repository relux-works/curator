# TASK-260910-1952mz — producer logbook rev3 (2026-09-17)

Rework of review-verdict-rev2 (changes_requested → corrections R1–R4).
Full transcripts in `TASK-260910-1952mz_results.md` (rev3 appendix).

## Decisions

- R1: malformed candidate record ⇒ `shell_hook_env_unapproved` (no valid
  record authorizes); unreadable/non-regular state ⇒ refuse under both
  profiles (read failure is never absence). Profile default baked per
  hook so an unset variable cannot degrade B to A.
- R2: POSIX hook strictly `sh`/`dash`-parseable; bash array handling moved
  behind a `BASH_VERSION`-guarded `eval`. Tests run every POSIX trust
  execution under sh+dash+bash+zsh. Windows POSIX skips KEPT (verdict
  asked to lift them; rework brief did not scope it; MSYS bridging
  unverifiable from macOS) — flagged for the orchestrator.
- R3: realpath identity both sides (Go `EvalSymlinks` + ancestor fallback;
  hook `cd -P`/`pwd -P` + `readlink` loop, `GetFullPath` + `.Target` loop).
  Warnings name the canonical path.
- R4: single-rename publish; `Revoke` reports `removed=false` when the
  publication fails (record still present).
- Seconds-60 rejected in hooks to match Go `time.Parse` (probed);
  absurd zone offsets rejected in hooks though Go accepts them (fail
  closed; real records always carry `Z`).

## Anomalies (host)

- Shared macOS host: fresh test binaries intermittently stall minutes at
  first exec with ~0 CPU (two `install.test` at 0:00 CPU; same-inode
  re-exec 0.022s) — first-exec security assessment delay. Mitigated with
  `go test -c` + warm-binary `-test.run` chunks. Full `./internal/install/`
  green across chunks; no assertion failed anywhere.
- Pre-existing corner kept: filesystem-root `//.agents/...` spelling
  mismatches Go `Clean` under bash (dash normalizes). Same in rev2;
  complete fix risks Cygwin UNC — reported, not patched.

## For the sibling (TASK-260910-3ungjy)

- `hookapproval` API unchanged in shape (`Upsert`/`ApproveFile`/`Revoke`/
  `List` + grammar consts); records now stored symlink-resolved.
  `Revoke` returns `(false, err)` when publication fails.
- Canonical warning path is the resolved path; approve UX should expect it.
