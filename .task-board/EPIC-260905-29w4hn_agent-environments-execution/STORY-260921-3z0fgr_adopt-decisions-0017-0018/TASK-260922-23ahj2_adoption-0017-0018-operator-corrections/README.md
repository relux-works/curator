# TASK-260922-23ahj2: adoption-0017-0018-operator-corrections

## Description
Operator corrections (2026-09-21) to the landed adoption of decisions 0017/0018 (PR curator-spec#77): C1 the provider flag spelling for yolo (--dangerously-skip-permissions / --dangerously-bypass-approvals-and-sandbox) is owned by agents-management as a LaunchRequest member for LaunchModeInteractive with goldens, the launcher only resolves and passes the mode (launcher SPEC section 1, Decision 0013 D5) — fix every attribution to the launcher SPEC and add follow-up F-M1 (skill-agents-management), narrow F-L1; C2 Pi root evidence: on the operator Mac the managed home links to ~/.pi/auth.json which does not exist while the real credential is ~/.pi/agent/auth.json, the link is dangling and env status is silent — record in 0017/environments 7.4 and add to F-C1 AC that env status/resolve report dangling or mis-targeted credential links; C3 absent cli_auth_credentials_store means effective file storage so isolated is admitted for codex_cli — state explicitly in 7.4, manager 12.4 and 0017 choice 4. Text/AC only; no schema, vector or generator change. Brief: 3qcjsy-rework-2.md.

## Scope
curator-spec decisions/0017, decisions/0018, protocol/environments.md 7.4/10.1/10.2, profiles/manager.md 12.4, cli/curator.md, follow-up table

## Acceptance Criteria
grep for launcher-owned flag spelling returns only genuine launcher ownership (mode resolution/provenance/record key); F-M1 added with a one-line AC and F-L1 narrowed; Pi dangling-link observation and the absent-key-means-file rule present in the named sections; make validate green
