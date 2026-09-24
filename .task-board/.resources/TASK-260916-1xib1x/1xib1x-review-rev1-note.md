# Review note for TASK-260916-1xib1x revision 1 (orchestrator, binding) — R4 script audit labels

Brief 1xib1x-brief.md (R1 labels are warnings in the production audit/validation output, never
errors/refusals, legacy install behaviour unchanged; R2 exact semantics: enforced never
declared-only; unfiltered-declared-network only for enforced commands with non-empty declared
hosts; declared-only with hosts gets only declared-only; R3 all four `audit_label_cases` driven
through `curator audit` and the install-time audit gate with exact label text, negative rows,
mutants per label, vector consumer reclassified; R4 docs + CHANGELOG). Spec at the pin:
`profiles/manager.md` audit-label section (~1025–1045: both classes REQUIRED, always `warn` in
every mode, never block, not subject to `fail_on`, never reported as an applied control).
Gate green: run 35698876152 — verify the gate commit resolves to the exact revision-1 tree.
Patch: internal/scriptpolicy/{labels.go,labels_test.go,conformance_test.go}, internal/audit/
{audit.go,auditlabels_test.go}, internal/install/{install.go,global.go,auditlabels_test.go},
cmd/curator/{main.go,auditlabels_test.go}, internal/skillcheck/auditlabels_test.go, docs, CHANGELOG.

Judge with your own reruns (disposable clone; build the binary; bounded commands):
1. Spec conformance: the label texts are exactly `script-command-declared-only` and
   `script-command-unfiltered-declared-network`; both surface as `warn` in every mode (strict too),
   never as errors, never trip `fail_on`, never appear in applied-control evidence (grep); the
   audit record carries the effective execution-policy identity or its absence per script command.
2. Semantics vs the four vector rows (schema7-script; schema8-declared-only-script;
   schema8-enforced-script → no labels; schema8-enforced-unfiltered-network) at BOTH production
   entries (`curator audit` CLI and the install-time gate incl. global scope — `global.go` changed);
   the declared-only-with-hosts case gets only declared-only; mutants: label enforced too; drop
   the network label; emit as error — each fails a row.
3. Legacy install behaviour unchanged for declared-only scripts (goldens/existing rows); the
   install.go/global.go changes add warnings only — diff them.
4. Vector consumer in `internal/scriptpolicy/conformance_test.go` now consumes
   `audit_label_cases`; ratio line honest; docs entries accurate.
Record exactly one verdict: accept_cr(TASK-260916-1xib1x, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write into the
control root's LOGBOOK.md.
