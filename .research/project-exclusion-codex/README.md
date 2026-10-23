# Codex 0.162.1 project-source exclusion probes

TASK-261021-boww1n. This research harness changes no product code and grants
no strict-context capability. Read the [qualification record](../../docs/qualification/project-source-exclusion-codex.md)
for P1-P8 verdicts and limits. A passing run records the retained-skill failure
witness as well as successful individual controls.

The only admitted vendor build is `codex-cli 0.162.1`, darwin/arm64, SHA-256
`74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0`.
Choose the exact standalone release, not the updating launcher. The harness
checks the digest before creating a fixture, copies those bytes into the run
root, then checks the version inside the sandbox. No auth command runs.

## Replay and tools

Go 1.25.5 runs the harness; macOS `sandbox-exec` confines every vendor process
and its children. `/usr/bin/env -i` supplies only synthetic HOME, CODEX_HOME,
four XDG roots, TMPDIR, system PATH, locale and isolated Git settings. The
sandbox denies all network, security-service lookups, Keychains-shaped paths,
the real native configuration directories and system Codex policy directories.
No real native file is opened to verify this boundary. All fixture writes and
copied executable bytes stay in the explicit task-local `.temp/` run root.

Git creates synthetic checkouts with separate metadata and a `.git` file.
`testdata/mcp.py` runs with `/usr/bin/python3` as a synthetic stdio MCP server;
it does not use network or authentication. It records its own PID in the
fixture. App-server runs as a foreground stdio child, never as a daemon. The
harness records process PIDs and waits for shutdown before finishing. `curl`
is used only for the synthetic loopback boundary check: a reachable endpoint
outside the sandbox must be unreachable inside it. No model turn runs.

```sh
# Default invocation: all vendor entries skip before fixture creation.
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-codex -count=1 -v

# Set CODEX_EXCL_BINARY privately to the exact standalone release.
# Each invocation requires a fresh run directory.
export CODEX_EXCL_RUN_ROOT="$PWD/.temp/project-exclusion-codex/replay-01"
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-codex -count=1 -v

# The runner succeeds only when every intended child failure is observed.
export CODEX_EXCL_MUTANTS_ROOT="$PWD/.temp/project-exclusion-codex/mutants-01"
sh .research/project-exclusion-codex/mutants.sh

env -u GIT_DIR -u GIT_WORK_TREE go vet ./.research/project-exclusion-codex
sh -n .research/project-exclusion-codex/mutants.sh
```

`CURATOR_GO`, when set, selects the Go executable for the mutant runner.
The caller can select task-local GOCACHE, GOMODCACHE and GOTMPDIR. The hidden
research directory is excluded from `go test ./...`; invoke it explicitly.
Do not interpret a skipped run as qualification. Platforms other than
macOS/arm64 are not qualified by this record.

`CODEX_EXCL_MUTANT` is reserved for four expected-red variants:
`drop-document-cap`, `drop-skill-disable`, `trust-project`, and
`allow-loopback`. Do not set it for an ordinary replay. The first three omit
one exclusion control without changing the matching expectation; the last
allows loopback and must fail the offline assertion. A setup error, skipped
plant, timeout or unrelated failure does not count as a killed control.

## Evidence and scope

Every command saves argv, cwd, PID, exit and output under
`$CODEX_EXCL_RUN_ROOT/<test>/evidence/`. API transcripts record each request,
response, notification and process wait result. `$RUN` and `$PRIVATE_HOME`
replace private locators in saved files. Raw runtime databases and synthetic
history stay ignored; they are never attached to the public board. The
committed `observations.json` is a bounded extraction of successful raw
observations, with input-file hashes, rather than a copied vendor prompt.

`TestIntakeBoundary` pairs qualified execution with missing opt-in, wrong
digest and invalid run-root refusals. `TestOfflineBoundary` pairs ordinary
reads and reachable loopback with specific denials. `TestInstructions` and
`TestSkills` require complete JSON prompt dumps before checking markers.
The API probes require matching JSON-RPC success responses, not model claims.

