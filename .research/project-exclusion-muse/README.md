# Muse project-source discovery replay

TASK-261021-4bo0a7 measures Muse Code 1.4.11-R6667.1 on macOS/arm64.
The executable SHA-256 is
`6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`.
See the [qualification record](../../docs/qualification/project-source-exclusion-muse.md)
for P1-P8 verdicts and gaps. No Curator adapter or product behavior changes.

## Replay

Set `MUSE_EXCL_BINARY` privately to a task-local copy of that exact executable.
The harness rejects another digest before creating a case directory. Never
point it at an updater-managed writable executable. Choose a fresh run root;
existing case directories refuse instead of reusing possibly trusted state.

```sh
# All tests explicitly skip before creating fixtures or starting Muse.
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-muse -count=1 -v

# MUSE_EXCL_BINARY must already identify the qualified binary.
export MUSE_EXCL_RUN_ROOT="$PWD/.temp/muse-project-exclusion/replay-new"
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-muse -count=1 -v -timeout 8m

# Expected-red narrowing and setup controls. The wrapper must exit 0.
export MUSE_EXCL_RUN_ROOT="$PWD/.temp/muse-project-exclusion/controls-new"
sh .research/project-exclusion-muse/mutants.sh

env -u GIT_DIR -u GIT_WORK_TREE go vet ./.research/project-exclusion-muse
env -u GIT_DIR -u GIT_WORK_TREE go build ./...
```

Use the repository's Go 1.25.5 toolchain. Keep Go build/cache/temp outputs in
the task's ignored workspace. Ordinary `go test ./...` excludes this hidden
directory; the explicit ungated command above verifies the default skip gate.
The default suite additionally needs a normal Git checkout and a short TMPDIR
on macOS. Its exact-head disposable-clone evidence belongs in the task outcome.

## Isolation and tools

Go's standard library drives the probes. Git creates synthetic checkouts with
`git init --separate-git-dir`: their `.git` is a file and their metadata is
outside the checkout. No real repository metadata is changed. `sandbox-exec`
enforces a network deny and denies the real account's native config, data,
credential directories and Keychains. Every vendor invocation is wrapped in
`env -i`, with disposable HOME and four XDG roots, fixture TMPDIR,
`MUSE_NO_AUTO_UPDATE=1`, file credential backend, and `--provider echo` for
session commands. No credential, auth command, remote provider or model is used.
Approval judging and reminder agents are disabled. Native child tool execution
is explicitly unmeasured, not delegated to another agent.

`tmux` provides a private background PTY for resume and late-file probes,
without loading a user tmux configuration. Its relative private socket avoids
macOS Unix-socket path limits. The harness records its server/pane PIDs,
exits normally, and cleans up only a server whose private socket still identifies
the recorded PID. It never focuses the user's terminal. `/rules` supplies native
rule-list observations; echo text is only a liveness control.

`curl` and a Go loopback HTTP fixture test the sandbox itself: an unsandboxed
positive reaches the fixture once; the same request under the vendor policy
exits 7 without another hit. A harmless file remains readable while a synthetic
Keychains file is denied and unchanged. The server closes before the test ends.

`mutants.sh` deliberately restores workspace trust for one excluded P1 case,
removes the network deny, and removes the synthetic Keychains deny. Each must
fail its specific invariant. It also supplies a wrong binary and an invalid run
root and requires refusal before vendor startup or fixture writes. A crash,
skip, missing plant, or unrelated setup failure cannot satisfy these controls.

## Evidence and limits

Each case leaves raw vendor output, native trace logs, process PIDs/exit codes,
terminal snapshots where applicable, and extracted startup observations under
`MUSE_EXCL_RUN_ROOT`. These raw files can contain local paths and remain ignored.
Only sanitized excerpts and aggregate test results are committed in `evidence/`.
The checker requires exactly one matching startup event, exact outcome and
non-duplicate integer fields. Missing/failed/duplicate/contradictory records fail.

Startup is bounded by the native `security_mode.resolve` event. Later root-only
provider/agent reloads and bootstrap copies are not mixed into startup counts.
Resume may retain root rules without emitting a fresh rules-loader record; its
root verdict therefore uses `/rules`, while live catalogs use loader diagnostics.
Do not interpret startup counters as proof about lazy tools, native children,
permission enforcement or all plugin capabilities.

“Approved” here means a deliberately selected synthetic copy placed in the
managed HOME/XDG fixture. It demonstrates a native delivery channel, not Curator
approval, protected-store integrity or hostile-code confinement. Native workspace
trust approves the checkout broadly. A fresh untrusted home is a measured exec
control for named startup sources, not a complete certified strict adapter.
