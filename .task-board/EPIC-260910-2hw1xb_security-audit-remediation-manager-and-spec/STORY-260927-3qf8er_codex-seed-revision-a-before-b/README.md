# STORY-260927-3qf8er: codex-seed-revision-a-before-b

## Description
rc.13 environments §7.4: a manager MUST ship Codex seed revision A (warning release) before revision B (flip). TASK-260916-33abdk implements B with registry CodexSeedRevision=B; curator has never released A. Ship A in the next curator release, flip to B in the release after it.

## Scope
(define story scope)

## Acceptance Criteria
A curator release ships CodexSeedRevision=A (warning, whole copy, codex_seed_record revision A, env status ungoverned list; A vectors driven, B bounds attributed); the following release flips to B with the B vectors driven.
