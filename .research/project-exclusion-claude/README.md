# Claude Code 2.1.305 project-source exclusion probes

TASK-261021-wt8ve7 measures the exact macOS arm64 native binary with SHA-256
`ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f`.
The [qualification record](../../docs/qualification/project-source-exclusion-claude.md)
separates bounded observations from strict-adapter admission. This package
changes no product code, registry or protocol requirement.

## Tools and replay

Go 1.25.5 drives the vendor parser, real native tools and a deterministic
loopback Anthropic-compatible SSE endpoint. Python 3 supplies a synthetic
stdio MCP server and projects sanitized evidence. Git supplies disposable
repository boundaries with separate metadata directories. Apple
`sandbox-exec` confines every vendor invocation. A POSIX shell runs the
expected-red controls. No package installation, login, auth command, real
credential, external inference, tmux or interactive terminal is needed.
The native Task tool is exercised with fixed local responses, including its
child request. This is not a delegated development worker or model judgment.

Privately copy the exact versioned binary into the task's `.temp/` directory;
never use the updating launcher. Set its locator as `CLAUDE_EXCL_BINARY`.
Without that variable every vendor leaf skips before reading a binary or
creating a fixture. A different digest refuses before fixture creation.
The run root must be a fresh absolute path containing a `.temp/` component.
Fixtures and private raw diagnostics remain there for inspection.

```sh
# Ungated: every vendor leaf skips. Ordinary ./... excludes hidden research.
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-claude -count=1 -v

mkdir -p .temp/claude-exclusion/tmp
export GOCACHE="$PWD/.temp/claude-exclusion/go-cache"
export TMPDIR="$PWD/.temp/claude-exclusion/tmp"
# Set CLAUDE_EXCL_BINARY privately to the frozen executable copy.
export CLAUDE_EXCL_RUN_ROOT="$PWD/.temp/claude-exclusion/replay-new"
env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-claude -count=1 -v -timeout=4m

# Separate fresh root; runner exits 0 only when each suite fails as intended.
export CLAUDE_EXCL_RUN_ROOT="$PWD/.temp/claude-exclusion/mutants-new"
sh .research/project-exclusion-claude/mutants.sh
sh -n .research/project-exclusion-claude/mutants.sh

env -u GIT_DIR -u GIT_WORK_TREE go build ./...
env -u GIT_DIR -u GIT_WORK_TREE go vet ./... ./.research/project-exclusion-claude

# Project a completed run into ignored scratch before reviewing publication.
python3 .research/project-exclusion-claude/collect.py \
  .temp/claude-exclusion/replay-new .temp/claude-exclusion/observations.json
```

Use the already configured `CURATOR_GO` toolchain for this checkout; the
control runner selects it when set. No inherited `GIT_DIR` or `GIT_WORK_TREE`
reaches a Go test or the disposable Git fixture. A full default-suite replay
needs a normal disposable validation clone, its exact handed-over head and
a short temporary-directory path. The task outcome records that run separately.

## Boundary and fixtures

Each vendor command is `sandbox-exec -> env -i -> pinned copy`. Its explicit
allowlist sets disposable HOME, CLAUDE_CONFIG_DIR, all four XDG parents,
TMPDIR and CLAUDE_CODE_TMPDIR, a system-only PATH, synthetic API credentials
and the fixture's exact loopback provider URL. Auto-update, nonessential
traffic and automatic memory are disabled. The sandbox denies all network
except that provider port, every write outside its fixture, Keychain files
and security-service lookups, real native configuration/library directories,
system managed-settings locations, and discovery files above the synthetic
ancestor. No native directory is read to establish a denial.

The harness creates all synthetic project and home files explicitly in
`replay_test.go`. `testdata/mcp.py` implements only initialize, tools/list
and an inert tools/call response. Hooks can write only task-local marker
files. Plugins use synthetic local marketplace metadata; no remote marketplace
is queried. Test comments state the actual claim and relevant bounds.

The default exclusion control is `--setting-sources user`. P5 independently
crosses it with `--strict-mcp-config` and an explicit managed-home MCP file,
including an empty one. The approved channels remain normal home instructions,
skills, settings, hooks, plugins and agents, plus that explicit MCP file.
Neither `--bare` nor `--safe-mode` is used as a selective exclusion recipe.

## Evidence and interpretation

`observations.json` preserves actual request marker presence, bounded request
and tool-result excerpts, init catalogs, requested model, exits, loader lines
and hook/MCP file effects. The collector never treats an empty directory as
evidence and refuses remaining ordinary or flattened private home locators.
Full sanitized request/debug records remain in the private task-local run root;
`evidence/` contains the public run and narrowing summaries. Synthetic session
IDs bind first/resume records. Generated native task-path encodings are redacted
as well as ordinary absolute paths.

A successful SSE turn says nothing about model obedience: responses are fixed.
Assertions inspect the real input and native tool effects. Missing requests,
missing init/result, errors and timeouts fail instead of proving absence.
P7 retains an unsafe-origin resume witness: exclusion does not scrub previously
loaded project text from history. P8's late instruction and skill probes are
discriminating. Its 800ms late-hook observation is not: the new project hook
never executes even without exclusion; only a startup-approved hook runs.
That row does not qualify arbitrary settings/plugin/MCP reload behavior.

## Narrowing controls

`CLAUDE_EXCL_OMIT=setting-sources` drops the source flag but leaves the
root exclusion assertion intact: the project marker must cause a failure.
`CLAUDE_EXCL_OMIT=strict-mcp` drops only strict MCP in the independent
all-settings-source row: the foreign server must start and fail the row.
`CLAUDE_EXCL_BOUNDARY_MUTANT=network` and `keychains` weaken only synthetic
curl/cat boundary probes. No vendor runs under those boundary mutants.
The runner requires exit 1 and the intended diagnostic, rejecting setup
failures, skipped controls, crashes and unrelated errors as evidence.
