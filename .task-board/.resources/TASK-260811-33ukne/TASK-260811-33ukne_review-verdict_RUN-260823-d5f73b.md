# Reviewer verdict for TASK-260811-33ukne

Verdict: **accepted**

Reviewer run: `RUN-260823-d5f73b` (Claude claude-opus-5, independent acceptance review).
Run is not goal-bound (`spawn goal` reports no active goal).
Reviewed state: the intentionally dirty worktree on branch `codex/legacy-board-repair`.
Nothing was staged, committed, reset, cleaned, or otherwise modified by this run.

## Per-scope-item verdict

### 1. Wrapper / process-family issue — accepted

The formerly-blocking defect is closed. Production Git acquisition and mirror
verification launch the real C0-bound Git executable:

- `internal/swiftpmsource/git.go:53` `newSharedGitAuthority` resolves
  `Config.GitToolRoot` with `filepath.EvalSymlinks`, joins
  `Toolchain.Git.ExecutableRelativePath`, and requires the resolved executable
  to remain below the resolved root (`pathWithinRoot`). Escape, relative, or
  linked paths reject with `closure_derivation_unauthorized` before any
  process start.
- `ToolIdentity.ProcessFamily` binds exact relative paths plus SHA-256 digests.
  Every member is symlink-resolved, containment-checked, read, and digest-
  verified at authority construction (`git.go:76-90`) and again in
  `authority.recheck` (`git.go:127-146`), which is the recheck callback passed
  to both `SourceAcquisitionExecutor.Execute` and `Executor.Execute`.
- The aggregate C0 fingerprint covers the family:
  `toolProcessFamilyFingerprint` (`git.go:110`) is required to equal
  `Toolchain.Git.Fingerprint` whenever a family is declared, and that
  fingerprint is what `toolNodeRecord` publishes into the C0 checkpoint.
- `internal/closureexec/acquisition.go:504-517` independently re-resolves the
  executable below the trusted `ExecutableRoot` and calls
  `verifyPortableExecutable`, which rejects non-regular/linked files and any
  byte difference from `permit.ExecutableSHA256`. A lying `Toolchain.Recheck`
  therefore still cannot get different bytes launched.
- `SourceAcquisitionExecutor.Execute` (`acquisition.go:467-503`) holds the
  causal mutex, single-use-consumes the permit, checks the causal head, and
  rechecks before and after the sole process seam.
- `exactGitTestTool` (`swift_integration_test.go:519`) binds the real
  `EvalSymlinks("git")` bytes and the real `git-upload-pack` digest when
  present. There is no hashed shell wrapper anywhere in the tree.

Independently confirmed non-skipped on this host:
`TestGitC0ExecutableMismatchRejectsBeforeAnyProcessStart`,
`TestGitC0RelativeAndSymlinkEscapeRejectBeforeAnyProcessStart`,
`TestGitC0ProcessFamilyDriftRejectsBeforeAnyProcessStart`,
`TestGitBrokerCapturesExactLocalRevisionTreeAndMirrorR01R07R08R09` — all PASS.

### 2. Cross-adapter process-seam allowlist — accepted, with one hardening note

`internal/rustsource/build_test.go:53` allows exactly the two shared seams
`acquisition.go` and `portable_runner.go`. An exhaustive search
(`exec.Command`, `os/exec`, `syscall.Exec`, `StartProcess`, `.Start()`) over
`internal/swiftpmsource/*.go` excluding tests returns nothing: the SwiftPM
adapter production code owns no direct process seam. The only production
`exec.Command*` calls in `internal/closureexec` are
`acquisition.go:528` and `portable_runner.go:153`.
(`portable_process_{unix,windows,other}.go` import `os/exec` only to set
`SysProcAttr` on an already-constructed `*exec.Cmd`.)

Note (non-blocking): the guard globs `*.go` and `../closureexec/*.go` only. It
does not scan `internal/swiftpmsource`, so nothing prevents a future direct
`exec.Command` there from regressing silently. See findings below.

### 3. Canonical schemas, C1/C2/C3 reachability, exact revision/tree, deterministic mirror closure — accepted

- Complete extracted trees are admitted through the real recursive artifact
  scan (`Policy.AdmitDependencyDirectory`, `capture.go:578-597`) before the
  manifest permit is committed; `evaluateAdmittedAt` runs only afterwards.
  `TestNoAffectedManifestOnIntakeToolPinAndMirrorRejectionCGN16CGN18` proves
  zero evaluator starts on compiled-payload, tool-drift, and origin-drift
  rejections.
- Receipt reachability is complete. `evidenceIDs` (`graph.go:238-267`)
  aggregates manifest permit/receipt, broker permits/receipts, source intake,
  mirror intake, commit-evidence intake, mirror derivation permit/receipt, and
  mirror verification receipts into C1 `JournalEntryIDs`, C2
  `IntakeReceiptIDs`/`OriginIDs`/`ProtectedHandleIDs`/`BrokerReceiptIDs`, and
  C3 `ArtifactManifestIDs`/`DerivationReceiptIDs`. Generated-lock evidence adds
  `resolutionJournalIDs`, `resolutionPermitID`, `resolutionReceiptID`
  (`capture.go:249-252`).
