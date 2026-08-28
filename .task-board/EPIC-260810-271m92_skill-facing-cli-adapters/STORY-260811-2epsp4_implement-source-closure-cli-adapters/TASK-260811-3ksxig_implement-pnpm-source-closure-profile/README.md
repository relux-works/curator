# TASK-260811-3ksxig: implement-pnpm-source-closure-profile

## Description
Implement the pnpm lock and materialization profile for the Node/TypeScript adapter. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Node/TypeScript; Source closure invariant items 1-4 and 6; Delivery completion. Accepted input: TASK-260810-2n3sbi pnpm package-manager profile.

## Scope
Parse a pinned pnpm-lock.yaml schema; bind importers, workspaces, snapshots, integrity, peer contexts, overrides, patches, target settings, and configuration; capture raw packages and contained local roots; derive a private read-only store; materialize frozen, offline, and scripts-disabled; reject pnpmfile hooks, custom resolvers or fetchers, undeclared patches, and side-effects cache.

## Acceptance Criteria
Supported pnpm graphs materialize and invoke offline from the admitted set with exact peer and target contexts; local dependencies are captured independently; missing store inputs, lock or metadata drift, extensions, side effects, lifecycle execution, native payloads, and ambient-store fallback fail closed; shared S01-S08 and pnpm N01-N13 variants pass.
