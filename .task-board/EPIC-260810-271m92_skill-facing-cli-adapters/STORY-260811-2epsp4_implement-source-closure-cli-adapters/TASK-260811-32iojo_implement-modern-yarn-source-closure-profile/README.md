# TASK-260811-32iojo: implement-modern-yarn-source-closure-profile

## Description
Implement the modern Yarn lock, cache, and extension-restriction profile for the Node/TypeScript adapter. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Node/TypeScript; Source closure invariant items 1-4 and 6; Delivery completion. Accepted input: TASK-260810-2n3sbi modern Yarn profile.

## Scope
Parse a supported yarn.lock plus manifests and .yarnrc.yml; bind exact Yarn release, built-in plugin set, patches, cache key and compression, linker, conditions, and checksums; admit raw or normalized package archives; create a private immutable cache; run with network disabled, immutable lock and cache, and skip-build; regenerate PnP or install state as output.

## Acceptance Criteria
Supported modern Yarn graphs materialize and invoke offline with deterministic cache and linker identity; local or downloaded plugins, custom resolvers or fetchers, Git pack behavior, undeclared patches, preseeded PnP or install state, lifecycle execution, native payloads, and ambient-cache input fail closed; shared S01-S08 and modern Yarn N01-N13 variants pass.
