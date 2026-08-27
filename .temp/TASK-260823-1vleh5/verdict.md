# TASK-260823-1vleh5 — review verdict: ACCEPTED

**Reviewer run:** `RUN-260823-2c49ba` (not goal-bound; `task-board spawn goal` → "run is not goal-bound").
**Under review:** `origin/main@77aafa0` (merge of PR
[#34](https://github.com/relux-works/curator/pull/34), rework commit `a74f1f5`, branched
from `62d578c`). `git diff a74f1f5 77aafa0` is empty — the merge carries no extra delta.
**Answers:** the CHANGES REQUESTED verdict of `RUN-260823-8f0f4e`
(`TASK-260823-1vleh5_review-verdict.md`), which reviewed `62d578c` / PR #33.
**Candidate suite:** `relux-works/curator-spec@6001dc3` (`candidate/schema-8-rc.9`).

## Verdict

Accepted. The blocking finding is fixed at the layer the previous review named, the
fix is spec-faithful rather than merely test-satisfying, the non-blocking finding is
fixed too, and every gate is green. The module-roots half accepted last round is
byte-unchanged by the rework.

---

## The blocking finding is closed

**What it was.** Admitting schema 8 removed the accident that used to stop an enforced
script command (`SupportedSchemaVersions` capped at 7). After `a27f78f` a manifest
carrying `execution_policy: "script-worker-v1"` loaded, `activeScriptCommands` picked
the command up on `command.Type == "script"` alone, and a plain shim ran the package
code with the caller's full ambient authority — the fail-open outcome
`profiles/manager.md` forbids by name.

**What landed.** `internal/scriptpolicy` (new) owns the closed diagnostic and the
refusal; `skillspec.Load` still accepts the manifest, unchanged from #33.

**The split is the one the spec asks for, not a compromise.** I checked the source
text at `6001dc3` rather than taking it from either artifact. `protocol/core.md:571`
and `profiles/manager.md:901` give the condition as *"a manager that does not
implement `script-worker-v1` **reads** a command that selects it"* — the trigger is
the read, not the activation. So refusing at `skillcheck.Validate` (which reads the
manifest, whether or not an edge activates the command) is literally the specified
boundary, and my initial worry that a declared-but-inactive enforced command should
not fail the whole install is wrong on the spec text. `profiles/manager.md:900` also
fixes the pair as `unsupported`/`error`; `scriptpolicy.StateUnsupported` /
`SeverityError` match exactly, and `TestAdmitRefusesEnforcedCommandWithClosedDiagnostic`
pins both. `skillcheck.Issue` carries no `state` field at all — no diagnostic on that
surface does — so dropping `State` there is consistent with the codebase, and the
`Error` still carries it for the surface that will need it.

**Both gates reach both install scopes.** `validateNodes` is called from
`internal/install/install.go:387` (project) and `internal/install/global.go:139`
(global); an `error`-severity issue calls `result.failf` and returns before any
staging. `cmd/curator/main.go:1050` (`curator skill check`) returns `exitFail`
through `skillcheck.HasErrors`.

**No second shim path exists.** `internal/install/targets.go` is the only site that
turns a skill command into a `runtimestore.ShimSpec`. `internal/globalbins/stage.go`
publishes forwarding shims from the already-installed canonical bin directory, i.e.
strictly downstream of the refusal.

### Mutation matrix — reproduced independently, not taken from the delivery note

Each guard disabled in a throwaway copy of the worktree, then restored:

| Mutation | `TestEnforcedScriptCommandIsRefusedAtInstall` |
| --- | --- |
| both guards active | pass |
| `skillcheck` guard removed | pass — the shim writer catches it |
| shim-writer guard short-circuited (`if false && …`) | pass — `skillcheck` catches it |
| **both disabled** | **fail** — `Status:ok`, `commands=[enforced-skill-tool]`, launcher published |

The both-disabled run reproduces the reviewed regression verbatim, so the test is not
vacuous and each layer is independently sufficient. Tree restored to `77aafa0`
afterwards (`git status --short` empty).

## The non-blocking finding is closed

`materializeManifestFixture` now materializes declared `runtime_roots`. I re-ran the
family probe myself over **both** schema-8 families (`agent-skill-v8` and
`csk-skill-v8`, 61 cases each side) and printed the actual rejection string for every
case. Every invalid case is now rejected by the rule it was written to test, including
the one the last review named:

| Case | Rejection reason |
| --- | --- |
| `invalid-script-worker-missing-path` | `commands.enforced-tool: script command requires 'unix_path' or 'win_path'` |
| `invalid-script-worker-missing-interpreter` | `…execution_policy: requires 'interpreter'` |
| `invalid-script-worker-interpreter-without-policy` | `…interpreter: requires 'execution_policy'` |
| `invalid-script-worker-unknown-interpreter` | `…interpreter: must be one of node-v1, python3-v1` |
| `invalid-script-worker-{compiled,hardened,null,opt-out,successor}-policy` | `…execution_policy: must be "script-worker-v1"` |
| `invalid-script-worker-on-build-command` | `…execution_policy: field is not supported for build commands` |
| `invalid-script-worker-on-system-command` | `commands.system-tool: has unsupported field(s): execution_policy, interpreter` |
| `invalid-script-worker-top-level-{execution-policy,interpreter}` | `<manifest>: has unsupported field(s): …` |
| all 15 `valid-*` | accepted |

The 13 `invalid-module-roots-*` cases are likewise each rejected by their own rule
(`build_module_root_declaration_invalid` for the five spelling ones, `has unsupported
field(s): modules` for the three misplacement ones, and so on). The probe file was
deleted after the run.

## Module-roots half — unchanged, still fails closed

`git diff a27f78f a74f1f5 -- internal/moduleroots internal/godriver internal/skillspec/parse.go
internal/skillspec/types.go internal/marker` is **empty**: the rework touched nothing the
previous review accepted. Re-confirmed the handoff is safe rather than re-deriving the
whole accepted analysis: `internal/skillspec/parse.go:938` is the only non-test consumer
of `moduleroots` (`ValidateDeclaration` at parse time), while
`internal/godriver/graph.go:306` still rejects every `item.Module.Replace != nil` with
`vendor_metadata_inconsistent`. A declared module root therefore cannot yet produce an
unvalidated build — the wrong diagnostic, but no fail-open. `EffectiveReplaceSet` /
`ValidateBijection` are implemented and tested but not yet wired into the driver; that
is `TASK-260823-1wvgw8`'s scope by design.

## Gates — this reviewer, worktree at `77aafa0`, each a standalone process

| Gate | Root | Result |
| --- | --- | --- |
| `gofmt -l cmd internal` | — | empty |
| `go build ./...` | — | 0 |
| `go vet ./...` | — | 0 |
| `golangci-lint run ./...` (v2.12.2) | — | 0 — "0 issues." |
| `go test ./internal/{install,scriptpolicy,skillcheck,skillspec}` | candidate `6001dc3` | 0 |
| `bash .github/ci/gate-selftest.sh` | — | 81 passed, 0 failed |
| `CI_REQUIRE_FULL_ROOT=1 GO_TEST_TIMEOUT=30m bash .github/ci/test-gate.sh` | candidate `6001dc3` | **served=43 deferred=0 excluded=0**, `go test exit=0`, `platform-case gate exit=0`, 10 skips, 0 failures |

served is 43, not the 42 of the last round, because `internal/scriptpolicy` is a new
package the root serves unconditionally — the count moved for the expected reason.

### Remote

- **PR #34** (`a74f1f5`): all eleven default required lanes green — Test and Race on
  ubuntu/macos/windows, Lint, Naming gate, Interop conformance gate, Gate self-test ×3.
- **Candidate matrix** (`workflow_dispatch` run `32668086905`, same sha `a74f1f5`):
  `Candidate suite` green on **ubuntu, macos and windows**, with
  `CANDIDATE_REF: 6001dc33281b94a4ec7442ab15278550dd0f51d9`, the immutable-revision gate
  accepting it, and `candidate-suite: manifest digest matches the supplied expectation`.
  Verified in the job log, not from the delivery note.

## The `main` push run: a known Windows flake, not this change

`push` run `32669423117` on `77aafa0` first reported `Test (windows-latest)` failed.
Diagnosed rather than waved off:

- The failing test is `internal/managerlock :: TestSubprocessContentionAndIndependentProjects`
  — `managerlock_test.go:465: independent project helper = "blocked", want acquired`,
  followed by a `TempDir RemoveAll` failure on a `.lock` file still held by another
  process. Recovered from the run's `test-evidence-windows-latest` artifact
  (`go-test-served.json`); the gate does not echo failing tests into the job log, only
  `stage served exit=1`.
- `internal/managerlock` is not in this task's diff (`git diff --name-only e17b0f1 77aafa0`
  → `.github/ci`, `internal/{install,marker,moduleroots,scriptpolicy,skillcheck,skillspec}`),
  and `GOOS=windows go list -deps ./internal/managerlock` contains exactly **one**
  `relux-works/curator` package — itself. There is no edge from the change to the failure.
