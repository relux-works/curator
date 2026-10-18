# Self-hosted runner setup

The `test-self-hosted` lane (`Test (rose-air)` in
[ci.yml](../.github/workflows/ci.yml)) runs the default gate on the
organisation's self-hosted macOS ARM64 runner (labels `self-hosted, macOS,
ARM64, rose-air`). The runner must carry the custom `rose-air` label: the
lane is pinned to it, so another self-hosted Apple-silicon runner without
that label never takes the job. It runs on `main` pushes only, and only while the repository
variable `ROSE_AIR_RUNNER` is `true`.

## One-time prerequisites on the runner

- **rustup.** Install once for the runner user, either with the installer at
  https://rustup.rs using its `--no-modify-path` form (verify
  `~/.cargo/bin/rustup --version` runs) or as the Homebrew formula
  (`brew install rustup`; verify `/opt/homebrew/bin/rustup --version`
  runs). An existing install of either kind is fine as is. The lane installs
  the pinned toolchain itself from [rust-toolchain.toml](../rust-toolchain.toml)
  on every run and never relies on a toolchain the runner happens to ship,
  so no toolchain needs installing or pinning by hand. Homebrew may link only
  `rustup` into `/opt/homebrew/bin` while keeping the `rustc` and `cargo`
  proxies in the versioned rustup formula keg; the lane resolves that keg
  from the installed `rustup` executable when the proxies are not already on
  `PATH`. No manual proxy links are needed. A lane on a runner without rustup
  fails in the `Install pinned Rust toolchain via rustup` step with this note
  named.

The runner service starts from launchd with a minimal PATH
(`/usr/bin:/bin:/usr/sbin:/sbin`, no `/opt/homebrew/bin`) and reads no
shell profiles, so `rustup` is not expected on the service PATH. The lane
locates it itself: it probes `PATH`, then `~/.cargo/bin` of the runner user
(or `$CARGO_HOME/bin`), then the Homebrew prefixes (`$HOMEBREW_PREFIX/bin`,
`/opt/homebrew/bin`, `/usr/local/bin`) and always invokes it by that
absolute path. It never adds a shared prefix bin such as `/opt/homebrew/bin`
to `PATH`, because that directory also holds Homebrew `go` and would shadow
the Go pinned by `go.mod`. Only directories holding the Rust proxies are
added: `~/.cargo/bin`, the symlink-resolved Homebrew keg of `rustup`, or,
for a non-symlinked `rustup` outside `~/.cargo/bin`, a lane-private
directory under `RUNNER_TEMP` linking exactly the Rust proxies. If `rustc`
or `cargo` is still missing after installing the pinned channel, it uses
`rustup which --toolchain <pinned-channel> rustc` to locate the pinned
toolchain's bin directory. All additions are written to `GITHUB_PATH` for
later lane steps. The step records where `go` and `node` resolve before any
change and fails, naming the shadowing directory, if an added directory
would change either.
The Homebrew prefixes are probed so a launchd-started
service can find `rustup` even when `/opt/homebrew/bin` is absent from its
initial PATH. The runner's `.path` file may also carry that prefix, but the
lane does not depend on it.

When no probed location holds `rustup`, the step fails with one
diagnostics block before the remedy note: the runner name (`RUNNER_NAME`),
`hostname`, `whoami`, `HOME`, `CARGO_HOME` (set or defaulted),
`HOMEBREW_PREFIX`, the `PATH` it searched, one
executable/exists-not-executable/absent line per probed candidate, a
bounded rust/cargo listing of the searched bin directories, and
`command -v` / `type -a` for `rustup`. The next red run is therefore
self-diagnosing: the block tells whether the job landed on a different
machine, a different service user, or a `rustup` outside the probed
paths.

Nothing else is installed by hand: Go and Node come from the
`actions/setup-go` / `actions/setup-node` steps, pnpm is installed per lane
into a lane-local prefix, and the Rust toolchain comes from rustup as above.

## After setup

No maintenance: toolchain versions are pinned in the repository
(`rust-toolchain.toml`, `PNPM_PIN`) and each lane verifies its pin against
the Go constant (`rust-pin-guard.sh`, `pnpm-pin-guard.sh`) before anything
installs.

## Linux runners (`relux-linux`)

The Linux entries of `Test`, `Race`, `Go family qualification`, `Gate self-test`, `Lint`,
`Interop conformance gate` and `Naming gate` can run on the repository's
self-hosted Linux runners (labels `self-hosted, relux-linux, curator`). Check
names do not change: a routed job is still `Test (ubuntu-latest)`.

Routing is decided per run:

- the repository variable `LINUX_SELF_HOSTED` must be `true`; otherwise every
  job stays on the hosted image;
- pushes, manual dispatches and pull requests whose head branch lives in this
  repository go to `relux-linux`;
- a pull request from a fork always runs on the hosted image. The guard sits
  in every routed `runs-on` expression, and `gate-selftest.sh` fails if any
  `runs-on` naming `relux-linux` loses it. Never route `pull_request_target`
  or fork-triggered work to the self-hosted label, and keep the repository
  setting that requires approval before fork workflows run for all external
  contributors.

Snapshot publication refuses a destination whose retained ancestry carries an
access or default ACL. The qualified self-hosted Linux runners meet that
requirement; the hosted Linux image does not. A fork pull request's hosted
Linux snapshot controls are therefore expected to refuse. This is a known
limit, not a pass.

One-time prerequisites on each Linux runner host (the runner account itself
needs no sudo and no Docker):

- a C toolchain (`gcc`, `cc`, `make`; on Debian or Ubuntu the
  `build-essential` package): the Rust lanes link with `cc`, and `Race` needs cgo and a C compiler;
- `rustup` for the runner account, installed per rustup.rs under that
  account's `~/.cargo/bin`; `install-rust-toolchain.sh` resolves it there;
- `git`, `python3`, `tar`, `xz`, `curl`, `zip` and `unzip`;
- outbound HTTPS to GitHub and the Go, Node and Rust distribution hosts.
  Toolchains still come from the pinned `setup-*` actions and the pinned
  installers, so the toolchain-identity rules hold on these runners as on
  hosted ones.

## Linux job isolation and scheduling

Every Linux-routable job runs `bash .github/ci/isolate-go-state.sh` immediately
after checkout, before `setup-go` or any Go command. The script creates a fresh
directory under `RUNNER_TEMP` with `mktemp`, then publishes `GOCACHE`,
`GOMODCACHE` and `GOPATH` through `GITHUB_ENV`. Even jobs sharing a runner
account, temporary directory or Go version get separate build and module
caches. The runner owns temporary-directory cleanup; isolation does not depend
on a successful final step. No shared cache is deleted.

On self-hosted Linux, `setup-go` cache restore/save is disabled. The lint
action also skips its cache and uses a private `GOLANGCI_LINT_CACHE`. This
prevents restored per-user state from reconnecting otherwise isolated jobs.
Hosted macOS/Windows caching and routing and the fork guard retain their
existing behavior. `GOTOOLCHAIN=local` and `GOENV=off` remain
mandatory. Shared Go state was implicated in intermittent cross-toolchain
link errors involving `golang.org/x/sys/unix.Syscall` and ABI0/ABIInternal.

Workflow concurrency is keyed by workflow and ref. Pull requests cancel
superseded runs; pushes (including `main` and `gate/**`) never cancel an active
run. This bounds overlapping runs of the same ref, not the number of jobs
inside a run, other refs, or work from other repositories. It does not impose
a host-wide scheduling limit.

The go-v1 probe has a fixed 15-second wall-clock bound in
`internal/godriver/session.go`. `ConfigFromEnvironment` reads only trusted Go
selectors; there is no existing documented CI timeout knob. `ProbeTimeout`
cannot raise the bound: values above 15 seconds are clamped. The workflow
therefore adds no timeout flag and makes no claim to qualify arbitrarily
loaded hosts. The test package timeout is a different bound.
Toolchain fingerprinting has a separate two-minute bound, and larger
`FingerprintTimeout` values are also clamped. Neither bound has an environment
override. A fingerprint timeout needs capacity/filesystem investigation or an
explicit product deadline-policy decision; cache isolation alone does not
establish loaded-host qualification.

Validate this contract with `go test -count=1 ./tools/ciworkflow/`, which parses
the real workflow, exercises narrowing plants, executes the initializer twice
with poisoned ambient caches, and evaluates fork/trusted routing expressions.
Inline `run` bodies reserve `GOCACHE`, `GOMODCACHE` and `GOPATH` for the
initializer: the gate rejects references to these names, including shell
exports, command-local assignments, environment-file writes and literal
quoted/escaped names. Reads and comments using these names also require a
reviewed repository script. This conservative source contract does not
interpret arbitrary shell programs: computed variable names and changes
inside called scripts or third-party actions require separate review. A
behavioral regression executes the actual first Go consumer with the review's
export plant to verify that Go consumes the unsafe paths and the gate rejects
that body.
Use `sh scripts/remote-gate.sh` for actual runner evidence. Keep logs in
`.temp/`. A workflow contract check does not reproduce the historical ABI
linker failure. Require three consecutive green `main` runs on the self-hosted
pool before declaring the pool qualified; a green `gate/**` snapshot alone
does not satisfy that rollout requirement.

### Temporary tree self-heal

Each Linux-routable job ends with `scripts/restore-runner-tmp.sh`, guarded by
`always() && runner.environment == 'self-hosted' && runner.os == 'Linux'`.
The host job-completed hook removes the runner's `TMPDIR` tree (including
stale test trees from previous jobs); `RUNNER_TEMP` holds Actions state and
is not a substitute for that root. The script restores owner write permission
with `chmod -R u+w`, never removes files, and always exits successfully.
Unset, empty, non-directory, symlink and filesystem-root values are refused.
Recursive chmod does not follow links inside the tree. Run
`go test -count=1 ./tools/ciworkflow/` for workflow, refusal and behavior checks.
