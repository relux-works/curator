# Review note — TASK-260923-2elcdc isolated environment lock (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2elcdc-brief.md` and curator-spec main text (environments §12.2, manager §1 rule 1, system-config-v2 environments.isolation;
F-S3 fe2d1c6). Verify each rule through the CLI production entry: `isolated` admitted; silence resolves to the locked direction; explicit
`shared` under an engaged isolated lock refuses with environment_isolation_lock_conflict; an already-provisioned shared passthrough fails
closed toward the explicit F-C2 migration (never silently migrates). Pinned-suite (dcc7f015) isolation vectors driven if present. Kill at
least two of: silent migration / explicit shared admitted under the lock / silence resolving to shared. Consistency with 0017 (cww1ov) code
on main. No CHANGELOG/LOGBOOK edit, no stray files, hosted gate green. Focused bounded runs (host memory). accept_cr or changes requested
with file:line. No LOGBOOK.md.
