# Producer brief: promote SPEC_PIN to the qualified released revision

## Where and what

- Repository `~/Developer/ReluxWorks/curator`. Branch and worktree named at spawn; base is curator
  main after PR #64 lands — the orchestrator confirms the exact OID. First run
  `git submodule update --init --recursive`.
- Files in scope: `.github/workflows/ci.yml` (the `SPEC_PIN` value and its comment),
  `.github/ci/root-artifacts.tsv`, `.github/ci/platform-cases.tsv`, and any skip reason that stops
  being reachable. Authority: curator-spec `v1.0.0-rc.11` = commit
  `87a0d0060bad64ab883d007dcdf35df7485368bf`.

## Why, precisely

`ci.yml` pins `SPEC_PIN: 0ed5c691e9208eea52f21db2fc05e226ce3516fd`, whose
`conformance/v1/manifest.json` hashes to `sha256:803918bf8672f76c…`. The protocol release record at
curator-spec main, `release/1.0.0-rc.9.json`, states:

```
downstream_consumption.required_manifest_sha256 = sha256:0e195ecd26af2fcb5e5afb5c3f…
downstream_consumption.committed_release_pin_advanced = false
```

and `0e195ecd…` is exactly the manifest of the tagged revision. So curator is currently testing
against a root the protocol release record **no longer names as required**, and the flag is the
release machinery saying the downstream pin has not caught up. This is what
`TASK-260720-38l1sy` (audit-curator-released-suite-pin) exists for, with `TASK-260720-25d05o`
(qualify-protocol-release-evidence) as its qualification half.

The policy in `ci.yml` is explicit and must hold: the pin carries **only a qualified released
revision**, and a candidate suite never enters through it. `v1.0.0-rc.11` is a signed, released tag,
so it satisfies that; pinning a branch or an untagged commit would not.

## What changes, and what must fall out of it

1. `SPEC_PIN` moves to `87a0d0060bad64ab883d007dcdf35df7485368bf`, and the comment above it is
   rewritten to say what is true of the new pin — the tag it is, the protocol version it publishes,
   and its manifest hash. The current comment's hashes and reasoning are all about the old pin; do
   not leave a single stale clause.

2. **The deferrals disappear.** Against the new root, `suite-plan.sh` serves what it used to defer:
   `internal/config`, `internal/envfragment`, `internal/envmarker` and
   `internal/interop/environments`. So every `root-unset` tolerance those packages' ledger rows carry
   becomes unreachable, and a tolerance that can never fire is a lie in the ledger. Remove the
   tolerated-skip columns from exactly the rows that lose them, and leave the rows required.

   This is the acceptance criterion the board already records for this task: **the three schema
   drivers lose their `root-unset` tolerance and must pass on every lane.**

3. Anything else that only made sense against the old root — a skip reason no longer reachable, a
   comment describing a deferral that no longer happens — goes with it. Enumerate what you changed
   and why; do not leave the ledger describing a world that ended.

## Acceptance

The default `Test` and `Race` lanes, on all three runners, now **serve** the environments families
and pass them — the coverage that until now existed only in a hand-dispatched candidate lane. Prove
it locally first with `test-gate.sh` against a materialized `v1.0.0-rc.11` root, and confirm
`suite-plan.sh` reports `deferred=0` for those packages. The orchestrator supplies the hosted
measurement.

Also confirm the candidate lane still behaves: dispatched against the same revision it must stay
green, and against a root missing an environments family it must still fail closed — the guard PR #63
landed must not be weakened by the pin move.

## Method, from this epic

- Materialize every root as a **plain checkout** verified file-by-file against `manifest.json`. Never
  `git archive`: `export-subst` is set on `conformance/v1/fixtures/byte-exact/subst.txt` and an
  archived root carries 65 bytes against 40 on disk.
- Run the two `test-gate` lanes **sequentially**, never concurrently and never alongside a `-race`
  suite — that contention over the machine-wide Go test lock has produced false reds twice.
- Anchor each `-run` level separately and count `=== RUN` lines; `go test -run '^(Parent/child)$'`
  splits on the slash, matches nothing and exits 0.
- A ledger row must describe what its test asserts. Three notes in the last two leaves claimed more
  than their scans enforced; do not add a fourth.

## Delivery

Small signed commits, human identity. **Do not write `LOGBOOK.md`.** Do not push, do not open a PR.
**Do not stage with `git add -A`.** Gates, each a standalone process with its observed exit code:
`go build ./...`, `go vet ./...`, `gofmt -l cmd internal`, `golangci-lint run ./...`,
`bash .github/ci/gate-selftest.sh`, `bash .github/ci/no-broad-suppression.sh`,
`bash .github/ci/ledger-consistency.sh`, and `test-gate.sh` against the new pinned root.

Attach `TASK-260906-284db9_drafting-report.md`: the old and new pin with both manifest hashes; the
rewritten comment; every ledger row that lost a tolerance and why; the `suite-plan` partition before
and after; the candidate-lane confirmation including the fail-closed negative; and the gate table.
Then `task-board handoff TASK-260906-284db9 --role developer`.
