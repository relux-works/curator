# Rework report 2 — stage (b), findings B1, B1b, M1, m2–m4

Task `TASK-260906-2g0bgq`, branch `feat/agent-environments-stage-b` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`.
Reviewed head: `7bca4d4b`. This revision: five signed commits on top,
`34d86781` (B1), `3a5a27cd` (B1b), `dfbc4435` (M1), `2794b5ed` (m2),
`1a936e77` (m3). Author `Ivan Oparin <oparin@me.com>`, all commits signed.
`git status` clean. No push, no tag, no PR action (PR #60 owned by the
orchestrator). No `LOGBOOK.md`, no control-root writes.

The cycle-2 rework verification (F1–F6, F8–F11 with the reviewer's own
narrowing mutants) stands and was not touched.

## Finding → resolution

| Finding | Resolution | Production call site | Named test |
|---|---|---|---|
| B1 (blocking): the three schema drivers `t.Fatal` against the `SPEC_PIN` root, which predates both families; all hosted Test/Race lanes red | Registered the mechanism `suite-plan.sh` exists for instead of the tolerated skip: two rows in `.github/ci/root-artifacts.tsv` (`internal/envfragment` → `schema-cases/launch-env-fragment-v1`; `internal/envmarker` → `schema-cases/agent-environment-marker-v1`); the three `platform-cases.tsv` rows now tolerate `root-unset` on `linux,darwin,windows` with the behaviour sentence extended to say a root which does not publish the family defers the package and records a `root-unset` skip. The three `t.Fatal`s are unchanged: a served root that lacks the family is still a genuine error and stays fatal | `suite-plan.sh` partition consumed by `test-gate.sh`; drivers `internal/envfragment/fragment_schema_test.go:312,327`, `internal/envmarker/marker_env_schema_test.go:78` | Ledger rows `TestFragmentAuthoritativeSchemaCases`, `TestFragmentEmissionMatchesReference`, `TestParseAuthoritativeEnvMarkerSchemaCases` (tolerated `root-unset` skip when deferred; must pass when served) |
| B1b (blocking): `TestCheckBoundary` feeds POSIX-rooted literals to `filepath.IsAbs` and cannot pass on Windows | Fixture rebuilt on a platform-absolute base: `root = filepath.Join(t.TempDir(), "environments")`, every value derived with `filepath.Join` (traversal keeps a real `..` segment via `root + Separator + ".." + Separator + "escape"`); `SystemPrompt.Path`/`MCP.Path` rewritten under the same root per case. `fragment_schema_test.go:checkPath` uses slash semantics (`strings.HasPrefix(value, "/")`) because the published vectors carry POSIX evident paths under `/manager/environments`; the host-path boundary stays covered by `CheckBoundary` + the rewritten test | `internal/envfragment/envfragment.go:CheckBoundary` | `TestCheckBoundary` |
| M1 (major): `TargetsFor`/`TargetFor` defined, never called; `switch.go:171` used global `TargetByID` | `useLocked` calls `envregistry.TargetFor(environment, target)` when `--env` is present, `TargetByID` when it is not. Narrowing mutant (below) kills the new CLI-level test | `internal/envprofile/switch.go:useLocked` via `cmd/curator/profile.go:cmdProfileUse` → `run()` | `TestProfileUseTargetBoundToAdapter` (new) |
| m2 (minor): `env status --json` published Go field names | Added explicit `snake_case` `json` tags to all nine status structs (`Status`, `HomeState`, `SurfaceState`, `ScopeHome`, `AdapterState`, `TargetState`, `ProfileState`, `MemberState`, `PrecedenceState`); `TestEnvStatusMatrix` now asserts all nine top-level `snake_case` members exist and `Homes` does not | `cmd/curator/env.go:cmdEnvStatus` → `json.MarshalIndent(status)` | `TestEnvStatusMatrix` |
| m3 (minor): write-through-link loop could go vacuous silently | Per-case `rendered` counter with `if rendered == 0 { t.Fatalf(...) }`; observed non-vacuous (see gates) | `internal/envprofile/managed.go:Resolve` via `TestResolveDriftRepair` | `TestResolveDriftRepair` |
| m4 (minor): prior report attested CI lanes it never read | This report's gate table (below) follows the rework-2 rule: exact standalone command, observed exit code, named root twice where root-sensitive, verbatim `gh pr checks 60` output for the old head, plain statement where CI was not consulted, unrun gates listed as not run | — | — |

## M1 narrowing mutant (applied, observed, reverted; tree clean after)

Mutant: `TargetFor` ignores its adapter argument and delegates to
`TargetByID`:

```go
func TargetFor(adapterID, id string) (Target, error) {
    return TargetByID(id)
}
```

Killed by `TestProfileUseTargetBoundToAdapter` through production `run()`:

```
--- FAIL: TestProfileUseTargetBoundToAdapter (0.39s)
    profile_test.go:249: use --env pi --target xcode-coding-assistant = 1, want environment_target_unknown
        stderr:
        curator: secondary fixed-home target "xcode-coding-assistant" writes are deferred: participation, consent, and status are implemented, surface writes are not
