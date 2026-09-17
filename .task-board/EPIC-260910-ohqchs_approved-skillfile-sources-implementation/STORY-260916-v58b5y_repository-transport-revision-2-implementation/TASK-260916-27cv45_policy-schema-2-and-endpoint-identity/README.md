# TASK-260916-27cv45: policy-schema-2-and-endpoint-identity

## Description
Source-policy schema 2 loader/validator (additive superset of schema 1: port-bearing endpoint/pin URLs, mirror_of attestation, operator alias table), canonical identity rules of §5; refusal rows tested.

## Scope
(define task scope)

## Acceptance Criteria
Source-policy schema 2 (port-bearing endpoint and pin URLs, mirror_of attestation, operator host-alias table) loads as an additive superset of schema 1 through the production policy loader; schema 1 policies load byte-identically (golden); canonical host/path remains the only portable identity and every refusal row of repository-transport.md §5 has a negative test at the loader entry; narrow tests and the remote gate are green.
