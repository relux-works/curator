# TASK-260823-1vleh5 — review verdict: ACCEPTED (`done`)

**Reviewer run:** `RUN-260823-047c36` (not goal-bound).
**Under review:** `origin/main@77aafa0` — merge of PR
[#34](https://github.com/relux-works/curator/pull/34), implementation commit `a74f1f5`,
branched from `origin/main@62d578c` (the reviewed state of PR #33).
`git diff a74f1f5 77aafa0` is empty: the merge carries no extra delta.
**Candidate suite:** `curator-spec@6001dc3` (`candidate/schema-8-rc.9`).
**Answers:** the CHANGES REQUESTED verdict of `RUN-260823-8f0f4e`.

## Verdict

The blocking finding is fixed, and fixed at the layer the previous review named.
The module-roots half was already accepted and independently verified in that
verdict; nothing in this delta touches it. Accepting.

---

## The blocking finding is closed

The previous verdict's finding was fail-open: admitting schema 8 removed the
accident (`SupportedSchemaVersions` capped at 7) that used to stop an enforced
script command, so a manifest selecting `script-worker-v1` loaded, was picked up
by `activeScriptCommands` as an ordinary script command, and got a plain shim
that ran package code with the caller's full ambient authority — the outcome
`protocol/core.md` §4.1.1 and `profiles/manager.md` (candidate `6001dc3`, lines
906-909) forbid by name.

The split the review asked for is exactly what landed:

- `skillspec.Load` **still accepts** the manifest. Verified: all 66
  `agent-skill-v8` cases still resolve valid/invalid correctly, including the
  five `valid-script-worker-*` cases.
- Admission moved one layer down into a new `internal/scriptpolicy`, which owns
  the closed diagnostic `script_execution_policy_unsupported` and the manager
  profile's fixed `unsupported`/`error` pair (`profiles/manager.md:901` —
  checked against the table, both values match).

### Both consuming layers verified, and each is load-bearing

| Layer | Reaches | Verified |
| --- | --- | --- |
| `skillcheck.executionPolicyIssues` | project install (`install.go:387`), global install (`global.go:139`), and `curator skill check` (`main.go:1050`) | `validateNodes` maps `severity: error` to `result.failf` and returns before any staging in both scopes |
| `install.activeScriptCommands` | the shim writer itself | both `stageRuntimeAndShims` call sites (`install.go:677`, `global.go:380`) propagate the error out of target staging |

I re-ran the implementer's mutation matrix myself in a scratch copy, disabling
each guard in turn and re-running `TestEnforcedScriptCommandIsRefusedAtInstall`:

| Mutation | Result |
| --- | --- |
| both guards active | pass |
| `skillcheck` guard disabled | pass — the shim writer catches it |
| shim-writer guard disabled | pass — `skillcheck` catches it |
| **both disabled** | **FAIL** — `Status:ok`, `commands=[enforced-skill-tool]`, launcher published |

The both-disabled run reproduces the reviewed regression byte for byte, so the
test is not vacuous and neither layer is decoration.

### No remaining path to an uncontained enforced shim

Checked exhaustively rather than assumed:

- `stageRuntimeAndShims` is the only shim producer, and has exactly two callers,
  both behind `validateNodes`.
- `globalbins/stage.go:110` only mirrors an **already published** canonical
  global shim into the user bin, so an enforced command that never got a
  canonical shim has nothing to forward.
- `skillspec.Command.ExecutionPolicy` is written only by `parse.go:314`, only for
  script commands, only at schema ≥ 8, and only with the closed value. The
  compiled `manager-worker-v1` identity lives on `marker.Build`/`buildmeta`, a
  different type — so `scriptpolicy.Enforced`'s "any non-empty policy" predicate
  cannot catch a `go-v1` build command. Confirmed by reading every
  `.ExecutionPolicy` site in the tree.
- Determinism holds on both layers: `Admit` sorts, and `node.ActiveCommandNames()`
  (`closure.go:99-107`) is already sorted, so the shim-writer refusal names the
  same command on every host and run too.

### Curator publishes no conformance claim

Worth stating, because a manager that admits schema 8 without implementing
§4.1.1 is non-conforming on that axis: `.github/ci/candidate-suite.sh:170` emits
`conformance_claim none` and `gate-selftest.sh:145` pins it. Nothing in the tree
claims what this refusal denies.

## Non-blocking finding of the previous review is also fixed, and I verified it

`materializeManifestFixture` now materializes declared `runtime_roots`. I
instrumented the whole `agent-skill-v8` family in a scratch copy and printed the
real rejection reason for all 66 cases. Every invalid case now fails for the
rule it was written to test, including the one the previous review named:

- `invalid-script-worker-missing-path` → `commands.enforced-tool: script command requires 'unix_path' or 'win_path'` (was `runtime_roots[0]: runtime root does not exist: scripts`)
- `invalid-script-worker-{missing-interpreter,interpreter-without-policy,unknown-interpreter}` → the co-requirement and closed-interpreter rules
- `invalid-script-worker-{compiled,hardened,null,opt-out,successor}-policy` → `must be "script-worker-v1"`
- `invalid-script-worker-on-build-command` → `field is not supported for build commands`
- `invalid-script-worker-on-system-command` / `top-level-*` → unsupported field
- all 15 `invalid-module-roots-*` → their own module-root diagnostics
- all 15 valid cases → accepted

The fixture change can only fail in the direction the test catches: over-materialising
flips an invalid case to accepted and reddens `TestReleasedSchemaCases`.

## Architecture fit

`internal/scriptpolicy` mirrors `internal/moduleroots` line for line in shape —
package doc quoting the normative sentence, closed diagnostic constants with
per-code comments, an `Error{DiagnosticCode, Path, Detail}` with the same
`Error()` rendering, and a `Code(err)` extractor that survives wrapping. Same
§4.2.3/§4.1.1 symmetry, one package per spec section. The refusal is
unconditional and documented as the single replacement point for
`STORY-260822-2h0v9j`, which is where the worker lands.

## Gates — this reviewer, worktree at `77aafa0`, each a standalone process

| Gate | Root | Result |
| --- | --- | ---: |
| `gofmt -l cmd internal` | — | clean |
| `go build ./...` | — | 0 |
| `go vet ./...` | — | 0 |
| `golangci-lint run ./...` (v2.12.2) | — | 0 — "0 issues." |
| `bash .github/ci/gate-selftest.sh` | released | 81 passed, 0 failed |
| `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` | candidate `6001dc3` | **0** — `served=43 deferred=0 excluded=0`, `go test exit=0`, `platform-case gate exit=0` |

`served=43` is one more than the 42 of the previous review: the new
`internal/scriptpolicy` package is served by the candidate root, not deferred.
10 skips recorded, every one a pre-existing class (`allowed-host-capability`,
`allowed-opt-in`, `tolerated-by-ledger`, `allowed-helper-process`) — no new skip
class hides this change.

### Remote

- PR #34 default lanes: **11/11 pass** (Test ×3, Race ×2, Lint, Naming gate,
  Interop conformance gate, Gate self-test ×3). Race is `ubuntu`+`macos` by
  design (`ci.yml:158`), so 11 is the complete default set, not a missing lane.
