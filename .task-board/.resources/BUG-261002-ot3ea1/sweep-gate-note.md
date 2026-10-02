# Gate note — BUG-261002-ot3ea1 rev1 (orchestrator)

Hosted run 36967359936 failed on two things:
1. **Every test lane:** `internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded` reports `internal/buildcache/process_linux.go:linuxExecutablePaths tests not-exist after os.ReadFile, os.Readlink without a seam route or reviewed allowlist reason`.

   /proc is OS process state, not manager-owned state. The right fix is most likely a reviewed allowlist entry with a precise reason in the guard's established allowlist format: "procfs enumeration of live process executables; absence means the process exited; never treated as manager-state absence". Alternatively, route the reads through the stateread seam if that fits better.

   Either way: an unreadable /proc entry must still fail SAFE (keep the build).
2. **Lint:** `internal/buildcache/process_linux.go:36:21: G304 Potential file inclusion via variable (gosec)`. Constrain the path to `/proc/<numeric pid>/exe` (validated digits) and use the repo's established gosec annotation pattern with a justification, or build the path from a validated integer.

Re-run locally with real exit codes: `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded -count=1`, the buildcache tests, and `golangci-lint run ./internal/buildcache/...` if available.
