# TASK-260811-twq9ad: implement-yarn-classic-source-closure-profile

## Description
Implement the Yarn Classic 1.x lock and offline-mirror profile for the Node/TypeScript adapter. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Node/TypeScript; Source closure invariant items 1-4 and 6; Delivery completion. Accepted input: TASK-260810-2n3sbi Yarn Classic profile.

## Scope
Parse the root yarn.lock with root and workspace manifests, resolved URL, integrity, exact Yarn version, and hoisting or layout configuration; capture and admit source tgz files; build a task-private offline mirror with empty ordinary cache; materialize frozen, offline, and ignore-scripts without checksum update or dependency-subtree lock authority.

## Acceptance Criteria
Supported Yarn Classic graphs materialize and invoke offline from the captured mirror; root and workspace graph identity is deterministic; missing or mutable artifacts, stale locks, checksum changes, lifecycle scripts, native payloads, and ambient-cache input fail before publication; shared S01-S08 and Yarn Classic N01-N13 variants pass.
