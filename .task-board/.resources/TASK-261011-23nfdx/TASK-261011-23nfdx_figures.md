# TASK-261011-23nfdx — R-CS3 figure handoff

Documentation-only authoring against the current CIPs after R-CS1/R-CS2. Audit source: B4 at curator revision 48789217; contract baseline: 177cb554, with the later recorded owner dispositions. No runtime support is inferred.

| Figure | File | Spec section | Audit row |
|---|---|---|---|
| Donor components (C4 retained) | `cips/diagrams/donor-side.dsl` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D3, D5, D12 | M-CIP1 |
| Donor components (PlantUML) | `cips/diagrams/donor-side.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D3, D5, D9, D12 | M-CIP1 |
| First deployment | `cips/diagrams/donor-first-deployment.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D4, D9–D10, D12 | M-CIP1 |
| Join handshake | `cips/diagrams/join-handshake.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D4, D6, D9, D12 | M-CIP1 |
| Remote-worker turn | `cips/diagrams/remote-worker-turn.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D5, D12; CIP-0008 R3/R5/R9 | M-CIP1 |
| Retirement sequence | `cips/diagrams/remote-worker-retire.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D7, D12 | M-CIP1 |
| Deployment variants | `cips/diagrams/donor-variants.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D3; CIP-0008 R9 | M-CIP1 |
| Independent lifecycle axes | `cips/diagrams/donor-lifecycle.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D7, D12 | M-CIP1 |
| State layout | `cips/diagrams/donor-layout.puml` | `CIP-0009-donor-side-deployment-and-bridge.md` — D0–D3, D5, D7, D12 | M-CIP1 |
| Credential source lifecycle | `cips/diagrams/credential-source-lifecycle.puml` | `CIP-0010-credentials-setup-token-and-inherited-auth.md` — C1–C6, Revision 2.2; CIP-0003 migration; CIP-0011 §2/§3 | M-CIP2 |
| Configuration publication | `cips/diagrams/configuration-publication.puml` | `CIP-0002-project-context-in-managed-launches.md` — CIP-0002 admission/leases/PATH; CIP-0004 recommendation/migration; CIP-0006 design/migration | M-CIP2 |
| Tool admission | `cips/diagrams/tool-admission.puml` | `CIP-0007-manager-provisioned-cli-tools.md` — Operator decision, option B, recommendation/security; shared D-PATH | M-CIP2 |
| Audit transport | `cips/diagrams/audit-transport.puml` | `CIP-0005-audit-backends-and-cli-secret-transport.md` — Recommendation §§1–7 | M-CIP2 |

All 12 PlantUML figures and the retained C4 figure are linked from the root README, source index and relevant CIP sections. CIP-0002 through CIP-0011 have operative summary tables for purpose, boundary, authority, inputs/outputs, phase and normative/proposed landing links; existing disposition blocks remain.

Second-review follow-ups: F1 restores identity deduplication and preservation of an existing dispatcher position in the shared disposition and CIP-0002; F2 removes the private input filename; F3 marks rule-derived transport/request/record names as pending matching module vectors.

Visible bounds: B remains disabled until the separate provisioner contract/qualification; A selection after B/C is open; channel/backend ownership, join schemas, exact service UIDs, state filenames, retention and detailed recovery formats remain unspecified where stated. Source-policy v1/v2, broker slices, current native behavior and SH2/later executor targets retain separate labels.

Validation: no tests, builds, renderer or standalone validation/lint commands run by the author, as explicitly required by the current task brief. Syntax and source correspondence were read manually. Configured CR validation is delegated only to `task-board handoff TASK-261011-23nfdx --role developer`; its actual result/exit is board evidence, not a claimed author-run test. No test/build/lint checklist pass is asserted here. No LOGBOOK edit and no commit.

Checklist scope reconciliation: the inherited role template’s test, lint, build/validation and LOGBOOK items were left unchecked while present and removed as obsolete under the explicitly exclusive current brief (“Docs only”, “Do not run tests or builds yourself”, “Never write LOGBOOK.md”). They are not recorded as passing. Configured CR validation remains the required handoff mechanism, with its own command evidence. Figure authoring, documentation acceptance and outcome attachment remain the applicable checklist requirements.
