# Review verdict — `CR-TASK-260906-2g0bgq-3` revision 3 (stage (b), cycle 3)

Reviewer run `RUN-260906-0644ed`. Full findings:
`TASK-260906-2g0bgq_review-findings-stage-b-3.md`. Probes:
`TASK-260906-2g0bgq_review-probes-stage-b-3.tar.gz`.

## ACCEPT

Zero blocking, zero major. Three minors (m5, m6, m7), none of them a rework-2 regression, none of
them grounds to return the work: m5 and m6 belong to the orchestrator's routing (a stage-(a)
inheritance and a `SPEC_PIN` bump), m7 needs no action at all.

**repeat-of: none.** The cycle-2 meta-finding — *a gate line attested that the command does not
print*, marked `repeat-of` cycle 1's F3 — is the class I looked hardest for, and it does not recur.
This revision's gate table gives every row its exact standalone command and observed exit code, runs
the root-sensitive gates against both roots, quotes `gh pr checks 60` verbatim while stating plainly
that those results are the *old* head's, and lists what it did not run. I re-ran 10 of its 13 rows
and every number matched, including the two skip counts (28 at the pin, 19 at main).

## Why an empty repository delta is the right outcome for this leaf

`CR-TASK-260906-2g0bgq-3` carries `repository_delta=empty` and a zero-byte patch
(`git diff f39f4a9309…c3e47989fa` → no paths; the patch resource is 0 bytes). **That is what this
leaf was specified to produce in this workspace, and it is correct.**

