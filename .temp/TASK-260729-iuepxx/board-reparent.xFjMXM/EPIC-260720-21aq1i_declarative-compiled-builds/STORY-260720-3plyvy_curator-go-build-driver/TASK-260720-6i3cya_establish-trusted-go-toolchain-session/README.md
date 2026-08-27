# Establish a trusted Go toolchain session

## Description
Implement the package-independent half of go-v1: trusted toolchain selection, private telemetry-off initialization, native target discovery, fixed environment construction, release-family gating, and curator-go-toolchain-v1 fingerprinting.

## Scope
Create internal/godriver with a narrow injectable process executor and platform helpers. Curator selects an absolute Go executable from CURATOR_GO first, then GOROOT/bin/go, then runtime.GOROOT()/bin/go; it never uses exec.LookPath, user PATH, project shims, runtime roots, or repository files. Resolve links and require the candidate to be the regular executable under the fingerprinted GOROOT. From a manager-owned empty CWD and bootstrap environment run only go telemetry off, go version, and the fixed go env -json probe, then freeze the native GOOS, GOARCH, and one tuning variable. Build the clean operation-private environment and exact toolchain digest, enforce the rc.4 tested release-family allowlist, bound output and time, and delete private probe state on every exit. Do not inspect package source or run go list or go build here.

## Acceptance Criteria
Missing CURATOR_GO and unusable GOROOT fallbacks, wrappers, repository-local fake Go, candidates outside derived GOROOT, pre-1.23 or non-allowlisted future families, telemetry failure or non-private telemetry paths, target mismatch, invalid version output, escaping toolchain links, special or duplicate paths, and toolchain mutation fail with stable diagnostics; a valid session executes the three fixed argv forms once with closed stdin and no inherited GOFLAGS, GOENV, GOWORK, PATH, compiler, proxy, auth, or locale state; the exact toolchain vector digest and LF or CRLF normalization pass; dry-run creates only an operation-private probe that is removed and never persists or refreshes a memo.
