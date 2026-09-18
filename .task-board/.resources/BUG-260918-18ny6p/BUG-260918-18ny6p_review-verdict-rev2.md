# Review verdict: ACCEPT revision 2

Task: BUG-260918-18ny6p. Reviewer independently reviewed exact candidate tree 96e61845041cc903feb63fb20af20c258f719e4d against base 6401d3c551fad8cacf73956c1aca5a889765fb4f. No product code modified.

## Identity and scope
`git diff 96e61845041cc903feb63fb20af20c258f719e4d --exit-code` exited 0: working tree matches candidate. Only the three assigned paths differ from base. `git rev-parse 68735a1129b5b232e8b3a49cf7b648bd5d7a4d19^{tree}` exited 0 and returned exactly the candidate tree. Attached rev2 validation records hosted run 35339413867 success, exit 0, with Linux/macOS/Windows test and gate-selftest jobs, lint, naming, interop and race jobs green. This hosted evidence is reused, not rerun. Rose-air was skipped.

## Independent validation
Shell: bash for syntax and test execution, launched via zsh; output redirected directly, no pipeline masking.
- `bash -n .github/ci/install-rust-toolchain.sh .github/ci/gate-selftest.sh`: exit 0.
- `git diff --check 6401d3c551fad8cacf73956c1aca5a889765fb4f`: exit 0.
- Extracted the unchanged gate-selftest harness prelude and complete install-rust-toolchain section, setting HERE to the candidate CI directory and adding a FAIL-count exit. `bash .temp/review-18ny6p/narrow.sh`: exit 0; 14/14 assertions passed, zero skipped. Drives the actual production install script: PATH rustup, CARGO_HOME-only, Homebrew-prefix-only, and absent-everywhere. Absence returned expected exit 1 and named runner setup documentation.
- Narrowing mutant in an isolated scratch copy removes ONLY the HOMEBREW_PREFIX candidate, preserving PATH, CARGO_HOME and both fixed prefix probes. Same section with IRS pointing to that copy: exit 1, 11 passed and 3 failed. The new Homebrew success, invocation and GITHUB_PATH assertions fail; existing CARGO_HOME and absence assertions remain green. Mutation detection: 1/1 attempted narrowing mutants killed.

## Assessment and bounds
The candidate preserves the prior CARGO_HOME prepend, then uses command -v and ordered executable fallback probes as explicitly required by the implementation brief. Both proxy and discovered rustup directories are added to current PATH/GITHUB_PATH before the first rustup call. Existing post-install rustc/cargo checks remain. Hosted PATH-found behavior and workflow topology are preserved. Documentation explains minimal launchd PATH, no shell profiles, automatic Homebrew discovery and optional runner .path configuration without requiring reinstall.

Dynamic discovery coverage is 3/5 location shapes: PATH, CARGO_HOME and configurable Homebrew prefix; fixed /opt/homebrew/bin and /usr/local/bin are inspected in the same ordered loop, not dynamically exercised on this host. The absent-everywhere row ran locally (no skip). On hosts with fixed-prefix rustup it is explicitly named as skipped. Real rose-air behavior remains unverified until the post-landing main push, as planned. The shipped row verifies GITHUB_PATH after completion; before-call ordering is additionally established by direct source inspection. No new architectural or external decision is needed.

Verdict: ACCEPT via accept_cr revision=2; route to integrating, not done. Run goal query reported no goal binding.

## Narrow test output
```
=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
ok    the installer finds rustup under CARGO_HOME/bin without PATH help
ok    the CARGO_HOME-only install names the filed channel
ok    the CARGO_HOME bin dir is recorded for the rest of the lane
ok    the installer finds rustup under the Homebrew prefix without PATH help
ok    the Homebrew-prefix install names the filed channel
ok    the Homebrew bin dir is recorded for the rest of the lane
ok    the CARGO_HOME bin dir (proxies) is still recorded alongside it
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

PASS=14 FAIL=0 SKIPPED=0
```

## Narrowing mutant output
```
=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
ok    the installer finds rustup under CARGO_HOME/bin without PATH help
ok    the CARGO_HOME-only install names the filed channel
ok    the CARGO_HOME bin dir is recorded for the rest of the lane
FAIL  the installer finds rustup under the Homebrew prefix without PATH help
      exit 1, want 0; output: rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane 
FAIL  the Homebrew-prefix install names the filed channel
      expected to find: toolchain install 1.92.0 --profile minimal
FAIL  the Homebrew bin dir is recorded for the rest of the lane
      expected to find: /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//gate-selftest.KY0eDT/fake-brew/bin
ok    the CARGO_HOME bin dir (proxies) is still recorded alongside it
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

PASS=11 FAIL=3 SKIPPED=0
```
