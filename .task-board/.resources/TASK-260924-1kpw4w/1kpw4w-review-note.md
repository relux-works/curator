# Review note — TASK-260924-1kpw4w `dependencies.skills[].directory`, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `1kpw4w-brief.md`, the operator memo and the ACCEPTED curator-spec amendment TASK-260924-2am4qa (patch resource
`TASK-260924-2am4qa_change-request_rev2.patch`: core §4.4, agent/csk-skill-v9 schemas, vectors). Disposable clone.
1. Grammar/containment exactly the Skillfile individual-selector `directory` rules (shared validator, no second grammar); root → ".".
2. Identity/closure/lock/audit carry the normalized directory; same-folder diamond unifies, different-folder same-name conflicts per spec;
   selected folder needs SKILL.md with matching name; symlink escape refused with the spec's diagnostic.
3. Absent `directory` byte-identical: the schema-8 lock-bytes golden (sha256:6855…) holds; v1-v8 manifests unchanged.
4. Vectors copied byte-for-byte from the accepted spec candidate (check the cited sha256 values against the patch) and driven.
5. Real CLI install/refresh/reinstall with lock, marker and audit; mutants killed — re-apply one yourself.
6. No CHANGELOG edit (policy) — the entry text is in the results; no stray files; validation log green.
accept_cr or changes requested with file:line. No LOGBOOK.md.
