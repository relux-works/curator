# TASK-261015-1yj9x7 — developer handoff

Ready for review. Candidate is uncommitted on the managed Story branch.

## Scope and identity
- Base: 3f5de6dff9aab4c037ba069323604b166bf0414e; fetched origin/main equals HEAD and local main; HEAD..main count 0.
- Changed files: internal/buildrepo/admission_test.go and internal/buildrepo/transport_test.go only.
- regularFalseStub writes a regular executable shell script (0755, exit 1) into t.TempDir(); all three hard-coded credential broker / SSH wrapper assignments now use it.
- Production code unchanged. Existing reviewer tests retained under their original names.
- No rework outside the named module. README Tools commands remain accurate; no new tool or dependency.
- Environment: Go 1.25.5, darwin/arm64; golangci-lint 2.12.2 built with Go 1.25.5.
- Candidate patch SHA256: 9e0ec94d435bebd7c39228a42754254775d01fb741d74776ce30f9826d064d34.
- admission_test.go SHA256: a37fd61d260046c440e184102b318910986f2f48e742b2336d3b7969fd3daa01.
- transport_test.go SHA256: 80f29effaa060555c2b3462b99fa25b862ee66d028f3732719ecdc37cda419fc.
- Unchanged admission.go SHA256: cfe024bf7013f75cdca2ff3680aeb5210422b8c547daa6b398cb0d6bc2d49c4a.

## Acceptance coverage and bounds
2 of 4 AC rows driven through production entry points, with row 2 bounded to the tested platform:
| AC row | Evidence / production call site | Bound |
| --- | --- | --- |
| No test relies on coreutils paths being regular | Repository-wide tracked *_test.go audit found zero /usr/bin/true, /usr/bin/false, /bin/true or /bin/false paths. Only the three reported assignments existed before the change. | Structural audit, not a production behavior row. |
| buildrepo passes with symlinked system utilities and on macOS | Full go test -count=1 ./internal/buildrepo/... green on macOS. TestNetworkAndLocalSHA1SHA256RawObjectParity exercises AcquireNetwork and AdmitLocal through the changed HTTP helper; TestResolvedTransportFallsBackOnAvailabilityThenVerifies exercises AcquireNetworkResolved through the changed transport helper. | No Linux/uutils host run was performed. The fixtures no longer reference the system true/false files. This is structural portability evidence, not a claimed Linux execution. |
| Symlinked broker still refused | TestGitToolRejectsSymlinkedBrokers/credential_broker calls ValidateGitTool; regular stub first accepted, link refused with CodeIdentityInvalid and the broker-specific message. SSH_wrapper subtest proves the same for SSHWrapper. | Controlled Unix symlinks; Windows subtest is explicitly skipped. |
| Production code unchanged | Exact diff audit: only the two scoped *_test.go files changed. | Structural audit, not a production behavior row. |

Brief gap: no attack-surface table was supplied; no catalog coverage map can be claimed. No AC rows treated as out of contract. No new source-text-inspecting gate exists, so the source-token behavioral-mutant requirement is inapplicable. No new incident or decision requires a product logbook change; task-scoped evidence is recorded here.

## Direct validation
All commands ran as standalone processes, redirected to complete local logs without tee. No earlier-run green evidence was substituted.
| Command | Exit | Result |
| --- | ---: | --- |
| go test -count=1 ./internal/buildrepo/... | 0 | ok github.com/relux-works/curator/internal/buildrepo 70.006s; includes existing reviewer suites and the new refusal test |
| make lint | 0 | golangci-lint run; 0 issues |
| go build ./... | 0 | All packages compile |
| bash .github/ci/no-broad-suppression.sh | 0 | no-broad-suppression: ok |
| git diff --check | 0 | Clean |
| Python tracked-test coreutils-path + changed-file scope audit | 0 | Zero paths; only the two scoped test files changed |

Scoped buildrepo property/refusal suites are green. The full hosted landing workflow (scripts/remote-gate.sh) was not invoked in this developer leaf; this note does not claim hosted checks passed. Orchestration owns integration and its landing gate.

## Narrowing mutants
Mutants ran with Go overlays stored under .temp/TASK-261015-1yj9x7; production source on disk remained unchanged. Each replaces the broker-loop Lstat result with Stat only for the named field. The gate remains: absolute paths, target regular-file validation, missing-file rejection and the other field's symlink refusal still apply. The Lstat token is preserved, and the actual behavioral suite executes.
| Mutant | What it narrows the gate to | Named failing test | Exit / survival bound |
| --- | --- | --- | --- |
| askpass: follow regular-file symlinks only for credential broker | Admits the credential-broker symlink class; all other object/path restrictions remain | TestGitToolRejectsSymlinkedBrokers/credential_broker | 1, killed; observed error nil instead of CodeIdentityInvalid |
| ssh: follow regular-file symlinks only for SSH wrapper | Admits the SSH-wrapper symlink class; all other object/path restrictions remain | TestGitToolRejectsSymlinkedBrokers/SSH_wrapper | 1, killed; observed error nil instead of CodeIdentityInvalid |

Commands (both truthfully expected-red):
- go test -count=1 -overlay .temp/TASK-261015-1yj9x7/askpass-overlay.json ./internal/buildrepo -run '^TestGitToolRejectsSymlinkedBrokers$'
- go test -count=1 -overlay .temp/TASK-261015-1yj9x7/ssh-overlay.json ./internal/buildrepo -run '^TestGitToolRejectsSymlinkedBrokers$'

Survivors: none. Raw mutant output:
```
--- FAIL: TestGitToolRejectsSymlinkedBrokers
    --- FAIL: TestGitToolRejectsSymlinkedBrokers/credential_broker
        admission_test.go:762: symlinked credential broker error = <nil>, want build_repository_identity_invalid refusal
FAIL
--- FAIL: TestGitToolRejectsSymlinkedBrokers
    --- FAIL: TestGitToolRejectsSymlinkedBrokers/SSH_wrapper
        admission_test.go:762: symlinked SSH wrapper error = <nil>, want build_repository_identity_invalid refusal
FAIL
```

Local scratch evidence: .temp/TASK-261015-1yj9x7/{buildrepo-test-01.log,lint-01.log,build-01.log,no-broad-suppression-01.log,diff-check-01.log,mutant-askpass-01.log,mutant-ssh-01.log,candidate.patch}. Readiness probes are private scratch and not published.

The initial required status command returned exit 1 because an estimate was required. An estimate of 1 was recorded, then development status succeeded before implementation.