FAIL
```

Restored tree: `go test -count=1 -run 'TestProfileUseTargetBoundToAdapter'
./cmd/curator/` → `ok`. `git status` clean.

## B1b sweep (stage-(b) delta `7320bc2a..HEAD`, what was looked for and found)

Looked for the Windows-portability class: a POSIX literal fed to
`path/filepath`, a hardcoded `/` separator in a path comparison, a
`strings.HasPrefix` over host paths.

- `internal/envfragment/envfragment_test.go:TestCheckBoundary` — the
  reported defect. Fixed (above). F4 mutant re-run against the rewritten
  fixture (`root := filepath.Dir(filepath.Clean(envRoot))`): killed —
  `--- FAIL: TestCheckBoundary ... a value outside the environments root
  must fail the boundary` (mutated root admits `base/evil`; restored tree
  `ok`).
- `internal/envfragment/fragment_schema_test.go:checkPath` — same shape in
  the driver's local check. Fixed to slash semantics with a comment naming
  the reason (vectors are POSIX evident paths; host paths stay with
  `CheckBoundary`). `filepath` import retained for the `filepath.Join`
  reads and `filepath.ToSlash` in the `..` split.
- `internal/envfragment` other POSIX literals (`testFragment`,
  `TestFragmentJSONCanonical`, `TestFragmentEnvAndShellFormats`,
  `TestFragmentCodexNameChannel`, `TestFragmentEnvCarriesOnlyEnvObject`) —
  not fed to `filepath`; they assert JSON bytes / `env|shell` rendering of
  string values. Portable as-is on all runners; no change.
- `internal/envprofile/managed.go:1798` (`SkillsDir + "/" + skill.name`) —
  marker surface record paths (slash domain), not host paths; no change.
- `internal/envprofile/managed_test.go:67` (`"/tmp/"+profile+"-source"`) —
  JSON body of a context source fixture, not a `filepath` argument at the
  site; the Windows lane's cycle-2 evidence failed only on the schema
  drivers + `TestCheckBoundary`, never here; no change.
- `cmd/curator/builds.go:581` (`root+"/"`) — pre-existing file outside the
  stage-(b) delta; out of scope, not touched.
- Git-config interpolation class (stage-(a) lesson): `git diff
  7320bc2a..HEAD | grep -E '^\+.*(git config|url\.|insteadOf)'` → empty;
  no new fixture interpolates a native path into a git config value.

## B1 suite-plan reproduction (patched table, real roots, own output quoted)

Pinned root materialized read-only:
`git -C /Users/iv/Developer/ReluxWorks/curator-spec archive
0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C
/tmp/specpin-v1` (the `0ed5c691` tree publishes no
`schema-cases/launch-env-fragment-v1` nor
`schema-cases/agent-environment-marker-v1`; full root at
`/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1` =
`f39f4a9` publishes both).

```
$ cd /Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b
$ bash -c "cd $W && CI_ROOT_ARTIFACTS=$W/.github/ci/root-artifacts.tsv \
  CI_PLATFORM_EXCLUSIONS=$W/.github/ci/platform-exclusions.tsv \
  bash $W/.github/ci/suite-plan.sh /tmp/specpin-v1/conformance/v1 /tmp/ev-pin"
suite-plan: GOOS=darwin
suite-plan: root=/tmp/specpin-v1/conformance/v1
suite-plan: platform qualification read from the root's own vectors/conformance-claim-v3-qualification.json

defer internal/envfragment
      the supplied root publishes none of: schema-cases/launch-env-fragment-v1
      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement
defer internal/envmarker
      the supplied root publishes none of: schema-cases/agent-environment-marker-v1
      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement

suite-plan: served=70 deferred=2 excluded=0
suite-plan: ok
```

```
$ bash -c "cd $W && CI_ROOT_ARTIFACTS=$W/.github/ci/root-artifacts.tsv \
  CI_PLATFORM_EXCLUSIONS=$W/.github/ci/platform-exclusions.tsv \
  CI_REQUIRE_FULL_ROOT=1 bash $W/.github/ci/suite-plan.sh \
  /Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 /tmp/ev-main"
suite-plan: GOOS=darwin
suite-plan: root=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1
suite-plan: platform qualification read from the root's own vectors/conformance-claim-v3-qualification.json


suite-plan: served=72 deferred=0 excluded=0
suite-plan: CI_REQUIRE_FULL_ROOT=1 -- every package must be served by this root
suite-plan: ok
```

The three `t.Fatal`s are unchanged: under registration a served root
that stopped publishing either family is a genuine error and stays fatal;
in the candidate lane (`CI_REQUIRE_FULL_ROOT=1`) any deferral is fatal,
so the lane fails closed instead of skipping.

## Gate table (each row: exact standalone command, observed exit code)

All commands run in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`
as standalone processes (no `tee`, no pipe chains hiding the status).

