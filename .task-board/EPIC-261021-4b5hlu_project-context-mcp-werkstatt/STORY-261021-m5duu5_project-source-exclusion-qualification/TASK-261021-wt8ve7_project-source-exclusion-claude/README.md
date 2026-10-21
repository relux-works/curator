# TASK-261021-wt8ve7: project-source-exclusion-claude

## Description
Qualify which project-native sources Claude Code 2.1.305 loads from a checkout and which controls exclude them, rows P1-P8, with approved-copy positive controls.

## Scope
curator: new .research/project-exclusion-claude harness and docs/qualification record; no product code.

## Acceptance Criteria
1. Rows P1-P8 measured with per-row verdict and exact control.
2. Positive control per row: an approved managed-home copy loads.
3. Exact digest-pinned binary, opt-in gate, offline sandbox, no real credentials.
4. A narrowing control fails its row.
5. Default suite green; gated run recorded.
