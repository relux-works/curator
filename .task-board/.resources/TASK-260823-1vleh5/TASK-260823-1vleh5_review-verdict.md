# TASK-260823-1vleh5 — review verdict: CHANGES REQUESTED (`to-dev`)

**Reviewer run:** `RUN-260823-8f0f4e` (not goal-bound).
**Under review:** `origin/main@62d578c` (merge of PR
[#33](https://github.com/relux-works/curator/pull/33), implementation commit `a27f78f`,
branched from `e17b0f1`). `git diff a27f78f 62d578c` is empty — the merge carries no
extra delta.
**Candidate suite:** `curator-spec@6001dc3` (`candidate/schema-8-rc.9`).

## Verdict

The module-roots half of this task is correct, well-tested, and independently
verified. It is **not** accepted, because the schema-8 script surface that shipped
alongside it puts curator into exactly the state `profiles/manager.md` forbids by
name: it admits `execution_policy: "script-worker-v1"`, installs the command
declared-only, and runs it uncontained.

Rework is small and well-defined, so this is `to-dev`, not `blocked`.

---

## Blocking finding — an enforced script command installs and runs uncontained

`profiles/manager.md`, immediately below the script diagnostics table (candidate
`6001dc3`, lines 908-911):

> A manager that does not implement this policy MUST reject such a command with
> `script_execution_policy_unsupported`. It MUST NOT install the command
> declared-only, downgrade it, or ignore the field, because the resulting shim
> would run package code the manifest says is contained.

`protocol/core.md` §4.1.1 is the matching MUST: "every conforming manager MUST
implement it on macOS, Linux, and Windows."

**What landed.** `internal/skillspec/parse.go:305` parses `execution_policy` and
`interpreter`, validates them against the closed constant and the closed
interpreter set, and stores them on `Command` (`internal/skillspec/types.go:53-59`).
Nothing else in the tree reads either field:

- `grep -rn '\.ExecutionPolicy\b' --include='*.go' cmd internal | grep -v _test`
  returns only `buildmeta`/`godriver`/`marker`/`install` hits for the unrelated
  compiled `manager-worker-v1` identity, plus the two `skillspec` definition sites.
- `grep -rn '\.Interpreter\b' --include='*.go' cmd internal | grep -v _test` returns
  **nothing**.
- `grep -rn 'script_execution' --include='*.go' cmd internal` returns **nothing** —
  the closed diagnostic `script_execution_policy_unsupported` does not exist in
  this codebase.
- Shim generation branches on the command type alone:
  `internal/install/targets.go:266` is `if command.Type == "script" && active[name]`.
  There is no policy branch anywhere on that path.

**Failure scenario.** A skill publishes `schema_version: 8` with

```json
"commands": { "tool": { "type": "script", "unix_path": "scripts/tool",
  "execution_policy": "script-worker-v1", "interpreter": "python3-v1" } }
```

(the suite's own `schema-cases/agent-skill-v8/valid-script-worker-enforced.json`).
Before this commit curator rejected the manifest outright — `SupportedSchemaVersions`
stopped at 7, so no enforced command could ever be installed. After this commit
`Load` succeeds, `activeScriptCommands` picks the command up, an ordinary shim is
written, and `scripts/tool` executes with the caller's full ambient authority. The
manifest says the command is contained; the manager silently runs it uncontained.
That is a fail-**open** regression introduced by this change, and it is the precise
outcome the profile sentence exists to prevent.

Contrast the module-roots half, which fails **closed**: `modules` is admitted at
parse time, and `internal/godriver/graph.go:306` still rejects every
`item.Module.Replace != nil` with `vendor_metadata_inconsistent`, so a declared
module root cannot produce an unvalidated build. That handoff to
`TASK-260823-1wvgw8` is safe. The script handoff to `STORY-260822-2h0v9j`
(status `backlog`, worker unimplemented) is not, and no board item owns the interim
behaviour.

**Required fix.** Keep `skillspec.Load` accepting the manifest — the schema-8
conformance cases mark `valid-script-worker-enforced.json` valid, and the task
description is explicit that parsing must not reject valid schema-8 manifests. The
rejection belongs one layer down, at command admission / shim installation: refuse
any command carrying `ExecutionPolicy == skillspec.ScriptExecutionPolicy` with the
closed diagnostic `script_execution_policy_unsupported` (state `unsupported`,
severity `error`) for as long as curator does not implement the worker. Add tests
pinning (a) the refusal and its diagnostic code, and (b) that a schema-8
**declared-only** script command still installs unchanged.

The task description's "declared-only behavior for now" was read as "parse and
ignore". Ignoring the field is one of the three things the profile names as
forbidden; parsing it and refusing the install satisfies both the task text and the
spec. This conflict should have been surfaced rather than resolved silently —
it is a stop-the-line-class contradiction between the instruction and the spec
being implemented, and the delivery artifact does not mention the profile rule at
all.

## Non-blocking finding — one schema case is rejected for the wrong reason

`schema-cases/agent-skill-v8/invalid-script-worker-missing-path.json` declares a
script command with `execution_policy`/`interpreter` and no `unix_path`/`win_path`.
Under `materializeManifestFixture` it is rejected by
`runtime_roots[0]: runtime root does not exist: scripts`, not by the rule it exists
to test. `TestReleasedSchemaCases` only compares valid/invalid, so it passes.

The parser does implement the rule correctly (`parse.go:296`,
`"script command requires 'unix_path' or 'win_path'"`), and `parse_test.go` covers
it directly, so this is precision of the suite consumption, not a defect. Suggest
materializing declared `runtime_roots` in the fixture so the case proves its own
rule. Worth folding into `TASK-260823-2u5xov` (suite-consumption assertions) if not
fixed here.

Every other module-roots and script-worker case is rejected for its own reason —
verified by instrumenting the family and printing the actual error for all 33
`module-roots`/`script-worker` cases.

## Findings acknowledged, correctly scoped out

Both were self-reported, are genuinely pre-existing, and are recorded in `LOGBOOK.md`
(entries 0130 and 0128). I verified the first independently: the five
`install-marker-v4` cases listed in `markerV4CasesThisReaderDoesNotModel` exist in
`schema-cases/install-marker-v3` with byte-identical bodies once `schema_version`
and `skill_schema_version` are removed, so marker v4 inherits the gap exactly rather
than widening it. The both-directions assertion is the right shape for the allowance.

Neither is covered by `TASK-260823-2u5xov`, whose scope is suite-consumption
assertions plus candidate qualification. **The marker cross-field gap needs its own
board item** — it is a reader that accepts five markers the published suite marks
invalid.

## What I verified and accept

Independently checked line by line against `protocol/core.md` §4.2.3 and the
`profiles/manager.md` diagnostics table:

- **Failure boundary.** `ValidateDeclaration` before `go list`;
  `EffectiveReplaceSet` + `ValidateBijection` after it and before `go build`.
  `TestModuleRootVectors` asserts each vector against the correct half rather than
  just the code, and hard-fails on a `link_paths` vector it does not materialise.
- **Selection-annotation reconciliation.** The premise the whole no-versioned-left
  rule rests on — that Go writes both the one-token directive annotation and the
  two-token selection annotation — I reproduced empirically rather than taking from
  the prose. A real `go mod vendor` (go1.25.5 darwin/arm64) over a module with
  `replace example.com/board => ../../pkg/board` writes both
  `# example.com/board v0.0.0 => ../../pkg/board` and
  `# example.com/board => ../../pkg/board`. The implementation's whole-annotation
  match (module path *and* right side) is also the correct trap-avoidance.
- **Containment.** Portable relative path other than `.`, per-component `Lstat` so a
  link cannot redirect the check outward, real regular `go.mod` directly inside,
  uniqueness, pairwise disjointness against other declarations / build roots /
  runtime roots under both exact and `NFD ∘ fold ∘ NFD` comparison. Component-wise
  `contains` so `pkg/boarding` is not read as below `pkg/board`.
- **Scoping.** `modules` is admitted only on local `go-v1` build commands
  (`go-repository-v1` returns before `rejectUnknownBuildFields`), rejected as
  unknown on schemas 2-7, rejected at top level, and ignored — never read as an
  enforcement claim — under schema 1's deployed extension tolerance. Per-command
  declarations are validated per command, which is right: the bijection forces
  commands sharing a build root into equal lists.
- **Marker v4.** Genuinely not optional — without it `marker.Write` fails closed on
  every schema-8 install. `TestMarkerV4BandIsExact` keeps v4↔schema-8 and
  v3↔schema-7 bound in both directions.
- **`.github/ci/root-artifacts.tsv`** declares all three newly-read artefacts, so a
  root that stops publishing them defers the package instead of going red on a
  missing file.

## Local gates (this reviewer, worktree at `62d578c`, each standalone)

| Gate | Result |
| --- | --- |
| `gofmt -l cmd internal` | empty |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run ./...` (v2.12.2) | 0 — "0 issues." |
| `go test ./internal/... -count=1 -timeout 30m` (candidate root exported) | 0, no failures |
| `bash .github/ci/gate-selftest.sh` | 0 — 81 passed, 0 failed |
| `bash .github/ci/suite-plan.sh <6001dc3 root> <evidence>` | 0 — served=42 deferred=0 excluded=0 |
| `CURATOR_CONFORMANCE_ROOT=<6001dc3> CI_REQUIRE_FULL_ROOT=1 GO_TEST_TIMEOUT=30m bash .github/ci/test-gate.sh <evidence>` | 0 — go test exit=0, platform-case gate exit=0 |

All 42 served packages recorded `pass`, none `fail`. The gate recorded 10 skips and
`platform-case gate: ok` — no new skip class. That reproduces the implementer's
reported evidence exactly.

PR #33 CI: all twelve required lanes green (Test and Race on ubuntu/macos/windows,
Lint, Naming gate, Interop conformance gate, Gate self-test ×3). "Candidate suite" is
`workflow_dispatch`-only (`ci.yml:312`) and skipping it on a PR is expected; the
candidate matrix for `6001dc3` was qualified separately under `TASK-260822-c0rxj7`.

The evidence is real and the gates are honest. The gap is that no gate can catch a
diagnostic the codebase never implements.
