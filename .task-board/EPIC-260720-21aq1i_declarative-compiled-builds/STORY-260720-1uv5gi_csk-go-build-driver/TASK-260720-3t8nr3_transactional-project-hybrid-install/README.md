# Integrate atomic project and hybrid builds

## Description
Integrate private go-v1 builds and all existing materialization surfaces into one project-scoped plan followed by one manager-home-isolated commit, including hybrid targets and consumer-last ordering.

## Scope
Own the project and hybrid materialization path in src/csk/installer.py plus only the adapter, environment, consumer, and hybrid integration changes required to express transaction targets. Build every miss into operation-private staging outside the home lock after all gates pass. Under the home lock recover, revalidate shared generations and cache trust, publish verified entries, and journal every project, hybrid, runtime, shim, environment, adapter, stale-removal, marker, and consumer target. Reuse the build, cache, marker, shim, and transaction APIs from predecessor tasks.

## Acceptance Criteria
Active builds run provider-first and command-lexically only after audit and before any persistent mutation. A build failure leaves live cache, project, hybrid, runtime, adapters, shims, environments, markers, and consumers unchanged. Commit revalidates closure, ownership, cache winners, target preimages, and generations under the home lock; stale plans restart. New or changed installs write marker v2, exclude build_roots from context and runtime, point shims to immutable cache artifacts, preserve mixed script and system behavior, and update the consumer ledger last. Any publish or target failure rolls back committed targets in reverse order while the home lock remains held; an unreferenced immutable publication may remain only when protocol-safe. Two-project success and success-versus-rollback vectors pass. Focused project, closure, hybrid, adapter, dry-run, rollback, concurrency, and strict-mypy gates pass.
