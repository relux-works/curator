# Review note — TASK-260927-4pv4au security_posture revision A (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 6bd98d49, tree 1390c2f5, 26 paths, gate green) against `4pv4au-brief.md`, `4pv4au-gatefix-1.md` and curator-spec rc.13:
profiles/manager.md §7.1 (security posture), environments.md (security_posture header row, passable_env_names_unbounded_refused, §12 knobs),
conformance/v1/vectors/security-posture.json (17 cases; profile_defaults, profile_refusals, diagnostics, status_gates, schema1_gates,
rollout_revisions). Verify through the production entry:
1. Revision A shipped (one switch): default permissive; `security_posture_permissive` warned once per operation on stderr only — never in
   `--json` stdout, script/launcher/worker protocol streams or a launched command's output; the legacy `status --json` shape unchanged.
2. Explicit hardened → hardened per-knob effective defaults; explicit knob beats profile default; system-locked value/posture beats explicit;
   hardened refusals (empty source allowlist explicit, empty MCP allowlist with declarations, passable_env_names null) and the warn-only case;
   schema-1 machine permissive; `env status` posture row; `env status --check` non-current on a hardened contradiction; hardened + absent
   passthrough follows s4-warn.
3. Vectors: 17 cases minus the unreachable-registry pair (owner TASK-260910-1sapuy, which is blocked_by this task — the effective-posture API
   it needs must exist and be usable from the install resolver) and the revision-B cases (owner TASK-260927-25hk87); gap rows attributed.
4. Mutants (default flipped to hardened; locked ignored; permissive warning dropped or duplicated; one refusal removed) killed, real exit
   codes. Manager-state reads via internal/stateread. No trunk revert, no CHANGELOG/LOGBOOK, no stray files (incl. binaries).
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
