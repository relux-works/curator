# TASK-260916-vygif0 — revision 2 amended review

Persisted verdict branch: CHANGES_REQUESTED for recoverable CR lifecycle repair only. Technical review: ACCEPT under rose-air-rust-amendment-2.md. This supersedes the previous revision-2 changes-requested verdict: the operator explicitly retained the closed target registry and amended the execution requirement. No implementation changes were needed or made by this reviewer.

Exact candidate: 0464af8889c9989fe5e18664f121af40f17a4dd6, base fcb4faf7ddf1e2fa0dc7faa70afcd403e423e529. Compared 804 non-board candidate blobs with worktree bytes (including symlinks): 0 mismatches. GitHub commit e54aae4e32a3f7461869a8b41b7f42b024193d5e resolves through the GitHub Git API to that exact candidate tree.

Findings:
- rust-toolchain.toml:12-13 pins 1.91.0 and minimal. internal/rustsource/toolchain_registry.go:23 exposes the existing release; pipeline.go:24 and the registry reuse it without changing admitted versions or targets.
- .github/ci/install-rust-toolchain.sh:31-47 reads the file, installs via rustup, prepends the Cargo shim directory through GITHUB_PATH, and checks the executables. ci.yml supplies guard then shared installer before all 4/4 test-gate job definitions (hosted matrix, rose-air, race matrix, candidate matrix). No lane-specific Rust release is installed.
- The missing-rustup branch at install-rust-toolchain.sh:34 names docs/self-hosted-runner-setup.md; its one-time prerequisite is documented at that note:9. Producer results explicitly state rose-air stays red until operator setup.
- platform-cases.tsv:491-492 and gate-selftest.sh:1068 onward require production execution on current Darwin lanes and ledger the Linux/Windows target-absence skips. The actual hosted macOS target executes both cases; no x86 macOS skip exception is needed for these configured lanes. Future target-matrix changes require revisiting this GOOS-based ledger.

Independent checks (zsh command host; Bash for shell gates):
- Focused extraction of the committed Rust guard/installer self-test section: exit 0, 16/16 rows passed. Includes equal pin, file drift, supported-value drift, absent inputs, floating channel, nonminimal profile, installer invocation/PATH, and no-rustup diagnostic.
- Narrowing mutant in an isolated temporary copy: replace equality refusal with refusal only for channel 1.93.0. Focused test exit 1, 14 passed / 2 failed; both file-drift and supported-value-drift rows killed this mutant. Coverage claim is 1/1 attempted narrowing mutants, not exhaustive parser or workflow mutation coverage.
- go test ./internal/rustsource -run 'TestSupportedRustToolchainVersionIsTheRegistrySource|TestCargoToolchainValidateAdmitsOnlyTheSupportedVersion' -count=1: exit 0.
- git diff --check: exit 0.
- rustc is absent locally (command-not-found), so production-manager cases were NOT rerun locally; no toolchain or host-state install was performed.

Exact remote evidence independently inspected, not replayed: https://github.com/relux-works/curator/actions/runs/35150638518 (success). Downloaded test-evidence-macos-latest/test/go-test-served.json: both TestProductionManagerCapturesRegistryFromRawPaths and TestProductionManagerCapturesGitWithoutCallerProjection have terminal Action=pass (2/2). Remote macOS race gate also reports both ok. All 3/3 hosted gate-selftest jobs report 133 passed, 0 failed, one explicit unrelated suite-plan/ledger group skipped because its prerequisite was absent; all three include passing no-rustup and diagnostic rows. Hosted test matrices, race jobs and lint succeeded. The existing rev2-validation.log ends exit 0, required command shards 1/1 green.

Bounds: rose-air was skipped on this candidate remote run and a green main-push rose-air result is NOT established here. Per the binding amendment, acceptance relies on the approved hosted macOS execution and leaves the documented one-time rustup prerequisite plus subsequent rose-air main-push verification to delivery. No registry widening, manual toolchain installation, commit, or done transition is authorized by this review. Findings recorded in this task-scoped outcome; LOGBOOK.md was not edited, per campaign rules.

Completed independent CI gate self-test: `bash .github/ci/gate-selftest.sh` exited 0, 180 passed / 0 failed. This includes ledger-consistency checks, workflow guard/install ordering and production skip-class negative cases. Attached as TASK-260916-vygif0_review-amended-selftest.log. This was the CI shell self-test, not a manual replay of the full landing suite.

Run goal queried before verdict: this run is not goal-bound. All live checklist items are checked. Attempted accept_cr(revision=2), but acceptance was refused; no acceptance or integration state was persisted.

## Lifecycle refusal and exact next action

The first acceptance attempt using the updated existing artifact was refused with change_request_evidence_missing because the run manifest had no original digest. Attached a new task-scoped verdict artifact and retried. That attempt was refused with change_request_state_conflict: revision 2 remains changes_requested and does not admit accepted. This is a recoverable board lifecycle failure, not a product blocker or a human-only decision. Route to-dev for the tracked developer/implementer to publish an unchanged fresh CR through handoff, retaining this exact tree and evidence, then obtain a new reviewer acceptance. No implementation changes requested. Do not infer acceptance from the technical verdict or bypass the board state machine.