- Exact revision/tree are cryptographically re-derived rather than trusted.
  The mirror transform (`closureexec/source_control_mirror.go`) recomputes the
  Git tree SHA-1 from the admitted source bytes and requires it to equal the
  acquisition tree, requires the admitted commit object to start with
  `tree <tree>\n`, and requires re-hashing that commit object to reproduce the
  pinned revision. Any of the three failing yields
  `closure_derivation_drift` before the mirror exists.
- Kind preservation is enforced at capture (`Mirror.OriginalKind`/`LocalKind`)
  and re-checked in `ReplayOffline` (`replay.go:58`), which also requires a
  bijection between mirrors and root-lock pins, re-verifies protected mirror
  bytes against `MirrorDigest`, and re-runs `fsck --full` through a
  `network=none` C4 derivation permit.
- Selection neutrality holds: no `target_platform`/`toolchain_component` node
  appears in capture; those exist only in `SelectionBinding`
  (`graph.go:179-230`). `TestCaptureAndOfflineReplayR01R05R06CGP05CGP11`
  proves Darwin and Linux destinations reuse one `GraphDigest` while producing
  distinct binding IDs.
- Lock parsing is strict: schema 2/3 only, `DisallowUnknownFields`, exactly two
  top-level keys, 40/64-hex revisions, no duplicate identities, registry kind
  rejected, `file://` rejected for `remoteSourceControl`.

### 4. Artifact binary denial and narrow mirror authorization — accepted, with a test-coverage finding

`internal/artifactpolicy/source_control_mirror.go` cannot authorize arbitrary
stores or verified binaries:

- The authorization is a sealed opaque interface whose only method
  (`artifactPolicySourceControlMirrorAuthorization`) is unexported, so no
  adapter package can implement or forge it.
- Issuance requires: exact origin/revision/tree agreement with the acquisition
  receipt; `VerifyIssuedReceipt` on that acquisition receipt;
  `InvocationSubtype == DerivationMirror`; `Network == "none"`;
  `InvocationKey == "source-control-mirror-v1:"+acquisitionID`; the admitted
  source receipt present in `AdmittedInputReceiptIDs`; exactly one
  `source-control-mirror-v1` local output at the exact mirror path;
  `VerifyIssuedDerivationChain`; exactly one receipt output whose SHA-256,
  artifact-manifest ID, and schema match; and a canonical evidence payload
  whose `mirror_digest`/`revision`/`git_tree`/`kind`/`acquisition_receipt_id`
  all match.
- `ValidateSourceControlMirrorAuthorization` re-checks the seal pointer plus
  every caller-supplied field via `reflect.DeepEqual` immediately before
  replay.
- `verified-binary-v1` remains explicitly unavailable
  (`artifactpolicy/codec.go:1039-1041`, `CodeBinaryAdmissionUnavailable`).
  `TestExtensionReachabilityAndBinaryPolicyP01P08` and the artifactpolicy
  compiled-vector suite pass.
- The two `source-control-mirror-tree-v1` digests are cross-checked: the
  adapter's `inventoryMirrorTree` result must equal the transform-emitted
  `mirror_digest`, so divergence is fail-closed.

I found no correctness defect. The finding below is coverage, not behavior.

### 5. Portable evidence honesty and fail-closed verified mode — accepted

`portableCapabilities` lists only `immutable-intake-recheck-v1`,
`immediate-toolchain-recheck-v1`, `declared-output-verification-v1` — no
lossless-observation claim. `SourceAcquisitionReceipt.Validate`
(`acquisition.go:229`) and `DerivationReceipt`/`Audit` validation
(`models.go:471`, `executor.go:522`) require `Network == "not-observed"` in
portable mode and `"none"` in verified mode, with empty
process/read/write sets in portable mode. `PreflightAssurance` refuses a
provider in portable mode and, in verified mode, resolves and negotiates the
configured provider with no fallback; `AssuranceConfig.validate` rejects
incomplete verified provider identity. `AssuredCacheInput.ID` namespaces
portable and verified cache identities disjointly, and
`DerivationReceipt.ValidateFor` rejects cross-mode and cross-provider reuse.
The SwiftPM Git authority constructs both executors explicitly as
`AssurancePortable`, so no verified claim is fabricated.

### 6. Final full/race/lint evidence — accepted and independently reproduced

Producer evidence verified:

| Artifact | Check | Result |
| --- | --- | --- |
| `TASK-260811-33ukne_full-go-06.log` | SHA-256 | `c3586a424a9ef0f850e02c613c66150a2e2aa07e7498335a4e1f8013e57e257d` — matches the brief exactly |
| same | contents | `EXIT:0`, 51 `ok` packages, 0 `FAIL`, `cmd/curator 405.794s` |
| `TASK-260811-33ukne_race_RUN-260823-27bddf.log` | contents | `ok closureexec 8.523s`, `ok swiftpmsource 14.790s` |
| `TASK-260811-33ukne_lint_RUN-260823-27bddf.log` | contents | `0 issues.` |

