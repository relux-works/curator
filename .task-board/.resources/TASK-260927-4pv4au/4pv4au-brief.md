# TASK-260927-4pv4au — security_posture model, revision A (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Spec: curator-spec v1.0.0-rc.13 (SPEC_PIN 23435129) profiles/manager.md §7.1
(security posture), environments.md (security_posture header row, passable_env_names_unbounded_refused, §12 knobs), and
conformance/v1/vectors/security-posture.json (17 cases: profile_defaults, profile_refusals, diagnostics, status_gates, schema1_gates,
rollout_revisions A→B). Ship REVISION A: knob admitted, default `permissive`, every operation under permissive warns
`security_posture_permissive` once; the manager reports the posture it runs in (env status row); explicit `hardened` applies the hardened
per-knob effective defaults; explicit knob beats profile default; a system-locked value/posture beats explicit; hardened refusals (empty
source allowlist explicit, empty MCP allowlist with declarations, passable_env_names null) and the non-refusing warning case; schema-1 machine
is permissive; `env status --check` non-current on a hardened contradiction; hardened with absent passthrough follows s4-warn.
The unreachable-registry pair (unreachable-registry-permissive-warns / -hardened-refuses) belongs to TASK-260910-1sapuy — expose the
effective posture to the install resolver via a small API it can use, but do not implement those two here. Revision-B cases
(revision-B-default-hardened-flip-install, posture-rows-flipped-revisions) are bounded with owner = the flip leaf
(flip-security-posture-default-hardened). One switch constant for the shipped revision (A).
Remove Story-owned posture gap rows that now pass; mutants (default flipped; locked ignored; permissive warning dropped; one refusal
removed) killed with real exit codes; new manager-state reads via internal/stateread. No CHANGELOG/LOGBOOK edit (entry in results).
Handoff runs the hosted gate — WAIT for it; hand off only green. Write only inside your Story worktree.
