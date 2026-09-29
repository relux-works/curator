# TASK-260910-2t0iun — gate fix 2, exact (THE ONLY CURRENT INSTRUCTION, with 2t0iun-decision-1.md)

Rev2 (tree 90c2ffc6) still fails the Windows platform-case gate with `FAIL skip with an unrecognised reason on windows: internal/install ::
TestInstallScriptSecurityRows` — you did not touch the two ledger files. The gate (.github/ci/platform-case-gate.sh) classifies a skip by
matching its printed reason against the regexes in .github/ci/skip-classes.tsv, and a skip is only tolerated for a case listed in
.github/ci/platform-cases.tsv with that class. Do exactly this:
1. .github/ci/skip-classes.tsv — add one row (TAB-separated, same shape as the existing platform-control rows):
   platform-control<TAB>install\.sh supports macOS and Linux<TAB>allow<TAB>install.sh is a POSIX shell installer; Windows installs use the release archive directly
2. .github/ci/platform-cases.tsv — add one row (TAB-separated, same column order as the existing rows):
   internal/install<TAB>TestInstallScriptSecurityRows<TAB>linux,darwin<TAB>windows<TAB>platform-control<TAB>install.sh verification rows (POSIX installer)
   (if the test has subtests that also skip, add the `TestInstallScriptSecurityRows/*` wildcard row the way other cases do).
3. Keep the t.Skip text byte-identical to the regex ("install.sh supports macOS and Linux").
Run `bash .github/ci/gate-selftest.sh` (if it runs locally) and `go test ./internal/install -run InstallScript` with real exit codes.
Set status development; update results (`git diff 90c2ffc6` shows both .tsv files); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