Independently rerun by this reviewer in this worktree:

| Gate | Command | Result |
| --- | --- | --- |
| Full suite minus `cmd/curator` | `go test -timeout 30m -count=1 $(go list ./... \| grep -v 'cmd/curator$')` | exit 0, 50 `ok`, 0 `FAIL` (`artifactpolicy 244.283s`, `rustsource 208.194s`, `swiftpmsource 57.263s`, `closureexec 32.257s`, `closuregraph 17.220s`) |
| Race | `go test -race -timeout 20m -count=1 ./internal/closureexec ./internal/swiftpmsource` | exit 0 (`14.714s`, `20.269s`) |
| Real-tool integration not skipped | `-v -run 'TestReal\|TestGitBroker\|TestGitC0\|TestBrokeredResolver' ./internal/swiftpmsource` | 7/7 PASS, 0 SKIP |
| Lint | `golangci-lint@v2.12.2 run` | `0 issues.` |
| Vet / build / gofmt / `git diff --check` | — | all clean |
| Binary-deny | `-run 'Binary\|Compiled\|Kotlin'` over artifactpolicy + swiftpmsource | all PASS |
| Kotlin exclusion | `cmd/curator` `kotlin-v1` unsupported-driver vectors | PASS (in the `TestBuild…` subset below) |
| `cmd/curator` subset | `-run 'TestBuild\|TestAssurance\|TestStatus' ./cmd/curator` | `ok 37.131s` |
| Canonical goldens | accepted Ruby verifier | `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2`, all references resolve |
| Tracked/untracked binaries | MIME scan of all non-board working-tree files | none |

Precisely what I did **not** rerun: the full uncached `cmd/curator` package
(~7 minutes on the producer host, and its cache-input scan of this dirty
worktree exceeds a single bounded call). I ran the `TestBuild`/`TestAssurance`/
`TestStatus` subset instead and accepted the remainder of `cmd/curator` from
the hash-bound `full-go-06` log, whose digest I verified matches the brief.

Prior verdicts `RUN-260823-5da0b5`, `RUN-260823-2a9dae`, `RUN-260823-5178c3`
each requested changes; I re-derived their named blockers independently and all
are resolved in the current tree.

### 7. Board validity — accepted

`task-board --no-update-check validate` → `Board is valid. No issues found.`

## Findings recorded for downstream work (none blocking)

1. **No negative-vector coverage for the `source-control-mirror-v1`
   authorization.** Coverage attribution over
   `./internal/{swiftpmsource,closureexec,artifactpolicy}` shows that every
   rejection branch of `IssueSourceControlMirrorAuthorization`
   (`artifactpolicy/source_control_mirror.go:57-100`), the
   `ValidateSourceControlMirrorAuthorization` drift branch (`:119`), and every
   permit-shape rejection in the transform runner
   (`closureexec/source_control_mirror.go:48-100`) is unexecuted. Only the
   happy path is covered, by the two Git/Swift-gated integration tests
   `TestRealSwiftPMManifestRunsThroughProductionManagerAndExecutor` and
   `TestBrokeredResolverGeneratesTransitiveLockBeforeMainCaptureR02`.
   `TestReplayRejectsCapturedMirrorByteDriftR08` and
   `TestMirrorAdmissionRejectsCheckoutManifestSubstitution` both run on the
   fixture path where `mirror.authorization == nil`, so they never exercise the
   production control. The logic reads correct and I could not construct a
   bypass, so this is a hardening gap rather than a defect — but the drift /
   tamper / malformed-authorization vectors belong in
   `TASK-260811-x611eq`'s cross-adapter conformance suite.

2. **The cross-adapter process-seam guard does not scan
   `internal/swiftpmsource`.** `internal/rustsource/build_test.go:41-64` globs
   only `*.go` and `../closureexec/*.go`. The SwiftPM adapter is compliant
   today (verified exhaustively), but the regression guard the brief relies on
   does not actually cover it.

3. **Silent downgrade to the fixture mirror path.** `captureMirror`
   (`capture.go:265-269`) falls back to `captureFixtureMirror` for any
   `Config.Broker` that is not exactly `*GitBroker`, and `replay.go:79` skips
   authorization revalidation whenever `mirror.authorization == nil`. Neither
   emits an error or a diagnostic. `internal/swiftpmsource` is not yet wired
   into `cmd/curator`, so this is unreachable from production today; the
   wiring task should make the production path assert an authorized mirror
   rather than infer it from a type switch.

4. **Two implementations of one canonical schema.**
   `closureexec/source_control_mirror.go:250` derives the `executable` field of
   `source-control-mirror-tree-v1` from `Perm()&0o100`, while
   `swiftpmsource/capture.go:468` uses `Mode()&0o111`. They agree today because
   the transform writes only `0o600`/`0o700`, and any disagreement is
   fail-closed via the digest cross-check, but the divergence is latent.

## Handoff

No product code was modified, staged, or committed by this reviewer run. As a
reviewer-archetype run it supplies no `commit_ack`; the accepted delivery is
still uncommitted in the shared dirty worktree and must be committed by the
commit-owning mover.
