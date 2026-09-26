# TASK-260923-2gt5f6 — publish the passthrough credential record (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. 0017 credential modes (STORY-260922-1cenbr, TASK-260922-cww1ov) are on curator
main; read that code first (stateread seam, credential link provisioning/repair/migration, marker handling). Spec: curator-spec main
(environments + agent-environment-marker-v2 schema, credential record members: isolation, strategy, source_role, backend, backend_version,
provenance; path only for linkable strategies).
1. On provision, repair and migration of a passthrough credential link, write the marker-v2 credential record under the manager-home lock
   via same-directory temp + atomic rename with journal protection; never rewrite a schema-1 marker just to add the record.
2. Production-entry rows through the CLI for provision / repair / migrate, schema-1 marker left untouched, crash between temp and rename
   (journal recovery), and the conformance vectors for the record if the pinned suite (dcc7f015) carries them (drive them; otherwise state).
3. Mutants: record written without the lock; schema-1 marker rewritten; path written for a non-linkable strategy — each killed.
No CHANGELOG/LOGBOOK edit (entry text in results). Focused bounded runs (host memory is tight). Attach results, check DoD,
`task-board handoff TASK-260923-2gt5f6 --role developer`. A write-boundary `policy warn` block is a warning.
