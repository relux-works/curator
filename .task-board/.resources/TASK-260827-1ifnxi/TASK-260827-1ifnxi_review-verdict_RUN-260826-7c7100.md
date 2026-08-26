# Review verdict — TASK-260827-1ifnxi

Verdict: **changes requested**. Route: `to-dev`.

Reviewed Change Request `CR-TASK-260827-1ifnxi-1`, revision 1, candidate tree
`84b1d2e558c9e5fd8e4c4261ca0df02ba5f2cd42`. The review was read-only and
limited to `docs/authoring-language-adapters.md`; the unrelated pre-existing
`cmd/curator` changes were excluded as required by the assignment.

The referenced adapter implementation and accepted spec are absent from this
Story's old base commit, so technical claims were checked read-only against
accepted integration commit `6f93b51bdcce209172dbb9224f44315aa601437f`.

## Required corrections

1. `docs/authoring-language-adapters.md:190-194` misidentifies the proof for
   macro-oracle replacement-body analysis as “the H24/R12 vectors in their
   tests.” `R12` is not a macro-oracle test vector: in the accepted SwiftPM
   research it means “Version-specific manifest/toolchain variants,” and no
   `R12` marker exists in `internal/swiftpminterop` tests. The source/build-
   setting macro-oracle tests are `H24` plus `H25` and its `Q1`–`Q6` cases in
   `internal/swiftpminterop/buildsettings_test.go`. Replace `R12` with the real
   code-backed test reference (or cite the research separately for its actual
   meaning).

2. `docs/authoring-language-adapters.md:329-330` presents
   `internal/{rustsource,...,swiftpmsource,...}/conformance_test.go` as a set of
   existing package-local suites. Two named paths do not exist:
   `internal/rustsource` uses `build_conformance_test.go`, while
   `internal/swiftpmsource` uses `swiftpmsource_test.go` and
   `swift_integration_test.go` (and has no `conformance_test.go`). Name the real
   files or phrase this as a naming recommendation without claiming those
   nonexistent files are existing proof.

Both defects violate the acceptance criterion that every normative claim name
the package or spec section that actually proves it. They are ordinary
documentation rework, not a Stop-The-Line condition.

## Checks that passed

- The supported-build predicate and C0–C7/C3a/C3b model match
  `.spec/skill-facing-cli-source-closure.md` and `internal/closuregraph`.
- The adapter process guard keeps exactly `internal/closureexec/acquisition.go`
  and `portable_runner.go` as the allowed adapter process-launch seams.
- The reject-by-default spelling, position, build-setting-kind, and macro-
  oracle-input axes otherwise match `internal/swiftpminterop`.
- The portable/verified observed-read boundary matches
  `internal/swiftpminterop` and `internal/swiftpmbuild/readset.go`.
- CGP05, CGP10, the seven obligations, and all 19 published rejection vectors
  match `internal/crossconformance`.
- The guide is author-oriented, English, and links the consumer conformance
  document instead of reproducing its support, diagnostics, migration, or
  rejection-matrix tables.
- `git diff --check` passes for the guide. No Go build or test was run, per the
  docs-only review brief.

