# TASK-260906-2g0bgq rework report 1 — stage (b) findings F1–F6 plus minors F7–F10

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch
`feat/agent-environments-stage-b`. Base: rebased onto `origin/main`
(`7320bc2a`, which records the landed stage (a) core after `981b1eeb`) with
`git rebase -S origin/main` (exit 0, 12/12 re-applied); identity proven by
`git range-diff origin/main...HEAD` showing the same 12 stage (b) subjects
re-applied. Story workspace carries an empty delta by design (implementation
lives in the curator repo; verified `git status --porcelain` clean).

Authority: curator-spec `f39f4a9` (verified `git rev-parse HEAD` on the story
worktree; `conformance/v1/manifest.json` present). Verification-sprint facts
used as given; nothing re-probed.

## Commits (5, all signed `Ivan Oparin <oparin@me.com>`, `Good "git" signature`)

- `70e872b5` — F1: gate the claude_code shared refusal on the pin.
- `198a45b0` — F2, F5, F6, F8: envprofile read-path hardening.
- `0933ea9f` — F3, F4, F10: fragment boundary case, env-only formats, lint.
- `0ee00d19` — F7, F9, F11: per-adapter targets, §12 status rows, deferred
  diagnostic, ledger rows.
- `7bca4d4b` — F9 follow-up: scan `shadow_acknowledged` envs for the
  unregistered row.

## Fixes (finding → production site → named test → narrowing mutant)

- **F1 (blocking).** `internal/envregistry/envregistry.go` `resolveIsolation`:
  the `shared` refusal is gated on `atOrAbovePinned`; below the pin the
  default and configured `shared` return `shared`, only `isolated` is refused
  (`environment_isolated_unsupported`). `TestIsolationMatrix` gains the two
  missing rows (`darwin,"",false` and `darwin,shared,false` → `shared`).
  Narrowing mutant (refuse `shared` on darwin regardless of pin):
  `TestIsolationMatrix` FAIL —
  `claude macOS default below pinned is shared, got "" (environment_shared_unsupported ...)`.
- **F2 (blocking).** `internal/envprofile/managed.go` `checkSurfaces`: after
  link-identity passes, targets under `contextstore.Root` (skills trees,
  referenced module files) keep the fast path; every other link (everything
  published through `p.docs`) is verified by `FileHash` of the target bytes
  (missing → missing, unreadable → unreadable, differ → drift, §8.4).
  `TestResolveDriftRepair` gains the write-through-link loop over 6
  (adapter, form) combinations × every rendered symlink surface.
  Narrowing mutant (keep link check, `if true` skip byte check):
  `TestResolveDriftRepair` FAIL —
  `claude_code  .agent-context/mcp/claude_code.json byte drift through an intact link must be stale, got <nil>`.
- **F3 (blocking).** `internal/envfragment/fragment_schema_test.go`
  `validHex`: De Morgan form. `golangci-lint run ./...` (v2.12.2, the CI pin)
  → `0 issues.`, exit 0. Note: this host's shared lint cache serves stale
  issues from a sibling worktree with the same module path; every green
  re-run here follows `golangci-lint cache clean`. The drafting report's
  green attestation is superseded by the re-run output below.
