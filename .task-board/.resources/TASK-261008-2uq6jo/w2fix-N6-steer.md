# Steer for TASK-261008-2uq6jo after revision 1 (read first)
Revision 1's hosted gate: all go tests passed on every platform; only the Windows **platform-case gate** failed:
- `internal/gitcred :: TestCredentialAnswerBoundRefusesOversized` skipped on windows with reason "fixture helper is a POSIX shell script";
- `internal/gitcred :: TestCredentialExactFrameBound` skipped on windows with reason "POSIX fixture";
→ "skip with an unrecognised reason on windows … add it to .github/ci/skip-classes.tsv with a class, or fix the case."
Preferred fix: make the regression run on Windows too (a Go-built helper or a `.cmd`/PowerShell helper chosen by GOOS, as other gitcred tests do — look at how existing gitcred tests provide a credential helper on Windows). If that is genuinely impossible, use the EXACT reason text of an existing class in `.github/ci/skip-classes.tsv` that fits (read the file) instead of inventing new wording; add a new class only if none fits, and say why in the results.
No local go test on the mini (R223); compile-only. Results resource: one paragraph on what you chose.
