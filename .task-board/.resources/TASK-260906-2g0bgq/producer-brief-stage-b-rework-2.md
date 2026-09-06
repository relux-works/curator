# Producer brief: stage (b) rework 2 — make the head green on every hosted lane, and wire the one gate that never runs

## Verdict

Review cycle 2 returned CHANGES REQUESTED: two blocking, one major, three minors. Read
`TASK-260906-2g0bgq_review-findings-stage-b-2.md` in full.

The rework itself is accepted. The reviewer re-drove **F1–F6 and F8–F11 with its own narrowing
mutants and watched each kill a named test** — that work stands and is not to be touched. What blocks
is that the reviewed head is red on **all five** hosted Test and Race lanes, for two defects the
pre-rework stage-(b) commits introduced and that no local run could see: one because every run used a
conformance root CI does not use, the other because no run was a Windows run.

## B1 (blocking) — the three schema drivers hard-fail against the root CI pins

`internal/envfragment/fragment_schema_test.go:312` and `:327`;
`internal/envmarker/marker_env_schema_test.go:78`; `.github/ci/platform-cases.tsv:283-285`.

`ci.yml:44` pins `SPEC_PIN: 0ed5c691` and exports its `conformance/v1` for the Test and Race jobs.
That tree predates the environments artefacts and publishes neither
`schema-cases/launch-env-fragment-v1` nor `schema-cases/agent-environment-marker-v1`. The three
drivers read those families unguarded, so every default lane fails on them and nothing else.

**Take the mechanism `suite-plan.sh` exists for, not the tolerated skip the reviewer proposed.**
The reviewer's fix — a `root-content` skip plus a tolerated class on the three ledger rows — turns CI
green and preserves the present-but-wrong failure, and its reasoning about that distinction is right.
But `skip-classes.tsv` gives `root-content` policy `allow` **in every lane, the candidate lane
included**, so under that fix a candidate root that stopped publishing either family would pass with
the case silently skipped. `CI_REQUIRE_FULL_ROOT=1` makes only a *deferral* fatal, and a deferral only
happens for artefacts declared in `.github/ci/root-artifacts.tsv`. Registration is therefore strictly
stronger and is what this stage ships:

1. Add two rows to `.github/ci/root-artifacts.tsv`:
   `internal/envfragment` → `schema-cases/launch-env-fragment-v1`;
   `internal/envmarker` → `schema-cases/agent-environment-marker-v1`, each with the sentence the
   table's third column wants.
2. Set `skip_allowed_on` to `linux,darwin,windows` and `class` to `root-unset` on the three
   `platform-cases.tsv` rows, and extend each behaviour sentence to say that a root which does not
   publish the family defers the package and records a `root-unset` skip.
3. **Leave the three `t.Fatal`s exactly as they are.** Under registration a *served* package is one
   whose declared artefacts the root publishes, so a served root that lacks the family is a genuine
   error and must stay fatal. The producer's stated intent — "a root that stops publishing it fails
   here instead of quietly narrowing the check" — is preserved in full and, in the candidate lane,
   strengthened: the lane now fails closed instead of skipping.

Both halves are already verified by the orchestrator at your reviewed head, with the real roots:

```
CI_ROOT_ARTIFACTS=<patched> suite-plan.sh <rc.9 root>
  defer internal/envfragment
  defer internal/envmarker
  suite-plan: served=70 deferred=2 excluded=0

CI_ROOT_ARTIFACTS=<patched> CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <curator-spec main root>
  suite-plan: served=72 deferred=0 excluded=0
  suite-plan: ok
```

Reproduce both yourself and quote them. Materialize the pinned root read-only with
`git -C <curator-spec> archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C <scratch>`.

## B1b (blocking) — `TestCheckBoundary` cannot pass on Windows

`internal/envfragment/envfragment_test.go:135-137`, gate at `internal/envfragment/envfragment.go:243`,
required on all three runners at `.github/ci/platform-cases.tsv:268` with no tolerated class.

The fixture feeds POSIX-rooted literals (`/manager/environments/...`) to a gate built on
`path/filepath`. On Windows `filepath.IsAbs` needs a volume name, so the *conforming* case is rejected
with "is not absolute" at the first assertion; the separator check would fail independently. Production
is unaffected — `buildFragment` passes `EnvRoot(req.Home)`, a real platform path.

