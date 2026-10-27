# Native-home test boundary

Only test code imports this package. The envprofile and CLI `TestMain` entry
points use `Supervise` before their test-running branches. The outer process
records protected preimages, reexecutes the same binary, waits, and checks the
postimage regardless of the inner test process's normal return, panic, testing
timeout, direct `os.Exit`, or isolation setup/cleanup failure. Fixture CLI
helpers retain the native HOME and provider overrides explicitly supplied by
their parent; their supervisor still observes the OS-account homes.

A parent PID claim alone never admits the inner process. The snapshot owner
opens a bounded loopback listener and passes a fresh capability to its child.
Both sides authenticate an HMAC challenge bound to the actual parent and spawned
child PIDs, with distinct request/response roles and a five-second deadline.
Missing, stale or unauthenticated handoffs refuse before test dispatch. This
checks a live handoff, not code attestation against a malicious account owner.
On Unix, diagnostic SIGQUIT is forwarded to the observed child so its blocked
stack is printed while the supervisor survives to check the postimage.

`internal/testgitenv.Isolate` is the single owner of test-process Git and HOME
isolation. Every process calls it; explicit fixture helpers then restore only
the roots deliberately supplied by their parent, before resolver/test dispatch.
Git isolation remains active even when a fixture selects its own HOME. This package
neither allocates a HOME nor changes provider variables. `CheckIsolation` fails
before tests if the neutral helper does not move HOME or if provider/data/state/
cache overrides remain ambient. This integration requires the neutral helper
from PR #196 (commit `227625df`, helper blob `b67f9277`). Until that dependency
lands, validation uses that exact source as an overlay, physically present for
CLI tests whose nested builds create their own Go overlays. The outer runner
also removes all `GIT_CONFIG*` variables, including count, parameters and indexed
key/value injections, and clears provider/data/state/cache overrides.

The protected inventory includes the OS-account home and, for an ambient test
process, inherited native-home overrides. It observes root identities/modes,
root-context files, provider settings/configuration, environment markers and
backups, and entries recursively under regular skill directories. The only
skill exceptions are the runtime projector's exact
`skills/synced/<id>/manifest.json` and `.last-complete-round` coordination files.
Actual `SKILL.md` payloads, nested manifests and ordinary copied-skill manifests
remain protected, as does every entry in `.agent-environment-backup`.
Regular noncredential files have content hashes as well as lstat identity,
size, mtime and mode. Symlinks have only their own identity and link text;
linked roots, skill links and instruction links are never traversed. Known
credential carriers (`auth.json`, credential/token/secret names, OAuth files,
and private-key/certificate extensions) have lstat metadata only, never content
reads or hashes. Public diagnostics print fingerprints, not contents or paths.
For private diagnosis, `CURATOR_TEST_HOME_GUARD_REPORT` may name an existing
directory outside the protected roots. Each changed postimage creates an
exclusive mode-0600 JSONL file containing exact changed paths and before/after
lstat/link metadata, never content or hashes. Unchanged runs create no report.

Bounds fail closed: 65,536 snapshot entries, depth 64, 128 MiB per hashed file,
and 1 GiB of hashed bytes across the inventory. Large installed skill assets
are included within these bounds. Native root inventory is a positive list of
rendered settings/config, root-context and env-state files (see `files` in
`guard.go`), plus skills and env-state backups. Root-level runtime churn is
explicitly excluded: sessions, history, sqlite databases and WAL/SHM files,
models_cache, global-state, shell-snapshots, todos, statsig, backups, logs and
projects. These exclusions do not apply inside protected skill payloads or
env-state backups. The account-level `$HOME/.claude.json` is runtime state;
the rendered `.claude.json` inside a native home remains protected. Directory sizes and mtimes
are ignored; directory identity, mode and complete recursive skill membership
remain checked. This is an observation boundary, not OS filesystem confinement.

`Control` and `Child` exercise the actual package boundary with synthetic
original homes only. The settings and copied-skill decoys already exist before
the snapshot; their mutations preserve file length and restore mtime. They must
fail with the exact synthetic-root fingerprint. Context and environment-state
plants also fail. Panic and direct-exit plants must still emit the changed
postimage; clean panic/exit controls preserve the original exit codes and emit
an unchanged postimage. A forged parent claim, including one with an oversized
synthetic preimage, must refuse before reaching the child test. An oversized
preimage without a claim also refuses. An isolated native write and an explicitly
supplied CLI helper HOME are positive controls. No plant selects the operator's home.

Example validation while #196 is pending (the overlay maps only the neutral
helper to its exact pinned source):

```sh
runner_home=$(mktemp -d)
# Also unset every inherited GIT_CONFIG_KEY_*/GIT_CONFIG_VALUE_* variable.
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS \
  -u GIT_CONFIG_COUNT -u GIT_CONFIG_PARAMETERS \
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
metadata are not detected. Detached writers that mutate after the immediate
observed process/package exits are outside this observer's lifetime. Concurrent
legitimate edits to protected payloads can produce a refusal
without proving which process made them. A failed or missing postimage is never
reported as evidence that the operator's homes stayed unchanged.
