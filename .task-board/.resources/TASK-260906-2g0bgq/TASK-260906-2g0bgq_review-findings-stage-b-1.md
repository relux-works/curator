# Review findings — stage (b) cycle 1, `feat/agent-environments-stage-b` @ `73174cc3`

Reviewer run `RUN-260906-ab4d71`. Subject: curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, 12 signed commits on `834b40f6`,
26 files, +6565/−40. Authority: curator-spec `f39f4a9` (verified: `git rev-parse HEAD` on the
main checkout, clean tree). Everything below was driven, not read; every claim carries the
command that produced it.

**repeat-of: none** (first cycle).

## Verdict: changes requested → `to-dev`

Three blocking findings and three major ones. The stage is close: the conformance subset is
byte-exact with zero skips, the fragment is well-formed for every adapter and both forms, the
marker records what §8.2 requires, and every gate except lint is green. The blockers are a
false refusal that makes `claude_code` unusable on a class of real machines, a drift hole that
lets a tampered MCP/system-prompt/root-context surface resolve as current, and a lint gate the
report attests green while it exits 1.

## On the empty repository delta

`CR-TASK-260906-2g0bgq-1` rev 1 carries `repository_delta=empty` and a zero-path patch. That is
**correct here and is not a finding**: the producer brief states the implementation lands in a
separate repository (`~/Developer/ReluxWorks/curator`, branch `feat/agent-environments-stage-b`)
and that "the story workspace carries an empty delta by design". The reviewable work is the
12-commit curator branch named by the review brief, and that is what this review drove. Worth
recording as a process note: the board's CR snapshot therefore holds no record of the 6565 lines
actually under review — the brief and this findings file are the only pointer.

---

## BLOCKING

### F1 — `claude_code` on macOS below the pinned release refuses *every* isolation value; no managed home can be provisioned

`internal/envregistry/envregistry.go:348-360` (`resolveIsolation`).

§7.4 splits the matrix on the pin:
- below 2.1.261 on macOS — `isolated` is `environment_isolated_unsupported`, and the default
  stays the document-wide default, `shared`;
