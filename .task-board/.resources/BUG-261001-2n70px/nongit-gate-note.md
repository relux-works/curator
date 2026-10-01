# Gate note — BUG-261001-2n70px rev1 (orchestrator)

Hosted run 36871917690: every lane is green except Test (windows-latest). There `go test` exits 0, but the platform-case gate fails:

    FAIL  skip with an unrecognised reason on windows: cmd/curator :: TestInstallGitignoreEntryPoint/unexpected-exit-128
    FAIL  skip with an unrecognised reason on windows: cmd/curator :: TestInstallGitignoreEntryPoint/unexpected-exit-2

The gate allows a skip only when `.github/ci/platform-cases.tsv` declares it with a recognised reason.

Pick one:
- (a) Make these rows run on Windows: a fake git that exits 128/2 works on Windows via a .exe/.cmd shim; use the established shim helpers.
- (b) If a row truly cannot run on Windows, add a ledger row with the established reason vocabulary (see existing rows and `.github/ci/gate-selftest.sh`), and make the skip message match it exactly.

Prefer (a). Run `bash .github/ci/gate-selftest.sh` (or its ledger subset) locally, with real exit codes, before re-handoff.
