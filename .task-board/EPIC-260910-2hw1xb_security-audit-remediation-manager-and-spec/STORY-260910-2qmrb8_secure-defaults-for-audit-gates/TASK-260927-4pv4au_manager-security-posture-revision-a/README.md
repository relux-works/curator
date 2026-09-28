# TASK-260927-4pv4au: manager-security-posture-revision-a

## Description
Implement the rc.13 security_posture model (profiles/manager.md §7.1, environments.md, vectors/security-posture.json): knob admitted with default permissive (revision A shipped first), effective per-knob defaults under hardened, explicit knob beats profile default, locked values/posture beat explicit, hardened refusals (empty source allowlist, empty MCP allowlist with declarations, passable_env_names null), permissive warns security_posture_permissive once per operation, schema-1 machine is permissive, env status posture row and --check non-current on hardened contradictions. Unreachable-registry cases stay with TASK-260910-1sapuy.

## Scope
(define task scope)

## Acceptance Criteria
security-posture vectors driven except the unreachable-registry pair (owned by 1sapuy) and the revision-B flip cases (bounded with an owner); Story-owned posture gap rows removed; mutants killed; hosted gate green