The fixtures are synthetic approved and unapproved copies. They do not drive
Curator admission, protected-store publication or hostile-code containment.
The resume source list fails its project-source positive baseline and cannot
prove project exclusion. Lazy reads, actual child agents, tainted persisted
model history, automatic in-turn reload, project execution-policy rules,
other plugin discovery mechanisms and TUI parity remain unqualified.


## Adapter rework regressions

`adapter_rework_test.go` calls production Render/Materialize/Verify APIs with a
manager lock and complete emitted argv. `TestAdapterTrustCapability` confirms
trusted native project prose/config plus the profile root; the untrusted-only
control suppresses project prose. Native config loading is permitted by the
owner's native-project ruling, not a permission-violation witness.

`TestAdapterTrustedPolicyPrecedence` pairs hostile repository config against
manager argv for four approval/sandbox combinations. Measured approval,
sandbox/write and memory settings hold, but repository MCP permission keys
survive. `TestAdapterRepositoryMCPPermissionEffect` proves the actual effect:
the same native MCP server initializes, but its tool disappears when only
repository `disabled_tools` changes; the unrestricted positive calls the tool.
The production permission gate must refuse, naming the surviving key.
These are startup/config/thread/catalog observations, not model approval-dialog,
model-turn, reload, resume or child-session qualification.

`TestAdapterNativeBundleLifecycle` covers empty/root/system-only profiles,
read-only/workspace-write policy, debug and app-server/thread startup, and
unchanged restart/reuse. The complete pinned native system bundle is published
under the ordinary receipt, so no live baseline is recaptured. The original
discovery probes above do not drive Curator admission; adapter regressions
exercise exported delivery APIs without granting launch authority.


`TestMCPCeilingParser` and `TestAdapterEnumeratedMCPReload` replay the owner's
read-only enumeration candidate without adding a product parser. Explicit
bare dotted overrides restore an admitted tool and prevent a known denied
server's process from starting. A late repository server added on the same
app-server still loads at the next thread; its `disabled_tools` affects the
actual catalog, and removing that restriction permits a synthetic call.
Malformed/unknown candidate inputs refuse. This bounded stdio parser is an
experimental fixture, not a complete native-config parser or launch authority.


## Config-change fence candidate (r3)

`config_fence_test.go` adds a bounded research preflight, not product admission
or complete native config discovery. `TestConfigFenceAdmission` pairs unchanged
thread/turn dispatch with late-server, disabled-tools, removed-server/file,
added/changed other config, linked file/parent and invalid/unknown-key refusals.
Each refusal is typed `project_config_changed`, requires relaunch and preserves
the baseline and input bytes without dispatch. `TestConfigFenceIntake` pairs
valid digests/dotted ceilings with invalid initial-input refusal.

`TestAdapterConfigFenceReadGap` proves why this cannot certify Permissions.
The unchanged native control calls the admitted MCP tool, and a quiescent late
server refuses before dispatch. Relaunch enumerates the late server, disables
it and still calls the admitted tool. On the original app-server, a deterministic
edit after a successful check but before `thread/start` loads an unenumerated
server. Its disabled-tools pair changes the actual catalog. Restoring original
bytes makes the next preflight pass while the unrestricted loaded tool still
returns its synthetic call marker. Every native thread start and the retained-state call pass the fence;
no model turn, network or credentials are used. Status listing re-reads config,
so the catalog is captured before restoration and the actual call follows it.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./.research/project-exclusion-codex -count=1 -v -run '^TestConfigFence(Admission|Intake)$'
# Set CODEX_EXCL_BINARY to the qualified frozen executable as above.
CODEX_EXCL_RUN_ROOT="$PWD/.temp/project-exclusion-codex/fence-01" \
  env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test ./.research/project-exclusion-codex -count=1 -v -run '^TestAdapterConfigFenceReadGap$'
```

Run through the authorized suite-slot wrapper. The adapter mutation lane adds
skipped-fence, filenames-only, linked-parent and refusal-path plants. The
production typed Permissions refusal remains; a preflight cannot atomically
freeze an independent native reader or erase already loaded runtime state.