- The identical tree passed the same lane twice on `a74f1f5` (PR run `32668077768`,
  dispatch run `32668086905`).
- `LOGBOOK.md` already records this exact failure mode as a Windows flake in the sibling
  test `TestSubprocessBuildKeyDeduplicationAcrossProjects` ("independent build key helper
  = blocked, want acquired"), same day, same runner class.

I re-ran the failed job (`gh run rerun 32669423117 --failed`) rather than accept on
inference. Result is recorded in the "Rerun" line below.

## Follow-ups (none blocking this task)

1. **The marker v3/v4 cross-field gap still has no board item.** `marker.Read` accepts
   five cases the published suite marks invalid, in both `install-marker-v3` and
   `install-marker-v4` (LOGBOOK 0130). The last review asked for its own item; a board
   search finds none. `TASK-260823-2u5xov`'s AC is suite-consumption assertions plus
   candidate qualification, which does not cover it. **Coordinator action**, not rework
   here: the change tightens acceptance of already-installed markers and needs its own
   review. `TASK-260823-1vleh5` handles it correctly in the meantime —
   `markerV4CasesThisReaderDoesNotModel` asserts the list in both directions, so a case
   that starts being rejected fails the test and names the entry to delete.
2. **Audit warning classes are owned.** `script-command-declared-only` and
   `script-command-unfiltered-declared-network` (§4.1.1) are not implemented, and the
   rework scoped them out correctly: `STORY-260822-2h0v9j`'s description names "audit
   warning class for declared-only legacy script commands" explicitly. No orphan.
3. **Gate legibility, cosmetic.** `test-gate.sh` prints `stage served exit=1` without
   naming the failing test; diagnosing the `main` failure required downloading the
   evidence artifact and parsing `go-test-served.json`. A one-line summary of failing
   `Test`/`Package` pairs would save that round trip. Not this task's scope.
4. **A logbook entry for the resolution** would close the loop on 0131, which currently
   records only the finding and the routing decision.

## Acceptance evidence for the commit-owning mover

Work is already on `main` (`77aafa0`), so there is nothing left to commit for this task.
Both PRs (#33, #34) are merged; the reviewer holds no uncommitted delta — the review
worktree at `77aafa0` is clean.
