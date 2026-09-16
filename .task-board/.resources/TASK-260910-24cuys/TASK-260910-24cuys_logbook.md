# TASK-260910-24cuys logbook

2026-09-15: Accepted source contract is present on local curator-spec HEAD
3535d63ea80f97bba2fcb6e1f06996cfc25cf7df. The task had no previous implementation
outcome. Default parser callers must remain frozen-v1; draft parsing is admitted
only by a reader-owned option, not package fields or a release capability claim.

Found a boundary difference: buildrepo.ValidRefName limits bytes while the draft
schema gitRefName maxLength limits Unicode scalars. Added an isolated draft
validator and production-byte-parser tests for 255/256 non-ASCII characters,
leaving the legacy validator untouched. No compensating transport/runtime work.

Host lacks golangci-lint (command exit 127) and Python jsonschema. No installation
attempted. Scoped Go tests, build, vet and formatting are green; lint remains
unverified. Three narrowing gate mutants were caught and reverted.

Stored as a task-scoped board logbook outcome because campaign rules prohibit
editing LOGBOOK.md or the control root.

Handoff subsequently refused (exit 1) because lint item 5 is unchecked.
Existing-binary search did not locate golangci-lint in the bounded existing roots.
The no-installs campaign boundary prevents autonomous provisioning. Task blocked
pending an approved binary path/provisioning or rerouting to a lint-equipped host.
Do not mark lint green from vet/format results or bypass the handoff checklist.


## 2026-09-15 resumed producer
Provisioned linter is available. Two QF1001 findings corrected by equivalent Boolean rewrites; scoped lint, five-package tests, build, vet and formatting exit 0. External lint blocker resolved. See resume-results outcome for exact commands and bounds. No LOGBOOK.md edit.