Build the fixture root and every value with `filepath.Join` over a platform-absolute base — `t.TempDir()`
or one shared helper — so the same assertions run on all three runners. Do **not** register a tolerated
skip: the boundary is not platform-specific, only the fixture is. `TestFragmentAuthoritativeSchemaCases`
carries the same shape in its own local `checkPath` at `fragment_schema_test.go:143`; fix it there too,
since B1 will stop hiding it in the candidate lane.

While you are there, sweep the stage-(b) delta for the same class once more and say in the report what
you swept and what you found. The reviewer already checked one neighbour — no new fixture interpolates a
native path into a git config value — so do not redo that; look for the remaining shapes: a POSIX
literal fed to `filepath`, a hardcoded `/` separator in a comparison, a `strings.HasPrefix` over paths.

## M1 (major) — `TargetsFor`/`TargetFor` are defined and never called

`internal/envregistry/envregistry.go:326` and `:338`; caller that should use them
`internal/envprofile/switch.go:171`.

The registry rows are right and `env status` prints both bindings, but the per-adapter *resolution*
half is unreachable: `switch.go:171` still calls the global `TargetByID`, so
`profile use --env pi --target xcode-coding-assistant` is admitted for adapters that declare no
target. The rework report claimed the diagnostic for `pi`/`opencode`; that is true of the helper and
false of every production path. This is the AC's named shape — a check present but uncalled.

Call `envregistry.TargetFor(environment, target)` from `useLocked` when `--env` is present, falling
back to `TargetByID` when it is not. Prove it through `run()`, not through the helper. Narrowing
mutant to apply and quote: make `TargetFor` ignore its adapter argument and delegate to `TargetByID`
— a CLI-level test must fail.

## Minors

- **m2** — `internal/envprofile/status.go:107-120`: `env status --json` publishes Go field names
  (`{"Adapters","Homes",…}`) while every other machine-readable surface in `cmd/curator` uses explicit
  snake_case. §12 does not name the members, so this is not a spec violation. Add `json` tags, or state
  the wire shape as a deliberate bound in the report; do not leave it undeclared.
- **m3** — `internal/envprofile/managed_test.go:303-338`: the write-through-link loop `continue`s on
  non-matching surfaces, so a future change to `storeDocPath` could empty it silently. It is not vacuous
  today — the reviewer measured 2–4 rendered surfaces per case — but the count is the evidence. Add a
  per-case `if rendered == 0 { t.Fatalf(...) }`.
- **m4** — the rework report's gate table ran every gate against the curator-spec main root and then
  stated a CI outcome it never read, while CI was already red.

## The meta-finding — read it as the main lesson

The reviewer marked B1 **repeat-of cycle 1's F3**, one level out: cycle 1 attested a gate line the
command does not print; cycle 2 attested a lane the report never read. Its words: *the next step is a
gate on how gate lines are produced, not a third revision that re-attests them.*

So, for this revision, the report's gate table is itself under review:

- every row carries the exact command, run as a standalone process, and its observed exit code;
- every row that concerns a conformance root names **which** root, and the root-sensitive gates appear
  **twice** — once against the `SPEC_PIN` tree and once against curator-spec main;
- any claim about hosted CI carries `gh pr checks 60` output pasted verbatim, or says plainly that CI
  was not consulted. A sentence like "left to CI" is not an outcome;
- a gate you did not run is listed as not run. That is an acceptable answer; an unread gate reported as
  green is not.

## Delivery

Continue the branch `feat/agent-environments-stage-b` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b` with small signed commits on top of
`7bca4d4b` — this is a code PR landing by fast-forward of the reviewed head, so a linear series is
right. Human identity, no `LOGBOOK.md`, nothing written into the control root.

Gates, each as its own process: `go build ./...`, `go vet ./...`, `gofmt -l cmd internal`,
`golangci-lint run ./...`, `bash .github/ci/gate-selftest.sh`, `bash .github/ci/no-broad-suppression.sh`,
`bash .github/ci/ledger-consistency.sh`, `bash .github/ci/test-gate.sh` against **both** roots (the
`SPEC_PIN` tree, and curator-spec main with `CI_REQUIRE_FULL_ROOT=1`), and
`go test -count=1 -timeout 30m ./cmd/curator` once at the end. Do not push and do not open a PR —
PR #60 already exists and the orchestrator owns its head.

Attach `TASK-260906-2g0bgq_rework-report-2.md`: finding → resolution table for B1, B1b, M1, m2, m3, m4;
the narrowing mutant for M1 with the test it killed; the two-root gate table described above; the
B1b sweep with what you looked for and what you found; and honest bounds. Then
`task-board handoff TASK-260906-2g0bgq --role developer`.
