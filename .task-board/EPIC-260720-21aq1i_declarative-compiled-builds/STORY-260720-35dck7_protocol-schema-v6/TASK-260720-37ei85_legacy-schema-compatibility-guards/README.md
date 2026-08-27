# Add legacy schema compatibility guards

## Description
Add explicit generator tests and regeneration checks that prevent schema 6 work from changing the structural meaning or generated evidence of schemas 1 through 5, install marker v1, conformance claim v1, fixed manager and system configuration surfaces, or registry and audit evidence boundaries. This closes the preservation requirement before behavioral build vectors are added.

## Scope
Work in curator-spec after all new metadata schemas and cases land. Own focused compatibility assertions in tools/generate-vectors/main_test.go, any non-product test fixture needed to compare frozen legacy semantics, and only the newly added rc.4 compatibility sentence in protocol/core.md needed to accurately describe schema-v1 extension behavior. Do not rewrite legacy schemas or duplicate new build behavior vectors.

## Acceptance Criteria
Tests assert that common command for schemas 1 through 5 still contains only script and system; both schema-v1 manifests preserve deployed additionalProperties extension behavior and assign no build semantics to an incidental build_roots field; each agent-skill and csk-skill schema 2 through 5 rejects build_roots; every schema 1 through 5 rejects type build; install-marker-v1 remains historical shape and claim-v1 remains schema 1 plus protocol rc.3; manager-config-v1 and system-config-v1 gain no driver, argv, environment, toolchain, output-path, hook, or build-policy override surface; registry and audit-record schemas gain no local artifact attestation or receipt-provenance field; regeneration retains legacy schema-case names and expected validity while adding new cases; origin/main comparison distinguishes intentional rc.4 inventory/hash changes from frozen wire semantics; go test ./tools/generate-vectors, make regenerate, make validate, and make regenerate-check pass twice without a diff.
