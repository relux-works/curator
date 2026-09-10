# TASK-260910-28kmef: service-recursionerror-400

## Description
curator-skill-registry: map RecursionError from load_json/canonicalization to ProtocolError so pathological deep JSON returns 400 invalid_json, not 500.

## Scope
(define task scope)

## Acceptance Criteria
Deep-nesting request test asserts 400
