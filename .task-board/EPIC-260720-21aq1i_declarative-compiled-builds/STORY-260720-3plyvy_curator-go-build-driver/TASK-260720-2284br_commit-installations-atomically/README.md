# Commit installations atomically across scopes

## Description
Integrate the staged install plan, protected cache, markers, shims, and target journal into one manager-home-isolated commit for project, global, and hybrid operations with the consumer ledger committed last.

## Scope
Refactor internal/install Project and Global plus narrowly required adapters, envfiles, globalbins, scopes, and cleanup APIs to produce staged targets rather than mutating piecemeal. Hold the canonical project operation lock from planning through handoff. After all private builds succeed, acquire the home lock, recover old journals, revalidate cache protection and every consulted generation or target preimage, and restart the earliest affected planning step when state changed. Atomically publish verified cache winners and commit deterministic target classes: project, global, or hybrid context and markers; script runtime, canonical and forwarding shims, env files; adapter and mirror ledgers; stale managed removals; consumer ledger last. Keep the home lock through reverse rollback and post-commit maintenance handoff.

## Acceptance Criteria
No shared target changes before the home lock and revalidation; a stale closure, activation, cache trust, target owner, preimage, or required key restarts rather than applying an old plan; injected failure at cache publication or every target class restores the prior project, global, hybrid, runtime, shim, env, adapter, mirror, and consumer state in reverse order while preserving pre-existing immutable cache entries; consumer state is absent after a failed first install and updates last after success; two concurrent project successes preserve both consumers and one project rollback cannot restore over another projects committed shared targets; recovery completes before new mutation; GC failure after commit is a warning and does not roll back the installation; existing install behavior and new race tests pass.