`producer-brief-stage-b.md` places every line of the implementation in a *separate repository* —
`~/Developer/ReluxWorks/curator`, worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`,
branch `feat/agent-environments-stage-b` — and says in terms: *"Never write LOGBOOK.md or anything
into the control root; the story workspace carries an empty delta by design."* A non-empty delta here
would have been a violation of the brief, not evidence of work. The reviewable artefact is the
22-commit curator branch the review brief names, and that is what this review drove: 28 files,
+7133/−40 against `origin/main`, PR #60 head `1a936e77490df3008efe08fa0aba473ff77575c6` — byte-equal
to the worktree head I reviewed. Every commit carries a good signature (`G`) and the repository's
human identity `Ivan Oparin <oparin@me.com>`; the delta touches `.github/ci/`, `cmd/curator/` and
`internal/` only — no board files, no `LOGBOOK.md`, no control-root writes.

Recorded as a standing process note (cycle 2 made the same one): the board's Change Request snapshot
holds no record of the 7133 lines actually under review. The board tracks the leaf; the code lives
in the other repository's PR.

## What was verified, by driving

Rework 2 is `7bca4d4b..1a936e77` (5 signed commits, 9 files, +138/−78). Coverage: **6 of 6 rows
driven**; **6 mutants applied, 5 killed, 1 survived and chased to a redundancy**.

- **B1** — the orchestrator's override was tested, not accepted. Registration fails **closed**: a
  candidate-lane invocation against a main root with either family removed exits 1
  (`suite-plan: FAILED`), and `test-gate.sh` then refuses to run at all. The rejected `root-content`
  alternative would have gone green — `skip-classes.tsv` gives it policy `allow` in every lane and
  `CI_REQUIRE_FULL_ROOT` is read by `suite-plan.sh` alone. The preserved fatal is live: a *served*
  root whose family directory exists but is empty still fails the three drivers. Pin lane exit 0
  (3 rows `tol root-unset`, 28 skips); main lane with `CI_REQUIRE_FULL_ROOT=1` exit 0 (3 rows `ok`,
  19 skips, 0 `root-unset`, 0 `stage-deferred`).
- **B1b** — the fixture still asserts what it did: the `..` clause mutant kills on the traversal
  case (so the value carries a real `..`), the root-parent mutant kills on the out-of-root case, and
  it still kills on F4's sibling case with the out-of-root assertion deleted. `TestCheckBoundary` is
  ledger-required on all three runners with no tolerated class and executes even when the package is
  deferred, so the hosted Windows lane proves it. The `fragment_schema_test.go:checkPath` slash
  decision is right — that function's only inputs are published vector strings against a literal
  POSIX root, host paths cannot reach it, and the old `filepath.IsAbs` rejected every *valid* vector
  case on Windows rather than catching anything.
- **M1** — driven through `run()` across the full matrix, not through the helper: `pi` and
  `opencode` get `environment_target_unknown … for <adapter>`, `claude_code` and `codex_cli` resolve,
  an undeclared target is unknown with or without `--env`, and `--env ghost-env` reports
  `environment_unknown` first. The producer's mutant kills `TestProfileUseTargetBoundToAdapter`. The
  `--env`-absent branch is **not** a remaining hole: §7.6's stated condition is "a target identifier
  not declared by the registry", the identifier alone, and revision 1's two rows each name their own
  adapter.
- **m2** — judged on the whole document, not the sampled top: 58 distinct key paths through `run()`,
  0 non-snake_case, all nine structs covered. **m3** — the vacuity guard fires when the loop is
  emptied.
- **Regression** — 25 environments conformance cases pass with zero skips at curator-spec `f39f4a9`,
  including all eight sets this stage was to deliver; `stage-deferred` occurs 0 times in either
  lane's `skips-observed.tsv`; zero package failures across both lanes.
- **Gates I re-ran myself**, each as its own process: `go build`, `go vet`,
  `GOOS=windows|linux|darwin go vet` (my addition), `gofmt -l cmd internal`,
  `golangci-lint run ./...` (`0 issues.`), `gate-selftest.sh` (`81 passed, 0 failed`),
  `no-broad-suppression.sh`, `ledger-consistency.sh` (`131 rows`), `test-gate.sh` against both roots,
  and `./cmd/curator` inside gate B (`ok 348s`). All exit 0.

## Hosted lanes on the exact accepted head

**Every lane green.** `gh pr checks 60` on head `1a936e77` (run 34020466962), verbatim:

```
Test (ubuntu-latest)            pass   3m8s
Test (macos-latest)             pass   8m2s
Test (windows-latest)           pass   40m30s
Race (ubuntu-latest)            pass   9m57s
Race (macos-latest)             pass   18m50s
Lint                            pass   41s
Gate self-test (ubuntu-latest)  pass   6s
Gate self-test (macos-latest)   pass   11s
Gate self-test (windows-latest) pass   22s
Interop conformance gate        pass   23s
Naming gate                     pass   9s
Candidate suite (${{ matrix.os }})  skipping  0     (conditional job; workflow_dispatch only)
```

From the Windows lane's own uploaded evidence (`test-evidence-windows-latest`, artifact
9985904326) — the two things that had to be true there and could not be checked locally:

```
suite-plan: GOOS=windows
suite-plan: root=D:\a\curator\curator/protocol-spec/conformance/v1
defer internal/envfragment    … publishes none of: schema-cases/launch-env-fragment-v1
defer internal/envmarker      … publishes none of: schema-cases/agent-environment-marker-v1
suite-plan: served=70 deferred=2 excluded=0 / ok

ok    internal/envfragment :: TestCheckBoundary
tol   internal/envmarker   :: TestParseAuthoritativeEnvMarkerSchemaCases (tolerated skip: root-unset)
tol   internal/envfragment :: TestFragmentAuthoritativeSchemaCases       (tolerated skip: root-unset)
tol   internal/envfragment :: TestFragmentEmissionMatchesReference       (tolerated skip: root-unset)
platform-case gate: 84 skips recorded … platform-case gate: ok
```

`TestCheckBoundary` **executed and passed on the real Windows runner** — B1b is hosted-proven, not
proven by construction; and B1's registration behaves on GOOS=windows exactly as it does locally.

## Routing

`accept_cr(TASK-260906-2g0bgq, revision=3, evidence=TASK-260906-2g0bgq_review-verdict-rev3.md)` —
the element moves to `integrating`. Not `done`: only the bound producer run may checkpoint or
integrate. No `commit_ack` supplied.

The orchestrator should route m5 (register `internal/interop`'s vector families in
`root-artifacts.tsv`, converting a candidate-lane-permissive `root-content` skip into a fail-closed
`root-unset` deferral) as its own CI leaf in or beside stage (c), and remember m6 (the three schema
drivers assert nothing on any automatic hosted lane until `SPEC_PIN` is bumped past `f39f4a9`).
