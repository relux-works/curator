# TASK-260728-1g0z69 — review cycle 7 verdict

## Verdict

**CHANGES REQUESTED** — route to `analysis`.

The design contract is otherwise coherent and the cycle-6 change removes the
original `go build ./...` acceptance-laundering path, but the attached boundary
probe still does not implement the contract's closed, fail-closed command
classifier.

## Blocking finding

### The `invalid GOTOOLCHAIN` rejection branch accepts an open-ended family of outcomes

The normative reference says the recognised `toolchain`-position set is
*exactly*:

- exit 0;
- `cannot find "v" in PATH`;
- `invalid toolchain "v" in <file>`;
- `invalid GOTOOLCHAIN "v"`;
- everything else is `unknown` and must fail.

See `docs/compiled-build-toolchain-requirements.md` lines 797-824.

The attached probe instead accepts both the exact form and **every** colon
suffix:

```go
case line == "invalid GOTOOLCHAIN "+q,
    strings.HasPrefix(line, "invalid GOTOOLCHAIN "+q+":"):
    return cmdRejected
```

See `TASK-260728-1g0z69_boundary-probe.go` lines 720-722.

That prefix is not one closed upstream outcome. On both qualified Go sources,
the colon-bearing `Select` diagnostics occur while parsing the initial
`GOTOOLCHAIN` setting and quote that setting (`local+path` in this probe):

- `invalid GOTOOLCHAIN %q: invalid minimum toolchain %q`;
- `invalid GOTOOLCHAIN %q: only version suffixes are +auto and +path`.

The toolchain-name-derived pre-search refusal is only the exact
`invalid GOTOOLCHAIN %q` form. This is visible in Go 1.25.1 and Go 1.25.5
`cmd/go/internal/toolchain/select.go` at the corresponding `Fatalf` sites.

Consequently an unrelated or future non-zero line such as
`invalid GOTOOLCHAIN "go1x": unrelated failure` is classified as `rejected`
instead of `unknown`. For an isolated-rejected value such as `go1x`, that
fabricated rejection agrees with the isolated result, so the agreement check
passes. This is the rejection-direction laundering that decision 0007 itself
explicitly forbids.

The existing `-red unrelated-command-failure` control does not catch this
branch: it exercises `default` and `go1`, whose unrelated `updates to go.mod
needed` outcome tests laundering toward acceptance only.

## Required rework

1. Remove the open-ended `invalid GOTOOLCHAIN "v":...` prefix branch, or replace
   it with a finite set of structurally exact, currently reachable outcomes.
   Under this probe's fixed `GOTOOLCHAIN=local+path`, colon-bearing diagnostics
   that quote the environment setting rather than the tested value must remain
   `unknown`.
2. Make the implementation match the reference's exact recognised set and keep
   every other non-zero outcome fail-closed.
3. Extend the unrelated-failure regression control (or add a focused classifier
   control) to prove both directions: an unrelated outcome cannot become
   `accepted` for an isolated-accepted value and cannot become `rejected` for an
   isolated-rejected value.
4. Re-run both supported Go toolchains, the green probe, all regression
   controls, validation/unit/Go/vet/gofmt/diff gates, clean regeneration, and
   frozen `1.0.0-rc.5` byte comparisons. Attach revised task-scoped evidence.

## Verified evidence

- `go version` is not one of the four Go 1.25.1/1.25.5 `toolchain.Select`
  shortcuts; it enters selection under `GOTOOLCHAIN=local+path`.
- Current reachable measured outcomes on both toolchains are exit 0,
  exact `cannot find "v" in PATH`, `invalid toolchain "v" in go.mod`, and exact
  `invalid GOTOOLCHAIN "v"`.
- Independent probe runs: green exit 0; `tidy-exit`,
  `patch-prerelease-compared`, `c-equals-upstream`, and
  `unrelated-command-failure` each exit 1 for their named reason.
- Project gates: validation passed (42 schemas / 422 vector files), 29 Python
  tests passed, `go test ./...`, `go vet ./...`, `gofmt -l tools`,
  vector generation, and `git diff --check` all passed.
- Candidate versus accepted predecessor is content-different only in
  `CHANGELOG.md`, decision 0007, and the implementation reference. Generated
  `conformance/v1` and `release/1.0.0-rc.5.json` remain byte-identical.
- No candidate artifact was modified, staged, committed, published, pinned, or
  used to claim Rust/Swift/Kotlin/JDK platform validation.
