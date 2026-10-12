# TASK-261012-2ff2r0 — apply-f23-operator-custody-cs: review handoff gaps

F-23 was applied to current text with the accepted 2026-10-12 owner amendment taking precedence over the original finding. This is a documentation disposition, not implementation or runtime qualification.

## Changed sections and figures

| Finding / brief item | Disposition and current location |
|---|---|
| F-23: separate operator and owner principal | CIP-0011 §5.1: no-harness `swarma-operator` is the configured broker operator/account owner, user-manager `operator` UID and holder of broker v0 root and all listed operator record authority. Interactive and agent accounts have no such authority; signing and direct-UID paths are covered. |
| Owner amendment: main phone path | CIP-0011 §5.2: non-exportable secure-element operator device key; biometric use only; key-only SSH over tailnet into `swarma-operator`; forced operator command, no shell; every act displayed and signed over exact canonical content, pinned-key verification, no cross-act reuse; single-biometric revocation. |
| Owner amendment: terminal fallback and sudo policy | CIP-0011 §5.2: password sudo into the separate account, no `NOPASSWD`, `timestamp_timeout=0`; widening also confirms exact content. No phone/additional hardware prerequisite for fallback. `NOPASSWD` stays only for dispatcher service helper/launcher rules. |
| Owner amendment: reads and command surface | CIP-0011 §§5.1/5.3: mandatory `swarma-observer` for status, grants, accounts and bindings reads without confirmation; enrol/grant require authentication plus content confirmation, revoke/disable authentication only. Broker metadata in `env status` follows the observer path. No grants-read or disable command spelling invented. |
| Owner amendment: exact widening list and agent refusal | CIP-0011 §5.3: grant issue, subscription enrolment, `team/1`, standing retirement or re-creation, explicit home retention, maintenance stop authority and standing-grant renewal. Standing retirement explicitly requires content confirmation under the owner rule. CLI widening refuses from an agent session; existing automatic narrowing/service authority stays bounded. |
| F-23: renewal, maintenance and loss consequences | CIP-0011 §5.3 references R-REBIND/R-MAINT: human operator authority or already-published bounded non-widening delegation, no inferred infinite re-grant; fresh authority for restart; fenced root replacement, no root restore into a session account. |
| Slice-0 acceptance and negative qualification scope | CIP-0011 Implementation plan item 2 and Test plan operator-custody row: custody/act gates, observer account, phone and terminal paths, publication and negative qualification required. Dedicated-account staging alone does not qualify the phone path. §6 removes the unqualified software-root slice-0 wording. |
| Shared cross-CIP principals/records | platform-contract-disposition.md Revision 4, Schema owner and version table R-OPERATOR row, Operator authority boundary, and open-question rows: fixed names and exact owner rules; device signer admitted to existing `team/1`/standing-retirement records, with format/admission mapping still unpublished. |
| Figures match changed operator acts | New operator-custody.puml; updated credential-source-lifecycle.puml, configuration-publication.puml, tool-admission.puml, audit-transport.puml, donor-side.puml, donor-first-deployment.puml, join-handshake.puml, donor-lifecycle.puml, donor-layout.puml, remote-worker-retire.puml and remote-worker-turn.puml. diagrams/README.md registers the new figure and revision. |
| Revision notes / Draft status | CIP-0011 Revision 6; shared disposition Revision 4; diagram source comments and figure README revision. CIP status remains Draft. |
| Fixed shared names / public hygiene | R-OPERATOR names used unchanged; earlier wave names/domains retained. New prose/figures have no personal first names, real host names, local absolute paths, session links or board-resource links. No new wire members, refusal codes, signature domains or algorithms allocated. |
| Docs-only scope / validation / logbook | Only CIP/spec Markdown, figure README and conservative PlantUML sources changed. No tests/builds/renderer run by the author, per brief. No LOGBOOK.md written, per explicit prohibition; material findings recorded here and in board notes. |

The existing donor-variants view has no operator act to change. The retained C4 component source describes a human project operator and unchanged component relationships without an operator OS-account/signing path; its semantics remain consistent and it is outside the instructed PlantUML edit scope.

## What stays open

CIP-0011 Open question 9 and shared disposition “Operator format and device-key lifecycle” explicitly state what would need publication:

- Forced operator command binary name and closed wire grammar.
- Signed operator record signature domain and canonical bytes.
- Device-key admission and mapping to existing broker root/grant/revocation verification, without an algorithm change implied by the new key class.
- Key enrolment, rotation and phone-loss policy; terminal re-enrolment under the authenticated fallback is a proposal, not a published recovery protocol.
- Joint phone secure-element/presence, SSH forced-command, terminal, observer and agent-session refusal qualification across both signing and direct-UID paths.

Existing delegation, E-rebind, maintenance, home-retention and fenced-root recovery publication/qualification questions remain with their component owners. No companion implementation or hardware capability is attested.

## Evidence and anomalies

- `git diff --check` ran directly after edits and returned exit code 0 (also returned 0 in the preceding three edit checkpoints).
- Manual readback covered CIP-0011, the shared table, the new operator sequence and every changed existing figure. This is source review only; no PlantUML renderer is available and rendering remains unverified.
- Tests and builds were not run by the author, as gw-CS5-brief.md explicitly forbids them; configured CR validation is delegated to `task-board handoff ... --role developer` and its own stored validation evidence. No pre-handoff suite result is claimed or accepted from older evidence.
- The audit figure previously launched an analyzer under “operator UID”; it now names the separate audit caller UID and keeps operator signing authority out of that harness path. Configuration/tool manager labels likewise no longer imply a shared operator/caller UID.
- Initial lifecycle status request returned exit code 1 because an estimate was missing. A Fibonacci estimate of 3 was recorded and the development transition then returned exit code 0. No external blocker remains.
- The generic logbook checklist is inapplicable because the current brief expressly says never to write LOGBOOK.md. Its evidence obligation is satisfied by this outcome and board notes without editing the logbook.
