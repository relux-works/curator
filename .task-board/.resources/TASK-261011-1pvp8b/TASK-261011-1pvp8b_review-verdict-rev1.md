# TASK-261011-1pvp8b — Reviewer B verdict, CR rev1 (deciding review)

Verdict: ACCEPTED. No P0/P1 remains. Two P2 notes (tb-R226) follow.

CR-TASK-261011-1pvp8b-1 rev 1, base 1d60804709d57d81710a5381e664125307d0d31e, candidate tree 252c43422e9f03557f701a27086ee0c82e98d3b2. Docs only; 7 of 7 changed paths are documentation (README, SPEC, four .puml, figure index). No code, tests, builds or LOGBOOK edits. No first names, host-local paths or session links in added lines. No tests or builds run on this host; PlantUML rendering is unverified (no renderer).

## Independent review (done before reading A)

1. Figures exist and show phase, process identities, state owners and refusal/recovery branches: 4 of 4 audit filenames present (components, launch-paths -> M-LR1; launch-states, artifact-layout -> M-LR2). Current interactive API, hosted Phase 1 and the target one-shot path (curator-run exec, dispatcher v0, root launcher, agent-UID executor, broker, helper exec.stop) are drawn in separate panels. Checked against SPEC sections 2, 4.1, 4.4b, 4.5, 4.6, 4.7, 4.8, 4.9, 6 and 10.1-10.5: ordering (config read before argv, resolve --repair, plan build, provider-limit gate, network admission, late checks), ax no-fallback (ax_handoff_failed), hosted FD3/FD4 classification (protocol_error 6 / unavailable 1 / receiver exit unchanged), receipt-before-UID-drop, FD3 plan only, 48/64/80/96/128 KiB bounds, D-PATH order, Claude setup-token no refresh, 8 s renewal, broker sole refresh owner, exec.stop with TERM then KILL and absence-before-release all match.
2. Owner decisions applied as stated in SPEC 10 (D-EXECUTOR, D-PATH, D-WIRE-NAMES, D-GOALS, D-IDENTITY, D-DONOR-PROVISIONING, D-RESUME). No invented numeric exits, wire formats, receipt transport or activation; SH2 ticket and donor B shown as open/disabled.
3. Spec gaps are visible notes ("unspecified in spec ..."): intake envelope/digest, entry invocation, receipt transport, numeric exits and signal encoding, service UIDs, hosted durable layout, native snapshot removal.
4. Links: 4 of 4 figures linked from README and SPEC 1.1 plus section-level links in 4, 4.6, 4.7, 4.9, 10; every figure links back to defining SPEC sections; anchors resolve against the headings of the candidate SPEC/README.
5. Outcome TASK-261011-1pvp8b_figures.md maps all four filenames to spec sections and audit rows and states the verification limits truthfully (no gate pass claimed).

## Reconciliation with reviewer A

- A-F1 (launch-states.puml:75, "optional" revocation): P2. SPEC 10.4 says D-IDENTITY "permits" revocation for one-shot-registered identities only; "optional" is a defensible reading of that wording but adds a qualifier the spec does not state. Note, not a blocker.
- A-F2 (launch-states.puml:89, recheck / renewal / auth failure routed to Target refusal): P2. SPEC 10.5 lists renew_timeout/renew_failed among the typed refusal names, so the edge is not contradicted, but a post-exec runtime failure is not a pre-exec refusal and the node label can blur that distinction. The panel already marks the outcome-evidence transport as unpublished. Note, not a blocker.
- A's other sweeps (launch-paths, components, artifact-layout, links, index, hygiene) agree with my checks; no P0/P1 raised.

## P2 notes for follow-up (tb-R226)

1. Drop "optional" from the Retired node: state "separate registered one-shot identity revocation after retirement; standing identities operator-only" without the qualifier.
2. Split Harness -> TargetRefusal into a pre-exec binding-recheck refusal and a runtime renewal/auth failure routed via execution outcome evidence, keeping the transport labelled unpublished.
3. Minor inferred arrows in launch-paths (agent -> dispatcher refusal evidence, root -> dispatcher typed refusal, Outcome -> Stop) are labelled transport-unpublished; consider adding "recipient unspecified in spec 10.4" wording.
