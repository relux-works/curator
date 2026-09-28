# Review note — TASK-260928-3ed9m3 curator global lock publication ordering (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 97e85642, tree 83a0c43a, 7 paths, gate green) against `3ed9m3-brief.md` and curator-spec main 4ad8042 (PR #113:
environments.md §9.4 + manager transaction rules; vectors/environments-global-lock-publication.json). Verify through the CLI:
1. global add / global install publish the extended profile lock BEFORE any in-place surface write; on environment_surface_unmanaged_conflict
   the published lock is KEPT (not rolled back) and the conflict names the surface; the five-operation takeover carrier set is unchanged (no
   --takeover on global ops).
2. After the conflict, `profile sync --takeover` (and `profile use --takeover`) materializes the updated skill set; a retry of the global op
   then succeeds / is a no-op. Rows mirror the three new vector cases and cite the spec text; the results note that the rows switch to the
   vector driver when SPEC_PIN moves past 4ad8042.
3. If the producer claims the code already conformed, check the evidence (order in code + a test that fails if the order is swapped).
4. Mutants (publish after surfaces; roll back the lock on conflict) killed with real exit codes. Manager-state reads via stateread; no
   CHANGELOG/LOGBOOK, no stray files/binaries.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
