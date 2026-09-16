# STORY-260916-2otjbn: umbrella-provider-trust-roots

## Description
Finding E4 (High, composition with S6): environments §11 refuses a curator-<name> provider only when it resolves inside a manager-published or managed directory (subcommand_provider_untrusted). A directory that a project-controlled shell hook (S6, .agents/env.sh) prepended to PATH is neither, so cd project && curator run claude_code executes the project binary as the launcher with the operator environment. The launcher SPEC inherits the same trust of PATH.

Implementation verification (TASK-260916-dv7xv5 rev2, curator main 80483355, launcher main b34e1e27, static): confirmed. findProvider resolves exec.LookPath on the ambient PATH (umbrella.go:30-38); providerUntrustedDir refuses only manager-published directories (:43-63); cmdUmbrella executes the result (:81-91); no ownership or writability test. The change is local to these two functions.

## Scope
curator-spec environments §11; curator cmd/curator umbrella.go; curator-agent-launcher SPEC §2 and stderr line-group

## Acceptance Criteria
Providers resolve only from the manager install directory plus an explicit machine-config provider directory list, never from ambient PATH, or at minimum providers in directories writable by anyone other than the operator are refused; the resolved provider path is printed in the launch stderr line-group; conformance vectors cover the S6-injected PATH case
