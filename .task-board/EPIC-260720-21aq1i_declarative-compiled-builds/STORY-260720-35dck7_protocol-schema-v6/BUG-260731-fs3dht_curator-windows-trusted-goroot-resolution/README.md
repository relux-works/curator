# BUG-260731-fs3dht: curator-windows-trusted-goroot-resolution

## Description
JUSTIFIED GAP from native Windows workflow-dispatch run 30623699047 artifact test-evidence-windows-latest: godriver.Probe rejects the actions/setup-go Windows GOROOT with go_toolchain_missing trusted GOROOT is not a real directory. BUG-260731-33v6zz proved its owned buildcache/globalbins/managerlock/staging packages are now failure-free, but seven existing cmd/curator compiled-command cases plus the new host-toolchain diagnostic remain blocked on this shared product boundary. Existing BUG-260731-lepevi is done and scoped to Linux; BUG-260731-33v6zz explicitly excludes internal/godriver. Fix the real Windows trusted-root resolution without searching PATH, downloading a toolchain, weakening provenance checks, or platform-skipping the compiled cases.

## Scope
Curator internal/godriver trusted GOROOT and launcher resolution on native Windows, with cmd/curator host-toolchain integration evidence. Preserve the fail-closed trust model and macOS/Linux behavior.

## Acceptance Criteria
On windows-latest with actions/setup-go, godriver.Probe(ConfigFromEnvironment(...)) selects the trusted host Go installation and reports the correct windows/amd64 target; the affected cmd/curator compiled status, repair, GC and dry-run cases no longer fail with go_toolchain_missing; no PATH search/download or trust relaxation; macOS/Linux tests and lint remain green.
