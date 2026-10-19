# Codex 0.162.1 qualification

TASK-261018-70enqt qualifies the existing `codex_cli` registry row on
`codex-cli 0.162.1`, macOS/arm64, executable SHA-256
`74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0`.
The candidate changes `VerifiedRelease` from `0.153.2` to `0.162.1`, adds
its pinning test, and records observations. Product behavior is unchanged.

The source checkpoint is `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`.
The refreshed remote main was `db65e39555b8dbac7e75f3b2d337daa33831f297`
(one commit ahead, changing only CHANGELOG). The managed Story branch is
left uncommitted at its checkpoint; integration belongs to the orchestrator.

## Registry facts and raw evidence

Paths below are relative to `evidence/`. Each vendor command log records cwd,
argv, actual exit and combined output. `$RUN`, `$TMP`, `$CODEX`, and
`$OPERATOR_HOME` replace
private locators; screen-truncated paths are sanitized too. Synthetic marker
text and loopback requests are retained. No operator credential was copied.

| Registry fact | Observation on 0.162.1 | Test and evidence |
| --- | --- | --- |
| `Probe`, `VerifiedRelease` | `codex --version` exits 0 with the exact release; digest recorded | Every replay test: `identity.txt`, `version.log`; `TestCodexCLI162VerifiedRelease` through production `envregistry.ByID` |
| `EnvVar=CODEX_HOME` | Selected home supplies context; changing only that variable removes both synthetic root and skill | `TestCodex162HomeRootSkills/{root-skills,empty-home-control}.log` |
| `RootTarget=AGENTS.md`, monolithic form/default | Global root content is in the real prompt; a production managed-root symlink also loads | `TestCodex162HomeRootSkills/root-skills.log`; `TestCodex162ManagedAdvisory/{small,large}-managed-prompt.log` |
| `SizeAdvisoryBytes=32768` | Production `Install -> Resolve` keeps a 33666-byte rendered root and emits `environment_context_size_exceeded`; small control does not warn | `TestCodex162ManagedAdvisory/{small,large}-resolve.txt` |
| `GlobalContextCap=none` | Global root tail survives above 32 KiB; this is not a claim of infinite model context | `TestCodex162RootSize/large-root.log`; managed-root logs above; tagged loader source below reads the complete file |
| `SkillsDir=skills` | Skill metadata is discovered in a regular directory and through a directory symlink; absent in another home | `TestCodex162HomeRootSkills/{root-skills,linked-skill,empty-home-control}.log` |
| `SystemPrompt`: `model_instructions_file`, replace | Exact synthetic replacement becomes the first developer message of an actual exec request | `TestCodex162ManagedChannels/{system-replacement.log,system-request.json}` |
| MCP: `-p curator-mcp`, argument is a name | Production `contextmaterialize.MCPFile` output loads from `curator-mcp.config.toml`; server absent without the layer | `TestCodex162ManagedChannels/{curator-mcp.config.toml,mcp-layer.log,mcp-no-layer.log}` |
| `Seeds=[config.toml]`, revision B, copied seed | Config selects a loopback provider/model; production managed-home seed is accepted too. Existing B stripping/invalid-TOML tests stay green | `TestCodex162ExecSeedProfile/{seed.log,seed-request.json}`; `TestCodex162ManagedAdvisory/*managed-prompt.log`; `managed-01.log` |
| `Passthrough=auth.json`, file-link fallback; `CredentialScope=home` | Linked synthetic file authenticates `login status`; a different empty CODEX_HOME and an unlinked home report `Not logged in`, exit 1 | `TestCodex162LinkedAuth/{linked-status,other-home-status,unlinked-status}.log` |
| `AuthWrite=in-place` | Synthetic `login --with-api-key` changes target bytes while preserving link and native inode | `TestCodex162LinkedAuth/{linked-login.log,auth-state.txt}` |
| `ExecFlags` | In a confirmed non-Git cwd, exec refuses without `--skip-git-repo-check` (exit 1); with it the loopback turn succeeds | `TestCodex162ExecSeedProfile/{confirm-outside-git,outside-git-refusal,seed}.log` |
| `ProfileFlag=-p (single, silent-if-missing)` | Existing layer changes requested model; missing layer selects base; a repeated flag refuses with exit 2 | `TestCodex162ExecSeedProfile/{layer,missing,repeated-profile}.log` and corresponding request JSON |

`SeedWritten={}` describes Curator's copy policy, not a promise that Codex
never edits its config: the trust probe records Codex's own additions.
Registry structural fields, keyring-preferred policy, and seed revision are
unchanged. Qualification covers the file credential path, not native stores.

## Pilot facts: CIP-0020 and the Werkstatt layout

