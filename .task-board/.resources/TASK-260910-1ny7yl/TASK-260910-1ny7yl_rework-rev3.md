# Rework brief — TASK-260910-1ny7yl, revision 3 (serve --checkpoint)

Revision 2 was rejected with three corrections
(`TASK-260910-1ny7yl_review-verdict-rev2.md`), all required as written.
Everything else passed; keep it unchanged unless a correction touches it.

1. **Startup posture/outcome events reach a real sink (P2).** The audit
   logger emits `startup_checkpoint` events at INFO, but `serve` builds the
   app before Uvicorn configures logging, so a real deployment emits nothing
   (0/2 in the reviewer's subprocess probe). Configure the audit/startup
   logging sink and level (structured, stderr or the configured audit sink)
   BEFORE the startup enforcement runs in `cli.py`, so that
   `checkpoint_not_configured`, the compared boundary and the comparison
   outcome are recorded in production; keep refusal events structured. Add
   CLI subprocess assertions (console entry point in a fresh process) for
   both the no-checkpoint posture and a successful comparison.
2. **Discriminating Merkle-root negative (P2).** Add a signed checkpoint that
   carries the genuine prefix head but a different `merkle_root` while the
   live store is above it; drive startup and assert
   `restore_inconsistent_with_checkpoint`, non-ready and writes disabled;
   confirm the narrowing mutant (drop the `merkle_root` comparison) now
   fails a committed test.
3. **The restore scenario traverses the CLI (P2).** Add an end-to-end test
   that restores the older backup and invokes the real
   `main([..., "serve", "--checkpoint", ...])` (or the console subprocess)
   with the factory and enforcement real, intercepting only Uvicorn's final
   run boundary to inspect the constructed app, and observes `/health` 503,
   write refusal and unchanged history. Keep the env-fallback and
   flag-precedence coverage.

## Validation and handoff
`python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT` set and `python -m mypy`
(exit codes, interpreter); update `TASK-260910-1ny7yl_results.md` with a
"Revision 3" section (per-correction file:line, transcripts incl. the
subprocess probe output) and hand off with
`task-board handoff TASK-260910-1ny7yl --role developer`. Worktree and rules
unchanged.
