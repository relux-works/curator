# TASK-260910-1xph2y Logbook

- Initial required set_status command exited 1 because an estimate was missing.
  Set estimate 8, then set_status development exited 0. No gate was bypassed.
- System Python lacked jsonschema (ModuleNotFoundError, exit 1). Installed only
  requirements-dev.txt into the task venv; pip installation exited 0.
- Existing corpus couples schema discovery, generator output and release hashes.
  Kept new draft schemas and cases isolated instead of rewriting frozen evidence.
- No logbook CLI/connector was available; scoped board schema lookup confirmed
  unknown operation logbook (exit 1). This attached task-scoped Logbook section
  and the separate logbook resource persist the findings through resource CRUD.
- Routine syntax chosen: explicit repository field for logical identity;
  source-policy.json ordered endpoints with optional pin; root_inputs for root
  package admission. Advanced ports/mirrors/aliases remain unsupported and are
  recorded in UNRESOLVED_QUESTIONS.md.
- Source-audit is a machine-local evidence binding, never a registry attestation.
  Required network evidence for a local source fails closed. Verified script
  operations are not invented; existing assurance build input fields bind receipt 3.

