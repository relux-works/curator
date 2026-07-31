# TASK-260720-1zntv0 — portable `manager-worker-v1` go-v1 preflight and build

Implementation of the accepted portable execution policy in Curator. This
document is the producer handoff; it makes no acceptance judgement about its own
work.

## 1. Provenance and boundaries

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
  (retained from the previous run; no new worktree was created and no state was
  reimported).
- `HEAD` is still exactly `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`;
  `git rev-list --count 17804cea..HEAD` is `0`; `git diff --cached --quiet`
  exits `0`; `git diff --check` exits `0`. Nothing was staged, committed,
  tagged, pushed, or published.
- Authoritative conformance input, read-only:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`.
  Independently re-verified: `shasum -a 256 conformance/v1/manifest.json` =
  `58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`, matching
  the accepted candidate pin. The host-execution vector
  `conformance/v1/vectors/go-host-execution-policy.json` is present at
  SHA-256 `c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de`.
- The pre-revision rc.4 candidate at
  `.temp/TASK-260720-3ag6pi/worktree/conformance/v1` was consumed only as
  additional test input for the source-aware argv and rejection vocabulary.
- No cache was published, no marker written, no live install mutated, and no
  built artifact was ever executed.

## 2. What was implemented

### 2.1 Execution-policy cache identity (`internal/buildmeta`)

`Policy` gained `ExecutionPolicy`, serialized as `execution_policy` in the
canonical CCJ-1 policy object with the fixed value `manager-worker-v1`.
`ReservedHardenedExecutionPolicy` (`hardened-worker-v1`) is defined and
explicitly rejected by `Input.Validate`, and a pre-revision input without the
field cannot be produced or decoded.

Independently recomputed against the accepted vector:

| Identity | Value | Source |
| --- | --- | --- |
| portable cache key | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` | matches `cache_identity.portable.cache_key` |
| receipt hash (golden) | `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd` | recomputed; pre-revision value was `sha256:750f5f75…` |

`internal/buildmeta` canonical receipt bytes are byte-identical to the accepted
`conformance/v1/schema-cases/build-receipt-v1/valid.json` after canonicalization
(`TestCandidateBuildReceiptSchemaCase`). The pre-revision key
`sha256:3fcd714a…` and the reserved hardened key `sha256:13736230…` are proved
distinct and non-producible, so old candidate entries miss rather than alias.

### 2.2 Native-control inventory and capability evidence (`internal/godriver`)

`controls.go` carries `rc5-native-control-inventory-v1` as an exhaustive, closed
per-platform table for exactly macOS and Windows and exactly five controls, with
`{availability, mechanism, unavailable_reason}` records and the single
`no-private-aggregate-domain` reason. Availability is probed **once per
operation, in the parent, before the worker exists**; no host label, build-time
constant, configuration value, or cached result is used.

`capability-evidence-v1` is closed: exactly `record_version`,
`execution_policy`, `platform`, `controls`; exactly one
`{name, availability, status, probed_at}` entry per inventory control; and the
eight consistency rules with their two stable diagnostics. The record is
result-only — it is returned in `godriver.Result.Evidence` and is absent from
the canonical build input and the canonical receipt (proved by
`TestCapabilityEvidenceIsNotACacheOrReceiptInput`).

Per-platform mechanisms, all applied inside the worker so the compiler and every
tool child inherits them:

| Control | macOS | Windows |
| --- | --- | --- |
| `descendant-domain-termination` | `setsid` private session; parent `killpg` before reaping | Job Object `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`, worker assigned to the job |
| `active-process-count-limit` | unavailable (`no-private-aggregate-domain`) | Job Object `JOB_OBJECT_LIMIT_ACTIVE_PROCESS` |
| `aggregate-memory-limit` | unavailable (`no-private-aggregate-domain`) | Job Object job + process memory limits |
| `per-file-size-limit` | `RLIMIT_FSIZE` lowered in the worker | unavailable (`no-private-aggregate-domain`) |
| `inherited-handle-restriction` | close-on-exec proved on the protocol descriptors, no `ExtraFiles` | `os/exec` `PROC_THREAD_ATTRIBUTE_HANDLE_LIST` with an empty `AdditionalInheritedHandles` |

The six deferred hardened guarantees appear in no mandatory control, no
inventory entry, and no evidence record; their absence produces no rejection,
no diagnostic, and does not block publication
(`TestDeferredHardenedGuaranteesNeverAppearAndNeverReject`).

### 2.3 Identity-verified worker session

- `identity.go` canonicalizes the installed manager executable to a real regular
  file, rejects symlink/reparse/hard-link substitution, records strong file
  identity, and hashes the bytes. `Verify` is called at the launch boundary (so
  a replacement race cannot widen the graph) and again after the last child
  exits.
