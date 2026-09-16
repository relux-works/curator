# spec: rename claude_code→claude and codex_cli→codex with deprecated aliases

## Description
curator-spec amendment: canonical environment ids claude and codex; claude_code and codex_cli accepted as deprecated aliases for one release in every surface that names an environment (manifests targets/forms, machine config knobs, launch fragments, CLI, defaults); marker migration rule for provisioned managed homes; decide the frozen-v1 launch-env-fragment schema path (additive enum + alias normalization rule vs versioned schema) and record it; update environments.md, manager.md, schemas, conformance vectors, docs, CHANGELOG, COMPATIBILITY.

## Scope
(define task scope)

## Acceptance Criteria
Text and schema/vectors consistent; make validate green; explicit compatibility statement for v1 consumers; deprecation timeline stated.
