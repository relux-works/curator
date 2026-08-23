# STORY-260822-2v23mv: scoped-build-ssh-config

## Description
build_ssh map in the global config: canonical-identity segment-prefix scopes select SSH credentials per external build repository. Longest matching scope wins; flags and CURATOR/CSK_BUILD_SSH_* environment values keep precedence; a package can never select credentials (spec core 12.2). Includes CLI management subcommand, install-time precheck with detected candidates, and docs.

## Scope
(define story scope)

## Acceptance Criteria
Config field parsed/serialized fail-closed; per-repository resolution wired into external installs; config build-ssh add/list/remove; precheck prompts on TTY and fails closed with ready-made commands otherwise; docs updated; go test green.
