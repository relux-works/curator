# Review note — TASK-261002-1ig5ev rev1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138). curator-spec CR, base e41c561b, 13 paths. Review it against `spec-history-brief.md`.

Verify, with real exit codes:
1. `git diff --quiet v1.0.0-rc.13 <candidate> -- release/1.0.0-rc.13.json` exits 0: byte-identical to the tag.
2. The generator no longer writes published release records.
   - Run the generator on the candidate and confirm the rc.13 record is untouched.
   - Say where candidate digest bookkeeping now lives, and check that downstream consumers (implementations.yml, curator digest dispatch) are not broken.
   - `make regenerate-check` exits 0.
3. The freeze guard fails when the rc.13 record is mutated: run the negative yourself. It covers every published record, not only rc.13. Released schemas are untouched.
4. `make validate` and the tools tests, or the relevant subset.
5. Hygiene: one CHANGELOG line; no LOGBOOK; no stray files; never spell any employer name.

accept_cr, or changes requested with file:line.