- Dispatched candidate run
  [32668086905](https://github.com/relux-works/curator/actions/runs/32668086905)
  on `a74f1f5` with `CANDIDATE_REF=6001dc33281b94a4ec7442ab15278550dd0f51d9`:
  **14/14 pass**, including **Candidate suite on ubuntu, macos, and windows**.
  I read the ref out of the job log rather than trusting the delivery note.
  This is stronger pre-merge evidence than PR #33 had.

## Findings carried forward — none blocking, none owned

1. **The two audit warning classes have no board item.**
   `profiles/manager.md:1008` — "Two warning classes are REQUIRED" —
   `script-command-declared-only` and `script-command-unfiltered-declared-network`.
   `grep` finds neither string anywhere in the tree. This is **not** a regression
   from this change: `script-command-declared-only` applies to every script
   command of schema 7 and earlier too, so the gap predates schema 8 and is not
   widened by it, and `script-command-unfiltered-declared-network` is unreachable
   while every enforced command is refused. Correctly scoped out here — but
   nothing on the board owns it. It is not in `TASK-260823-2u5xov`'s scope
   (suite consumption + qualification) and it is not the worker.
2. **The marker v3/v4 cross-field gap still has no board item**, as the previous
   verdict required. `task-board grep` over the whole board finds no item for
   the five `install-marker-v{3,4}` cases the reader accepts.
3. **The two guards disagree about scope, deliberately.** `skillcheck` refuses a
   package that *declares* an enforced command; the shim writer only refuses one
   that is *active*. Since `skillcheck` runs first and is stricter, a
   declared-but-inactive enforced command fails the whole install. That is
   fail-closed and defensible — a manager that cannot honour the manifest should
   not partially honour it — and it cannot regress anything, because schema 8 is
   new. Stating it so the worker story knows both layers exist with different
   scopes.
4. **`Admit` reports only the first enforced command per skill**, so an operator
   with three of them fixes one and re-runs. Deliberate (it buys the lexical
   determinism that is tested), and a nit, not a defect.
5. **Repo-wide doc staleness, pre-existing, not this task's.** `README.md:36`
   still says "schemas 1 through 5" while `SupportedSchemaVersions` now spans 1-8;
   it has been stale since schema 6. `CHANGELOG.md`'s `Unreleased` section is
   empty and the file has not been touched in 68 commits on `main`. Neither is a
   fair charge against this delta, but admitting a manifest schema version is the
   kind of thing that CHANGELOG's own first line promises to record.

## On the stop-the-line question

The previous verdict said the first pass silently resolved a contradiction
between the task text ("declared-only behavior for now") and the profile
sentence that names "ignore the field" as forbidden. The rework's answer is
correct: the two are reconcilable — parse the field (task text: "parsing must
not reject valid schema-8 manifests"), refuse the install (profile) — so this
was ordinary rework, not a stop-the-line. The delivery artifact says so
explicitly and quotes the sentence it previously did not mention.

## Acceptance evidence for the commit-owning mover

Nothing to commit: `a74f1f5` is already merged to `main` as `77aafa0`, and this
reviewer-archetype run supplies no `commit_ack`. The task's AC — "Schema-8
manifests parse and validate per the candidate rules; schema-case families for
agent-skill-v8 consumed by tests; merged to main green" — is met on all three
clauses, with the candidate suite green on all three platforms at the exact
implementation commit.
