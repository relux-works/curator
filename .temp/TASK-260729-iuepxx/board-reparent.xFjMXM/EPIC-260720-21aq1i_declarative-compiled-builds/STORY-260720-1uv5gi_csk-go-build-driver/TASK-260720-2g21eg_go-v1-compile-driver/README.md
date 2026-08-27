# Implement fixed go-v1 compile driver

## Description
Implement the source-aware go-v1 preflight and compile engine with manager-owned argv, environment, dependency-graph validation, and output verification, without cache or install concerns.

## Scope
Own src/csk/builds/go_v1.py, its injected process executor boundary, and focused unit and fixture tests. Starting from the frozen snapshot and trusted toolchain descriptor, build the exact clean environment and run the fixed go list and go build argv directly. Parse the complete go list JSON stream, enforce module and path containment for main and transitive dependencies, and validate the staged executable. Do not publish caches, write markers or shims, or launch the output.

## Acceptance Criteria
The only source-aware argv are the normative vendor-only go list and go build forms with trimpath, buildvcs disabled, compiler gc, pgo off, internal linking, libgcc none, and a manager-derived output path. The environment starts empty and fixes private Go caches, GOPROXY and VCS off, GOWORK off, GOTOOLCHAIN local, CGO disabled, GO_EXTLINK_ENABLED 0, native target and tuning, locale, temp roots, and empty executable PATH. Exactly one non-DepOnly main package is accepted. Missing or inconsistent vendor data, workspace or toolchain switching, graph paths outside build root or GOROOT, load errors, cgo or native fields, any SysoFiles, nonstandard SFiles, escaped embeds, and active nonstandard //go:cgo_import_dynamic fail before build. Output is one bounded regular executable in staging, is hashed and permissioned, and is never run. Focused pytest, a valid real Go fixture build, poisoned-environment negatives, and strict mypy pass.
