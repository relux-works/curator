# TASK-260910-2t0iun — gate fix (THE ONLY CURRENT INSTRUCTION, with 2t0iun-decision-1.md)

Rev1 (tree 68c8cb3c) passes go test everywhere; only the platform-case gate fails on Windows (run 36477364094):
`FAIL skip with an unrecognised reason on windows: internal/install :: TestInstallScriptSecurityRows`. install.sh is POSIX-only, so the
Windows skip is legitimate — register it in .github/ci/platform-cases.tsv with the correct skip class (use the existing class other
POSIX-only shell cases use, e.g. platform-control) and make the t.Skip message match that class's recognised reason format (read
.github/ci/platform-case-gate or equivalent to see the accepted reasons). Required on linux/darwin. Run the gate's platform-case check
locally if a script exists, plus `go test ./internal/install -run InstallScript` with a real exit code. Set status development; update
results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
