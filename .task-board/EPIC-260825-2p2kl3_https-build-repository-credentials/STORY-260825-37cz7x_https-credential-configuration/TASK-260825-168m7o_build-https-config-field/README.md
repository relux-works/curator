# TASK-260825-168m7o: build-https-config-field

## Description
Add the build_https field to the global config (internal/config, mirroring internal/config/buildssh.go): a map of canonical-identity scope to {token?, token_env?, username?}. Reuse the build_ssh scope grammar and longest-prefix matcher rather than duplicating them. 'token' accepts only the enumerated source names; a literal secret in that field must be rejected with a message saying secrets never live in the config. Exactly one of token or token_env per entry. Extend the manager key set; keep the field out of the lockable keys, consistent with build_ssh and the ratified rule.

## Scope
(define task scope)

## Acceptance Criteria
Parse and serialize roundtrip; literal-secret rejection tested; exactly-one-source rejection tested; grammar rejections and longest-prefix match tested including a segment-boundary case; go test ./internal/config green.
