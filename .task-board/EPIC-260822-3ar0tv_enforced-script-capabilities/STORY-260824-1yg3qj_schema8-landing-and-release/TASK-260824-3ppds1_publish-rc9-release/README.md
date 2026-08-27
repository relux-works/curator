# TASK-260824-3ppds1: publish-rc9-release

## Description
curator-spec: publish 1.0.0-rc.9 per GOVERNANCE.md and RELEASE.md from the landed main: version metadata and vector manifest already in tree from the landing; regenerate twice and prove the clean second run; create the annotated SIGNED tag (git signing is configured on this machine — verify with git config user.signingkey and a test signature first; the release workflow verifies against maintainers.allowed_signers); push the tag; confirm the release workflow packages schemas and the conformance archive with checksums; verify the GitHub release artifacts exist. Executor: claude only.

## Scope
(define task scope)

## Acceptance Criteria
Tag 1.0.0-rc.9 (repo tag convention verified against existing tags) signed and pushed; release workflow green; release artifacts published with sha256 checksums.