- `workerproto.go` defines the length-prefixed bounded framing, the closed
  eight-kind message vocabulary, and the HMAC-SHA-256 build permit bound to the
  session secret, the session nonce, and the exact build argument vector.
- `workerclient.go` is the parent: probe → identity → launch
  `os.Executable() __curator-go-worker-v1` over anonymous inherited pipes → one
  canonical request → require the identity proof and the evidence record →
  `list` → parent graph validation → one authenticated permit → `build` →
  artifact verification → post-exec identity re-verification → terminate and
  join the whole domain. The parent bounds the domain at
  `Limits.Timeout + 5s` and kills the group **before** reaping, so a group kill
  can never reach a reused process identifier.
- `workerserver.go` is the worker: it re-proves its own identity against the
  request, re-validates the roots, the exact `go list`/`go build` vectors, the
  manager-derived staging output, the complete closed offline environment, the
  limits, and the probe set; applies the controls; releases unrelated
  descriptors; emits the evidence; then serves exactly one list and, after one
  authenticated permit, exactly one build. It has exactly one
  process-creation site, guarded by a launcher identity check, and reports how
  many programs it started so the parent can require exactly one per phase.
- `cmd/curator/main.go` dispatches the hidden mode before any parsing, requires
  exactly the one manager-owned argument, and keeps it out of the usage surface.

### 2.4 Package-influence exclusions

`BuildRequest.CommandObject` carries the exact package-declared build-command
object. `validatePackageCommandSurface` admits exactly `type`, `driver`, and
`source_dir` with their fixed values and rejects every other key with
`build_execution_package_influence_forbidden`, naming the surface (executable,
argv, environment, output, flags, hooks, plugins, generators) before any
private state is created and before the worker starts.

### 2.5 Removed

The declarative `HostPolicy`/`HostExecution`/`guardedHostPolicy` adapter the
previous review rejected is gone. Its useful content — the closed environment,
argv, root, and staging validation — now lives inside the worker, where it
governs a real process instead of a mock. The former `source_not_readonly`
permission-bit gate was **removed on purpose**: rc.5 defines the portable source
mechanism as the frozen snapshot plus identity re-verification, and a
permission-bit rejection would have been an extra containment claim the portable
profile is not allowed to make.

## 3. Test strategy

Every worker test drives a **real** hidden-mode process. `TestMain` gives the
test binary the same fixed worker mode the installed manager has, so
`internal/godriver` tests exercise the shipped protocol, controls, identity
checks, and teardown, not an in-process stand-in.

`internal/godriver/testdata/stubgo` is a real native executable compiled with
the real toolchain at test time and installed into a fingerprintable fake
`GOROOT`. It is driven exclusively by a manager-owned script file inside that
`GOROOT`, never by an argument or environment value the driver does not already
fix, and it records every invocation so tests assert the exact argument vector,
working directory, and environment the worker used.

Coverage of the rc.5 case inventories:

- all 11 `capability_evidence_cases` are recomputed through the implementation's
  validator, with the expected diagnostic derived from the case fields;
- all 8 `package_influence_cases` surfaces are driven through `Build`;
- the `identity_and_protocol_cases` families are driven against a real worker
  with hand-built frames: identity digest/path/size mismatch, unexpected program
  below the worker, tool directory outside `GOROOT`, mutated list vector,
  mutated build vector, escaped output, poisoned and widened environment,
  working directory outside the snapshot, overlapping private root, unsupported
  protocol version, contradictory probe, foreign platform, permit before list,
  replayed nonce, out-of-order message, unknown kind, oversize frame,
  unauthenticated permit, permit bound to another vector, second build, second
  list;
- `failure_boundary`, `deferred_capability_rejection_guards`,
  `native_control_inventory`, `capability_evidence_record`, `process_graph`,
  `session_states`, `mandatory_controls`, and `cache_identity` are compared
  field-by-field against the vector.

Native-control effect is proved, not declared:

- `TestBuildTerminatesTheCompleteWorkerDomain` starts two long-lived tool
  children below the compiler and requires both to be gone after `Build`
  returns;
- `TestPerFileSizeLimitIsReallyApplied` (macOS) requires a private write above
  the bound to fail *inside* the Go child under `RLIMIT_FSIZE`.

## 4. Gates and exact exit codes

Every command below was run directly as a standalone process. No gate was piped
through `tee` or a pipe chain.

### 4.1 macOS arm64 — local host (Darwin 25.5.0, Go 1.25.5 darwin/arm64)

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `test -z "$(gofmt -l internal cmd)"` | 0 |
| `go test ./... -count=1` | 0 |
| `go test -race ./... -count=1` | 0 |
| `go test ./internal/godriver/ -count=1` | 0 |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1VendoredBuildIsBoundedAndNotLaunched` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<rc.5> CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver/ ./internal/buildmeta/` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<rc.4 candidate> go test ./internal/godriver/ ./internal/buildmeta/` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=arm64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./internal/godriver/` | 0 |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/godriver/` | 0 |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/godriver/` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |
| `git rev-list --count 17804cea..HEAD` = `0` | 0 |

