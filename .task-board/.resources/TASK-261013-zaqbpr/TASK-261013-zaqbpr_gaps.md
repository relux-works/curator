# TASK-261013-zaqbpr — apply-wave5-transfer-untrusted-launcher

Documentation handoff for W-LR6, owner decisions of 2026-10-13.
Changes remain uncommitted in the assigned Story worktree. No runtime
implementation or platform qualification is claimed.

## Brief coverage

| W-LR6 item | Applied text and figures |
|---|---|
| 1 — R-TRANSFER plan and workspace contract | SPEC §§10.1–10.2, 10.4–10.5; README target mode; all four target figures. Base B, manifest digest, write scope and exec-profile reference are digest-covered; exact fields remain unpublished with intake. Artifact layout now names the dispatcher's workspace contract. |
| 2 — R-STANDING surface checks | SPEC §10.2; README; all four target figures. Settings/hooks/MCP/instructions checked by digest before exec, or root-owned outside home; launch/restart/native-resume coverage where admitted; drift refuses and repair is explicit. |
| 3 — R-EXEC wrappers and enforcement owner | SPEC §10.2; README; all four target figures. D-PATH protected manager dispatcher holds fixed-subcommand light-check wrappers, no exposed heavy entry points; OS allowlist owned by user manager. |
| 4 — R-NET-TOOL proxy, resources and search | SPEC §§10.1–10.2, 10.5–10.6; README; all four target figures. Bound provider proxy only; read-only task resources covered by plan digest; provider-side search switched off without per-identity operator opt-in. The plan is the only control for search inside provider API calls. Qualified provider search/platform Unix-socket web service are trusted-circle opt-ins only. |
| 5 — figures | artifact-layout.puml, components.puml, launch-paths.puml, launch-states.puml and diagrams/README.md updated. Sequence/state figures show materialization before work, confirmed account-wide absence before export, recorded export or explicit operator waiver before home deletion. |

## Shared names and changed sections/figures

All figures below refer to the target panels only; current interactive APIs
remain scoped to operator-machine behavior. Exact unsupported platform
profiles are expressly denied in the target text.

| Shared name | Changed sections | Changed figures |
|---|---|---|
| R-TRANSFER | SPEC §§10.1–10.2 (full transfer path, ownership, deterministic import, exact-C review, rework and signed integration), 10.4 (export cleanup obligation), 10.5 (remaining formats); README target mode | All four figures; artifact/component ownership and sequence/state materialize/export ordering; diagrams index |
| R-STANDING | SPEC §§10.2, 10.5, 10.6; README target mode | All four launch-obligation notes; sequence/state drift refusal |
| R-EXEC | SPEC §§10.1–10.2, 10.5–10.6; README target mode | All four launch-obligation notes |
| R-NET-TOOL | SPEC §§10.1–10.2, 10.5–10.6; README target mode | All four launch-obligation notes; sequence plan/surface check, artifact digest coverage, component broker/proxy boundary |
| R-UNTRUSTED | SPEC §10.6 admission dependencies; README target mode | All four target admission notes, explicit Slice 0/user-manager v0 refusal |
| R-VERIFY | SPEC §§10.2, 10.6; README target mode | All four launch-obligation notes; sequence exact-C review runs code only in controlled verification |
| R-CEILINGS | SPEC §10.6 host matrix, Linux selected unit/kernel baseline, macOS qualification gaps and open soft-row decision; README | All four target admission notes reference owner section |
| R-WATCH | SPEC §10.6 per-UID least-privilege telemetry, per-account journaled narrowing, phone alerts; README | All four target admission notes reference owner section |
| R-ALLSTOP | SPEC §10.6 manual operator narrowing, range deny, journaled stop/confirmed absence, lease freeze, host snapshots and explicit recovery; README | All four target admission notes reference owner section |
| R-RANGE-DENY | SPEC §10.6 static range rule, v0 listener-only/v1 narrowing, all-protocol coverage and active precedence probes; README | All four target admission notes reference owner section |
| R-SOCKETS | SPEC §10.6 declared-set test-UID socket qualification; README | All four target admission notes reference owner section |
| R-SHARED | SPEC §10.6 no shared services on untrusted-admitting hosts; README | All four target admission notes reference owner section |

Earlier names already satisfied and retained: R-FD3 in SPEC §10.1;
R-START, R-SESSION and D-PATH in §10.2; R-FATE and R-TIME in §10.4.
Earlier operator custody is referenced as R-OPERATOR in §10.6 using
swarma-operator, swarma-observer, operator device key and operator command.
No alternate custody path or new signature format is defined.

## Superseded text and public hygiene

Added the dated revision note at the top in the existing style. Removed the
open transfer-owner claim, inherited/fresh hooks claim, general broker-network
wording and metadata-only description of all exports. Retirement metadata
remains a separate restricted export; it does not substitute for workspace
results. Existing interactive direct/cooperative profiles stay confined to
operator machines and cannot bind on a platform host.

Only SPEC.md, README.md and diagrams sources/index changed. Added text has no
first names, host names, private absolute paths, session links or board-resource
links. No new wire fields, refusal codes, numeric exits or cap values are
assigned. Shared design labels remain unchanged and marked unpublished.

## What stays open

- SPEC §10.1: exact intake envelope/schema, digest preimage/placement, framing,
  read deadline, fixed entry argv and per-platform FD3 mechanism/qualification.
- SPEC §10.5: canonical base archive; exact workspace, exec-profile and resource
  plan fields; resource/export caps; workspace-result schema/retention;
  deterministic import command; candidate trailer grammar; optional shared-object
  provenance, authorised-object access and retention.
- SPEC §§10.2/10.5: harness-surface digest coverage and drift refusal code;
  qualified protected wrappers and OS exec profiles; opt-in records, provider
  search configuration and qualified tuples; donor/platform web-tool formats.
- SPEC §10.6: qualified host matrix/class minimums/probes, watcher thresholds,
  sampler/alert interfaces, allstop act/record formats; soft-row acceptability is
  an open owner decision per host. No OS-wide exclusion is imposed.
- SPEC §10.6: controlled runner pool/job contract/results format belongs to
  board/infrastructure lanes and remains unpublished.
- SPEC §§10.4–10.5: existing consumed-plan/exit transport, ownership across exec,
  numeric exits/signals, helper signatures and host admission publication stay
  open. Donor provisioning variant B stays disabled pending its separate spec.

## Inspection evidence and limits

- Standalone `git diff --check`: exit 0 after edits (also exit 0 on prior passes).
- Text-only inspection via Python: exit 0; changed-path scope is seven doc files;
  every figure has one start/end delimiter and all four brief shared names.
  Added-text literal scans for private-home paths, local-host markers, session/
  app/resource links and LOGBOOK.md returned zero. This is bounded text
  inspection, not a renderer, formal policy proof or behavioral test.
- Tests/builds were not run by the producer: explicitly forbidden by W-LR6.
  PlantUML was not rendered: the brief states this host has no renderer.
  Configured CR validation is sanctioned at handoff; its actual results belong
  to the handoff evidence, and are not pre-claimed here.
- Initial requested development transition exited 1 because an estimate was
  missing. The estimate mutation exited 0; the retried transition exited 0.
- No prior test evidence is claimed or reused as proof for this documentation.

Checklist scope: the generic code/tests deliverable is satisfied as docs-only
per W-LR6. LOGBOOK.md is expressly prohibited; findings and remaining questions
are persisted in this outcome and board notes, with no logbook edit.
