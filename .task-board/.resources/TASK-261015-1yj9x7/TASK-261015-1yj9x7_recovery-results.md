# TASK-261015-1yj9x7 developer recovery handoff

Ready for review. This report supersedes the earlier result's candidate identity. The managed Story candidate remains uncommitted.

## Change and recovery cause

The candidate changes only internal/buildrepo/admission_test.go and internal/buildrepo/transport_test.go. regularFalseStub creates a regular executable shell script with mode 0755, contents "#!/bin/sh\nexit 1\n", in t.TempDir(). All three credential-broker and SSH-wrapper assignments now use that helper.

TestGitToolRejectsSymlinkedBrokers first admits the regular fixture, then refuses its symlink separately as AskPass and SSHWrapper through production ValidateGitTool. It checks CodeIdentityInvalid and the field-specific refusal message. Production source and all prior reviewer test names remain unchanged.

The predecessor's configured hosted validation failed: native Windows go test exited 0, but platform-case-gate exited 1 because the new Unix fixture's Windows skip was unclassified. Recovery changes only that skip message to "controlled broker symlink fixture is exercised on unix runners", matching the existing platform-control class. No CI class, ledger, gate, or production behavior was changed.

## Identity

- Base HEAD: 3f5de6dff9aab4c037ba069323604b166bf0414e. HEAD..local main count: 0. No branch, index, or checkpoint changes made.
- Candidate binary patch SHA256: 1bea09633f1b6b6d38fe395ae2a61895b5f0f3e8a4a9b79cb861d9eab4c354ae.
- admission_test.go SHA256: 4b84e89d17a8e8f64ff54479b80b5627e9aef74037ece8c0e89241235430e559.
- transport_test.go SHA256: 80f29effaa060555c2b3462b99fa25b862ee66d028f3732719ecdc37cda419fc.
- Unchanged admission.go SHA256: cfe024bf7013f75cdca2ff3680aeb5210422b8c547daa6b398cb0d6bc2d49c4a.
- Environment: Go 1.25.5, darwin/arm64; golangci-lint 2.12.2 built with Go 1.25.5.
- README Tools commands remain accurate; no new dependency or project tool.
- Rework stays inside internal/buildrepo; no paths outside the named module changed.

## Acceptance coverage

2 of 4 AC rows driven through production entry points; the portability row has the explicit platform bound below. The remaining two rows are structural properties.

| AC row | Named test / production call site or structural check | Bound |
| --- | --- | --- |
| No test relies on system true/false being regular | Repository-wide audit of all tracked Go tests finds zero hard-coded /usr/bin/true, /usr/bin/false, /bin/true or /bin/false paths; broader non-document source search also finds none. | Structural property, not counted as a production call. |
| buildrepo passes on the affected layouts and macOS | Full buildrepo suite on macOS. TestNetworkAndLocalSHA1SHA256RawObjectParity drives AcquireNetwork and AdmitLocal; TestResolvedTransportFallsBackOnAvailabilityThenVerifies drives AcquireNetworkResolved using the changed helper. | No native Linux/uutils run in this recovery. The fixture is independent of system true/false paths; do not infer native platform execution from that fact. |
| Symlinked broker remains refused | TestGitToolRejectsSymlinkedBrokers/credential_broker and /SSH_wrapper drive ValidateGitTool's broker loop in internal/buildrepo/admission.go. | Controlled Unix symlinks; Windows fixture is explicitly skipped and classified as platform-control. |
| Production unchanged | Exact HEAD diff permits only the two scoped test files; admission.go bytes equal HEAD. | Structural property, not counted as a production call. |

Brief gap: no attack-surface table was supplied, so no catalog coverage map is claimed. No AC rows are treated as out of contract. No new source-text-inspecting gate exists.

## Direct validation

All listed commands ran in this recovery, as standalone gate processes without tee. Expected-red commands retain their actual nonzero exits. No earlier green result substitutes for a current validation.

| Command | Exit | Evidence |
| --- | ---: | --- |
| go test -count=1 ./internal/buildrepo/... | 0 | Full package suite, including prior reviewer tests and new symlink refusals; 94.025s |
| make lint | 0 | golangci-lint run; 0 issues |
| go build ./... | 0 | All packages compile |
| bash .github/ci/no-broad-suppression.sh | 0 | no-broad-suppression: ok |
| git diff --check | 0 | No whitespace errors |
| gofmt -d internal/buildrepo/admission_test.go internal/buildrepo/transport_test.go, followed by an empty-output assertion | 0 | No formatting delta |
| Tracked-test path and changed-file scope audit | 0 | Zero coreutils paths; only the two test files changed |
| Production platform-case-gate with original synthetic Windows skip | 1 | Expected red: unrecognised skip reason |
| Production platform-case-gate with candidate synthetic Windows skip | 0 | Exactly one skip, classified allowed-platform-control |

The synthetic Windows fixtures contain only this case and use an empty tier-1 ledger, while reading the repository's actual skip-class table and executing its actual platform-case gate. They establish classification only, not native Windows test execution or full-ledger coverage.

Landing-gate property self-tests: bash .github/ci/gate-selftest.sh exited 0 with 265 passed, 0 failed, 2 skipped after loading the repository-prescribed agent Git environment. The first attempt exited 1 because fixture commit signing failed without that environment, and its ancestry property therefore failed. The retry is diagnostic, not a hidden rerun. The two local skips are rustup-absence fixtures on a platform with a fallback rustup installation; those two rows were not executed here. Full output and bounds: TASK-261015-1yj9x7_gate-selftest-recovery.log.

The configured full hosted landing command, sh scripts/remote-gate.sh, was not rerun by this recovery. Its previous failure is recorded as a failure, not accepted green evidence. The Change Request flow reruns that configured command against the new snapshot; no new hosted-matrix result is claimed here.

## Narrowing mutant evidence

Both overlays preserve the Lstat token and all path/object checks. Each follows links only in one broker field, leaving the other field's gate intact. They execute the actual behavioral test through production ValidateGitTool; the production file on disk remains unchanged.

| Mutant | What it narrows the gate to | Named failing test | Exit / survival bound |
| --- | --- | --- | --- |
| AskPass overlay: replace broker Lstat result with Stat only for credential broker | Admits regular-target credential-broker symlinks while retaining absolute-path, missing-path, regular-target, and SSH-wrapper restrictions | TestGitToolRejectsSymlinkedBrokers/credential_broker | 1, killed; got nil instead of refusal |
| SSH overlay: replace broker Lstat result with Stat only for SSH wrapper | Admits regular-target SSH-wrapper symlinks while retaining absolute-path, missing-path, regular-target, and credential-broker restrictions | TestGitToolRejectsSymlinkedBrokers/SSH_wrapper | 1, killed; got nil instead of refusal |

Commands:
- go test -count=1 -overlay .temp/TASK-261015-1yj9x7/askpass-overlay.json ./internal/buildrepo -run '^TestGitToolRejectsSymlinkedBrokers$'
- go test -count=1 -overlay .temp/TASK-261015-1yj9x7/ssh-overlay.json ./internal/buildrepo -run '^TestGitToolRejectsSymlinkedBrokers$'

Survivors: none.

## Evidence and logbook

TASK-261015-1yj9x7_recovery-validation.log contains direct command outputs and actual exit codes. TASK-261015-1yj9x7_recovery-logbook.md records the hosted failure, correction, and platform bounds without editing the control root. Raw local logs remain under .temp/TASK-261015-1yj9x7/. The candidate is ready for review and for the orchestrator's snapshot validation.