Under the rc.5 conformance root the `internal/godriver` suite reports **222
passing cases, 4 skipped, 0 failed**. Measured statement coverage of
`internal/godriver` is **72.1 %**; the uncovered remainder is dominated by
worker-side statements that execute in a real subprocess (and so are not
attributed to the parent profile) and by the other platform's control adapter.

### 4.2 macOS amd64 — `ssh relux` (macOS 15.7.4, Go 1.25.5 darwin/amd64)

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1` | 0 (all 36 packages `ok`) |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1VendoredBuildIsBoundedAndNotLaunched` | 0 (`PASS`) |

Both remote runs above were repeated end to end after the lint pass of section
4.5, on the exact bytes handed off.

Go 1.25.5 darwin/amd64 was installed at `~/curator-ci/go` on that host for this
validation.

### 4.3 Windows amd64 — `ssh win` (Windows 10 19045.6456, Go 1.25.5 windows/amd64)

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./internal/godriver/` | 0 (`ok`, 93.1s) |
| `go test ./internal/buildmeta/` | 0 (`ok`) |
| `go test ./cmd/curator/` | 0 (`ok`, after installing Git) |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1VendoredBuildIsBoundedAndNotLaunched -v` | 0 (`PASS`, 45.28s) |
| `go test ./... -count=1` | **1** — see 4.4 |

Go 1.25.5 windows/amd64 (`C:\curator-ci\go`) and Git 2.51.0
(`C:\curator-ci\git`) were installed on that host for this validation. The
source was transferred as a vendored tarball, so the Windows run resolved no
module over the network.

### 4.4 Reported red: pre-existing Windows gaps outside this task

`go test ./...` on Windows exits 1. Every failure is in a package this task did
not create or modify, and every one of them fails before any go-v1, worker,
policy, or evidence code is reached:

| Package | Failing tests | Reported cause |
| --- | --- | --- |
| `internal/buildcache` | `TestPublishAndInspectExactProtectedHit`, `TestInspectRejectsCorruptReceiptAndArtifactState`, `TestPublishQuarantinesCorruptEntryBeforeReplacement`, `TestAtomicPublicationIdenticalRace`, `TestAtomicPublicationConflictingRace`, `TestWindowsProtectedStateMatrix`, `TestExplicitQuarantineMovesEntryAndMissingIsNoop` | `prepare protected cache root: untrusted cache provenance: … owner does not match the effective user` |
| `internal/buildsource` | `TestRejectsInvalidPathsLinksAndCollisions/invalid_protocol_path`, `TestFrozenTokenRejectsRootReplacement` | Windows path grammar and open-file rename semantics |
| `internal/globalbins` | `TestSafeSelectionFeedsStagedForwardingTargetWithoutLiveMutation` | `script command is not executable: tool` (POSIX exec-bit assumption) |
| `internal/runtimestore` | `TestPrepareScriptRuntimeStagesIncompleteReplacementWithoutBuildRoots`, `TestPrepareScriptRuntimeReusesOnlyCompleteManagedTree`, `TestPrepareSingleScriptStagesManagedBinWithoutCopyingSnapshotTree`, `TestShimTransitionMatrixIsDeterministicAndManagerScoped`, `TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode` | same POSIX exec-bit assumption plus wrapper forwarding |
| `internal/shell` | `TestPowerShellHookRunsOnEveryPrompt` | PowerShell hook expectation |

These packages pass on both macOS hosts. This is the first native Windows run of
the full suite in this worktree, so these are newly *observed* pre-existing
platform gaps, not regressions introduced here. They are reported rather than
fixed because they are outside this task's ownership.

### 4.5 Lint

