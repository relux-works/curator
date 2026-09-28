# Review note — TASK-260928-q5100t spec: global add/install lock publication ordering (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 17d88795 = spec main, tree d4880af0, 9 paths, validators green) against `q5100t-brief.md`, the operator decision
(alternative 1 of `TASK-260906-1xbrz6_global-recovery-decision-packet.md`) and the existing text of environments.md §9.2 (steps 4–5), §9.4,
§9.5, §8.3 and the manager transaction rules. Verify:
1. The new normative text says: global add / global install publish the extended profile lock BEFORE in-place materialization; on
   environment_surface_unmanaged_conflict the published lock is KEPT; the §9.4/§9.5 recovery (profile sync --takeover or profile use
   --takeover, then retry) is now true; the five-operation takeover carrier set is unchanged (no new flag); consistent with §9.2's
   publish-before-rematerialize wording; no contradiction elsewhere (grep every mention of global add/install, takeover, and rollback).
2. Vectors cover: lock published before surfaces; lock preserved on conflict; sync --takeover after the conflict materializes the updated
   skill set; the manifest/schema/generator wiring is consistent (tools/generate-vectors regenerate check, validate.py).
3. The validator test that pins the takeover closed-set text (TakeoverClosedSetTextTests) was adapted without weakening it: it must still
   fail when the §9.5 sentence moves into §9.4 or the carrier set widens — show a mutant.
4. CHANGELOG entry under Unreleased; no released-text edits to rc.13 records; no stray files.
Run tools/validate.py + the unit tests with real exit codes. accept_cr or changes requested with file:line. No LOGBOOK.md.
