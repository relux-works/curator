# Add cross-platform Go build end-to-end tests against rc.6

## Description
Add a minimal real Go skill fixture and end-to-end tests proving CocoaSkills install, cache, activation, launch, repair, recovery, concurrency and rollback behavior for compiled commands against the accepted rc.6 candidate.

## Scope
Own CocoaSkills fixtures, process-level tests and CI changes. Use curator-spec commit 432eb2ee1fe2d6b271e37269f867c8851c325539 through an explicit caller-supplied candidate root and authenticate manifest sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071. Preserve the committed byte-equivalent suite pin 0c81c1f8d5321d822be2a2817b05aea03e656e15 and emit no release claim. Real go-v1 builds run natively on macOS and Windows with accepted Go family 1.25 and protocol floor 1.23. Ubuntu runs portable non-driver coverage and proves source-aware go-v1 fails closed; no Linux driver-success claim. Exercise project, global, hybrid, mixed script/build commands and platform shims. Use injected failures or fake toolchains only for negative paths.

## Acceptance Criteria
A vendored Go command builds without network under manager-worker-v1, launches only on explicit invocation, forwards exact arguments and exit status, and works through project/global shims on macOS and Windows. Evidence covers cache hits, relevant and irrelevant mutations, dry-run purity, build-two isolation, target-swap rollback, interrupted recovery, concurrent publisher identity and two-project preservation. Native macOS and Windows CI runs real builds across the supported Python matrix. Ubuntu runs portable coverage and asserts unavailable-control with no worker launch. Exact rc.6 root commit and manifest digest are recorded; no committed release pin, tag, GitHub Release or conformance claim changes. Full pytest, strict mypy, build, twine and diff checks pass.
