# STORY-260916-wgt8vz: path-kind-admission-and-boundary

## Description
Finding E6 (Medium): path-kind sources (overlays per Decision 0012 §5, onboarding imports per environments §9.6) are pinned by state hash and carry no canonical source identity, so the MCP package allowlist and any future signer allowlist (E1) cannot name them, and the missing store boundary (S5) applies to the directory itself. Whether a path-kind package may carry MCP declarations or class: system modules is not stated.

## Scope
curator-spec decisions/0012 §5, environments §2.2/§6/§9.6; curator contextpkg path sources

## Acceptance Criteria
The spec states which source kinds may carry MCP declarations and class: system modules (git only unless a reviewed exception); path overlays must pass the same ownership/permission/containment validation S5 specifies for the store; diagnostics and vectors cover a path-kind package that declares MCP or a system module
