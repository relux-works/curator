# TASK-260927-31gaka: ship-codex-seed-revision-a

## Description
Make the shipped Codex seed revision A (warning release) on trunk after TASK-260916-33abdk lands: whole-copy seed, mcp_native_servers_ungoverned warning with the migration hint, codex_seed_record revision A, env status ungoverned list. Keep the B implementation behind the registry constant for the flip leaf.

## Scope
(define task scope)

## Acceptance Criteria
rc.13 A-revision provisioning/posture vectors driven; B-shipped cases bound with an owner (the flip leaf); mutants killed; hosted gate green
