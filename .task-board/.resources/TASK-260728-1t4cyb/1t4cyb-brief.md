# TASK-260728-1t4cyb — external-repository guide: driver threat-review checklist (THE ONLY CURRENT INSTRUCTION; re-scoped 2026-09-30)

curator-spec repository. The reconciliation audit (curator TASK-260930-2mtgv7) found that docs/external-build-repositories.md already
covers package authoring, development substitutions, operator requirements and shared-suite consumption. The one gap is that it has
no threat-review checklist for FUTURE build drivers. Ivan approved only this part: cross-manager parity is dropped.
1. Add a section to docs/external-build-repositories.md, "Threat review for a new build driver". It is a concise checklist, and every
   item cites the normative clause it enforces: the build-driver security model, the external-repository rules, and the
   skill-build-v1 / schema 7 rules. Cover at least:
   - network during build;
   - source/toolchain read-only exposure;
   - write roots;
   - process tree and resource bounds;
   - executable allowlist and argv provenance;
   - environment scrubbing;
   - cache identity inputs;
   - receipt fields;
   - symlink and hard-link handling;
   - substitution and dry-run behaviour;
   - fail-closed ordering (admission → audit → cache → compiler).
2. Keep it informative: add no new MUST, change no vectors, no schemas.
3. Make sure the existing examples still validate against schema 7 (the task AC). Run tools/validate.py and the doc/link checks, and
   record the real exit codes. CHANGELOG entry under Unreleased. No LOGBOOK.md. Never spell any employer name.
Update the results, then run `task-board handoff TASK-260728-1t4cyb --role developer`, then END YOUR TURN.
