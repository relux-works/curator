# TASK-260822-3pkc80: cli-config-build-ssh

## Description
Subcommand 'curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]' plus list/remove (cmd/curator). add validates via the config grammar and replaces an existing scope; remove of a missing scope is an error; list prints one line per scope sorted.

## Scope
(define task scope)

## Acceptance Criteria
CLI tests for add/list/remove incl. validation failures; help text documents precedence (flags > env > config scopes); go test green.
