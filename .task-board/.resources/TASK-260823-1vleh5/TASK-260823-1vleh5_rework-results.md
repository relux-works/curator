# TASK-260823-1vleh5 — rework: enforced script commands are refused, not installed

**Status:** ready for review (board `to-review`).
**Branch:** `task/TASK-260823-1vleh5-script-policy-admission`, commit `a74f1f5`,
branched from `origin/main@62d578c` — PR
[relux-works/curator#34](https://github.com/relux-works/curator/pull/34).
**Candidate suite:** `relux-works/curator-spec@6001dc3` (`candidate/schema-8-rc.9`),
`protocol/core.md` §4.1.1, `profiles/manager.md` script diagnostics table.
**Answers:** the CHANGES REQUESTED verdict on PR #33 (`RUN-260823-8f0f4e`).

## The blocking finding, and why it was a real regression

Admitting schema 8 removed the accident that used to stop an enforced script
command. With `SupportedSchemaVersions` capped at 7, a manifest carrying
`execution_policy: "script-worker-v1"` was rejected outright, so no shim could
ever be written for one. Once schema 8 parsed, the same manifest loaded,
`activeScriptCommands` picked the command up as an ordinary script command
(`internal/install/targets.go` branched on `command.Type == "script"` alone),
and a plain shim ran the package code with the caller's full ambient authority.

That is the exact fail-open outcome the profile forbids by name:

> A manager that does not implement this policy MUST reject such a command with
> `script_execution_policy_unsupported`. It MUST NOT install the command
> declared-only, downgrade it, or ignore the field, because the resulting shim
> would run package code the manifest says is contained.

Before this change the diagnostic did not exist anywhere in the tree.

The task description's "declared-only behavior for now" was read on the first
pass as "parse and ignore". Ignoring the field is one of the three things the
profile names as forbidden. Parsing the field and refusing the install satisfies
both the task text ("parsing must not reject valid schema-8 manifests") and the
spec, so this is not a stop-the-line: the instruction and the spec are
reconcilable, they were just reconciled wrongly the first time.

## What changed

### `internal/scriptpolicy` (new package)

Owns the manager side of §4.1.1, in the shape `internal/moduleroots` already
established for §4.2.3: the closed diagnostic plus the manager profile's fixed
`unsupported`/`error` pair, an `Error` type carrying them, a `Code(err)`
extractor that survives wrapping, and `Admit`.

`Admit` refuses the first enforced command in **lexical** order, not map order,
so the same manifest names the same command on every host and every run.
`Enforced` is the single predicate: only script commands can carry a policy and
`script-worker-v1` is the single closed value, so any non-empty policy is an
enforced command.

Refusal is unconditional, because this manager has no worker. The package
documents that this is the one place the refusal is replaced when the worker
lands (`STORY-260822-2h0v9j`), and that nothing above it decides the policy.

### The split the review asked for

`skillspec.Load` **keeps accepting** the manifest. A schema-8 document that
selects the policy is a valid document — the published cases mark
`valid-script-worker-enforced.json` valid — and document validity is a separate
question from what this manager will admit. Parsing was left exactly as it
landed in #33.

Admission is now answered one layer down, by two consumers:

| Layer | Surface it protects |
| --- | --- |
| `skillcheck.Validate` → `executionPolicyIssues` | both install scopes, through the shared `validateNodes`; also `curator skill check` |
| `install.activeScriptCommands` | the shim writer itself |

The shim writer refuses on its own rather than trusting the preflight that
normally runs before it, because the only thing it could do with an enforced
command is write the uncontained launcher — it must not be the layer that
decides. An enforced command that no edge activates is not staged at all, so the
guard does not disturb an unrelated declared-only command in the same skill.

`skillcheck` emits the closed protocol code verbatim rather than a
`skill.`-namespaced one; namespacing it would take it out of the closed set.

### Non-blocking finding: the schema case that proved nothing

`materializeManifestFixture` now materializes declared `runtime_roots`.
`invalid-script-worker-missing-path.json` was previously rejected by
`runtime_roots[0]: runtime root does not exist: scripts` — a reason that is true
of the **valid** cases too, so the case proved nothing about its own rule and
`TestReleasedSchemaCases`, which only compares valid/invalid, stayed green
regardless.

Verified by instrumenting the family and printing the actual error for all 18
`script-worker` cases. Every one now fails for the rule it was written to test:

| Case | Rejection reason after the fix |
| --- | --- |
| `invalid-script-worker-missing-path` | `commands.enforced-tool: script command requires 'unix_path' or 'win_path'` |
| `invalid-script-worker-missing-interpreter` | `…execution_policy: requires 'interpreter'` |
| `invalid-script-worker-interpreter-without-policy` | `…interpreter: requires 'execution_policy'` |
| `invalid-script-worker-unknown-interpreter` | `…interpreter: must be one of node-v1, python3-v1` |
| `invalid-script-worker-{compiled,hardened,null,opt-out,successor}-policy` | `…execution_policy: must be "script-worker-v1"` |
| `invalid-script-worker-on-build-command` | `…execution_policy: field is not supported for build commands` |
| `invalid-script-worker-on-system-command` | `commands.system-tool: has unsupported field(s): execution_policy, interpreter` |
| `invalid-script-worker-top-level-{execution-policy,interpreter}` | `agent-skill.json: has unsupported field(s): …` |
| all five `valid-script-worker-*` | accepted |

## Tests

New: `internal/scriptpolicy/scriptpolicy_test.go` (diagnostic code, the
`unsupported`/`error` pair, field path, lexical determinism over 32 iterations,
`Code` through wrapping, declared-only control),
`internal/install/scriptpolicy_test.go` (end-to-end refusal, declared-only
control, shim-writer guard), and two cases in
`internal/skillcheck/skillcheck_test.go`.

### Mutation check — each guard was disabled in turn

A test that passes because *something* rejects is not evidence that the guard
under test rejects. Each layer was disabled and the end-to-end test re-run:

| Mutation | `TestEnforcedScriptCommandIsRefusedAtInstall` |
| --- | --- |
| both guards active | pass |
| `skillcheck` guard disabled | pass — the shim writer catches it |
| shim-writer guard disabled | pass — `skillcheck` catches it |
| **both disabled** | **fail** — `Status:ok`, shim published |

The both-disabled run reproduces the reviewed regression exactly (`Status:ok`,
`commands=[enforced-skill-tool]`, launcher written), so the test is not vacuous
and each layer is independently sufficient.

## Gate evidence

Every command run as a standalone process; these are the real exit codes.

| Gate | Root | Exit |
| --- | --- | ---: |
| `golangci-lint run` | — | 0 (0 issues) |
| `gofmt -l cmd internal` | — | clean |
| `go vet ./...` | — | 0 |
| `go test ./internal/...` | candidate `6001dc3` | 0 |
| `make check-ci` | released pin `00b1688` | 0 |
| `make candidate-test` (`CI_REQUIRE_FULL_ROOT=1`) | candidate `6001dc3` | 0 |
| `make race` | released pin `00b1688` | 0 |
| `make gate-selftest` | released pin `00b1688` | 0 (81 passed, 0 failed) |
| `make no-broad-suppression` | — | 0 |

`internal/skillspec` defers against the released pin (it publishes no
`schema-cases/agent-skill-v8`), so the candidate lane is the one that actually
exercises this change; it was run with `CI_REQUIRE_FULL_ROOT=1`, where a
deferral is fatal.

Remote: PR #34 default lanes, plus a manually dispatched candidate-conformance
run on `candidate_ref=6001dc3…` with the manifest digest expectation wired.

## Scoped out, deliberately

- **The worker itself.** `script-worker-v1` containment stays with
  `STORY-260822-2h0v9j`. This change makes the interim behaviour safe and gives
  that story a single place to replace.
- **The audit warning classes** `script-command-declared-only` and
  `script-command-unfiltered-declared-network` (§4.1.1, audit surface). Not named
  by this task's AC and not part of the review's required fix.
- **Suite-consumption assertions** beyond the fixture fix — that is
  `TASK-260823-2u5xov`'s scope. What is fixed here is only the fixture defect the
  review named, so the case now proves its own rule.
