# BUG-261018-k2tn36: managed-codex-reads-user-agents-skills

## Description
A Codex session launched on a Curator managed home still discovers user skills from $HOME/.agents/skills, because the launch shifts CODEX_HOME but not HOME. Field report (another operator, rc.7, Codex 0.163.0): a child on --profile default lists 42 skills, 1 from the profile, 37 from $HOME/.agents/skills, 4 Codex system skills. Required: a profile launch receives only the skills its profile and project lock declare, plus the harness built-in system skills listed by name; the same check for Claude user-level skills and instructions. Proposed contract (pending owner agreement): profile-only context is a MUST, not an optional qualification.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
