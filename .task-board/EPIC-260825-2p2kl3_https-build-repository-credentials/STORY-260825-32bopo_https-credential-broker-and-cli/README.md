# STORY-260825-32bopo: https-credential-broker-and-cli

## Description
The manager credential broker for one pinned HTTPS fetch, the operator-side credential reads that feed it, and the command surface that manages scopes.

## Scope
(define story scope)

## Acceptance Criteria
Broker answers only the two prompts Git asks and only for the pinned host, failing closed otherwise; credentials are read by the manager outside the fetch process graph through the operator credential machinery on every supported platform; a store that persists nothing is detected rather than trusted; config build-https add/login/list/remove behave and validate; go test green.
