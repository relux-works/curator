# Curator

Curator is an agent environment manager (AEM): a single tool that manages what an AI coding agent gets in a project. Skills and their transitive dependencies, executable commands, MCP server requirements, per-agent delivery, and the security gates around all of it. Declarative, reproducible, verifiable.

Curator is implemented in Go and follows the [Curator Specification](https://github.com/relux-works/curator-spec), an open protocol for skill packages, project manifests, installation semantics, and the audit registry; sections are cited across this repository as `Spec §N.M`.

## Status

v0.1 development complete: all twelve phases of [docs/implementation-plan.md](docs/implementation-plan.md) are done. CI consumes the authoritative schemas and conformance vectors from `curator-spec` on ubuntu, macos, and windows, plus lint and the naming gate. Work is tracked on the in-repo task board under [.task-board/](.task-board/).

## Install

```bash
# Homebrew (macOS, Linux)
brew install relux-works/tap/curator

# installer script (macOS, Linux)
curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh

# Scoop (Windows)
scoop bucket add relux-works https://github.com/relux-works/scoop-bucket
scoop install curator

# Go toolchain
go install github.com/relux-works/curator/cmd/curator@latest
```

Debian and RPM packages ship with every [release](https://github.com/relux-works/curator/releases), together with SBOMs and cosign signatures. macOS binaries are Developer ID signed (Relux Works, LLC). Verify any downloaded artifact:

```bash
gh attestation verify <artifact> --owner relux-works
```

## What Curator manages

- **Skill packages**: `SKILL.md` plus context directories, with an implementation-neutral machine manifest (`agent-skill.json`, schemas 1 through 5) declaring commands, runtime layout, capabilities, and dependencies. The legacy `csk-skill.json` filename remains readable.
- **Project manifests**: `Skillfile.json` with exact git references; non-committed development substitutions.
- **Resolution**: transitive dependency closures unified to one commit and one source identity per name, with activation modes.
- **Installation**: context and runtime separation, install markers with content hashes, a commit-keyed runtime store, command shims, managed per-agent adapters.
- **Scopes**: project, global, and hybrid (machine-stored, per-project activation).
- **MCP requirements**: read-only verification of declared MCP servers against agent configuration surfaces.
- **Security**: source allowlists, declared capabilities, no code execution at install time, and an audit registry client (Ed25519 signed records, deny-wins federation, snapshot verification).
- **Operator credentials**: per-repository SSH selection and scoped HTTPS token sources for external build repositories, matched by canonical source identity and never selectable by a package. Private HTTPS fetches use a manager-owned, host-pinned askpass broker; public HTTPS can remain anonymous. See [SSH credentials](docs/build-ssh.md) and [HTTPS credentials](docs/build-https.md).

## Registry client guarantees

Curator binds persisted rollback and equivocation state to the canonical
registry URL, so signing-key rotation never resets the highest accepted
snapshot. This durable state lives under the Curator home `state/registry`
directory, outside the disposable `cache/registry` responses; upgrades migrate
legacy state without lowering it, and corruption or write failure is
fail-closed. A protected catalog distinguishes first use from deletion of a
previously accepted registry state. Record pagination rejects repeated or oversized cursors, more than
10,000 records per artifact query, and responses larger than 16 MiB.

Registry requests use bounded per-attempt and total deadlines. GET requests
retry network failures, `429`, and `503` at most twice after the first attempt.
Publication retries the exact body only with its deterministic
`Idempotency-Key`; other client errors and unsafe requests are never retried.
Redirects are rejected so a registry cannot move a request or bearer token to
another endpoint.

## Commands without profile setup

Shell profile changes are not required. After `curator install`, agents can
invoke project commands through `.agents/bin/<command>` on Unix and
`.agents\bin\<command>.cmd` on Windows. Global installation publishes
non-destructive forwarding shims to a safe user directory already on `PATH`
when one is available; otherwise Curator reports the canonical global bin
location. Installed launchers carry the project command directory and resolved
system dependency directories themselves, preserve the inherited `PATH`, and
return the child command status.

Repository bootstrap can remain idempotent and non-interactive:

```bash
curator bootstrap --if-missing --non-interactive --skills-root "$HOME/src/skills"
curator upgrade .
```

`upgrade` fetches only the selected project's direct and transitive dependency
closure. `upgrade --dry-run` plans with temporary sources and snapshots without
changing source checkouts, caches, security state, runtime state, or project
artifacts.

Interactive users who want bare command names and automatic project switching
can cache the optional hook once:

```bash
curator shell-init --install
# Add the source command printed above to .zshrc or .bashrc.
```

Automatic detection selects zsh or bash from `SHELL`, preserves Git Bash on
Windows, and otherwise selects PowerShell on Windows. The cached hook does not
start Curator during later shell launches. Curator never edits a profile
automatically. Set `CURATOR_AUTO_ENV=0` to retain global activation while
disabling project-directory scans.

## Compiled-command status, diagnostics, and repair

`curator status` is read-only. It reports one code per declared skill and, when
the closure activates compiled (`go-v1`) commands, one diagnostic line per
active build command. `status --json` carries the same values in a `builds`
array; a closure without compiled commands produces the historical document
unchanged, with no `builds` key at all.

The codes are stable and machine-readable. Only `up-to-date` (skills) and
`current` (compiled commands) mean "exactly current"; `status --check` exits
non-zero for every other value, including a state it does not recognize.

Reporting and verdict are separate. Plain `status` exits zero when it produced
the report it was asked for — including when the plan itself refused, as long as
every active compiled command still got a row (any raw detail is a `warning:` on
standard error). A refusal that leaves some command undescribed still exits
non-zero, as does any failure in a scope without compiled commands. `--check` is
the surface that turns a non-current verdict into a non-zero exit.

| Code | Meaning |
|---|---|
| `up-to-date` / `current` | every check passed |
| `not-installed` | the declaration has no installed skill |
| `invalid-marker` | the install marker is absent, unreadable, or invalid |
| `unsupported-marker` | the marker schema cannot be read by this manager |
| `needs-install` | the installation is behind its declaration, or its marker schema cannot describe a compiled command |
| `content-drift` | installed content no longer hashes to the recorded value |
| `unresolvable` | the declared ref cannot be resolved in the source repository |
| `build-context-exposed` | a build root reached agent-facing context |
| `build-command-drift` | the recorded and activated compiled command sets differ |
| `build-source-drift` | the recorded build-source identity no longer matches the raw snapshot |
| `build-input-drift` | the recorded logical key was derived from a different build input |
| `unsupported-build-driver` | a recorded or planned driver outside `go-v1` |
| `unusable-build-toolchain` | the trusted Go toolchain could not be resolved or verified, so nothing could be planned |
| `missing-build-artifact` | the protected cache holds no entry for the recorded key |
| `corrupt-build-receipt` | the entry's canonical receipt differs from the recorded one |
| `build-artifact-drift` | the entry's artifact path or hash differs from the recorded one |
| `corrupt-build-cache` | the protected entry cannot be interpreted |
| `untrusted-build-cache` | candidate bytes are outside a provable protected boundary |
| `unsupported-build-platform` | this host cannot prove protected build cache state |
| `build-state-changed` | the install marker or the protected cache evidence moved while status was classifying it |
| `unknown-build-state` | a planner outcome this manager does not know; it fails closed |

A row may carry a `cause`, a stable subcode that refines the state without
widening the state vocabulary. `unusable-build-toolchain` carries the `go-v1`
boundary code that refused the operation. `build-input-drift` carries one of:

| Cause | Meaning |
|---|---|
| `build-root` | the marker does not record the build root the closure now activates |
| `target` | the marker's recorded artifact path is not the one this target derives |
| `unattributed` | the key differs, and the marker records no prior input to attribute it |

The logical cache key is one opaque digest over the complete build input —
schema version, driver, build source, build root, command, source directory,
native target and tuning, trusted toolchain identity, and the fixed manager
build policy. An install marker records no prior input, so a key mismatch is
reported as input drift and attributed only as far as the marker's own recorded
build roots and artifact path can prove. Curator does not guess which input
moved.

A compiled verdict is bracketed on both sides. `status` fingerprints the
install markers before it plans and re-reads them afterwards, and it re-takes
the exact protected-cache lookup every row was classified from once
classification is done. Either half moving — a marker rewritten by a concurrent
install, or an entry removed, corrupted, replaced, or stripped of its provable
protection — reports `build-state-changed` instead of publishing a verdict that
was already stale.

Each diagnostic reports the driver, build root, source directory, build-source
identity, native target and tuning, logical cache key, manager-derived artifact
path, and the read-only cache outcome. A command the trusted toolchain refused
before any logical identity existed still reports everything that was already
established — driver, build root, source directory, and the validated
build-source digest — and reports no target, key, or cache outcome, because
none was derived. Every path in a diagnostic is
protocol-relative: manager home, cache, staging, and probe locations are never
published. Untrusted details — cache reasons and compiler output — are
collapsed onto one line, stripped of anything non-printable, path-redacted, and
length-bounded before they are printed or serialized.

`install --dry-run` and `upgrade --dry-run` run no compiler. Per active build
command they report `cache-hit`, `would-preflight-and-build`,
`would-rebuild-untrusted-cache`, `corrupt`, `unsupported`, or
`toolchain-unavailable` — a plan, never a completed compiler check.

There is no separate repair command: `install` and `upgrade` are the
reconciliation path. They rebuild a missing, corrupt, drifted, or untrusted
entry into new protected state, and only after every manifest, closure,
collision, requirement, audit, registry, and moved-tag gate has passed. An
unusable entry is quarantined and replaced under the manager-home lock, never
adopted by changing permissions or rewriting a marker. A failed gate,
preflight, build, or commit leaves the previous installation, its consumers,
and the live cache unchanged, and the run says so.

The cache is not a transaction target — a launcher can only point at an entry
that is already live — so a replacement is selected before the installation
that needs it is durable. A run that then fails puts the cache back before it
releases the manager-home lock: the replacement is withdrawn and the entry it
displaced is restored, both by renaming inside the protected cache root, so
nothing is deleted and the ordinary sweep collects the leftovers.

Selecting a replacement is itself several moves — quarantine the unusable
predecessor, rename the new entry into the freed slot, validate it, sync the
cache root — and a publication that fails part way through puts back what it
already moved before it reports. Each of those moves holds to the same rule on
its own: quarantining an entry is a rename plus the sync that makes it durable,
and a sync that fails returns the entry to the live slot rather than reporting a
failure with the slot already empty. That return is synced too, because it is a
move like any other; only if its own sync also fails does the run report the
cache as changed, with the entry still live and readable. So a failed
publication leaves the cache exactly as it found it, and the reversal above is
only ever needed for a publication that fully succeeded.

That reversal is refused in exactly one direction — when an incomplete
transaction still references the published entry, or when a journaled target is
no longer at the state the run found it. Restoring an unusable predecessor over
an entry a recovered commit is about to point at would turn a reported failure
into a broken installation, so the rebuilt entry is kept instead.

A restoration that cannot complete is reported rather than assumed. Whenever a
run leaves the live cache changed — a kept entry, a publication that could not
undo its own moves, or a reversal that stopped part way — it says so per command
instead of repeating the ordinary "the live build cache is unchanged" claim, and
the warning names which of the three it was. The installation and its consumers
are unchanged either way; nothing on these paths is ever deleted, so the state
left behind is always one a later `install` or `upgrade` repairs.

Two states a rebuild cannot resolve fail closed instead of being repaired: a
host that cannot prove protected cache state at all (`unsupported`), and a
trusted Go toolchain that cannot be resolved or verified
(`toolchain-unavailable`). Both refuse before any mutation.

`curator global status` reports the same thing for the machine-wide scope, in
the same stable vocabulary: one code per declared skill and one diagnostic line
per active compiled command. It accepts `--check` and `--json`, and both mean
exactly what they mean for `curator status` — `--json` carries the same values
in a `builds` array, and `--check` exits non-zero for every code that is not
`up-to-date` or `current`.

Deriving compiled currentness needs the current logical build input, which only
a plan produces, so `global status` runs the same read-only plan
`curator global install --dry-run` runs. That resolves the machine-wide closure
and passes the read-only audit and registry gates. It runs no compiler and
writes no installation target, cache entry, or trust state.

Two things differ from the project scope, both deliberate:

- The machine-readable document carries `alias`, `skills`, and — only when the
  closure activates compiled commands — `builds`. It carries no `path`: the
  scope has no operator-supplied root, and the manager home is never published.
- Plain `global status` keeps its historical contract of always reporting and
  always exiting zero. The declared-skill report is read straight from install
  markers and never from the plan, so a scope without compiled commands prints
  the lines it always printed even when the plan refuses; the refusal is a
  `warning:` on standard error. `--check` is the only surface that turns a
  verdict into a non-zero exit, and it fails closed twice over: once for every
  non-current code, and once when the plan refused before it could describe
  every compiled command, because such a run cannot prove the scope is current.

A machine-wide scope with no `Skillfile.json` declares and activates nothing:
it prints nothing and passes `--check`.

Curator selects the trusted Go installation only through `CURATOR_GO`
(an absolute `<GOROOT>/bin/go`, `bin/go.exe` on Windows) or `GOROOT`. It never
searches `PATH` and never downloads a toolchain, and it accepts only release
families it has tested against the `go-v1` vectors. A missing, untrusted, or
untested toolchain reports the failing boundary together with those
mechanisms and the tested families.

## Maintenance and the build-cache grace period

`curator gc` runs one serialized maintenance pass. It acquires the exclusive
manager-home mutation lock, recovers any incomplete install transaction, and
only then marks and sweeps, so it cannot race an install, a rollback, or a
recovery, and cannot lose a consumer registry update. The same pass runs at the
end of every installation, under the lock that installation already holds.

Marking reads the live project, global, and hybrid scopes once. Runtime store
entries are marked from every supported install marker schema; compiled build
cache entries are marked from marker v2 build state and from every in-flight
transaction journal.

Anything the pass cannot prove keeps its artifacts, and keeps them across
passes. A consumer registry that exists but does not match the exact shape
Curator writes is reported and left untouched rather than rewritten; a
registered checkout is unregistered only once its scope is proven absent or
proven valid and empty. An *ambiguous* registry counts as unreadable: a document
that states `schema_version` or `consumers` more than once does not say what it
means, so it is refused by every reader and writer rather than resolved to
whichever value happens to come last. A skills directory or installed skill that
is a symbolic link or a reparse point is refused instead of followed, and any
marker that exists but cannot be read or validated blocks the build sweep. A
later pass therefore sees the same uncertainty and makes the same refusal,
instead of inheriting a registry the earlier pass had quietly emptied.

The sweep removes a protected build cache entry only when all of the following
hold: no marker and no journal references it, the cache root and the entry are
still verifiable manager-protected state, the entry is structurally exact and
self-consistent with the logical key its directory encodes, and it was
published more than **24 hours** ago — Curator's documented grace period.
Everything else is retained and reported as a maintenance warning, including
corrupt receipts, untrusted roots, symlink or reparse escapes, and ownership or
DACL failures. Entry content is never executed, adopted, or permission-repaired,
and a receipt alone is never treated as proof of provenance or of a live
consumer. Retaining an unreferenced entry is always safe: the only cost of a
removal is one rebuild.

Every decision and every removal is bound to the directory object the pass
proved, not to the pathname it proved it under. The decisive classification of a
candidate — its exact members, its receipt, its artifact bytes and size — is
read through the descriptor of the entry the pass resolved, and the rename that
retires it and the deletion behind it resolve through the proven cache root; an
entry whose parent is no longer that object is retained and reported. Exchanging
the cache-root path after validation can therefore neither redirect a removal
outside the Curator cache root nor let a planted replacement supply the verdict
for the entry that is actually being removed.

## An open protocol

The specification is an open protocol, not an internal contract: any manager
built from it interoperates with the same skills, the same project manifests,
and the same audit registries. That matters when internal security policies
rule out adopting an external binary and require an in-house implementation
instead. One such independent implementation of the protocol is
[cocoaskills](https://github.com/ivanopcode/cocoaskills) (Python); Curator's
conformance against the shared wire formats is enforced directly from the
versioned protocol suite in CI; this repository carries no private copy of the
expected protocol values.

The registry-service profile is implemented by
[Curator Skill Registry](https://github.com/relux-works/curator-skill-registry),
which serves signed audit and revocation records plus a verifiable transparency
log for any conforming Curator manager.

## Development

The repository uses an in-repo task board (`.task-board/`, epics, stories, and tasks as files) and the agent tooling connected under `agents/`. Go testing follows the closed-loop tooling of `skill-go-testing-tools` (including `tuitestkit` for terminal UI phases).

### Gates and tooling

Every gate below is a script under `.github/ci/`, called directly by
[.github/workflows/ci.yml](.github/workflows/ci.yml) and mirrored by a `make`
target for local use. CI calls the scripts rather than `make` because `make` is
not a guaranteed tool on the Windows runner. Each gate writes its raw stream and
its report under `EVIDENCE` (default `.temp/ci-evidence/`), which CI uploads per
runner, so any claim about a gate can be checked against the run that produced
it.

| Tool | What it gates | How to run it | Where its output goes |
| --- | --- | --- | --- |
| `test-gate.sh` | plans the run from the supplied conformance root, executes it, then enforces the platform-case ledger; every status is fatal | `make ci-test` | `$EVIDENCE/test/` — `go-test*.json`, `suite-plan.txt`, `platform-cases.txt`, `skips-observed.tsv` |
| `test-gate.sh` (`-race`) | the same gate under the race detector | `make race` | `$EVIDENCE/race/` |
| `suite-plan.sh` | decides, from the root alone, which packages it serves, which it cannot, and which this platform does not qualify | called by `test-gate.sh` | `$EVIDENCE/*/suite-plan.txt`, `plan-*.txt` |
| `platform-case-gate.sh` | requires every case [`platform-cases.tsv`](.github/ci/platform-cases.tsv) names on this runner, and classifies every skip against [`skip-classes.tsv`](.github/ci/skip-classes.tsv) | called by `test-gate.sh` | `$EVIDENCE/*/platform-cases.txt`, `skips-observed.tsv` |
| `ledger-consistency.sh` | proves each ledger row against the real per-GOOS builds via `go list` — no runner needed | `make ledger-check` | `$EVIDENCE/ledger/ledger-consistency.txt` |
| `excluded-packages.sh` | the one resolver of "which packages does this platform not execute, and on whose authority" | called by the two gates above | stdout (TSV) |
| `candidate-suite.sh` | rejects a non-immutable candidate revision; records a candidate root's identity as candidate-only evidence | `make candidate-verify-ref CANDIDATE_REF=…` / `make candidate-record CANDIDATE_ROOT=…` | `$EVIDENCE/candidate/candidate-suite-identity.txt` |
| `toolchain-identity.sh` | asserts the resolved Go toolchain is exactly `go.mod`'s, with `GOTOOLCHAIN=local` and `GOENV=off` read back | run by every Go-consuming job | job log |
| `no-broad-suppression.sh` | rejects bare `//nolint`, bare `//#nosec`, production-path lint exclusions, wholesale disabling, and unrecorded `gosec` exclusions | `make no-broad-suppression` | job log |
| `gate-selftest.sh` | drives every gate above against synthetic inputs and asserts a real exit code for each negative case | `make gate-selftest` | job log |

`make ci-test`, `make race` and `make check-ci` require `CURATOR_CONFORMANCE_ROOT`
to point at a materialised `<curator-spec>/conformance/v1`; they refuse to run
without it, because a gate that runs with the conformance suite unset is a
smaller gate wearing the same name.

#### The compiled-build platform carve-out

`rc5-native-control-inventory-v1` defines native control records for exactly
macOS and Windows, so the go-v1 driver refuses a compiled build on any other
host before a worker starts. That carve-out reaches the suite at two
granularities, and neither is a silent omission:

* **whole package** — `internal/godriver` is not executed where the supplied
  root's own qualification vector marks the platform `excluded` (linux, with
  `until_task: TASK-260728-1skseh`). `test-gate.sh` still runs
  `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` on that very runner, so
  the exclusion is asserted rather than obeyed;
* **individual case** — the `cmd/curator` cases that need a *completed*
  compilation are carved out by `requireNativeControlInventoryPlatform`, which
  reads `godriver.InventoryPlatform` rather than a GOOS list, so it cannot drift
  from the inventory. Their skip reason names the inventory and is classified
  `platform-control` by `skip-classes.tsv`, and
  `TestCompiledInstallFollowsTheNativeControlInventoryExactly` runs on every
  runner to prove the boundary from whichever side that runner is on: a covered
  host installs and publishes exactly one protected cache entry, an uncovered
  host is refused with `build_execution_control_unavailable`, publishes nothing,
  and fails `status --check`.

When the inventory gains a record for a platform, the guard stops skipping there
on its own; only `must_run_on` in `platform-cases.tsv` needs widening.

The committed protocol-suite pin is declared once, as `SPEC_PIN` in the workflow
`env:` block, and every job reads it from there. A schema v6 candidate suite is
never committed and never a default: it enters only through the
`candidate-conformance` job, on an explicit `workflow_dispatch` that supplies a
full 40-character revision or a pre-materialised root. That job sets
`CI_REQUIRE_FULL_ROOT=1`, so a candidate must serve the whole package set, and
everything it emits is stamped in the artifact itself as candidate-only evidence
— neither a published release nor a conformance claim.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the working agreements: board-first workflow, discrete signed commits, spec-first rule.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
