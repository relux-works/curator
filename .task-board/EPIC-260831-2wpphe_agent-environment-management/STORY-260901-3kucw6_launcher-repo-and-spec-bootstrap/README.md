# STORY-260901-3kucw6: launcher-repo-and-spec-bootstrap

## Description
Bootstrap the launcher (curator run) per Decision 0010: new implementation repository consuming agents-management as a Go module and Curator/ax as CLI contracts; specification per OQ1 layout (recommended: new launcher repo + sibling -spec repo at stabilization). Composes spawn/context/session planes; configured ax integration is always used; no session state of its own; applies system-prompt channels with warnings.

## Scope
(define story scope)

## Acceptance Criteria
Repository name and spec home decided by operator; skeleton repo with SPEC draft covering flags, agents-management consumption, fragment consumption, ax handoff, system-prompt warnings; curator-run discovery verified against the umbrella convention; skill-agents-management left unsplit.
