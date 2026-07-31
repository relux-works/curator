# STORY-260728-1ojb1p: curator-external-build-repositories

## Description
Implement schema-7 external build repositories in the Go Curator manager after the rc.5 protocol contract and local go-v1 foundations are accepted. Resolve exact Git source, admit a complete raw snapshot, audit it independently, compile with go-repository-v1, publish protected caches and mixed receipts/markers, install manager-derived shims into PATH, and preserve transactional rollback.

## Scope
Go Curator repository models, Git transport and local-substitution boundary, raw object reader, snapshot audit/cache, mixed receipt-v2 and marker-v3 planning, lifecycle, docs, and macOS/Windows native validation. No cocoaskills implementation and no generic future-language executor.

## Acceptance Criteria
Curator accepts only schema-7 closed inputs; exact tagged and untagged locks resolve with the accepted failure classes; no ambient Git config, helper, hook, filter, submodule, LFS hydration, alternate, replace, promisor, or package signing behavior can influence the build; audit precedes artifact cache lookup and compiler execution on real, cache-hit, dry-run, repair, and coverage-audit paths; installed command name/path is manager-derived and structurally verified; rollback and protected-cache currentness tests pass on macOS and Windows; schema-6 go-v1 behavior remains unchanged.
