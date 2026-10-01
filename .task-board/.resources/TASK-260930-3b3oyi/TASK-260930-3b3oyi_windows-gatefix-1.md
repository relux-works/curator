# TASK-260930-3b3oyi — Windows gate fix 1

Hosted run 36778472331 failed its platform-case gate, not the Go package tests. The downloaded go-test.json contains zero fail events. All six other first-run regression tests passed; TestPostureWarningOncePerProcessTree was skipped with reason `no sh on this platform`. The hosted wrapper reported Go test exit 0 and platform-case gate exit 1. The attached hosted log and extracted events preserve that evidence; they describe the previous candidate, not a Windows runtime validation of this revision.

Test file SHA-256: e3a35de700fa97fb394ed387f0686736bab8163262b83e0879adbe04e56c56df.

Changed cmd/curator/firstrun_ux_test.go: removed the Windows skip and shell fixture. Build the real manager as curator-run (curator-run.exe on Windows), then execute it with `run env status --json`. The real umbrella dispatcher starts the same native executable with `env status --json`. No test assertions were weakened. The test now also requires exit 0, a JSON status document from the child, and exit 0 from both standalone and marked-child controls.

Observed production binary rows on macOS:
- Nested `run env status --json`: exit 0, one warning.
- Standalone `env status --json`: exit 0, one warning.
- Marker CURATOR_INTERNAL_POSTURE_WARNED=1, `env status --json`: exit 0, empty stderr.

Exact newline-terminated warning for the first two rows:
`warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default`

Validation directly run in this developer session:
- `gh run view 36778472331 --log-failed`: exit 0.
- `gh run download 36778472331 -p 'test-evidence-windows*'`: exit 0.
- `go test ./cmd/curator -run 'Bootstrap|GlobalInit|EnvStatus|EnvResolve|Posture'`: **exit 1**, Go default 10m timeout (601.694 seconds), active TestEnvStatusMatrix in native version detection; no assertion failure printed. Full log attached.
- `go test ./cmd/curator -count=1 -v -run '^TestPostureWarningOncePerProcessTree$'`: exit 0 (36.670 seconds; test 34.72 seconds).
- `go test ./cmd/curator -count=1 -timeout 4m -v -run '^(TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree)$'`: exit 0 (112.533 seconds), 7/7 task-specific regression tests passed. Complete verbose log attached. This bounded rerun covers the seven first-run tests; it is not a passing rerun of the entire combined mask.
- `GOOS=windows go vet ./cmd/curator ./internal/...`: exit 0.
- `GOOS=windows go test -c ./cmd/curator -o /tmp/TASK-260930-3b3oyi-gatefix-1/curator.test.exe`: exit 0; compile validation only.
- `golangci-lint run`: exit 0, 0 issues (v2.12.2).
- `go build ./...`: exit 0.
- `git diff --check`: exit 0.

Runtime: Go 1.26.0, darwin/amd64. Windows runtime tests and the full hosted platform-case gate were not rerun locally because this is a macOS host; the updated test compiles for Windows and has no skip. CURATOR_CONFORMANCE_ROOT is unset locally; the targeted checks do not establish full protocol conformance.

CHANGELOG/LOGBOOK entry text (files not edited, as instructed): Replaced the shell-dependent first-run posture test with native manager/provider execution on all platforms. Hosted Windows failure was an unclassified skip, with Go tests green; the revised row now proves nested execution, successful exit codes, one warning across the process tree, and the standalone/marked-child controls.