- at or above — `isolated` is the platform default and a configured `shared` is
  `environment_shared_unsupported` ("For `claude_code` on macOS **at or above** 2.1.261 the
  evidence is positive and the restriction is lifted... The same fact removes `shared`").

The implementation applies the `shared` refusal on darwin unconditionally, so when
`atOrAbovePinned` is false there is no admissible value at all:

```
default below pinned -> mode="" err=environment_isolated_unsupported: isolated is unsupported for claude_code on macOS below 2.1.261
shared  below pinned -> mode="" err=environment_shared_unsupported: shared is unsupported for claude_code on macOS at or above 2.1.261: there is no Keychain item a manager could link
```

Two things are wrong: the unconfigured default is refused with a diagnostic about a mode the
operator never chose, and the `shared` message asserts "at or above 2.1.261" in exactly the
state where the release is *below* it.

**Reachability.** `EffectiveIsolation` has one production call site,
`internal/envprofile/managed.go:269-272` in `assembleHome`, which every `Resolve` and every
`StatusOf` row goes through. `cmd/curator/env.go:52,84` passes `envregistry.DefaultMachineConfig()`,
so `configured` is always `""` in production. An operator on macOS with claude 2.1.200 installed
gets `environment_isolated_unsupported` from `curator env resolve claude_code` and from
`curator env status`, with no configuration that can unblock it.

**Failure scenario.** macOS, `claude --version` reports `2.1.200`, no machine configuration.
`curator env resolve claude_code --repair` fails; the managed home is never created.

**What no test drives.** `TestIsolationMatrix`
(`internal/envregistry/envregistry_test.go:88-125`) drives `("darwin","",true)`,
`("darwin",shared,true)` and `("darwin",isolated,false)`. It never drives `("darwin","",false)`
or `("darwin",shared,false)` — the two cases that are wrong.

**Evidence.** Probe `TestProbeClaudeDarwinBelowPinned` against the production `resolveIsolation`
(`.temp/review/probes/zz_probe_test.go`); output above.

**Fix.** Gate the `shared` refusal on `atOrAbovePinned`; below the pin fall through to the
`shared` default and keep only the `isolated` refusal. Add both missing matrix rows to
`TestIsolationMatrix`, and add a narrowing mutant (refuse `shared` on darwin regardless of the
pin) that the new rows kill.

---

### F2 — byte drift of a linked manager-authored surface is invisible; `env resolve` emits a fragment for a tampered home

`internal/envprofile/managed.go:1128-1148` (`verification.checkSurfaces`, the `linked` branch).

§8.4 is explicit: "For `linked` surfaces, drift is a link that no longer targets the expected
store path **or a target whose bytes fail the recorded hash**." §10.1 grants the no-rehash
exemption only to "a symlinked surface whose link targets an entry of **the immutable profile
store** (the store entry's integrity is the store's own invariant, section 4)".

Manager-authored surfaces are **not** store entries. `storeDocPath`
(`internal/envprofile/managed.go:206-208`) writes them to
`<manager>/profiles/<profile>/rendered/<env>/<rel>` — a mutable path derived from
(profile, env, relative path), with no content addressing and no store invariant. The surfaces
linked there are:

| surface | adapters |
|---|---|
| root context (`AGENTS.md`) | `codex_cli`, `opencode`, `pi` |
| `opencode.json` (referenced form) | `opencode` |
| `.agent-context/system-prompt.md` | all four |
| `.agent-context/mcp/<adapter>.json`, `curator-mcp.config.toml` | `claude_code`, `codex_cli`, `opencode` |

`checkSurfaces` verifies these with `os.Readlink(full) == target` and nothing else. Writing
*through* the intact link changes the target's bytes, leaves the link identical, and the home
verifies as current.

**Evidence.** Probe `TestProbeDriftThroughLink` / `TestProbeDriftSystemPromptAndMCP`
(`.temp/review/probes/zz_drift_probe_test.go`), driving production `Resolve`:

```
codex_cli: .../environments/acme/codex_cli/AGENTS.md -> .../profiles/acme/rendered/codex_cli/AGENTS.md
SPEC §8.4: byte drift through an intact link is NOT detected; resolve emitted a fragment of 937 bytes
opencode:  ... NOT detected; fragment 869 bytes
pi:        ... NOT detected; fragment 773 bytes
.agent-context/system-prompt.md    byte drift through an intact link NOT detected (fragment 1080 bytes)
.agent-context/mcp/claude_code.json byte drift through an intact link NOT detected (fragment 1080 bytes)
```

**Failure scenario, security-relevant.** `buildFragment` sets
`mcp.path = <home>/.agent-context/mcp/claude_code.json` (`managed.go:1413`) and the launcher
passes it to `--mcp-config`. Appending a server object to that file — either directly through
the home symlink or at the rendered path — changes which processes the agent spawns.
`curator env resolve claude_code` reports the home current and hands the launcher the tampered
path. The same holds for the system-prompt file, whose bytes are the agent's instructions.
Fail-closed is the stated purpose of the check: "so that a launcher never runs an agent in a
home the manager knows is wrong without saying so" (§10.1).

**What no test drives.** `TestResolveDriftRepair`
(`internal/envprofile/managed_test.go:230-274`) covers exactly two shapes: the copied
`CLAUDE.md` verified by hash, and a link **replaced** by a regular file. Neither touches the
§8.4 clause "or a target whose bytes fail the recorded hash". No test writes through a live
link.

**Fix.** Verify the recorded hash of the link target for every surface whose target is a
rendered document (i.e. everything the plan publishes through `p.docs`), keeping the
link-identity-only fast path for targets under `contextstore.Root` — skills trees and
referenced module files, which genuinely are immutable store entries. Add the write-through-link
case to `TestResolveDriftRepair` for every adapter and every managed surface, plus a narrowing
mutant that keeps the link check and drops only the byte check.

---

### F3 — the lint gate is red on the head under review; the drafting report attests it green

`internal/envfragment/fragment_schema_test.go:132` — a file this stage adds (+347 lines, commit
`aa0bfbb2`).

```
$ golangci-lint run ./...            # v2.12.2, the version .github/workflows/ci.yml:222 pins
internal/envfragment/fragment_schema_test.go:132:6: QF1001: could apply De Morgan's law (staticcheck)
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f') {
		   ^
1 issues:
* staticcheck: 1
REAL-EXIT=1
```

Run in the producer's own worktree, with the repo's `.golangci.yml`. The config's only
`_test.go` exclusion is for `gosec`; staticcheck applies to test files.

`TASK-260906-2g0bgq_drafting-report.md` states: "`golangci-lint run ./...` (v2.12.2, the CI pin)
→ exit 0, 0 issues. Includes 4 mechanical fixes in `internal/interop/context_resolution_test.go`
(unused params, De Morgan) that predated this stage and kept the gate red."

So the exact linter and the exact rule were known and fixed once in commit `1db3f08a`, and a new
instance was introduced afterwards in `aa0bfbb2` and re-attested green without a re-run. This is
a capability claim that does not reproduce; it also means the hosted lint job fails on this head.

**Fix.** Fix the finding, re-run the gate, and correct the report line with the re-run's output.

---

## MAJOR

### F4 — a narrowing mutant of the §10.3 root-containment gate survives the entire suite

`internal/envfragment/envfragment.go:257-260` (`CheckBoundary.checkPath`).

The gate stays present and is weakened to admit exactly the sibling class — everything one level
above the environments root:

```go
- root := filepath.Clean(envRoot)
+ root := filepath.Dir(filepath.Clean(envRoot))
```

```
$ go test -count=1 ./internal/envfragment/ ./internal/envprofile/ ./cmd/curator/
ok  github.com/relux-works/curator/internal/envfragment  0.292s
ok  github.com/relux-works/curator/internal/envprofile   28.955s
ok  github.com/relux-works/curator/cmd/curator          280.337s
```

`TestCheckBoundary` passes under the mutant. The class the mutant admits is not hypothetical: it
is `<manager>/profiles/<profile>/rendered/...` — where every linked surface's real bytes live —
and `<manager>/store/...`. §10.3 says "Fragment values are absolute paths **below** the
manager-owned environments root."

The drafting report's mutant for this gate is "`CheckBoundary` root check removed", a delete-only
mutant. The AC row is explicit that a delete-only mutant "proves only that the gate exists and is
not accepted as evidence". Of the five entries in the report's mutant table, four are deletions
or constant-true rewrites of the whole predicate; only the `codexKeyring` entry is a genuine
token-preserving narrowing. The table's closing line, "Delete-only mutants were not used as
evidence", does not describe the table.

For contrast, I confirmed the reserved-name bound *does* have real narrowing evidence: weakening
`BoundEnvNames` to admit exactly `HOME` (`ReservedEnvName(name) && name != "HOME"`) is killed by
`TestBoundEnvNames`. And narrowing `SupportsReferenced` to additionally admit exactly `pi` is
killed by `internal/contextmaterialize`. So the pattern is achievable; it is missing on this gate.

**Fix.** Add a `CheckBoundary` case asserting that a value below the environments root's *parent*
is refused, and re-run the root-parent mutant to see it fail.

### F5 — the §8.1 always-copied `claude_code` root context is untested under the `referenced` form

`internal/envprofile/managed.go:344` (`homePlan.rootContext`, referenced branch).

§8.1 states the rule as the one per-surface exception that "holds in every mode and every home
and is not overridable by the registry or by machine configuration: the `claude_code`
root-context surface is always a copied regular file", and §5.3 repeats it "whatever the form",
because the tool's guard "skips a user-level `$CLAUDE_CONFIG_DIR/CLAUDE.md` that is itself a
symbolic link or a hard link".

Disabling exactly the referenced-form copy branch so `CLAUDE.md` links into the store instead:

```go
- if path == target && p.adapter.ID == envregistry.ClaudeCode {
+ if false && path == target && p.adapter.ID == envregistry.ClaudeCode {
```

```
$ go test -count=1 ./internal/envprofile/
ok  github.com/relux-works/curator/internal/envprofile  29.442s
```

The mutant survives. The production behaviour is correct — my probe `TestProbeReferencedMarker`
confirms `CLAUDE.md` is a regular file with the marker copy reason `claude-code-root-context`
under the referenced form — but the invariant §5.3 wrote the referenced form around has no guard.

**Fix.** Assert `os.Lstat(<home>/CLAUDE.md).Mode()&os.ModeSymlink == 0` and the recorded copy
reason under **both** forms, so the mutant above fails.

### F6 — the referenced-form staleness check tests project-entry presence, not the approval key

`internal/envprofile/managed.go:1286-1300` (`verification.checkClaudeProject`).

```go
if plan.form == envregistry.FormReferenced && !entries[req.LaunchDir] {
```

`claudeProjects` returns only the *keys* of `projects`. §5.3 states what actually gates the
content: an `@path` target outside the launch directory "is loaded only when the managed
`.claude.json` project entry for that launch directory sets
`hasClaudeMdExternalIncludesApproved: true`, otherwise it is **dropped silently**", and "the
`referenced` form for `claude_code` therefore requires the project entry of section 7.4
**carrying that key** for the launch directory, and a home lacking it for the launch directory is
stale".

`.claude.json` is tool-owned after provisioning (§7.4: "thereafter owned by the tool"), so the
tool can rewrite the entry and lose the key while keeping `projects.<dir>`.

**Evidence.** Probe `TestProbeReferencedApprovalKeyDropped`
(`.temp/review/probes/zz_claude_probe_test.go`) — provision referenced, delete only
`hasClaudeMdExternalIncludesApproved` from the launch directory's entry, then call production
`Resolve`:

```
SPEC §5.3: the referenced home has no external-includes approval for /var/.../007, yet resolve
reports it current and emits a 1095-byte fragment; the tool will drop every @-reference silently
```

**Failure scenario.** The agent launches with a `CLAUDE.md` whose only content is the generation
header and `@`-lines that the tool discards — an empty root context, no diagnostic anywhere.

**Fix.** Check the key, not the key's parent. Extend `claudeProjects` to report the flag and add
the negative case to `TestResolveClaudeProjectEntry`.

---

## MINOR

### F7 — §7.6 declares two secondary targets; the registry declares one, bound to no adapter
`internal/envregistry/envregistry.go:274-282`. §7.6: "Revision 1 declares exactly two", one row
for `claude_code` (home as `CLAUDE_CONFIG_DIR`) and one for `codex_cli` (same directory as
`CODEX_HOME`). `Target` carries no adapter field and `TargetByID` resolves globally, so
`xcode-coding-assistant` resolves for `pi` and `opencode`, which declare no targets.
`TestUnknownTarget` only drives an undeclared id, so the missing per-adapter binding is unguarded.

### F8 — an unreadable native `config.toml` is treated as the `file` credential store
`internal/envprofile/managed.go:527-534` (`codexKeyring`) returns `false` on any read error; the
doc comment on `effectivePassthrough` states it ("an unreadable one is treated as file with no
warning"). §7.4 makes `file` the value for *absence* of the setting; a failed read is a third
fact. The AC names this shape: "A failed, partial, or malformed read is never a legitimate
absence, and a fallback defined for absence must not fire on a read failure." Consequence: for a
keyring operator whose `config.toml` is unreadable, provisioning creates a symlink to an
`auth.json` that does not exist and the liveness row reports `environment_passthrough_detached`
permanently. Report the read failure or warn; do not fold it into the default.

### F9 — `env status` omits three §12 rows
§12 requires the report to carry "the lock's context members with weights and the precedence
primitives per activation" and "unregistered adapters found in machine configuration". Neither
appears in `HomeState`/`Status` (`internal/envprofile/status.go:29-90`) nor in either renderer,
so `--json` cannot carry them either. `Mode`, `Form`, `SeededProjects` are populated but the text
renderer never prints them (`cmd/curator/envstatus.go`); `--json` does. The brief's item 6
enumerates a narrower row list, so if this narrowing is intended, declare it as a stated bound in
the report rather than leaving it silent.

### F10 — `--format env|shell` emits variables the fragment's `env` object does not carry
`internal/envfragment/envfragment.go:143-169`. `variables()` appends every variable-kind channel
value after the `env` members, so an `opencode` `--format env` prints
`OPENCODE_CONFIG=<mcp path>` — a value §10.3 says "never enter[s] the fragment's `env`", emitted
by a command §10.2 says "activates nothing". The three formats stop being renderings of one
object. Separately the function takes `adapter` and discards it (`_ = adapter`, line 167) while
its doc comment and §10.1 promise "the adapter's declared variable order"; the order is
alphabetical over `Env`. Either honour the parameter or drop it and state the ordering rule.

### F11 — report and board accuracy
- The board note says "13 signed commits"; `git rev-list --count 834b40f6..73174cc3` is 12, as
  the review brief says. All 12 verify `G` with author `Ivan Oparin <oparin@me.com>`.
- The report names two stale unindexed fragment cases (`valid-no-composition`,
  `valid-local-state-pin`). Five valid-shaped files are unindexed: `valid-config-key-channel.json`,
  `valid-empty-channels.json` and `valid-file-channels.json` are equally stale (all five carry
  `profile.commit`/`state_sha256` and no `precedence`). The class judgement is right; the
  enumeration is incomplete, and an incomplete enumeration of what a driver skipped is the shape
  the AC asks reviewers to catch.
- `internal/envprofile/switch.go:170` still returns
  `environment_target_unknown: secondary fixed-home targets are stage (b)` for a target the
  registry *declares*. §7.7 reserves that diagnostic for an **undeclared** target, and the message
  is stale now that stage (b) is the current stage.
- `.github/ci/skip-classes.tsv:106` keeps the `stage-deferred` / `deferred to stage \(b\)` class
  although nothing in the tree emits it (verified: 0 occurrences in the gate's
  `skips-observed.tsv`). The brief asked to "leave the class only where a case is still deferred";
  as it stands it is a standing permission for a reason no test uses.
- `internal/envprofile/managed.go:57-64` (`ManagedParent`) has two identical branches.

---

## What I verified green, at `73174cc3`, on darwin/arm64, go1.25.5

Every command below was run by me as a standalone process; exit codes are real.

| gate | result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./...` (v2.12.2) | **exit 1, 1 issue — F3** |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed, exit 0 |
| `bash .github/ci/no-broad-suppression.sh` | ok |
| `bash .github/ci/test-gate.sh` (`CURATOR_CONFORMANCE_ROOT=f39f4a9`, `GO_TEST_TIMEOUT=30m`) | `go test exit=0, platform-case gate exit=0`; 19 skips, all in registered classes, **0 `stage-deferred`** |
| `bash .github/ci/ledger-consistency.sh <evidence>` | 126 rows checked across linux darwin windows, ok |
| `go test -count=1 -race` on contextmaterialize, envregistry, envfragment, envmarker, envprofile, interop | all ok (envprofile 37.5s) |
| `go test ./internal/interop` at the authority root | ok, 102 subtests, **zero skips** |
| `TestConformanceEnvironmentsMonolithic` | **19 of 19 pass**, including `referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`, `system-prompt-composed`, `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none`; the stage-deferred skip is gone |
| commits | 12, every one `G` (good signature), author `Ivan Oparin <oparin@me.com>`; scope confined to stage (b) |

Behaviour I drove and found correct:

- **Composed fragment**, production `Resolve --repair --format json`, all four adapters and both
  forms (probe `TestProbeFragmentComposedArtifact`): closed top-level member set;
  `profile` exactly `{name, lock_sha256}` with a bare 64-hex pin; `precedence` always present;
  `env` arity 1 carrying the registry variable with a POSIX-absolute value; `system_prompt.channels`
  byte-equal to the §7.3 registry descriptors; `mcp.channels` byte-equal to the §7.8 descriptor
  without `semantics`; `env_names` sorted; **no `mcp` section for `pi`**; `codex_cli` `mcp.path`
  is `<home>/curator-mcp.config.toml` with `{"argument":"name","flag":"-p","name":"curator-mcp"}`;
  `path_prepend` never emitted.
- **Marker** under the referenced form: `form: referenced` on `root-context` only, sorted surface
  keys, `copies: [{path: CLAUDE.md, reason: claude-code-root-context}]`, `passthrough`/`seeds`
  arrays present even when empty, `seeded_projects` sorted, `CLAUDE.md` a regular file.
- **Stale refusal**: an unprovisioned or drifted home returns `environment_home_stale` with the
  reasons and **no fragment** (`Resolve` returns before `buildFragment`).
- **Repair lock**: lock-free verification runs first and a current home emits without taking the
  lock; the wait is bounded at 30s (`internal/envprofile/lock.go:19`, inside §10.1's 1–60s band)
  and a contended repair fails with `environment_lock_unavailable`, distinct from
  `environment_repair_failed`.
- **Passthrough liveness**: a severed `file-link` (link replaced by a regular file) is detected —
  `os.Readlink` errors — and `--repair` re-links it.
- **Seeds**: `environment_seed_unreadable` stops provisioning before the first write; an absent
  seed is skipped, not reported; copy seeds are gathered only at provisioning so repair never
  refreshes tool-owned state.
- **Umbrella dispatch**: identifier-shaped unknown names reach discovery, non-identifier names
  stay a usage error, a missing provider names `curator-<name>` with installation guidance, and a
  provider inside the user-bin shim directory or below the environments root is refused with
  `subcommand_provider_untrusted`. Note §11 also names "a managed skill bin directory"; the
  implemented check covers the shim directory and the environments root only, and the §9.4
  `curator-*` reservation is enforced on the skill **member name**
  (`internal/envprofile/managed.go:207`) rather than on a published command name. §9.4 is outside
  this brief's authority list, so this is an observation, not a finding.
- **MCP bytes**: CCJ-1 for `claude_code`/`opencode`, the fixed TOML layer for `codex_cli`,
  `written=false` for `pi` and for an empty set; `env_names` grammar and the reserved-name
  exclusion are bounded twice — at package parse
  (`internal/contextpkg/contextpkg.go:485-495`) and again in `BoundEnvNames`.
- **Shadowing**: an unacknowledged declared shadowing path makes the `env status` row
  non-current; an acknowledged one downgrades to a warning
  (`internal/envprofile/status.go:221-233`).

## Probe artifacts

`.temp/review/probes/` in the story worktree (never written into the producer's worktree or the
control root; the producer's tree is clean at `73174cc3`):

- `zz_probe_test.go` — F1, drives `envregistry.resolveIsolation`
- `zz_drift_probe_test.go` — F2, drives `envprofile.Resolve`
- `zz_claude_probe_test.go` — F6 and the referenced-form marker check, drives `envprofile.Resolve`
- `zz_fragment_probe_test.go` — the composed-artifact check, drives `envprofile.Resolve`

Mutants were applied and reverted inside a throwaway `rsync` copy under `.temp/review/scratch`,
never in the producer's worktree.
