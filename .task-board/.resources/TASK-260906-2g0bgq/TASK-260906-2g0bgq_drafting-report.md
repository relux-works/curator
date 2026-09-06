# TASK-260906-2g0bgq drafting report — stage (b): managed homes, seeds and passthrough, read-only resolve and the fragment, MCP channels, umbrella dispatch, env status

Branch `feat/agent-environments-stage-b` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, from stage (a) head `834b40f6`. All commits signed (SSH key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cMhi`, `commit.gpgsign=true`).

Authority: curator-spec worktree at `f39f4a9` (protocol/environments.md rev 1.1). Verification-sprint facts used as given (Keychain service-name suffix over `CLAUDE_CONFIG_DIR`; no codex global-`AGENTS.md` truncation; codex/pi `auth.json` in-place rewrite; `@path` project-approval guard incl. linked user-level `CLAUDE.md`; codex `-p` single silent-if-missing). Nothing re-probed.

## Surface → package map

- Referenced root-context bytes (`@`-refs for claude_code, header-only root + manager-authored `opencode.json` for opencode, exact module files, `environment_form_unsupported` for codex_cli/pi, `environment_path_collision` detection) → `internal/contextmaterialize/referenced.go`.
- MCP launch-channel bytes (CCJ-1 for claude_code/opencode, fixed TOML layer for codex_cli with `curator-mcp` name, sorted set resolution, sorted `env_names` union, `written=false` for pi and empty sets) → `internal/contextmaterialize/mcp.go`.
- Closed revision-1 adapter table (home variables incl. opencode parent semantics, targets, forms, §7.3/§7.8 channel descriptors, per-platform passthrough, seeds, isolation matrix, XDG allowlist default, targets, advisories, recorded releases) + `MachineConfig` knobs with §12.1 defaults + form/isolation/consent gates → `internal/envregistry/envregistry.go`.
- Managed homes (layout below `<manager>/environments`, provision, lock-free verify over exactly the marker's surfaces, repair under the mutation lock, seeds/passthrough/XDG, `.claude.json` merge, `environment_home_stale`/`environment_repair_failed`/`environment_lock_unavailable` paths, first-resolve notice) → `internal/envprofile/managed.go`.
- `launch-env-fragment-v1` build, `json|env|shell` renders, `BoundEnvNames` double bound, `CheckBoundary` §10.3 → `internal/envfragment/envfragment.go`.
- `env resolve` / `env status` rows and text/JSON render → `cmd/curator/env.go`, `cmd/curator/envstatus.go`, `internal/envprofile/status.go`.
- `curator run` umbrella dispatch (`curator-<name>` on PATH, `subcommand_provider_missing` + guidance, `subcommand_provider_untrusted` for the shim dir and environments root) → `cmd/curator/umbrella.go` + `main.go` wiring.
- Marker reader rules the schema families require (sorted seeded_projects, source-ordered surface keys, closed passthrough strategies, root inclusion, no duplicate members, `imported_from_native` exactly-as-true-on-path) → `internal/envmarker/envmarker.go`.
- `profile update` retains `lock.prev.json` (§9.2); `profile remove --purge` removes the profile's environments directory (§9.2).

## Conformance

- `referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`, `system-prompt-composed`, `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none`: 19 of 19 materialization cases pass byte for byte with surface hashes, resolved sets, and `env_names` unions checked; the stage-deferred skip is removed (zero skips in the family) and the ledger tolerance row for it is deleted.
- Marker family: all 53 indexed `agent-environment-marker-v1` cases decided correctly (10 valid accepted incl. managed-home/passthrough/seed/XDG-link/fallback members, 43 invalid rejected).
- Fragment family: all 49 indexed `launch-env-fragment-v1` cases decided correctly against the registry model; the emitted canonical claude fragment re-encodes the published `valid.json` reference bytes exactly.
- Unindexed stale files (not conformance; logged by the drivers, skipped): `agent-environment-marker-v1/{valid-no-composition,valid-linked-copy-fallback}.json` + 10 invalid-`composition`/`copy-fallback`/`ref` files, and `launch-env-fragment-v1/{valid-no-composition,valid-local-state-pin}.json` — all in the obsolete composition/copy_fallback/source_identity/commit-pin shape that contradicts the published schemas (marker schema requires version/profile/members/precedence/mode/surfaces; fragment schema requires profile `{name, lock_sha256}`). Recommend spec remove or re-index them.

## AC coverage — 7 of 7 rows driven through production entry points

1. Managed homes + marker records → `TestResolveProvisionRepair` (+ drift/link/orphan/purge tests) via `envprofile.Resolve`.
2. Seeds/passthrough (liveness, keyring, XDG, unreadable, merge) → `TestResolvePassthroughLiveness`, `TestCodexKeyringAmbient`, `TestOpencodeXDGSeeds`, `TestSeedUnreadableStopsBeforeFirstWrite`, `TestClaudeSeedMergePreservesToolState` via `Resolve`.
3. Read-only resolve + fragment (`--repair` lock, formats, §10.3) → `TestResolveStaleUnprovisioned`, `TestResolveLockContention`, `TestResolveFormats`, `TestCheckBoundary` via `Resolve`.
4. MCP channels + allowlist → interop `mcp-*` cases via `MCPFile`, plus existing `checkMCPCommand` gate tests (stage a).
5. `curator run` umbrella dispatch → `TestUmbrellaMissingProvider`, `TestUmbrellaDispatchesToProvider`, `TestUmbrellaPropagatesExitCode`, `TestUmbrellaInvalidNameIsUsage`, `TestImplementedCommandWinsOverProvider` via `run()`.
6. Env status matrix → `TestEnvStatusMatrix` (CLI) + `TestStatusOrphanRetention`, `TestStatusShadowAcknowledgment` via `StatusOf`.
7. Conformance subset + gates + signed commits + this report → interop/schema drivers, gate outputs below, `git log --show-signature`.

## Mutant evidence

| mutant | narrows the gate to | named test that fails |
|---|---|---|
| `SupportsReferenced` admits everything | form gate admits codex_cli referenced | `TestReferencedRejectsCodex` — FAIL observed, restored |
| `BoundEnvNames` skips reserved exclusion | §10.3 bound admits PATH/HOME | `TestBoundEnvNames` — FAIL observed, restored |
| `codexKeyring` treats any value as keyring (token preserved, behavior changed) | strategy gate never links | `TestCodexKeyringAmbient/file` — initially SURVIVED, then killed after adding the file/auto/keyring matrix (behavioral suite `internal/envprofile` executed) |
| opencode `isolated` refusal removed | isolation gate admits opencode isolated | `TestIsolationMatrix` — FAIL observed, restored |
| `CheckBoundary` root check removed | §10.3 admits out-of-root values | `TestCheckBoundary` — FAIL observed, restored |

No surviving mutants. Delete-only mutants were not used as evidence.

## Gate outputs

All commands run as standalone processes (no `tee`/pipes); exit codes real.

- `git submodule update --init --recursive` → exit 0 (checked out `agents/skills/skill-go-testing-tools`).
- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l cmd internal` → clean.
- `golangci-lint run ./...` (v2.12.2, the CI pin) → exit 0, 0 issues. Includes 4 mechanical fixes in `internal/interop/context_resolution_test.go` (unused params, De Morgan) that predated this stage and kept the gate red.
- `bash .github/ci/gate-selftest.sh` → 81 passed, 0 failed, exit 0.
- `bash .github/ci/ledger-consistency.sh` → 126 rows checked across linux/darwin/windows, ok.
- `bash .github/ci/no-broad-suppression.sh` → ok.
- `go test -count=1 -race` on touched packages → all ok (contextmaterialize, envregistry, envfragment, envmarker, envprofile, interop-with-vectors).
- Vector families through `CURATOR_CONFORMANCE_ROOT` (spec worktree `conformance/v1`): interop full package ok; 19/19 environments materialization cases byte-exact, 0 skips; 53/53 indexed marker cases; 49/49 indexed fragment cases.
- `bash .github/ci/test-gate.sh` (served+deferred+assert stages, darwin lane): first run `go test exit=1` on exactly one package (`cmd/curator TestRunUnknownCommand`, the pre-umbrella contract) with the platform-case gate ok; after the §11 contract update + full `go test -count=1 -timeout 30m ./cmd/curator/` → ok (276s). A second full re-gate failed environmentally (`acquire package host GOROOT test lock: context deadline exceeded` — a concurrent `go test` from another session on this shared machine; zero tests ran, left untouched per process rules); after it cleared, the final re-gate → `go test exit=0, platform-case gate exit=0`, evidence in `.temp/ci-evidence-local/stage-b-final`.
- `go test -count=1 -timeout 30m ./cmd/curator/` → ok, exit 0.
- Platform-case gate linux/windows lanes: not executable on this darwin host; ledger rows in place (126 rows, consistency-checked for all three GOOS), darwin lane green; linux/windows left to CI.

## Deferred with reasons (not in stage-b scope or a later batch by spec)

- Manager-config schema 2 persistence of the §12.1 knobs: the spec carries these "under one environments object" in "the next batch". `MachineConfig` implements every knob with its stated default; no file surface added.
- Secondary-target (Xcode) surface writes: participation evaluation, one-time consent gate, and status rows implemented; writes deferred (the default probe path requires an operator to record; no consent CLI surface exists yet).
- `env unmanage` / `env backups` / `env config` commands: not in the brief's six items; purge is covered by `profile remove --purge`.
- Retained previous lock (`lock.prev.json`) is written on update; its GC drop belongs to future profile-GC work (no profile GC exists yet).
- `curator run` provider execution on Windows (`.bat`/`.exe` lookup) and the launcher itself: the launcher is a separate repository by brief design; dispatch resolution is tested, exec handoff is POSIX-proven.
- Tool release detection (`claude|codex|opencode|pi --version`, first dotted token) is best-effort and warns `environment_tool_version_unverified` on any failure — never a match.
- In-flight `--repair` exclusion holds the manager-wide mutation lock, which is stronger than the specified per-profile-×-environment exclusion.

## Spec observations for the story (not blockers)

- Stale unindexed files contradicting the published schemas (listed under Conformance above); recommend removal/re-index in curator-spec.
- `TestRunUnknownCommand` encoded the pre-§11 contract; updated to assert provider-refusal for identifier names and usage error for non-identifier names (spec-mandated change, both paths tested).
