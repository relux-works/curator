# TASK-260822-3j4bcb — review verdict: ACCEPTED

Reviewer run `RUN-260822-bc9f3a`. Read-only; no code changed.

## Scope reviewed

One production edit, `internal/closure/closure.go:288-291` (`resolveNode`):

```go
-  return nil, fmt.Errorf("%s: %w", item.name, err)
+  return nil, fmt.Errorf("invalid skill manifest for %s %s %s -> %s (via %s): %w",
+      item.name, resolved.Kind, resolved.Ref, short(resolved.Commit), item.chain, err)
```

plus tests in `internal/closure`, `internal/install`, `internal/skillcheck`.

Working tree also carries unrelated uncommitted changes from other tasks
(`internal/skillspec/parse.go` + `parse_test.go` → `TASK-260822-5wfdfx`;
`internal/config/config.go` → cocoaskills parity). Not this task's scope; the
commit-owning mover must split them.

## AC verification (independent, not taken from the implementer's notes)

| AC clause | Verdict | Evidence |
|---|---|---|
| Test with a broken transitive manifest asserts name+ref+chain | PASS | `TestTransitiveManifestErrorNamesNodeRefAndChain` (`closure_test.go:365`) asserts `invalid skill manifest for leaf`, `tag v1`, `(via <project> -> mid -> leaf)`, and the underlying verr text. `TestBrokenTransitiveManifestNamesTheBrokenSkill` (`install_test.go:598`) pins the same provenance through `install.Project` → `Result.Errors`, status `failed`. Both re-run here with `-count=1`: PASS. |
| Protocol error codes unchanged | PASS | `skillcheck.Validate` still emits `skill.manifest_invalid` / path `csk-skill.json` with the bare `err.Error()`; the wrap is at the closure call site only. `TestValidateMissingSkillAndInvalidManifest` now pins that boundary explicitly. Grepped repo-wide: `manifest_invalid` appears only in `skillcheck.go:34` and its test. |
| `go test` green | PASS | Full `go test ./...` clean; `-count=1` re-run of `internal/closure`, `internal/install`, `internal/skillcheck`, `internal/skillspec`, `cmd/curator` all `ok`. |

## Gates re-run by the reviewer (not trusted from the implementer log)

- `go build ./...` — exit 0
- `go test ./...` — no failures
- `go test -count=1 ./internal/{closure,install,skillcheck,skillspec} ./cmd/...` — all `ok`
- `go vet ./...` — exit 0
- `gofmt -l ./cmd ./internal` — empty
- `golangci-lint run` — `0 issues.`, exit 0

Not run: `make ci-test` / `make check-ci`. `CURATOR_CONFORMANCE_ROOT` is unset in
this environment and the gate script hard-fails without it. Residual risk checked
directly instead: `internal/closure/conformance_test.go` (53 lines) asserts
provider order only and contains no error-string assertion, and no vector or doc
pins the closure manifest-error text (`grep` over `docs/`, `README.md`, repo).
`internal/skillspec/conformance_test.go` is untouched by this change because
`skillspec.Load`'s own message is unchanged. Residual conformance risk: low but
not zero — CI's `Implementations` lane is the real gate.

## Architecture fit

- The `<kind> <ref> -> <short commit> (via <chain>)` shape is copied from the
  sibling `version conflict for ...` (`closure.go:229`) and `cannot resolve ...`
  (`closure.go:225`, `:276`) messages in the same function family, so closure
  diagnostics stay uniform.
- Placement at the call site rather than inside `skillspec.Load` is correct and
  load-bearing: the other caller (`skillcheck.go:32`) backs standalone
  `curator skill check`, which has no closure and whose `skill.manifest_invalid`
  payload is a protocol surface. Pushing the wording down would have leaked
  `via <project> -> ...` into a command with no closure and mutated a protocol
  message.
- Coverage claim verified: `skillspec.Load` has exactly two call sites repo-wide;
  all three `closure.Build` callers (`install.Project` at `install.go:179`,
  `install.Global` at `global.go:59`, `cmdAudit` at `main.go:1139`) propagate the
  error verbatim via `result.failf("%v", err)` / `[]string{err.Error()}`, adding
  no prefix. `install.validateNodes` → `skillcheck.Validate` is genuinely
  unreachable for manifest errors: `closure.Build` runs first over the same
  snapshot directory with the same deterministic parser. Leaving it alone rather
  than adding a second divergent format is the right call.
- `install_test.go` references `closure.ProjectEdge` rather than hardcoding
  `<project>`, so the assertion tracks the constant.

## Non-blocking observations (no rework required)

1. `closure.go:291` — the message omits `node.Substituted`. A dev-substituted
   node renders as `... revision HEAD -> <commit> (via ...)` with no hint the ref
   came from `Skillfile.dev.json`. `substitution.Describe()` is already in scope
   at that point. Outside the AC, and the short commit still identifies exactly
   what was parsed, so this is a future polish item, not a defect.
2. `install_test.go:600-624` builds two skill repos inline rather than extending
   `env.skill`. Justified — the helper has no knobs for dependencies or a broken
   manifest — but a third such test should push the knobs into the helper.
3. `closure_test.go:388` — the guard
   `!strings.Contains(message, "manifest for mid")` is vacuous against the *old*
   code (which produced `leaf: commands: must be an object`). It is a fine
   forward-looking regression guard; the red-green signal comes from the positive
   assertions, and the implementer captured the red run.

## Handoff

Acceptance evidence recorded. Changes are uncommitted per repo policy. The
commit-owning mover must commit only `internal/closure/closure.go`,
`internal/closure/closure_test.go`, `internal/install/install_test.go`,
`internal/skillcheck/skillcheck_test.go` (plus the LOGBOOK entry for this task)
and then make the final `done` transition with `commit_ack=scope_committed`.
