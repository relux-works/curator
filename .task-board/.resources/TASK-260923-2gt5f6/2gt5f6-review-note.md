# Review note — TASK-260923-2gt5f6 passthrough credential record (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev2; rev1 failed macOS Test and a successor fixed it — name that fix and judge it) against `2gt5f6-brief.md`
and curator-spec main (agent-environment-marker-v2 credential record: isolation, strategy, source_role, backend, backend_version, provenance;
path only for linkable strategies). Through the CLI production entry: record written on provision / repair / migrate under the manager-home
lock via same-directory temp + atomic rename with journal protection; schema-1 marker never rewritten just to add the record; crash between
temp and rename recovers via the journal. Pinned-suite vectors for the record driven if present. Kill at least two of: record without lock /
schema-1 rewritten / path for non-linkable strategy. Consistent with 0017 (cww1ov) on main. No CHANGELOG/LOGBOOK edit, no stray files,
hosted gate green. Focused bounded runs (host memory). accept_cr or changes requested with file:line. No LOGBOOK.md.
