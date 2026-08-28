# TASK-260811-1u42b9: implement-npm-source-closure-profile

## Description
Implement the npm lock and materialization profile for the Node/TypeScript adapter. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Node/TypeScript; Source closure invariant items 1-4 and 6; Delivery completion. Accepted input: TASK-260810-2n3sbi npm package-manager profile.

## Scope
Parse supported package-lock.json and npm-shrinkwrap.json versions; reconcile root and workspace manifests, selected and pruned dependency edges, resolved locators, integrity, and embedded package metadata; capture and admit raw tgz bytes; derive a private cache; materialize with npm ci, offline, frozen semantics, and scripts disabled; detect implicit binding.gyp and bundled dependencies.

## Acceptance Criteria
Supported npm graphs materialize and invoke offline from exact admitted tarballs with no resolver or ambient-cache input; stale locks, integrity or metadata drift, mutable locators, bundled dependencies, native payloads, implicit node-gyp, and undeclared lifecycle execution fail before build or publication; shared S01-S08 and npm N01-N13 variants pass.
