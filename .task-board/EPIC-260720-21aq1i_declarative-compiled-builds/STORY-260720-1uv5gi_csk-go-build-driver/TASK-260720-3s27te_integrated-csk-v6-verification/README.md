# Verify integrated CocoaSkills schema v6 and go-v1 implementation

## Description
Perform final clean integration verification of the independently implemented CocoaSkills schema v6 and go-v1 behavior against the accepted rc.6 portable manager-worker-v1 contract, with reproducible evidence mapped to every delivery criterion and protocol vector cluster.

## Scope
Verify from landed dependency heads in a clean task worktree using caller-supplied curator-spec commit 432eb2ee1fe2d6b271e37269f867c8851c325539 and manifest sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071. This is verification, not redesign. Run repository-supported full tests, strict typing, packaging, exact-root conformance and hosted CI. Native real go-v1 verification covers macOS and Windows. Ubuntu covers portable non-driver behavior and fail-closed source-aware go-v1 only; no Linux success claim. Make only narrow integration corrections and preserve the committed byte-equivalent suite pin 0c81c1f8d5321d822be2a2817b05aea03e656e15.

## Acceptance Criteria
Full pytest against the exact authenticated rc.6 candidate, strict mypy, build, twine, diff check and all hosted CI jobs pass with no unexpected skips. Native macOS and Windows real Go fixture evidence and Ubuntu portable fail-closed evidence are recorded. A task report records exact CocoaSkills and curator-spec SHAs, manifest digest, commands, logs and a matrix mapping all story criteria and required rejection, lifecycle, cache, dry-run, rollback, status, launch, execution-policy, native-control and capability cases. No tag, GitHub Release, conformance claim or committed release-pin change is made.
