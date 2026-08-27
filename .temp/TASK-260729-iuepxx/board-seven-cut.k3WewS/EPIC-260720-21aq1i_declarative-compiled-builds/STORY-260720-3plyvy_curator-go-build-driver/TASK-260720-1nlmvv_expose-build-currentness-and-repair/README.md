# Expose build diagnostics and repair behavior

## Description
Surface compiled-command planning and currentness through Curator CLI install, upgrade, dry-run, status, status --json, status --check, and gc behavior, with clear missing-toolchain and repair guidance.

## Scope
Own cmd/curator command wiring, result structures, status drift classification, CLI tests, and only the install result fields needed for presentation. Report each active build command with driver, build_root, source_dir, build-source digest, native target, logical key, and cache outcome without leaking absolute private paths or unbounded compiler output. Map invalid marker, context exposure, untrusted cache, missing or corrupt receipt, artifact drift, wrong target or toolchain, unsupported driver or Go family, and concurrent state changes to stable machine-readable codes and concise human text. Existing install and upgrade act as repair by rebuilding invalid compiled state; do not add a new repair command unless the landed rc.4 CLI contract requires it.

## Acceptance Criteria
Dry-run miss says would-preflight-and-build and never claims a completed compiler check; untrusted protected state says would-rebuild-untrusted-cache; cache hits and every non-current class are distinct in JSON and human output; status --check exits nonzero for every invalid, unsupported, missing, corrupt, or unknown compiled state and zero only for exact current state; missing or incompatible Go diagnostics name the accepted selection mechanisms and tested family without suggesting PATH or automatic download; install and upgrade repair invalid cache and marker state only after gates and preserve the old install on failure; compiler diagnostics are bounded and redacted; legacy status output and CLI exit codes remain compatible.
