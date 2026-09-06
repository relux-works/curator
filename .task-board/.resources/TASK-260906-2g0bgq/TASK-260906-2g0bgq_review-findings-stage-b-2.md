# Review findings — stage (b) cycle 2, `feat/agent-environments-stage-b` @ `7bca4d4b`

Reviewer run `RUN-260906-5ed60e`. Subject: curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, 17 signed commits on `origin/main`
(`7320bc2a`), PR https://github.com/relux-works/curator/pull/60. Authority: curator-spec
`f39f4a9` (verified `git rev-parse HEAD` on the story worktree, clean tree). Everything below
was driven; every claim carries the command that produced it.

**repeat-of: `TASK-260906-2g0bgq_review-findings-stage-b-1.md` F3 (class: a gate line attested
green that the command does not print).** B1 is the same class one level out: every gate in the
rework report was run against one conformance root, and the lane the repository actually gates on
was never read. Two consecutive same-class findings — the next step is a gate on how gate lines
are produced, not a third revision that re-attests them.

## Verdict: changes requested → `to-dev`

Two blocking findings, one major, three minors. The rework itself is good: **F1–F6 and F8–F11 are
fixed, each with a narrowing mutant I applied myself and watched kill a named test.** What blocks
is not the rework — it is that the head under review is red on **all five** hosted Test and Race
lanes, for two defects introduced by the pre-rework stage-(b) commits: one never surfaced because
no run ever used the conformance root CI pins, the other never surfaced because no run was ever a
Windows run.

## On the empty repository delta

`CR-TASK-260906-2g0bgq-2` rev 2 carries `repository_delta=empty` and a zero-path patch. **Correct
here, not a finding**: `producer-brief-stage-b.md` places the implementation in a separate
repository (`~/Developer/ReluxWorks/curator`, branch `feat/agent-environments-stage-b`) and states
that "the story workspace carries an empty delta by design". The reviewable work is the 17-commit
curator branch the review brief names, and that is what this review drove
(`git diff 7320bc2a..HEAD` — 27 files, +7073/−40, no board files, no `LOGBOOK.md`, no control-root
writes). Recorded again as a process note: the board's CR snapshot holds no record of the 7073
lines actually under review.

---

## BLOCKING

### B1 — three tests this stage adds hard-fail against the conformance root CI pins; four hosted lanes are red on this head

`internal/envfragment/fragment_schema_test.go:312` and `:327`;
`internal/envmarker/marker_env_schema_test.go:78`;
`.github/ci/platform-cases.tsv:283-285`.

`ci.yml:44` pins `SPEC_PIN: 0ed5c691` and checks that tree out as
`protocol-spec/`, then exports
`CURATOR_CONFORMANCE_ROOT=${{ github.workspace }}/protocol-spec/conformance/v1` for the Test and
Race jobs (`ci.yml:130`, `ci.yml:197`). `0ed5c691` is *Land schema 8…*, 2026-08-24 — it predates
the environments artefacts:

```
$ git -C curator-spec ls-tree --name-only 0ed5c691 conformance/v1/schema-cases/ | grep -E 'launch-env|agent-environment'
(nothing)
$ git -C curator-spec ls-tree --name-only f39f4a9  conformance/v1/schema-cases/ | grep -E 'launch-env|agent-environment'
conformance/v1/schema-cases/agent-environment-marker-v1
conformance/v1/schema-cases/launch-env-fragment-v1
```

The three schema drivers this stage adds treat that root as a failure, not as the registered
`root-content` condition:

```go
// fragment_schema_test.go:312
if seen == 0 { t.Fatal("the root publishes no launch-env-fragment-v1 cases") }
// fragment_schema_test.go:327 and marker_env_schema_test.go:78 — a raw os.ReadFile / os.ReadDir
// on schema-cases/<family>, whose error goes straight to t.Fatal
```

**Reproduced locally**, extracting the pinned root read-only
(`git -C curator-spec archive 0ed5c691 conformance/v1 | tar -x -C .temp/review2/specpin`):

