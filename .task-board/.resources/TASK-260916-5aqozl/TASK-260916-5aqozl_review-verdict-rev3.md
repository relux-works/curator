# TASK-260916-5aqozl revision 3 — ACCEPT

Reviewed exact candidate tree 02d5499d24c4662d04207c8cf510bcf3d0f96e4d against base 81fd85b2f721a81e4bd00f499834967f3779be73. All six changed worktree files match candidate blobs byte-for-byte. GitHub gate commit bb52b77f9e04e73338827b2b091c1ce98a2907a8 has this exact tree (verified via GitHub API). No production code changes. No blocking findings.

Scope follows pnpm-ci-decision.md, decision-2, windows-failure-2 and reviewer addendum; these explicitly supersede corepack-only and all-three-Windows-executed wording. npm installs the self-contained package; exactly two Windows cases are deferred to BUG-260916-2f3xbf.

1. Workflow .github/workflows/ci.yml:128,253,360,571 provisions the shared PNPM_PIN after setup-node in all four suite job definitions: Test, rose-air, Race, candidate-conformance. Each checks the Go constant, installs into RUNNER_TEMP/pnpm-prefix, and appends both platform bin layouts to GITHUB_PATH before tests. Existing Race matrix is Ubuntu/macOS only. The macbook-iv incident and mechanism are documented at lines 61–81. Guard lines 27–37 fail on unset, unreadable, absent or mismatching pin.
2. conformance_test.go:1248 resolves Windows launcher extensions to adjacent node_modules/pnpm/bin/pnpm.cjs, verifies existence, then the existing version probe validates 10.33.0. Ordinary POSIX executable paths pass through unchanged. privatedir.MakeAll correctly creates private Windows output directories in the test harness without weakening production privacy checks.
3. conformance_test.go:873,1127 calls the Windows deferral only after pnpm availability/version checks; helper at 1308 guards runtime.GOOS == windows. The ledger requires both cases on Linux/macOS and tolerates only stage-deferred on Windows; no third deferral. Missing pnpm retains its explicit existing reason.
4. YAML parses successfully (Ruby YAML); existing pull_request/gate and main-push conditions are intact. rose-air is main-push-only and remains unverified until integration, not claimed passing. Candidate-conformance is conditionally skipped in this gate.

## Hosted evidence inspected independently

Attached rev3 validation log quotes: `remote gate: run 35105225871 finished: success`, `[exit 0]`, `required=1 green=1 failed=0 missing=0`. Run: https://github.com/relux-works/curator/actions/runs/35105225871 . I downloaded the three test-evidence and two race-evidence artifacts and read their go-test-served.json and Windows skips-observed.tsv, rather than relying on job color alone.

Measured real-pnpm outcomes: 13/15 PASS, 2/15 authorized Windows deferrals across five hosted Test/Race jobs. All 7/7 non-deferred Test cases PASS; all 6/6 Race cases PASS. This is NOT three executed tests on Windows.

Windows skip evidence for both LockSupersetSnapshotDependencies and PrivateStoreAndOfflineMaterialization: `stage-deferred tolerated-by-ledger deferred on windows pending BUG-260916-2f3xbf: pnpm writable store registry contains an undeclared member`. TargetPrunedUnreachableRejectsBeforeInstall PASS on Windows. Exact extracted events follow:

race-evidence-macos-latest/race/go-test-served.json: TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall pass
race-evidence-macos-latest/race/go-test-served.json: TestRealPinnedPNPMLockSupersetSnapshotDependencies pass
race-evidence-macos-latest/race/go-test-served.json: TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization pass
race-evidence-ubuntu-latest/race/go-test-served.json: TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall pass
race-evidence-ubuntu-latest/race/go-test-served.json: TestRealPinnedPNPMLockSupersetSnapshotDependencies pass
race-evidence-ubuntu-latest/race/go-test-served.json: TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization pass
test-evidence-macos-latest/test/go-test-served.json: TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall pass
test-evidence-macos-latest/test/go-test-served.json: TestRealPinnedPNPMLockSupersetSnapshotDependencies pass
test-evidence-macos-latest/test/go-test-served.json: TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization pass
test-evidence-ubuntu-latest/test/go-test-served.json: TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall pass
test-evidence-ubuntu-latest/test/go-test-served.json: TestRealPinnedPNPMLockSupersetSnapshotDependencies pass
test-evidence-ubuntu-latest/test/go-test-served.json: TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization pass
test-evidence-windows-latest/test/go-test-served.json: TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall pass
test-evidence-windows-latest/test/go-test-served.json: TestRealPinnedPNPMLockSupersetSnapshotDependencies skip
test-evidence-windows-latest/test/go-test-served.json: TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization skip

Ubuntu Race stream includes one non-JSON Go dependency-download diagnostic; it was identified and excluded from event parsing, not treated as test evidence.

## Independent local verification

Bash with pipefail: PATH=/tmp/pnpm-prefix/bin:$PATH go test -count=1 -run 'TestRealPinnedPNPM|TestResolveWindowsPNPMEntrypoint' -v ./internal/pnpmsource/ — exit 0, all three integration cases PASS plus resolver tests; package 32.149s. Uses existing task-local npm-installed pnpm, whose existing version probe enforces the pin.

Bash with pipefail, default PATH absent pnpm: go test -count=1 -run '^TestRealPinnedPNPM' -v ./internal/pnpmsource/ — exit 0, exactly three SKIP with `pinned pnpm executable unavailable`; no hidden passes. git diff --check passes. No full landing-suite rerun; hosted lint/build/platform results accepted only for the exact verified tree.

Coverage bounds: local host is Darwin; Windows execution comes from hosted artifacts. Workflow self-test uses textual wiring checks, not a complete semantic GitHub Actions interpreter. Future main-push rose-air execution remains an integration follow-up. BUG-260916-2f3xbf owns restoration of the two Windows tests. No source files modified by reviewer.

CI self-test completed: `gate-selftest: 157 passed, 0 failed`. Independently exercised negative pin drift/unset/unreadable/missing-declaration cases and narrowed synthetic streams: Windows bug deferrals accepted, same skips on Linux rejected, absent-pnpm class on deferred Windows cases rejected. These are gate-entry behavioral checks, not mutation of production code. The shell wrapper returned 0; self-test completion summary confirms zero failed assertions.
