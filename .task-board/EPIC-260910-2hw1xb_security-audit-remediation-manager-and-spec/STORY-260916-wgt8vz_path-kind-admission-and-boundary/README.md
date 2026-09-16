# STORY-260916-wgt8vz: path-kind-admission-and-boundary

## Description
Finding E6 (Medium): path-kind sources (overlays per Decision 0012 §5, onboarding imports per environments §9.6) are pinned by state hash and carry no canonical source identity, so the MCP package allowlist and any future signer allowlist (E1) cannot name them, and the missing store boundary (S5) applies to the directory itself. Whether a path-kind package may carry MCP declarations or class: system modules is not stated.

Implementation verification (TASK-260916-dv7xv5 rev2, curator main 80483355, launcher main b34e1e27, static): partially confirmed. MCP half NOT APPLICABLE: dependency declarations are git-only (contextpkg.go:289-305), MCP members load only from lock members (managed.go:163-170), and the MCP allowlist applies at resolution to every MCP dependency regardless of the root kind (contextresolve.go:477-483). System-module half confirmed: class parsing is kind-agnostic (contextpkg.go:263). Directory boundary confirmed within the path ingestion flow (envprofile.go:609, state-hash pin :928-952; marker records no source identity for non-git kinds, switch.go:760-777, path provenance :793-797). Drop the MCP half from this story.

## Scope
curator-spec decisions/0012 §5, environments §2.2/§6/§9.6; curator contextpkg path sources

## Acceptance Criteria
The spec states which source kinds may carry MCP declarations and class: system modules (git only unless a reviewed exception); path overlays must pass the same ownership/permission/containment validation S5 specifies for the store; diagnostics and vectors cover a path-kind package that declares MCP or a system module
