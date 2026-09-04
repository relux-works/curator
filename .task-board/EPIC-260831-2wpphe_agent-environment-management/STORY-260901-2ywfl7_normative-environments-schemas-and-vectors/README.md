# STORY-260901-2ywfl7: normative-environments-schemas-and-vectors

## Description
JSON Schemas under schemas/v1/ (profilefile-v1, context-manifest-v1, agent-environment-marker-v1, launch-env-fragment-v1) plus positive/negative conformance vectors and byte-exact determinism vectors for header/chapters/monolithic/referenced forms, wired into the conformance manifest and CI validate tooling.

## Scope
(define story scope)

## Acceptance Criteria
Four schemas validate; vectors regenerate deterministically twice; validate.py green; determinism vectors cover header, chapters, both forms, zero-modules case.
