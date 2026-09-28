# TASK-260928-3ed9m3 — curator: global add/install lock publication ordering (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. The spec side landed on curator-spec main 4ad8042 (PR #113, TASK-260928-q5100t): environments.md §9.4 +
manager transaction rules now say `global add` / `global install` publish the extended profile lock BEFORE in-place materialization and KEEP
it when a surface write meets environment_surface_unmanaged_conflict, so `profile sync --takeover` / `profile use --takeover` then a retry
recovers; the takeover carrier set is unchanged. New vectors: conformance/v1/vectors/environments-global-lock-publication.json (spec main, not
yet in curator's pinned rc.13 suite).
1. Read curator's global add / global install path (cmd/curator, internal/envprofile / install / globalbins as applicable): establish the
   current order of lock publication vs in-place surface writes and what happens to the lock on a surface conflict. If it already conforms,
   say so with evidence; otherwise make it conform (publish-before-materialize, keep the lock on conflict, no rollback of the lock).
2. Production-entry rows (through the CLI) mirroring the three new vector cases: lock published before surfaces; lock preserved on
   conflict; sync --takeover after the conflict materializes the updated skill set then a retry succeeds. Cite the spec main clause text.
   Mutants: lock published after surfaces; lock rolled back on conflict — killed, real exit codes.
3. No gap-ledger rows are needed (the vectors are not in the pinned suite); note in the results that the rows must be switched to the vector
   driver when SPEC_PIN moves past 4ad8042.
No CHANGELOG/LOGBOOK edit (entry text in results). Manager-state reads via internal/stateread. Update the results resource, handoff, END
YOUR TURN. Write only inside your Story worktree.