| Gate (exact command) | Root / lane | Observed output | Exit |
|---|---|---|---|
| `go build ./...` | — | (no output) | 0 |
| `go vet ./...` | — | (no output) | 0 |
| `gofmt -l cmd internal` | — | (no output; clean) | 0 |
| `golangci-lint run ./...` | — | `0 issues.` | 0 |
| `bash .github/ci/gate-selftest.sh` | — | `gate-selftest: 81 passed, 0 failed` | 0 |
| `bash .github/ci/no-broad-suppression.sh` | — | `no-broad-suppression: ok` | 0 |
| `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev` | — | `ledger-consistency: 131 rows checked across linux darwin windows` + `ledger-consistency: ok` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 go test -count=1 -race ./internal/envfragment/ ./internal/envmarker/ ./internal/envregistry/` | curator-spec `f39f4a9` | `ok` all three packages | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 go test -count=1 -race ./internal/envprofile/` | curator-spec `f39f4a9` | `ok ... 37.205s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/tmp/specpin-v1/conformance/v1 bash .github/ci/test-gate.sh /tmp/tg-pin` | SPEC_PIN `0ed5c691` | `test-gate: go test exit=0, platform-case gate exit=0`; three schema rows `tol ... (tolerated skip: root-unset)`; 28 skips, 0 `stage-deferred` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/tg-main` | curator-spec `f39f4a9` | `test-gate: go test exit=0, platform-case gate exit=0`; all three schema rows `ok` (served, not skipped); 19 skips, 0 `stage-deferred`, 0 `root-unset` | 0 |
| `go test -count=1 -timeout 30m ./cmd/curator` | — | `ok ... 271.933s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test -count=1 -race -run 'TestProfileUseTargetBoundToAdapter\|TestProfileUseTargetIsStageB\|TestEnvStatusMatrix' ./cmd/curator/` | curator-spec `f39f4a9` | `ok ... 11.523s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test -count=1 -run 'TestConformanceEnvironments' ./internal/interop/` | curator-spec `f39f4a9` | `ok` (`TestConformanceEnvironmentsHeader` + `TestConformanceEnvironmentsMonolithic` pass; stage-(b) sets observed passing by name: `referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`, `system-prompt-composed`, `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none`) | 0 |

## Hosted CI (verbatim; not re-attested)

`gh pr checks 60` at report time — this is the CI state of the
*reviewed* head `7bca4d4b`, not of this revision (the five commits above
are not pushed, per the brief's do-not-push rule, so no lane has run on
them):

```
Race (macos-latest)	fail	16m29s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882638
Race (ubuntu-latest)	fail	9m22s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882673
Test (macos-latest)	fail	8m46s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882675
Test (ubuntu-latest)	fail	3m34s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882653
Test (windows-latest)	fail	39m39s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882654
Gate self-test (macos-latest)	pass	10s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882550
Gate self-test (ubuntu-latest)	pass	7s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882629
Gate self-test (windows-latest)	pass	24s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882690
Interop conformance gate	pass	17s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882641
Lint	pass	41s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882607
Naming gate	pass	9s	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882676
Candidate suite (${{ matrix.os }})	skipping	0	https://github.com/relux-works/curator/actions/runs/34017171892/job/101442882935
```

No claim is made about linux/windows lanes for this revision: all local
runs above are darwin/arm64 (`go1.25.5`). The Windows half of B1b is
established by construction (platform-absolute `filepath.Join` fixtures;
slash semantics only for POSIX evident vectors) plus the stdlib
`volumeNameLen` reading the reviewer confirmed on the hosted lane — not
by a Windows execution here. The `linux`/`windows` platform-case rows are
ledger-checked (`ledger-consistency: 131 rows ... ok`) but executed here
only on darwin; CI owns the other two runners.

## Bounds and non-goals

- Full `go test -race` on `cmd/curator` was not run (full non-race run:
  272s; a race run exceeds the single-call time budget). Race coverage:
  full `-race` on `internal/envfragment`, `internal/envmarker`,
  `internal/envregistry`, `internal/envprofile`, plus `-race` on the
  touched CLI areas (`TestProfileUseTarget*`, `TestEnvStatusMatrix`).
- No candidate-root lane (`CI_REQUIRE_FULL_ROOT=1` against a supplied
  candidate tree): no candidate root was in scope for this rework.
- Nothing deferred: `stage-deferred` occurs 0 times in both lanes'
  `skips-observed.tsv`; the `stage-deferred` class usage this stage
  removed stays removed.
- Scope: B1, B1b, M1, m2, m3 only. F1–F6/F8–F11 behavior untouched beyond
  the fixture/test changes named above; `git diff 7bca4d4b..HEAD --stat`
  (9 files, +138/−78) is confined to stage (b).