```
$ CURATOR_CONFORMANCE_ROOT=.../specpin/conformance/v1 go test -count=1 \
    -run 'TestFragmentAuthoritativeSchemaCases|TestFragmentEmissionMatchesReference|TestParseAuthoritativeEnvMarkerSchemaCases' \
    ./internal/envfragment/ ./internal/envmarker/
--- FAIL: TestFragmentAuthoritativeSchemaCases (0.00s)
    fragment_schema_test.go:312: the root publishes no launch-env-fragment-v1 cases
--- FAIL: TestFragmentEmissionMatchesReference (0.00s)
    fragment_schema_test.go:327: open .../schema-cases/launch-env-fragment-v1/valid.json: no such file or directory
FAIL	github.com/relux-works/curator/internal/envfragment
--- FAIL: TestParseAuthoritativeEnvMarkerSchemaCases (0.00s)
    marker_env_schema_test.go:78: open .../schema-cases/agent-environment-marker-v1: no such file or directory
FAIL	github.com/relux-works/curator/internal/envmarker
```

**Hosted evidence, same three tests, every lane that has reported** (extracted from the run's own
uploaded gate evidence, `actions/artifacts/{9984307207,9984380254,9984389803}`):

| hosted job | result | failing tests |
|---|---|---|
| Test (ubuntu-latest) | **fail** 3m34s | the three above |
| Test (macos-latest) | **fail** 8m46s | the three above |
| Race (ubuntu-latest) | **fail** 9m22s | the three above |
| Race (macos-latest) | **fail** 16m29s | the three above |
| Test (windows-latest) | **fail** 39m39s | the three above **plus** `TestCheckBoundary` — see B1b |
| Lint, Gate self-test ×3, Interop conformance, Naming | pass | — |

The repository already answers this exact condition, and `internal/interop` uses the answer in
five places:

```go
// internal/interop/context_materialization_test.go:75
t.Skipf("conformance root %s publishes no vectors/environments.json (pre-environments suite; root-content)", root)
```

`skip-classes.tsv` registers it (`root-content  publishes no ` → allow; `root-content  is a
pre-revision root` → allow), and the ubuntu evidence shows six such skips recorded and tolerated
in the same run. The three new drivers are the only environments tests that do not do this.

The comment above `TestParseAuthoritativeEnvMarkerSchemaCases` defends the fail-closed choice —
"a root that stops publishing it fails here instead of quietly narrowing the check" — and the
intent is right. But the distinction the repository draws is *absent family* (registered
`root-content` skip) versus *family present and wrong* (failure). Collapsing both into a failure
does not harden anything; it just makes the pinned root unusable.

**Fix.** For each of the three: when the root publishes no `schema-cases/<family>` directory,
`t.Skipf` with the registered `root-content` phrasing naming the family, exactly as
`internal/interop` does; keep `t.Fatal` for a family that exists but is empty, unindexed, or
mismatched — that is the check the comment wants and it stays. Then give the three
`platform-cases.tsv` rows a tolerated class of `root-content` (column 4/5), or the platform-case
gate rejects the now-correct skip. Re-run with `CURATOR_CONFORMANCE_ROOT` pointed at the
`SPEC_PIN` tree — not only at curator-spec main — and paste that run's output.

### B1b — `TestCheckBoundary` cannot pass on Windows: POSIX-rooted fixture paths through `filepath.IsAbs`

`internal/envfragment/envfragment_test.go:135-137`, gate at
`internal/envfragment/envfragment.go:243`; registered required on all three runners at
`.github/ci/platform-cases.tsv:268` with **no** tolerated class.

This is the class the review brief asked me to look for — stage (a) shipped a Windows-only fixture
defect no Unix lane could see. The test feeds POSIX-rooted literals to a gate built on
`path/filepath`:

```go
root := "/manager/environments"
good.Env = map[string]string{"CLAUDE_CONFIG_DIR": "/manager/environments/companyA/claude_code"}
if err := CheckBoundary(adapter, root, good); err != nil {
    t.Fatalf("a conforming fragment must pass the boundary: %v", err)
}
```

`checkPath` calls `filepath.IsAbs(value)` and then `strings.HasPrefix(value, root+string(filepath.Separator))`.
On Windows, `$(go env GOROOT)/src/internal/filepathlite/path_windows.go:184`:

