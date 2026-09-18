# Brief — TASK-260910-3u9t1e: `verify-backup` never trusts the live home's keys silently (R7, service)

Story `STORY-260910-2xe3n2` (registry-robustness-hardening, `EPIC-260910-16qce1`),
wave 4; third of four sequential leaves (after R5 and R8, both checkpointed
on the Story branch — build on them). Rules:
`remediation-registry-producer-rules.md` (attached). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-2xe3n2/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` R7 (Low): when `--public-key` is omitted,
`verify-backup` verifies the candidate backup and checkpoint with the
registry's own keyring under `--home`; if the live home is compromised, so
is the backup check. Require or loudly warn on implicit key resolution.

## Settled decision (do not reopen)
Explicit key is the documented expectation: `verify-backup` without
`--public-key` still works (operators' scripts must not break) but prints a
prominent warning on stderr AND records it in the command's structured
output/exit summary ("keys resolved from the live home <path>; supply
--public-key from an out-of-band copy for an independent check"), and
`README.md`/`SECURITY.md` document the out-of-band key expectation for
backup verification (where the pinned public key comes from, how to keep
it). No behaviour change with `--public-key`. Exit code unchanged
(warning, not refusal) — a refusal would be a breaking CLI change outside
this finding's scope.

## Deliverable
1. The warning path in `cli.py` `_cmd_verify_backup` (+ any shared
   key-resolution helper), covered by tests: implicit resolution warns
   (stderr text + structured field), explicit key does not; the verdict
   itself unchanged in both cases.
2. Docs as above; `CHANGELOG.md` Unreleased entry "R7: …".

## Out of scope
R4 docs (next leaf), key management (R6, another story), spec edits.

## Checklist and handoff
Tick the checklist items; attach `TASK-260910-3u9t1e_results.md` (file:line,
transcripts: `python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT`,
`python -m mypy`), then `task-board handoff TASK-260910-3u9t1e --role developer`.
