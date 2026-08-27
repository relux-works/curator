# TASK-260822-3j4bcb — closure-error-provenance

## Problem (reproduced before the fix)

`internal/closure/closure.go` wrapped a failed `skillspec.Load` with a bare
`fmt.Errorf("%s: %w", item.name, err)`. For a broken **transitive** manifest the
whole diagnostic was:

```
leaf: commands: must be an object
```

No resolved ref, no requirement chain, and the `leaf:` prefix reads like another
verr field-path segment rather than a node name — so the operator cannot tell
which skill in the closure is actually broken or which ref pinned it.

Captured red run before the change:

```
closure_test.go:385: closure error = "leaf: commands: must be an object",
    want it to contain "invalid skill manifest for leaf"
```

## Change

One production edit, `internal/closure/closure.go:288` (`resolveNode`):

```go
spec, err := skillspec.Load(snap)
if err != nil {
    return nil, fmt.Errorf("invalid skill manifest for %s %s %s -> %s (via %s): %w",
        item.name, resolved.Kind, resolved.Ref, short(resolved.Commit), item.chain, err)
}
```

Rendered text now:

```
invalid skill manifest for leaf tag v1 -> d0775b5f70d0 (via <project> -> mid -> leaf): commands: must be an object
```

The `<kind> <ref> -> <short commit>` shape matches the existing `version conflict
for ...` and `cannot resolve ...` messages in the same file, so closure
diagnostics stay uniform. The short commit is included on purpose: for a
substituted or branch-pinned node the ref alone ("revision HEAD") is not enough
to identify what was actually parsed.

## Placement decision (why not inside `skillspec.Load`)

Provenance is added at the **call site**, not inside `skillspec.Load`. The other
`skillspec.Load` caller is `internal/skillcheck/skillcheck.go`, which backs
standalone `curator skill check` and emits the protocol issue
`skill.manifest_invalid` with `err.Error()` as its message. Pushing closure
wording down into `skillspec` would have leaked closure-only context ("via
<project> -> ...") into a command that has no closure, and changed the protocol
issue payload. `skill.manifest_invalid` and its `csk-skill.json` path are
untouched.

## Coverage of the install planning path

`skillspec.Load` has exactly two call sites repo-wide (verified by grep):
`closure.go:288` and `skillcheck.go:32`. Every install planning path
(`install.Project`, `install.Global`, `cmdAudit`) reaches manifest parsing only
through `closure.Build`, so the single closure wrap covers all of them.

`install.validateNodes` -> `skillcheck.Validate` re-parses the same snapshot, but
that branch is unreachable for manifest errors: `closure.Build` runs first over
the same directory with the same deterministic parser, so it always fails before
`validateNodes` is entered. It was therefore left alone rather than given a
second, divergent provenance format.

## Tests

| Test | Package | What it pins |
|---|---|---|
| `TestTransitiveManifestErrorNamesNodeRefAndChain` | `internal/closure` | name + `tag v1` + `(via <project> -> mid -> leaf)` + underlying verr text; asserts the message does not read as if the declaring skill `mid` were broken |
| `TestBrokenTransitiveManifestNamesTheBrokenSkill` | `internal/install` | same provenance surfaces through `install.Project` -> `Result.Errors`, status `failed` |
| `TestValidateMissingSkillAndInvalidManifest` (extended) | `internal/skillcheck` | protocol code `skill.manifest_invalid` and path `csk-skill.json` unchanged, and closure wording does **not** leak into the standalone check |

New harness helper `harness.brokenSkill(name, payload)` in
`internal/closure/closure_test.go` commits an unparsable `csk-skill.json` at tag
`v1`, matching the existing `harness.skill` shape.

## Gate results (each run as a standalone process, real exit codes)

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go test ./...` | 0 (log: `.temp/TASK-260822-3j4bcb/go-test-01.log`) |
| `go vet ./...` | 0 |
| `golangci-lint run` | 0 — `0 issues.` (log: `.temp/TASK-260822-3j4bcb/golangci-lint-01.log`) |
| `gofmt -l ./cmd ./internal` | 0, no output |

`gofmt -l .` at repo root is not a usable signal here: `.temp/` holds an
unpacked go1.25.1 source tree and a vendored worktree from earlier tasks, so it
reports thousands of unrelated files. Scoped to `./cmd ./internal` it is clean.

The conformance gates (`make ci-test`, `make check-ci`) were **not** run: they
require `CURATOR_CONFORMANCE_ROOT` pointing at a materialised
`<curator-spec>/conformance/v1` checkout, which is not present in this
environment. `internal/closure/conformance_test.go` self-skips without it.

## Files touched

- `internal/closure/closure.go` — the wrap (production change)
- `internal/closure/closure_test.go` — `brokenSkill` helper + new test
- `internal/install/install_test.go` — install-path test
- `internal/skillcheck/skillcheck_test.go` — protocol-code boundary assertions
