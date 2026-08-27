# STORY-260825-37cz7x: https-credential-configuration

## Description
The build_https configuration surface and its per-repository resolution: canonical-identity scopes reusing the build_ssh grammar and longest-prefix match, mapping a scope to a token SOURCE and never to a secret, with the run-wide environment override bindable to a host as core 12.2 requires.

## Scope
(define story scope)

## Acceptance Criteria
Config parses and serializes fail-closed and rejects a literal secret; resolution picks run-wide (host-bound) then longest scope then anonymous; every selection reaches the fetch; go test green.