- **F4 (major).** `TestCheckBoundary` gains the parent-sibling case
  (`/manager/profiles/...` below the root's parent, outside the root).
  Root-parent mutant (`filepath.Dir`): FAIL —
  `a value below the environments root parent but outside the root must fail the boundary`.
- **F5 (major).** New `TestClaudeRootAlwaysCopied` asserts
  `Lstat(CLAUDE.md)` is not a symlink and the recorded
  `claude-code-root-context` reason under monolithic and referenced.
  Mutant disabling the referenced copy branch: FAIL —
  `referenced CLAUDE.md is a symlink, want a copied regular file`.
- **F6 (major).** `claudeProjects` reports
  `hasClaudeMdExternalIncludesApproved` per launch directory; the referenced
  staleness reason names the missing approval. `TestResolveClaudeProjectEntry`
  gains the tool-drops-the-key negative. Presence-only logic as mutant:
  FAIL — `a referenced home whose approval key the tool dropped must be stale, got <nil>`.
- **F7 (minor).** `Target` gains `Adapter`; exactly two rows share
  `xcode-coding-assistant` (claude_code as `CLAUDE_CONFIG_DIR`, codex_cli as
  `CODEX_HOME`); new `TargetsFor`/`TargetFor`
  (`environment_target_unknown` for `pi`/`opencode`).
  `TestSecondaryTargetsBoundToAdapters` asserts the two rows and the
  per-adapter binding. Narrowing mutant (third row binding `pi`): FAIL —
  `revision 1 declares exactly two secondary targets, got 3`.
  `TargetState` carries `Adapter`; the renderer prints `target <id> (<adapter>)`.
- **F8 (minor).** `codexKeyring` → `(bool, error)`; absent means file store,
  unreadable fails provisioning and stales verification. New
  `TestCodexUnreadableConfigFails` (directory at the config path,
  deterministic even as root). Pre-fix behavior is the killed mutant
  (provisioning must fail where it linked).
- **F9 (minor).** `Status` gains `Profiles` (lock members with weights +
  precedence primitives per installed profile) and
  `UnregisteredEnvironments` (env-ids from `forms`, `isolation`,
  `in_place_mode`, and `shadow_acknowledged` that the registry does not
  declare); the text renderer prints `mode`, `form`, `seeded-projects`, and
  the new sections (`--json` carries them as struct members).
  `TestStatusProfileMembersAndUnregistered` (via `StatusOf`) and the
  `TestEnvStatusMatrix` CLI assertions. Narrowing mutant (skip exactly
  `cursor`): FAIL — `unregistered []`.
- **F10 (minor).** `variables()` renders exactly `f.Env` (`EnvVar` first,
  then sorted); channel paths and `env_names` never enter `--format
  env|shell`. `TestFragmentOpenCodeVariableOrder` replaced by
  `TestFragmentEnvCarriesOnlyEnvObject`. Pre-fix behavior is the killed
  mutant (extra `OPENCODE_CONFIG=` line).
- **F11 (minors).** `switch.go` reserves `environment_target_unknown` for
  undeclared targets; declared targets name the deferred writes
  (`TestProfileUseTargetIsStageB` covers both; pre-fix message as mutant:
  FAIL — `a declared target must not report unknown`).
  `ManagedParent` identical branches collapsed. `skip-classes.tsv`:
  `stage-deferred` permission removed (0 emitters; gate skips confirm).
  `platform-cases.tsv`: +5 rows (ledger-consistency → 131 rows ok).
  Stale-file enumeration corrected: FIVE valid-shaped unindexed fragment
  files, all with `profile.commit`/`state_sha256` and no `precedence`
  (`valid-no-composition`, `valid-local-state-pin`,
  `valid-config-key-channel`, `valid-empty-channels`,
  `valid-file-channels`; index carries 9 valid files). Class judgement
  stands — recommend spec remove/re-index. Windows-fixture class: no new
  fixture interpolates a native path into a git config value (only
  `t.TempDir` homes, directory-at-path unreadability, JSON edits).

No surviving mutants. No delete-only mutant used as evidence.

## AC coverage — 7 of 7 rows driven through production entry points

1. Managed homes + marker → `TestResolveProvisionRepair`,
   `TestResolveDriftRepair`, `TestClaudeRootAlwaysCopied` via
   `envprofile.Resolve`.
2. Seeds/passthrough → `TestResolvePassthroughLiveness`,
   `TestCodexKeyringAmbient`, `TestCodexUnreadableConfigFails`,
   `TestOpencodeXDGSeeds`, `TestSeedUnreadableStopsBeforeFirstWrite`,
   `TestClaudeSeedMergePreservesToolState`, `TestResolveClaudeProjectEntry`
   via `Resolve`.
3. Read-only resolve + fragment → `TestResolveStaleUnprovisioned`,
   `TestResolveLockContention`, `TestCheckBoundary` via `Resolve` /
   `envfragment.CheckBoundary`.
4. MCP channels + allowlist → `mcp-*` conformance via
   `contextmaterialize.MCPFile` + `TestBoundEnvNames`.
5. Umbrella dispatch → `TestUmbrellaMissingProvider`,
   `TestUmbrellaDispatchesToProvider`, `TestUmbrellaPropagatesExitCode`,
   `TestUmbrellaInvalidNameIsUsage`, `TestImplementedCommandWinsOverProvider`
   via `run()`; `TestProfileUseTargetIsStageB` via `UseWithPolicy`.
6. Status matrix → `TestEnvStatusMatrix` (CLI) +
   `TestStatusOrphanRetention`, `TestStatusShadowAcknowledgment`,
   `TestStatusProfileMembersAndUnregistered` via `StatusOf`.
7. Conformance subset + gates + signed commits + this report → below.

## Gates (each run as a standalone process; real exit codes)

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l cmd internal` → clean.
- `golangci-lint run ./...` (v2.12.2, post-`cache clean`) → `0 issues.`,
  exit 0.
- `go test -count=1 -race ./internal/envregistry/ ./internal/envfragment/`
  → ok, exit 0.
- `go test -count=1 -race ./internal/envprofile/` → ok, exit 0 (re-run on
  the final tree after the F9 follow-up).
- `bash .github/ci/gate-selftest.sh` → 81 passed, 0 failed, exit 0.
- `bash .github/ci/no-broad-suppression.sh` → ok, exit 0.
- `bash .github/ci/ledger-consistency.sh` → 131 rows checked across linux /
  darwin / windows, ok.
- Vector families (`CURATOR_CONFORMANCE_ROOT=.../curator-spec/conformance/v1`):
  `TestConformanceEnvironmentsMonolithic` 19/19 (incl.
  `referenced-claude-code-composed`, `referenced-opencode`,
  `referenced-opencode-zero-modules`, `system-prompt-composed`,
  `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none`);
  `./internal/interop/ ./internal/envmarker/ ./internal/envfragment/` → ok.
- `bash .github/ci/test-gate.sh` (darwin lane, evidence
  `.temp/ci-evidence-local/rework-1`) → `go test exit=0, platform-case gate
  exit=0`; 19 skips (7 host-capability, 8 opt-in, 2 platform-control, 1
  root-content, 1 helper-process), 0 `stage-deferred`.
- `go test -count=1 -timeout 30m ./cmd/curator/` → ok (280s), exit 0.
- Platform-case gate linux/windows lanes: not executable on this darwin
  host; ledger rows in place and consistency-checked for all three GOOS;
  darwin lane green; linux/windows left to CI.

Scope note: the full `test-gate.sh` and `cmd/curator` runs above executed
on `0ee00d19`; the final `7bca4d4b` delta (6-line `shadow_acknowledged`
scan + test) was re-verified with build, vet, gofmt, lint, and
`-race ./internal/envprofile/` green on the final tree. The delta cannot
alter CLI output for default machine configs (no shadow acks), and the
`TestEnvStatusMatrix` assertions still hold by construction.

## Deferred (unchanged; still out of scope)

Manager-config schema 2 persistence; secondary-target surface writes;
`env unmanage`/`backups`/`config`; `lock.prev.json` GC; Windows provider
exec + the launcher itself; best-effort tool detection; per-profile repair
exclusion (manager-wide lock held instead).
