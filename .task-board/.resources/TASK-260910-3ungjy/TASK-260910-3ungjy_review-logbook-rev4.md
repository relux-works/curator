# Review logbook — TASK-260910-3ungjy revision 4

R1–R3 independently pass through the production status/env status entry points, text and JSON, normal and --check. Revision 3→4 is only a JSON assertion repair; no Windows folding/skip or production behavior change. No new product defect found.

Shared-host validation anomaly: the first 24-test CLI invocation exited 1 before tests at the TestMain host-GOROOT lock deadline (301.014s). A bounded retry passed (256.322s). Two initial mutant attempts were killed while waiting in TestMain → AcquireHostGOROOT → AcquireHomeOnly; neither was counted as a killed mutant. The retry allows eight minutes for the existing five-minute lock acquisition. No lock bypass, process termination, or test-harness weakening.

PowerShell was missing from PATH, but the sibling results documented /tmp/pwsh/app/pwsh. Using that installation closed the local capability gap: all 14 shell-trust vectors passed without skips; sh/bash/pwsh fresh-binary approve/revoke E2E passed. Hosted revision-4 log separately records all required Linux/macOS/Windows lanes green.

Repository LOGBOOK.md intentionally untouched per campaign rules; this task-scoped board outcome is the review logbook.

Final adversarial result: 2/2 class-narrowing mutants caught by named committed CLI tests; original bytes restored and both tests pass. Review accepted; route to integrating through accept_cr revision 4.
