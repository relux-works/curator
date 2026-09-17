# Brief — TASK-260910-1ny7yl: `serve --checkpoint` — startup comparison against the operator checkpoint (R3 + P2, service)

Story `STORY-260910-35tbgb` (serve-time-checkpoint-gate, `EPIC-260910-16qce1`),
wave 3; the spec revision landed as curator-spec `47c3c8c`
(`profiles/registry-service.md` §6 "Backup and restore" — the startup
checkpoint comparison — plus §5/§9/§10/§11 and registry §5 cross-references;
vectors `registry-service.json` `checkpoint_cases`, pinned by the spec
validator). Single leaf of the story. Rules:
`remediation-registry-producer-rules.md` (attached). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-35tbgb/worktree`
(forked from `main` `131952d`, which carries R1/P1 and R2).

## Finding (read it first)
`docs/security-audit-2026-09.md` "R3 + P2": `Store.__init__` verifies internal
integrity and `verify-backup` compares a candidate backup against a signed
checkpoint, but `serve` never compares live state against the operator's
out-of-band checkpoint, so a silently restored older database passes startup
and serves stale state.

## Deliverable (exactly the landed §6 rule)
1. `serve --checkpoint <path>` (and the environment-variable form the CLI
   uses for other options, if any): the file is a signed
   `registry-snapshot-v1` object; at startup, AFTER the §5 integrity
   verification (incl. the R2 boundary/frontier rebuild) and BEFORE the
   listener binds or `/health` can report ready, verify the checkpoint
   signature against the accepted signing keys (the staged-rotation set) and
   compare the live boundary with it:
   - live `version`/`log_size` below the checkpoint → refuse
     `restore_below_checkpoint`;
   - equal version with different `head`, `merkle_root` or `log_size` →
     refuse `restore_inconsistent_with_checkpoint`;
   - above the checkpoint → serve only if the live log reproduces the
     checkpoint boundary at its `log_size` (head and Merkle root at that
     prefix — use the R2 `boundaries` rows), else
     `restore_inconsistent_with_checkpoint`;
   - signature failure → refuse `checkpoint_signature_invalid`.
   Refusal = the process stays up non-ready (`/health` 503 through the
   common integrity path, writes disabled), never exits with a truncated or
   repaired store, never invents history; the diagnostic code is in the
   structured startup log and the error envelope.
2. Without `--checkpoint`: start as today and record
   `checkpoint_not_configured` in the structured startup log/audit events
   (posture); with one: record the compared boundary (version, log_size,
   head) and the outcome. `health-response-v1` schema unchanged.
3. `verify-backup` stays; README/SECURITY document which is normative for
   "before the service becomes ready" (the startup comparison) and how to
   produce/rotate the checkpoint file.
4. Conformance: drive `registry-service.json` `checkpoint_cases` through the
   real startup path (every case: below, equal-consistent, equal-inconsistent,
   above-consistent, above-inconsistent, bad signature, none configured) in
   `tests/test_protocol_conformance.py`; unit tests for the CLI flag and the
   AC scenario "restore an older database and prove serve refuses";
   existing `recovery_cases` and R2 tests stay green.
5. `.github/workflows/ci.yml` protocol-suite `ref` → `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`
   (the vectors live there); every pre-existing conformance test green at
   the new pin.
6. `CHANGELOG.md` Unreleased entry "R3/P2: …"; `compose.yaml` example of
   mounting a checkpoint file if the service is started through compose.

## Out of scope
P4 import high-water (`STORY-260910-stz5f0`), R4–R8, spec edits, key
rotation changes.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-1ny7yl_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-1ny7yl --role developer`; the runtime runs
the hosted gate once at handoff.
