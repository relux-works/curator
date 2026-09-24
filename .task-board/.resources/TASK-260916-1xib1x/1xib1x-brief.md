# TASK-260916-1xib1x brief (orchestrator, binding) — R4 script audit labels

Story STORY-260822-2h0v9j; R1–R3 checkpointed on the Story branch (worker, capabilities/controls,
probes/evidence/preflight). Independent of R3 internals; touches the audit/validation output only.
Spec at the CI pin 87a0d006: `profiles/manager.md` audit-label section (~1031: `script-command-
declared-only` — a script command that does not declare `execution_policy: "script-worker-v1"`,
i.e. every schema-7 script command and every schema-8 script command without the field;
`script-command-unfiltered-declared-network` — an enforced script command whose declared network
hosts are reporting-only, i.e. not filtered by the manager); vector
`script-host-execution-policy.json` `audit_label_cases` (4): `schema7-script` → [declared-only];
`schema8-declared-only-script` → [declared-only]; `schema8-enforced-script` → []; `schema8-enforced-
unfiltered-network` → [unfiltered-declared-network]. Code: `internal/audit` (`Gate`/`GateReadOnly`
:101/:107, warning classes), `cmd/curator/main.go` `auditTarget` (:1851), `internal/skillcheck`
(validation issues), the reconciliation note (`.research/TASK-260916-3gcc00_reconciliation.md`:
"declared-only warning class for schema7 and schema8 script commands and the unfiltered-declared-
network warning; enforced scripts are never labelled declared-only; drive all four audit vectors
through the production audit/validation output; preserve legacy install behavior").

## Rulings
R1 Labels are WARNINGS in the production audit/validation output (`curator audit`, the install
audit gate, `skillcheck` issues as the existing warning classes are surfaced) — never errors, never
install refusals: legacy install behaviour for declared-only scripts is unchanged (schema 7 and
schema-8-without-policy skills install exactly as today; goldens/rows prove it).
R2 Exact semantics: an enforced command (`execution_policy: "script-worker-v1"`) is NEVER labelled
declared-only; `script-command-unfiltered-declared-network` applies only to enforced commands whose
declared `network` hosts are non-empty (reporting-only per Protocol Core §4.3 — the manager does
not filter hosts); a declared-only command with network hosts gets ONLY declared-only (the
network label is about enforcement, not declaration) — confirm against the vector rows and say so.
R3 Evidence: all four `audit_label_cases` driven through the production audit output (CLI `audit`
and the install-time audit gate), each label text pinned exactly as the vector names it; negative
rows: enforced ⇒ no declared-only label (mutant: label enforced too → fails); declared-only with
hosts ⇒ no network label; mutants per label; the vector consumer in `internal/scriptpolicy/
conformance_test.go` reclassifies `audit_label_cases` from "not implemented" to consumed (R5 will
finish the rest). Windows: nothing platform-specific expected; rows run everywhere.
R4 Docs: `docs/troubleshooting.md`/audit docs entries for both labels (what they mean, what to do:
adopt `script-worker-v1`; declare only the hosts you need — reporting only); CHANGELOG `### Added`.
results.md: label rules table, row table, mutant table, ratio line (vector key consumed), bounds.
Publish only on a green gate.
