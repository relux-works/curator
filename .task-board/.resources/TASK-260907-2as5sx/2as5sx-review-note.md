# Review note — TASK-260907-2as5sx §8.4 absence-vs-failed-read as a class, CR revision 5 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

First content review. Final leaf of STORY-260906-1a2i5a (with checkpointed TASK-260907-187z6x). Review against `2as5sx-brief.md`, reworks 1-3
and the results; source of truth curator-spec environments §8.4/§8.4.1 and the marker/state sections. Read-only disposable clone.
1. The shared seam `internal/stateread`: three outcomes kept distinct where the spec does — absent (proven not-exist), read-failed
   (unreadable, typed `manager_state_unreadable`), read-OK-but-invalid (per-reader rule, e.g. re-derive for the external-evidence marker case).
2. The site inventory is complete (the five review-cycle sites + the rest); the AST inventory test really fails on a new
   `IsNotExist`-then-default pattern outside the seam (re-apply that mutant yourself).
3. Rework regressions fixed: crossconformance `external-evidence-mismatch-execution_policy` re-derives; `TestSurfacingEmittedDespitePublicationFailure`
   emits §2.3 rows before the failure (check the chosen order against the spec clause the producer cites).
4. Windows: every POSIX mode-bit unreadable test has its own `linux,darwin  windows  platform-control` ledger row; no shared DACL helper touched;
   internal/godriver byte-identical to base.
5. No stray files; gofmt clean; mutants per site killed (table) — re-apply one yourself; validation log green.
accept_cr or changes requested with file:line. No LOGBOOK.md.
