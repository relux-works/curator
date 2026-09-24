# TASK-260916-1xib1x rework 1 (orchestrator, binding) — R4

Verdict rev1: CHANGES_REQUESTED (TASK-260916-1xib1x_review-verdict-rev1.md); the label semantics,
warn-only behaviour, fail_on bypass and docs are verified (4/5 reviewer mutants killed). Continue
from the revision-1 tree (no checkout/clean/stash); fix exactly:

1. Per-command audit identity (spec manager §7: "for every script command the record carries the
   effective execution-policy identity or its absence"): `AuditLabelsForCommands` (labels.go:78)
   drops commands with no warning, and `Report`/the persisted verdict (audit.go:75/:168/:329)
   carry no per-command policies — an enforced no-host skill audits as only `app: audit clean`.
   Add a per-command record independent of warning eligibility (command name → exact
   `script-worker-v1` or explicit absence), expose it through the production audit record (CLI
   output line(s) + stored verdict), and test mixed skills and no-warning commands without adding
   spurious warnings; fix the CLI row that currently asserts the bare "audit clean" output.
2. Global-install coverage: the four vector rows + the declared-only-with-hosts negative row must
   also run through `install.Global` with its DEFAULT production gate (global.go:189 wiring); assert
   each audit warning on the same message line as the exact label, distinct from the skillcheck
   warning, no errors, legacy shims unchanged; the reviewer's global-wiring mutant (Subject.Commands
   → nil in global.go only) must now be killed.
Append "Revision 2" to results.md; republish only on a green gate.
