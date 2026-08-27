# TASK-260825-2gyhq8: config-build-https-command

## Description
Add 'curator config build-https' with add, login, list and remove, alongside the existing build-ssh command. add takes a scope plus exactly one source; login reads a token from a hidden prompt (or stdin when not a terminal) and stores it through the operator helper, then selects it; list prints one line per scope sorted, and shows whether a stored token is actually present; remove drops the scope and its stored token. A token is never accepted as a command-line argument. Help text documents the precedence and the disclosure warning core 12.2 requires.

## Scope
(define task scope)

## Acceptance Criteria
add, login, list, remove behave with validation failures covered; no token reaches argv; help documents precedence and exposure; go test green.
