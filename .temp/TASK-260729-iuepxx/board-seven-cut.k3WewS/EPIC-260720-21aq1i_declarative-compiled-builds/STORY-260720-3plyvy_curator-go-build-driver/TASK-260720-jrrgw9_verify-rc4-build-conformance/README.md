# Verify rc.4 build-driver conformance end to end

## Description
Consume the authoritative protocol rc.4 build-drivers and manager-lifecycle vectors and add Curator unit, integration, concurrency, and executable-launch coverage without creating a private copy of portable expected values.

## Scope
Extend internal/interop and package conformance consumers to read CURATOR_CONFORMANCE_ROOT. Add Curator-only temporary Git skill fixtures and mock or fake executors for fault injection, plus bounded real Go integration where the CI toolchain is allowlisted. Cover project, global, hybrid, dependency-closure, cache-hit, repair, rollback, recovery, GC, and shim launch behavior. Preserve the existing script golden fixture and registry hashes. The task may make only test-focused fixes to implementation seams; route substantive defects to their owning implementation task.

## Acceptance Criteria
Every positive build-driver identity, environment, process, cache, context, marker, and lifecycle vector has an executable assertion; every minimum rejection cluster from the rc.4 suite maps to a stable Curator error without package code execution; tests prove corrupt and stale artifacts rebuild, a self-consistent untrusted cache is not adopted, dry-run mutates nothing, build two failure preserves the previous install, cache hits avoid go list and go build, and installed binaries receive forwarded arguments and return exact exit status; concurrent project success and rollback vectors pass under go test -race; schemas 1 through 5 and the existing golden fixture remain unchanged; go test ./... and go test -race ./... pass with the authoritative suite.
