# TASK-260922-1t551d rework 3 (orchestrator, binding) — F-C1

Revision 3 gate FAILED only on `Test (windows-latest)` platform-case gate (run 35688466491):
`FAIL skip with an unrecognised reason on windows: internal/envprofile ::
TestCredentialLinkTargetInspectionFailure`. Linux/macOS lanes are green. Every skip reason must
match `.github/ci/skip-classes.tsv` (see `.github/ci/platform-case-gate.sh` tier 2): use the exact
reason text/class the sibling envprofile symlink rows use on Windows (privilege / POSIX-mode
semantics), or register the case in `.github/ci/platform-cases.tsv` with the sibling's
must/skip shape. Check every other new row's Windows skip the same way. Do not widen the
vocabulary. Continue from the revision-3 tree (no checkout/clean/stash); append "Revision 4" to
results.md; republish only on a green gate. No other change.
