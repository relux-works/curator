# Brief — TASK-260910-2g5v17: per-upstream high-water for `import-bundle` (P4, service)

Story `STORY-260910-stz5f0` (import-upstream-high-water, `EPIC-260910-16qce1`),
wave 4 of the 2026-09 security-audit remediation; single leaf. Rules:
`remediation-registry-producer-rules.md` (attached; `main` is now `c7ef32c`,
the CI protocol-suite pin is already `dced9b8` — do not move it). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-stz5f0/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` P4 (Low): `import_bundle` verifies
signatures, chain, head, size and Merkle root, but never compares the
upstream snapshot against a persisted high-water for that upstream key — an
old but validly signed bundle imports as new.

## Settled decisions (do not reopen)
- The store persists, per upstream public key (`key_id` of the
  `--upstream-key`), the highest accepted upstream boundary
  (`version`, `log_size`, `head`, `merkle_root`) in the same transactional
  store discipline the R2/R3 boundary state uses (a table, not a side file;
  protected exactly like the rest of `registry.db`).
- Import compares the bundle's upstream snapshot with that high-water under
  the client §5 rollback rules: version below → refuse; equal version with a
  different `head`/`merkle_root`/`log_size` → refuse; equal and identical →
  accepted with nothing persisted (a no-op re-import); higher → import, then
  persist the new high-water in the same transaction as the imported
  records (never before the import commits, never partially).
- Fail closed by default: refusal is a non-zero exit with a closed
  diagnostic naming the upstream `key_id`, the persisted and the offered
  boundary (working spellings `import_upstream_rollback` and
  `import_upstream_inconsistent`; keep them if nothing in the landed spec
  names these situations — check `profiles/registry-service.md` first and
  reuse an existing code if one applies). An explicit operator flag
  `--accept-older-upstream` turns the version-below refusal into a warning
  and imports WITHOUT lowering the persisted high-water; the inconsistent
  (same-version-different-body) case is never overridable. This is the
  orchestrator's settled reading of the audit's "reject (or warn under a
  flag)" consistent with the campaign's fail-closed rule.
- No protocol/wire change, no spec edits; the persisted state is an
  implementation detail of the service and is included in `backup` and
  compared by `verify-backup` only if that is cheap — state your choice.

## Deliverable
1. Store: the per-upstream high-water table + accessors, transactional
   update with the import.
2. CLI `import-bundle`: the comparison, the diagnostics, the flag; the
   structured log records the compared boundaries and the outcome.
3. Tests: rollback bundle (older version) refused; same version different
   body refused even with the flag; identical re-import no-op; newer bundle
   imports and advances the high-water; the flag imports an older bundle
   with a warning and leaves the high-water untouched; a failed import
   (e.g. chain break) leaves the high-water untouched; first import of an
   unknown upstream establishes the high-water.
4. Docs: `README.md` import section (the rule and the flag), `SECURITY.md`
   (what the high-water protects against), `CHANGELOG.md` Unreleased
   entry "P4: …".

## Out of scope
Client-side (curator) behaviour, R4–R8 tasks, spec edits, mirror-group
comparisons (S2 is a client-side rule).

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-2g5v17_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-2g5v17 --role developer`; the runtime runs
the hosted gate once at handoff.
