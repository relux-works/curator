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
Missing, stale or unauthenticated handoffs refuse before test dispatch. The
observer consumes authentication failure concurrently with process exit; the
five-second deadline covers both accept and handshake. On failure it terminates
its owned Unix process group (Windows: private Job Object), reaps the child and
checks the final postimage before returning refusal. A never-connecting child
with a pipe-holding descendant is the bounded control; an early-exit variant
checks that an unauthenticated leader cannot leave its descendant behind. Both
require actual output EOF. This checks a live
handoff, not code attestation against a malicious account owner.
On Unix, diagnostic SIGQUIT is forwarded to the observed child so its blocked
stack is printed while the supervisor survives to check the postimage.

`internal/testgitenv.Isolate` is the single owner of test-process Git and HOME
isolation. Every process calls it; explicit fixture helpers then restore only
the roots deliberately supplied by their parent, before resolver/test dispatch.
Git isolation remains active even when a fixture selects its own HOME. This package
neither allocates a HOME nor changes provider variables. `CheckIsolation` fails
before tests if the neutral helper does not move HOME or if provider/data/state/
cache overrides remain ambient. This integration requires the neutral helper
from PR #196 (commit `6dc56ac6`, helper blob `d830b041`), adopted byte-for-byte
in this branch. The outer runner removes every inherited `GIT_*` selector and
`GIT_CONFIG*` variable, including count, parameters and indexed key/value
injections, and clears provider/data/state/cache overrides. Hostile controls
then inject only synthetic Git selectors and verify the decoy is unchanged.

The protected inventory includes the OS-account home and, for an ambient test
process, inherited native-home overrides. It observes root identities/modes,
root-context files, provider settings/configuration, environment markers and
backups, entries recursively under regular skill directories, and the rendered
`.agent-context` tree, including `modules`. The context container is observed
before descent, so linked containers are never followed.

The only skill exceptions are regular coordination leaves in the native runtime
layout: `skills/synced/<uuid>_<uuid>/manifest.json` and `.last-complete-round`.
The `synced` and bucket entries must themselves be directories, and a regular
`skills/synced/.bucket-<uuid>_<uuid>` registration must exist. Registration and
directory identities stay protected. These are observed layout checks, not
authentication of runtime provenance. Unknown/unregistered layouts stay fully
protected. A `skills/synced/SKILL.md` entry identifies a copied skill and disables
the exception altogether, including UUID-shaped payload collisions. A directory
or symlink at a coordination leaf is always protected; directories are traversed.
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
`guard.go`), plus skills, referenced context and env-state backups. Root-level
runtime churn is explicitly excluded: sessions, history, sqlite databases and WAL/SHM files,
models_cache, global-state, shell-snapshots, todos, statsig, backups, logs and
projects. These exclusions do not apply inside protected skill payloads or
env-state backups. The account-level `$HOME/.claude.json` is runtime state;
the rendered `.claude.json` inside a native home remains protected. Directory
sizes and mtimes are ignored; directory identity, mode and complete recursive skill membership
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

The regression pairs exercise the real skill projector/copier with a skill named
`synced` and both regular and directory `references/manifest.json` collisions.
A referenced-context pair uses the actual renderer and managed writer. Each
pair admits a write in the neutral helper's isolated HOME and refuses the same
write in a sentinel-checked synthetic original HOME, with exact-path metadata
in the private report. All guard and envprofile checks use an absolute fresh
runner HOME; the runner scrubs Git selectors, Git configuration, SSH agent,
GOFLAGS and provider/XDG aliases before invoking Go.

Residual limits: killing/crashing the supervisor, whole-process-group kills,
host loss, and the pre-TestMain fixture transport/worker/tripwire dispatches
cannot guarantee a final postimage. Killing a supervisor may also leave its
inner process alive. The authentication-timeout process group/job is not a
general sandbox: a deliberately detached Unix process can escape its group.
Other unsupported Go platforms fall back to killing only the immediate child;
macOS/Linux use process groups. On Windows the child starts suspended, enters
its private Job Object, and only then resumes, including package initialization.
Windows code has a cross-compile check here; runtime qualification belongs to
the Windows CI runner.
A mutation restored before the postimage, an unchanged symlink's destination
outside the inventory, and credential bytes changed while restoring all their
metadata are not detected. Detached writers that mutate after the immediate
observed process/package exits are outside this observer's lifetime. Concurrent
legitimate edits to protected payloads can produce a refusal
without proving which process made them. A failed or missing postimage is never
reported as evidence that the operator's homes stayed unchanged.
