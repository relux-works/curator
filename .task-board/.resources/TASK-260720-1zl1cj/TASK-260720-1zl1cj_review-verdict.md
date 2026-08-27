# TASK-260720-1zl1cj review verdict

## Verdict

Changes requested; route to implementation rework.

## Blocking correctness finding

The first-use canonical manager-home identity is not stable when the configured home does not yet exist below a symlinked ancestor. In internal/managerlock/identity.go:63-72, filepath.EvalSymlinks failure on a missing leaf falls back to the unresolved absolute spelling. The first real acquisition then creates the home and lock tree. A later New call for the exact same configured path successfully resolves the symlink and produces a different Manager.home. Because managerlock.go:293-295 hashes Manager.home into the home-lock filename, the two managers open different files in the same physical lock directory and can simultaneously hold what must be one exclusive manager-home lock. stateForHome also keys process-local order tracking by the unstable spelling.

Concrete Darwin proof: before creation, EvalSymlinks for .../alias/home failed and fallback identity was .../alias/home; after mkdir through alias, the resolved identity was .../real/home. manager-home domain hashes were 65dc597326ed19a58ca23e48e40ec381c97cc3dbdf890c99f47f3008cf9a9386 and 9d2c6c4a12443694eca80f0470766685e81f1401e7161c69872d5d3cd6595c81. This violates canonical identity and same-home cross-process exclusion.

Required rework: make canonicalAbsolute stable across creation by resolving the longest existing ancestor and appending the cleaned nonexistent suffix, with platform-correct canonicalization. Add a regression that constructs manager A for a missing home through an aliased existing ancestor, acquires home-only so state is created, then constructs manager B from the same configured path and proves contention before A releases and acquisition after release. Cover stable Home identity before and after creation; add Windows-appropriate canonical-prefix coverage where symlink privileges are unavailable.

## Passing evidence

Independent review passed go test -race -cover ./internal/managerlock -count=1 -v with 82.1 percent coverage; go vet ./internal/managerlock; focused gofmt; git diff --check; Linux amd64 and Windows amd64 managerlock test compilation; and make check across the repository. Native Windows runtime is unavailable on this Darwin host; the repository windows-latest CI matrix runs go test ./....

No product code was modified during review.