`TestCodex162ProjectHookTrust` launches the real TUI in a private background
tmux server, using project `.codex/config.toml` SessionStart hooks. The screen
says **Hooks need review**, **Trust all and continue**, and
**Continue without trusting (hooks won't run)**. Selecting trust persists
`hooks.state.<project-config>:session_start:0:0.trusted_hash` in the disposable
user config. A fresh invocation in that project skips hook review; a second
project with identical hook bytes prompts again. See
`TestCodex162ProjectHookTrust/{first-screen.log,trusted-config.toml,repeat-screen.log,second-screen.log}`.
This qualifies startup consent and source/project scope. It does not qualify
execution of a trusted hook, every hook event, or arbitrary subdirectory scope.
Project trust and hook trust are separate: the fixture pre-trusts both projects.

`TestCodex162GitSandbox` runs `codex sandbox` with **workspace-write** against
real disposable Git repositories. An ordinary file is writable (exit 0), but
`.git/sentinel` at the primary root and another writable root is refused
(exit 1, `Operation not permitted`); original bytes remain unchanged.
**A descendant repository's `nested/.git/sentinel` is writable in this probe.**
Therefore the pilot's phrase "every .git under writable roots" must mean
the root-level `.git` for each root to match this observation. The probe
contradicts a recursive interpretation; it does not establish the pilot's intent.
An explicit additional writable root at `.git` can also override this protection,
as demonstrated by the killed `root-git-extra-root` control.

With `git init --separate-git-dir <metadata> .`, metadata outside those roots,
`GIT_DIR`/`GIT_WORK_TREE`, and an explicit writable metadata root, real
`git config` succeeds inside the same workspace-write sandbox. See
`TestCodex162GitSandbox/{ordinary,root-git,nested-git,additional-root-git,separate-init,separate-write,separate-read}.log`.
This is the measured basis for a separated Werkstatt metadata layout; neither
its product implementation nor CIP-0020/spec text is changed here.

## Tools and replay

Go 1.25.5, Git, Codex 0.162.1, and tmux are used. tmux supplies a private PTY
for the trust screen; macOS Seatbelt supplies the sandbox probe. Go's standard
library serves a loopback Responses endpoint and drives assertions, not a live
model. No additional language, library, package install, or real auth is needed.
The sandbox test explicitly skips on non-macOS hosts; that is not qualification
evidence for another platform. The other opt-in tests require Codex; the trust
test also requires tmux. Replay targets this macOS host; the non-sandbox
commands also assume Unix Git and temporary-directory paths.

```sh
# Explicit default invocation: every vendor probe skips before creating a home.
go test ./.research/codex162 -count=1 -v

# Choose fresh ignored output directories; set the binary locator privately.
mkdir -p .temp/codex162
export CODEX162_BINARY="$(command -v codex)"
export CODEX162_EVIDENCE="$PWD/.temp/codex162/replay-new"
export TMPDIR="$PWD/.temp/codex162"
go test ./.research/codex162 -count=1 -v

# Expected-red individual suites; the control runner itself must exit 0.
export CODEX162_MUTANTS_ROOT="$PWD/.temp/codex162/mutants-new"
sh .research/codex162/mutants.sh

go test ./internal/envregistry -count=1 -v
go vet ./internal/envregistry ./.research/codex162
golangci-lint run ./internal/envregistry/... ./.research/codex162/...
go build -o .temp/codex162/curator ./cmd/curator
```

Evidence maintenance also uses `sh -n .research/codex162/mutants.sh` for shell
syntax, `shasum -a 256` over the listed source files for
`evidence/source-sha256.txt`, and `rg -n '/Users/|Bearer |sk-proj-' .research/codex162/evidence`
as a bounded publication check (exit 1 means no matches, not a failed probe).
Read-only `curl -fLsS <tagged-source-url> -o .temp/codex162/<source-file>`
lookups keep upstream source in ignored scratch. Before publishing outer Go
failure logs, `CODEX162_PRIVATE_HOME="$HOME" perl -pi -e
's/\Q$ENV{CODEX162_PRIVATE_HOME}\E/\$OPERATOR_HOME/g' <log>` redacts paths
printed by test diagnostics, in addition to the harness's vendor-log redaction.
These are evidence utilities; their scratch outputs stay in `.temp/codex162/`.

Ordinary `go test ./...` excludes this hidden research directory. Every vendor
process receives an explicit HOME, CODEX_HOME, XDG paths and temporary directory,
with no inherited auth/proxy/session environment. Temporary homes remain under
the chosen TMPDIR and are removed by Go. Non-Git exec uses the pre-existing
system temporary directory only as cwd; it writes no project file there.
No operator Codex auth/config is imported or changed. Root README is untouched.

The seven controls and the historical survivor are reported in the
[handoff/gaps record](TASK-261018-70enqt_gaps.md), alongside actual command exits,
coverage ratio, source identity, excluded rows, and integration limits.
That record also documents the revision-1 hosted failure and the five dependent
credential-test expected values updated during recovery. No credential
production code changed.

Supporting official references, checked after local tool discovery:
[configuration](https://learn.chatgpt.com/docs/config-file/config-advanced),
[AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md),
[tagged global loader](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/codex-home/src/instructions/mod.rs),
[tagged trust UI](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/tui/src/startup_hooks_review.rs).
The installed executable's raw observations above are the qualification evidence;
documentation and upstream source do not replace those runs.
