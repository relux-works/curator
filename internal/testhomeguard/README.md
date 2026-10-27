# Native-home test boundary

Only test code imports this package. The envprofile and CLI `TestMain` entry
points use `Supervise` before their test-running branches. The outer process
records protected preimages, reexecutes the same binary, waits, and checks the
postimage regardless of the inner test process's normal return, panic, testing
timeout, direct `os.Exit`, or isolation setup/cleanup failure. Fixture CLI
helpers retain the native HOME and provider overrides explicitly supplied by
their parent; their supervisor still observes the OS-account homes.

`internal/testgitenv.Isolate` is the single owner of test-process Git and HOME
isolation. Ambient test processes call it once; explicit fixture helpers inherit
the parent's Git boundary and their intended fixture environment. This package
neither allocates a HOME nor changes provider variables. `CheckIsolation` fails
before tests if the neutral helper does not move HOME or if provider/data/state/
cache overrides remain ambient. This integration requires the neutral helper
from PR #196 (`bc36ebaa` or its accepted successor). Until that dependency lands,
validation uses its exact source as a Go overlay. The helper at that pin leaves
provider and XDG data/state/cache overrides unchanged; the outer test runner must
clear them, or the package deliberately refuses. The shared helper needs to own
that remaining isolation; there is no second pinning implementation here.

The protected inventory includes the OS-account home and, for an ambient test
process, inherited native-home overrides. It observes root identities/modes,
root-context files, provider settings/configuration, environment markers and
backups, and **all entries recursively under regular skill directories**.
Regular noncredential files have content hashes as well as lstat identity,
size, mtime and mode. Symlinks have only their own identity and link text;
linked roots, skill links and instruction links are never traversed. Known
credential carriers (`auth.json`, credential/token/secret names, OAuth files,
and private-key/certificate extensions) have lstat metadata only, never content
reads or hashes. Diagnostics print fingerprints, not file contents or paths.

Bounds fail closed: 65,536 snapshot entries, depth 64, 128 MiB per hashed file,
and 1 GiB of hashed bytes across the inventory. Large installed skill assets
are included within these bounds. Native runtime databases, session histories,
logs and unrelated root-level sidecars are excluded. Directory sizes and mtimes
are ignored; directory identity, mode and complete recursive skill membership
remain checked. This is an observation boundary, not OS filesystem confinement.

`Control` and `Child` exercise the actual package boundary with synthetic
original homes only. The settings and copied-skill decoys already exist before
the snapshot; their mutations preserve file length and restore mtime. They must
fail with the exact synthetic-root fingerprint. Context and environment-state
plants also fail. Panic and direct-exit plants must still emit the changed
postimage; clean panic/exit controls preserve the original exit codes and emit
an unchanged postimage. An isolated native write and an explicitly supplied CLI
helper HOME are positive controls. No plant selects the operator's home.

Example validation while #196 is pending (the overlay maps only the neutral
helper to its exact pinned source):

```sh
runner_home=$(mktemp -d)
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS \
  -u CODEX_HOME -u CLAUDE_CONFIG_DIR -u PI_CODING_AGENT_DIR \
  -u XDG_DATA_HOME -u XDG_STATE_HOME -u XDG_CACHE_HOME HOME="$runner_home" \
  go test -overlay /path/to/neutral-helper-overlay.json -count=1 -v \
  ./internal/envprofile -run '^TestNativeHomeGuard'
```

Residual limits: killing/crashing the supervisor, whole-process-group kills,
host loss, and the pre-TestMain fixture transport/worker/tripwire dispatches
cannot guarantee a final postimage. Killing a supervisor may also leave its
inner process alive; this wrapper is not a process-tree containment facility.
A mutation restored before the postimage, an unchanged symlink's destination
outside the inventory, and credential bytes changed while restoring all their
metadata are not detected. Concurrent legitimate edits can produce a refusal
without proving which process made them. A failed or missing postimage is never
reported as evidence that the operator's homes stayed unchanged.
