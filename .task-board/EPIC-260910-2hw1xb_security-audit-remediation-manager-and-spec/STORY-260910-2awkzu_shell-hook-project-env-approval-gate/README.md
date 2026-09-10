# STORY-260910-2awkzu: shell-hook-project-env-approval-gate

## Description
Finding S6 (High): the cached shell hook auto-sources any .agents/env.sh found walking up from PWD with no approval or digest check, so a hostile project checkout achieves code execution on cd in hooked shells. Introduce a direnv-style trust boundary.

## Scope
curator-spec manager profile + internal/shell

## Acceptance Criteria
Hook sources only env files recorded/digested by the manager or explicitly allowed by the operator; unknown file warns and is not sourced; tests cover hostile-project case
