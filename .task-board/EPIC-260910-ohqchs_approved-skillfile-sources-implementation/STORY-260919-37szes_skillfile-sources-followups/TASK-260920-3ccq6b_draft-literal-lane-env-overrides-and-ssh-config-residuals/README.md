# TASK-260920-3ccq6b: draft-literal-lane-env-overrides-and-ssh-config-residuals

## Description
Residuals from the BUG-260920-3ukdk4 review (N2): per-invocation environment overrides (GIT_SSH_COMMAND, GIT_SSH, GIT_PROXY_COMMAND, GIT_EXEC_PATH, GIT_ASKPASS) are still honoured on the draft literal-URL lane although repository-transport lists environment overrides among non-imports; the user's ~/.ssh/config (Host aliases, ProxyCommand) is still read by the real ssh for literal ssh:// declarations (contract §5 'remain NOT imported'). Corpus row v2-user-ssh-alias-ignored covers only the logical declaration. Decide and implement the contract-consistent isolation (whitelist of honoured env; ssh config isolation or explicit documented bound), with production-entry rows and mutants.

## Scope
internal/gitops environment construction for the draft literal lane; cmd/curator resolve/refresh; docs

## Acceptance Criteria
Only the contract-listed environment reaches git on the draft literal lane (SSH agent socket, askpass if kept by ruling); GIT_SSH_COMMAND/GIT_PROXY_COMMAND/GIT_EXEC_PATH cannot redirect or hijack the clone (production-entry rows, mutants killed); ssh config isolation implemented or the bound documented with the operator-facing consequence; legacy v1 lane unchanged.
