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
