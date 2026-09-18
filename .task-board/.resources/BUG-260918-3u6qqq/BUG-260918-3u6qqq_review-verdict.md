# BUG-260918-3u6qqq independent review — revision 1

Verdict: CHANGES_REQUESTED — reviewer runtime binding repair only. Technical review passed; no code changes requested. repeat-of: none.
Reviewer run: RUN-260918-95f53a. Goal query: run is not goal-bound.

Exact candidate: 62e0d5a6e93922631e52d68915e997e8f45b5fbe.
Gate commit: 1eeb7433fd9bbbdff443497b90a3bf273d160dda; git rev-parse commit^{tree} equals the candidate. Full tracked-worktree diff against the candidate is empty; no untracked nonignored files. Only the three scoped files differ from base a4a3bcee268b39edffc2f84de37bc9b869c1dd57.

Implementation: install-rust-toolchain.sh:38-45 computes CARGO_HOME with HOME fallback, requires GITHUB_PATH, appends the bin directory and exports PATH before command -v rustup and the installation call. It has no RUSTUP_HOME dependency. The exact docs-named refusal, channel parser, minimal-profile install, shim checks and diagnostics remain unchanged. gate-selftest.sh:993-1015 drives the real installer with fake binaries, including isolated CARGO_HOME and an unrelated RUSTUP_HOME. docs/self-hosted-runner-setup.md accurately describes launchd/no shell profiles and lane discovery; no workflow topology changes.

Independent verification (bash, local macOS): extracted the original common self-test harness plus the contiguous rust-pin-guard and install-rust-toolchain sections, changing only HERE to the actual CI script directory. No assertions were replaced. Ran bash narrow.sh: exit 0, 19/19 assertions passed. Installer behavior classes driven: 3/3 (PATH rustup, CARGO_HOME-only rustup, absent in both). Absent row expects exit 1 and verifies the docs note; channel/guard cases stayed green.

Narrowing attack: a scratch copy moves only the rustup availability check back before cargo_home/PATH setup, restoring the original PATH-only bug. The same self-test sections target that copy: exit 1, 16 passed / 3 failed. The new CARGO_HOME-only row exits 1 instead of 0 with the original docs-named error; its install-log and GITHUB_PATH assertions also fail. Required mutant killed: 1/1; this is bounded regression proof, not exhaustive mutation coverage. Repository source was never modified by the reviewer.

Checks:
- `git diff --exit-code 62e0d5a6e93922631e52d68915e997e8f45b5fbe --`: exit 0
- `git diff --check`: exit 0
- `bash -n .github/ci/install-rust-toolchain.sh .github/ci/gate-selftest.sh`: exit 0
- shellcheck unavailable (command lookup exit 1); not claimed as run.

Existing hosted validation reused, not rerun: BUG-260918-3u6qqq_change-request_rev1-validation.log reports sh scripts/remote-gate.sh exit 0, command coverage 1/1. Independently queried GitHub run 35300299336 with gh (exit 0), confirmed successful conclusion and exact head SHA above. All 11 executed jobs succeeded, including Linux/macOS/Windows gate self-tests, all hosted Test jobs, both Race jobs, Lint, Naming and Interop. Candidate suite and rose-air are skipped. https://github.com/relux-works/curator/actions/runs/35300299336

Bounds: reviewer reran only the 19 narrow assertions, syntax and whitespace checks; did not rerun the full landing suite or producer's full local 183-case suite. Real rose-air ARM64 runner operation remains unverified until a main push after landing, as specified by the brief. No new product or architecture decision, anomaly or regression requires a logbook change; campaign rules prohibit LOGBOOK.md edits.

Lifecycle finding: accept_cr(BUG-260918-3u6qqq, revision=1, evidence=BUG-260918-3u6qqq_review-verdict.md) exited 1 with change_request_acceptance_unauthorized: this run was handed revision 0 and cannot attest revision 1. This is a recoverable reviewer/runtime binding failure, not a human blocker. No acceptance was persisted.

Required rework: coordinator must route a new tracked reviewer run explicitly bound to published CR revision 1; retain the exact candidate and existing hosted validation. That reviewer must perform its required independent review and accept the revision it was actually handed. No source edits or duplicate full landing suite requested by this review. Routing to to-dev to release reviewing and carry the runtime repair; do not mark done or supply commit_ack.

## Narrow test output
```
=== rust-pin-guard.sh: the committed toolchain file cannot drift ===
ok    the committed rust-toolchain.toml agrees with SupportedRustToolchainVersion
ok    a bumped toolchain file is rejected
ok    a changed supported value is rejected
ok    a missing toolchain file fails closed
ok    a toolchain file with no channel fails closed
ok    a floating channel is rejected
ok    a non-minimal profile is rejected
ok    an unreadable Go source fails closed
ok    a Go source with no declaration fails closed

=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
ok    the installer finds rustup under CARGO_HOME/bin without PATH help
ok    the CARGO_HOME-only install names the filed channel
ok    the CARGO_HOME bin dir is recorded for the rest of the lane
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

review-rust: 19 passed, 0 failed
```

## Mutant output
```
=== rust-pin-guard.sh: the committed toolchain file cannot drift ===
ok    the committed rust-toolchain.toml agrees with SupportedRustToolchainVersion
ok    a bumped toolchain file is rejected
ok    a changed supported value is rejected
ok    a missing toolchain file fails closed
ok    a toolchain file with no channel fails closed
ok    a floating channel is rejected
ok    a non-minimal profile is rejected
ok    an unreadable Go source fails closed
ok    a Go source with no declaration fails closed

=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
FAIL  the installer finds rustup under CARGO_HOME/bin without PATH help
      exit 1, want 0; output: rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane 
FAIL  the CARGO_HOME-only install names the filed channel
      expected to find: toolchain install 1.92.0 --profile minimal
FAIL  the CARGO_HOME bin dir is recorded for the rest of the lane
      expected to find: /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//gate-selftest.mbG3Ge/fake-cargo-only/bin
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

review-rust: 16 passed, 3 failed
```
