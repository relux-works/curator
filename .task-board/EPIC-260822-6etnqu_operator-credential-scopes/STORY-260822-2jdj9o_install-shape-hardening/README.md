# STORY-260822-2jdj9o: install-shape-hardening

## Description
Verify and, where confirmed, fix three installation shapes against the spec: (1) manager executable invoked through a package-manager symlink must canonicalize before identity verification (profiles/manager.md: resolve to a canonical regular installed file, reject substitution); (2) external repository static audit must not block on non-executable text below vendor/ of third-party modules (the fixed session runs only go list / go build); (3) a present-but-unreadable or newer-schema canonical manifest must fail loud and never fall back to agents/runtime.json.

## Scope
(define story scope)

## Acceptance Criteria
Each item has a reproducing or proving test; confirmed defects fixed; disproved items documented with evidence in task notes; go test green.