```go
func IsAbs(path string) (b bool) {
	l := volumeNameLen(path)
	if l == 0 { return false }
	...
}
```

`volumeNameLen("/manager/environments/companyA/claude_code")` is 0 — no drive letter, no `\\`
UNC prefix — so `IsAbs` returns false and the `good` case is rejected with *"is not absolute"* at
the test's first assertion. The separator check would fail for the same input independently
(`\` vs `/`). **Production is unaffected**: `buildFragment` passes `EnvRoot(req.Home)`, a real
platform path (`internal/envprofile/managed.go:1464`). This is purely a fixture-portability defect,
and it is the only one of its class I found — no new fixture interpolates a native path into a git
config value (`git diff 7320bc2a..HEAD | grep -E '^\+.*(git config|url\.|insteadOf)'` → empty).

**Confirmed on the hosted runner**, word for word, from the Windows lane's own uploaded evidence
(`actions/artifacts/9984817963`), which reported `fail` at 39m39s after this reading was written:

```
=== RUN   TestCheckBoundary
    envfragment_test.go:139: a conforming fragment must pass the boundary:
        fragment env CLAUDE_CONFIG_DIR "/manager/environments/companyA/claude_code" is not absolute
--- FAIL: TestCheckBoundary (0.00s)
```

`TestFragmentAuthoritativeSchemaCases` shares the shape in its own local `checkPath`
(`fragment_schema_test.go:143`) but fails earlier under B1.

**Fix.** Build the fixture root and values with `filepath.Join` over a platform-absolute base
(`t.TempDir()`, or a `runtime.GOOS`-aware literal in one shared helper), so the same assertions
run on all three runners. Do not register a tolerated skip for it — the boundary is not
platform-specific, only the fixture is.

---

## MAJOR

### M1 — the F7 fix ships a gate no production path calls; `--env pi --target xcode-coding-assistant` is still admitted

`internal/envregistry/envregistry.go:326` (`TargetsFor`) and `:338` (`TargetFor`);
caller that should use them: `internal/envprofile/switch.go:171`.

The registry half of F7 is genuinely fixed and genuinely reachable: two rows, one per adapter, and
`env status` prints both bindings (driven through the CLI —
`target xcode-coding-assistant (claude_code)` / `(codex_cli)`). The narrowing mutant that adds a
third row binding `pi` is killed by `TestSecondaryTargetsBoundToAdapters`
(`revision 1 declares exactly two secondary targets, got 3`).

The per-adapter *resolution* half is not wired in:

```
$ grep -rn 'TargetsFor(\|TargetFor(' --include='*.go' . | grep -v _test.go
internal/envregistry/envregistry.go:326:func TargetsFor(adapterID string) []Target {
internal/envregistry/envregistry.go:338:func TargetFor(adapterID, id string) (Target, error) {
```

Definitions only. `switch.go:171` still calls the global `envregistry.TargetByID(target)`, and
`status.go:464` iterates `envregistry.Targets` directly. Driven through the production `run()`
entry point (`.temp/review2/artifact/zz_r2_target_probe_test.go`):

```
profile use --env pi           --target xcode-coding-assistant -> exit=1 "secondary fixed-home target ... writes are deferred"
profile use --env opencode     --target xcode-coding-assistant -> exit=1 "secondary fixed-home target ... writes are deferred"
profile use --env claude_code  --target xcode-coding-assistant -> exit=1 "secondary fixed-home target ... writes are deferred"
profile use --target ghost-target                              -> exit=1 "environment_target_unknown: undeclared target"
```

`pi` and `opencode` declare no target, yet the target resolves for them; the rework report's claim
—"new `TargetsFor`/`TargetFor` (`environment_target_unknown` for `pi`/`opencode`)" — is true of
the helper and false of every production path. This is the AC's named shape: *the check present but
uncalled from production*; `TestSecondaryTargetsBoundToAdapters` calls `TargetFor` directly, which
proves it compiles and returns what it is told to.

**Blast radius is small** — target surface writes are deferred, so the operator gets an error
either way, just the wrong one. It is major because it is an attested gate that has never run.

**Fix.** Call `envregistry.TargetFor(environment, target)` from `useLocked` when `--env` is
present (falling back to `TargetByID` when it is not), and drive the refusal through `run()`, not
through the helper. Narrowing mutant: make `TargetFor` ignore the adapter argument and delegate to
`TargetByID` — a CLI-level test must fail.

---

## MINOR

### m2 — `env status --json` emits Go field names
`internal/envprofile/status.go:107-120`. `Status` and every nested struct carry no `json` tags, so
`curator env status --json` publishes `{"Adapters","Homes","NonCurrent","Notes","Orphans",
"Profiles","Scopes","Targets","UnregisteredEnvironments"}` (driven through `run()`), while every
other machine-readable surface in `cmd/curator` uses explicit snake_case (`cmd/curator/builds.go:267-276`).
§12 does not name the members, so this is not a spec violation — but the wire shape is currently
hostage to a Go identifier rename, and it is inconsistent with the rest of the CLI. Add tags or
declare the shape as a stated bound.

### m3 — the F2 write-through-link loop can go vacuous without saying so
`internal/envprofile/managed_test.go:303-338`. The loop selects surfaces with
`strings.Contains(filepath.ToSlash(target), "/rendered/")` and `continue`s otherwise; a case that
yields no match asserts nothing and passes. It is **not** vacuous today — I measured it
(`.temp/review2/artifact/zz_r2_probe_test.go`): claude_code/monolithic 2, claude_code/referenced 2,
codex_cli 3, opencode/monolithic 3, opencode/referenced 4, pi 2 rendered symlink surfaces, plus 1
store-target surface in each referenced case. But the count is the evidence and the test does not
assert it. Add a per-case `if rendered == 0 { t.Fatalf(...) }`, so a future change to
`storeDocPath` cannot silently empty the loop.

### m4 — the rework report's gate table has the F3 shape at one level out
Every line in `TASK-260906-2g0bgq_rework-report-1.md` §Gates is the real output of a real command —
that part of F3 is fixed and I re-ran the whole table. But every one of them was run against
`CURATOR_CONFORMANCE_ROOT=.../curator-spec/conformance/v1`, and the closing line
"Platform-case gate linux/windows lanes: … left to CI" states an outcome the report never read.
CI was already red. A gate line that names CI must carry `gh pr checks <n>` output, or say
plainly that CI was not consulted.

---

## What I re-drove and found correct at `7bca4d4b`, darwin/arm64, go1.25.5

Every mutant below was applied by me to a throwaway `rsync` copy under
`.temp/review2/mut/<name>/` (never in the producer's worktree, which is clean at `7bca4d4b`),
built, and run against the named test. Harness and logs in the probes archive.

| finding | how I drove it | narrowing mutant | result |
|---|---|---|---|
| F1 | `Resolve` with `Detect → "2.1.200"` on darwin: default **provisions** (1032-byte fragment), configured `shared` admitted, `isolated` → `environment_isolated_unsupported`; at `2.1.261` default provisions and `shared` → `environment_shared_unsupported` | refuse `shared` on darwin regardless of the pin | **killed** — `claude macOS shared below pinned is shared, got ""` |
| F2 | write through the intact link into **every** rendered surface of 6 (adapter, form) combos — 16 surfaces — via production `Resolve`: all 16 report `environment_home_stale` and emit **no** fragment; store-target links keep the link-identity fast path | keep the link check, skip the byte check | **killed** — `.agent-context/mcp/claude_code.json byte drift through an intact link must be stale, got <nil>` |
| F3 | `golangci-lint cache clean && golangci-lint run ./...` in the producer's own worktree → `0 issues.`, **exit 0**; De Morgan rewritten at the source, `.golangci.yml` untouched by this stage, zero `nolint` in the delta; hosted **Lint pass** | — | fixed |
| F4 | `CheckBoundary` parent-sibling case | `root := filepath.Dir(filepath.Clean(envRoot))` | **killed** — `a value below the environments root parent but outside the root must fail the boundary` |
| F5 | `Lstat(<home>/CLAUDE.md)` regular file **and** recorded `claude-code-root-context` reason under both forms, via `Resolve` | `if false &&` on the referenced copy branch | **killed** — `referenced CLAUDE.md is a symlink` |
| F6 | provision referenced, delete only `hasClaudeMdExternalIncludesApproved` (project entry kept), `Resolve` → `environment_home_stale: launch directory … has no external-includes approval`, no fragment | `entries[name] = true` (presence-only) | **killed** — `a referenced home whose approval key the tool dropped must be stale, got <nil>` |
| F7 | registry rows + `env status` bindings correct; resolution helper unreachable | third row binding `pi` | **killed** (`got 3`) — but see **M1** |
| F8 | directory at the native `config.toml`, `Resolve --repair` → `environment_seed_unreadable: seed config.toml: … is a directory`, provisioning stops | `return false, nil` on read error | **killed** — `an unreadable native config must be stale, got <nil>` |
| F9 | `env status` through the CLI carries `mode`, `form`, `profile … precedence winner=… placement=…`, `member context … weight …`, and the unregistered rows | skip `ghost` in the `ShadowAcknowledged` scan → **killed** (`unregistered [cursor]`); skip `cursor` in `consider` → **killed** (`unregistered [ghost]`) | both killed |
| F10 | `env resolve --format env` and `--format shell` through `run()` for **all four** adapters, compared name-for-name against the same resolve's JSON `env` object: exact match every time (`CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `XDG_CONFIG_HOME`, `PI_CODING_AGENT_DIR`) | re-append MCP channel variables in `variables()` | **killed** — `env format carries more than the env object: "XDG_CONFIG_HOME=…\nOPENCODE_CONFIG=…"` |
| F11 | `stage-deferred` row deleted from `skip-classes.tsv` (3 deletions) and **0 occurrences** in observed skips; `switch.go:171` reserves `environment_target_unknown` for undeclared targets (`ghost-target` → the diagnostic, declared → the deferred message); `ManagedParent` single branch; +30 stage-(b) `platform-cases.tsv` rows | — | fixed |

Gates I ran myself, each as a standalone process, real exit codes, in a clean `rsync` copy:

| gate | result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./...` (v2.12.2, post-`cache clean`) | `0 issues.`, **exit 0** |
| `bash .github/ci/gate-selftest.sh` | **81 passed, 0 failed** |
| `bash .github/ci/no-broad-suppression.sh` | ok |
| `bash .github/ci/test-gate.sh` (`CURATOR_CONFORMANCE_ROOT=curator-spec f39f4a9`, `GO_TEST_TIMEOUT=30m`) | `go test exit=0, platform-case gate exit=0`; **19 skips, all registered, 0 `stage-deferred`** |
| `go test -count=1 ./cmd/curator -run 'TestEnv…\|TestUmbrella…'` | ok |
| `TestConformanceEnvironmentsMonolithic` | **19 of 19 pass**, zero skips, incl. `referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`, `system-prompt-composed`, `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none` |
| `bash .github/ci/test-gate.sh` with `CURATOR_CONFORMANCE_ROOT=SPEC_PIN` | **not green — B1** |
| commits | **17**, every one `G` (good signature), author `Ivan Oparin <oparin@me.com>`; scope confined to stage (b) — no board files, no `LOGBOOK.md`, no control-root writes |

Also confirmed reachable from production, not merely defined:
`envfragment.CheckBoundary` ← `managed.go:1464`; `envfragment.BoundEnvNames` ← `managed.go:1460`;
`Adapter.ResolveIsolation` ← `EffectiveIsolation` ← `assembleHome`. `TargetsFor`/`TargetFor` are
the exception — **M1**.

Also observed, not findings: the `env status` text renderer prints `form ,` (empty) for
unprovisioned rows; `curator profile use --target` accepts `--env` for an adapter the target is not
bound to (the operator-visible half of M1).

## Probe artifacts

`TASK-260906-2g0bgq_review-probes-stage-b-2.tar.gz` — the in-package probes
(`zz_r2_probe_test.go` driving `envprofile.Resolve`; `zz_r2_cli_probe_test.go` and
`zz_r2_target_probe_test.go` driving `cmd/curator`'s `run()`), the mutant harness (`mutant.sh`),
all ten mutant logs, the local darwin `test-gate.sh` log and observed skips, and the ubuntu CI
observed skips. Scratch lived under the story worktree's `.temp/review2/`; the producer's worktree
was never written to and is clean at `7bca4d4b`.