`golangci-lint` was not installed on this machine. It was installed for this
task at `.temp/TASK-260720-1zntv0/tools/golangci-lint` with
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.0`, and
run against the repository's own `.golangci.yml`.

| Command | Exit | Result |
| --- | --- | --- |
| `golangci-lint run ./internal/godriver/... ./internal/buildmeta/... ./cmd/curator/...` | **0** | `0 issues.` |
| `golangci-lint run` (whole repository) | 1 | 45 findings, all outside this task's scope |

The first run is this task's lint gate: every package this task owns or touched
is clean. Getting there fixed 24 findings in `internal/godriver`,
`internal/buildmeta`, and `cmd/curator`, including several that pre-dated this
task — unchecked `Close` returns, two `G115` conversions (one now guarded by an
explicit non-negative check), five missing exported-method comments, three
deprecated `runtime.GOROOT` uses replaced with `go/build`.Default.GOROOT, a
builtin-shadowing `copy` variable, and two quickfix suggestions.

The remaining 45 whole-repository findings are in packages this task does not
own (`internal/runtimestore`, `internal/install`, `internal/registry`,
`internal/closure`, and others) and were left alone. `go vet ./...` exits 0 on
all three hosts. The Makefile `lint` target runs the whole-repository form, so
it still exits 1 and no green is claimed for it.

### 4.6 Expected-red gates during development, reported truthfully as failures

1. `go test ./internal/godriver -run TestBuild` after the first end-to-end wiring:
   exit 1, `cannot decode stub call log: unexpected EOF` and
   `go_build_failed … want artifact_size_limit`. Root cause: the tests set
   `FileBytes` to 64–1024 bytes, and the worker really applied `RLIMIT_FSIZE`,
   truncating the stub's own writes. Closed by giving those tests realistic
   per-file bounds and adding a dedicated macOS test that asserts the limit
   fires. This failure is positive evidence that the control is real.
2. `go test ./internal/buildmeta` after adding `execution_policy`: exit 1, the
   golden canonical bytes, cache key, and receipt hash no longer matched. Closed
   by recomputing the goldens to the rc.5 values, not by weakening the assertion.
3. First native Windows run: exit 1 on
   `TestBuildRunsExactlyOneFixedListAndBuildInsideTheWorker` (the test pinned
   `bin/golden-tool` instead of the derived `bin/golden-tool.exe`) and on
   `TestFingerprintRejectsEscapingAndAbsoluteLinks/absolute`. The second was a
   real portability defect: `filepath.IsAbs("/etc/passwd")` is false on Windows,
   so a rooted toolchain link was classified `toolchain_link_dangling` instead of
   `toolchain_link_absolute`. Both were fixed.
4. First native Windows `cmd/curator` run: exit 1,
   `git: executable file not found in %PATH%`. Environmental; closed by
   installing Git on that host, after which the package is `ok`.

### 4.7 Not run, and why

- **Linux runtime.** `rc5-native-control-inventory-v1` covers exactly macOS and
  Windows, so the portable policy is not defined for Linux. The implementation
  rejects there with `build_execution_control_unavailable` before the worker
  starts, which is proved by `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`
  and `TestProbeSetGuardMatchesTheInventoryExactly`. Linux coverage in this task
  is therefore compile, vet, test-binary link, and that rejection path only. No
  Linux host was used and none is claimed.
- **`make lint` as configured.** The Makefile target invokes a repository-wide
  `golangci-lint run`, which exits 1 on the 69 pre-existing findings above. The
  scoped result is reported instead of claiming a green repository-wide lint.

## 5. Files

New in `internal/godriver`: `controls.go`, `controls_darwin.go`,
`controls_windows.go`, `controls_other.go`, `identity.go`, `workerproto.go`,
`workerclient.go`, `workerserver.go`, `controls_test.go`, `worker_test.go`,
`guards_test.go`, `main_test.go`, `process_alive_unix_test.go`,
`process_alive_windows_test.go`, `testdata/stubgo/main.go`.

Rewritten: `build.go`, `build_test.go`, `build_conformance_test.go`,
`platform_unix.go`, `platform_windows.go`.

Removed: `host_policy.go`, `host_policy_unix.go`, `host_policy_windows.go`.

Changed elsewhere: `internal/buildmeta/models.go` (execution policy),
`internal/buildmeta/codec.go` (policy decode), `internal/buildmeta/buildmeta_test.go`
(rc.5 goldens, receipt schema-case check), `internal/godriver/session.go`
(package doc, exported-method comments, `go/build`.Default.GOROOT, tagged
switch, checked `Close`), `internal/godriver/fingerprint.go` (rooted-link
rejection, guarded size conversion, checked `Close`),
`internal/godriver/fingerprint_test.go` (pre-revision-root skip),
`internal/godriver/graph.go` (checked `Close`),
`internal/godriver/graph_test.go` (fixture), `internal/godriver/executor.go`
(locked `Bytes`, doc comment), `internal/godriver/executor_test.go`,
`internal/godriver/session_test.go`, `cmd/curator/main.go` (hidden mode,
bounds-check annotation), `cmd/curator/main_test.go`.

## 6. Boundaries preserved

- No cache published, no marker written, no live install mutated, no ref, tag,
  pin, release, or platform claim created or edited.
- The built artifact is never started; the happy-path test asserts the stub
  launcher recorded exactly the three package-independent probes plus one list
  and one build.
- No package-selected program, argv, environment value, output path, flag, hook,
  plugin, or generator is reachable.
- The predecessor worktree, the curator-spec candidate worktrees, and the rc.4
  candidate were read only.
