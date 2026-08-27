# Implement trusted Go toolchain identity

## Description
Resolve an operator-trusted native Go executable before entering package-controlled directories, initialize private telemetry-off state, freeze the native target, and compute curator-go-toolchain-v1 exactly.

## Scope
Own src/csk/builds/toolchain.py and its focused tests. Capture the operator process search path before project shim augmentation, reject repository or project-managed Go candidates, resolve the real GOROOT bin/go executable, run only the three package-independent probe forms from an empty bootstrap environment, validate the accepted Go release-family allowlist and native tuning tuple, and fingerprint the complete GOROOT tree. Add no package or manager-config build-policy fields and run no go list or go build.

## Acceptance Criteria
The probe invokes direct argv for go telemetry off, go version, and the fixed go env -json field list, verifies GOTELEMETRY off and GOTELEMETRYDIR inside the operation-private root, and deletes that root on every exit. GOROOT, GOHOSTOS, GOHOSTARCH, GOOS, GOARCH, and exactly one applicable tuning variable are validated and frozen. The framed tree digest handles files, directories, internal relative links, normalized LF or CRLF version output, duplicate and invalid paths, escaping links, executable checks, and tree mutation exactly as shared vectors require. Pre-Go-1.23, unknown release families, wrappers, repository-local Go, and mismatched GOROOT fail closed. Focused pytest and strict mypy pass on POSIX and Windows-safe imports.
