# TASK-260910-16k7xy — gate failures rev1 AND rev2 (identical; rev2 changed nothing that mattered)

Runs 35146117950 and 35151271983 fail on exactly two things:
1. Lint (golangci revive indent-error-flow): internal/snapshot/capture.go:334:10 — the `else` after a returning `if` must go (outdent the block; move the short variable declaration to its own line).
2. Windows platform-case gate: internal/snapshot TestCaptureMatchesNormativeVectors/{base,build-edit,runtime-edit} skip with the undeclared reason "platform does not preserve POSIX execute bits; normative executable vectors unverified here".

Required: the three normative vector cases must RUN on Windows. The vector's executable flags are part of the declared inventory, so on Windows derive the executable flag from the declared source (the vector manifest / Git index mode for tracked files, an explicit flag list for untracked fixture files) instead of the filesystem mode, and make the fixture create files accordingly. Do NOT add a skip class. Only if one sub-case truly cannot be expressed on Windows, use the existing platform-control reason text verbatim: "Windows does not expose portable executable permission bits" (.github/ci/skip-classes.tsv:66) and explain in results.md. Hand off rev3 only after `golangci-lint run ./internal/snapshot/...` (or the repo's lint make target) is clean locally.
