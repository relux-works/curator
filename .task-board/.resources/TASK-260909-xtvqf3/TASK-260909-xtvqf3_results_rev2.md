# TASK-260909-xtvqf3 results rev2: diagnostics CR2 conformance gate

Status: ready for review (developer handoff).
Scope: task-scoped gate only; no production changes. Supersedes rev1;
rev1 artifacts and verdict stay available on the board.

## Baseline identity (exact candidate, unchanged)

- Tree: `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`
- Base: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`
- SPEC.md sha256: `5a7ccf0ba95708cb573a586977eb0ba4bad1e46233a540dc99784c1d4d922d48`
- Blobs: diagnostics.go `4baf919afcb56981a1cdc7f208eb5dafe340aedb`,
  diagnostics_test.go `dcfbc7e1f8369cff46a6b54aa84b0c1b205996cb`,
  helpers_test.go `a46ad729092296ac9be3dcc4619ac825a61c0b30`,
  main.go `f2a23748d8eb3f3f8bca389d390373e6cf4fdc10`,
  main_test.go `849e6aaee17c4040e88a7ab2d7b8c01e4ad2ca70`,
  mutants.sh `5ff00da1d66e5def3bfc43e12b7b01708f0cefc3`,
  README `2fee2deebdccfb374fc7fa81952f3ecd666b6158`,
  SPEC `997f00651692999577a3e775d28c72747f83e4d2`
- Verified at runner start via `git cat-file`, 8 blob OIDs, SPEC sha256;
  disposable copies via `git archive` from toplevel; original workspace
  untouched (`git status --short` empty, `.temp/` gitignored).

## Gate artifacts rev2 (attached, task-scoped)

- `TASK-260909-xtvqf3_rev2_gate-conformance_test.go`
- `TASK-260909-xtvqf3_rev2_gate-framing_test.go`
- `TASK-260909-xtvqf3_rev2_vectors.json`
- `TASK-260909-xtvqf3_rev2_run-gate.sh`
- `TASK-260909-xtvqf3_rev2_adoption.md`
- `TASK-260909-xtvqf3_rev2_gate-evidence.log` (full runner + self-check)
- `TASK-260909-xtvqf3_results_rev2.md` (this file)

Vectors derived at runtime from SPEC section 6 Codes column plus owner
constants (`fragment.Code*`, `composition.Code*`, `systemprompt.Code*`);
foreign = normative minus own; extras `environment_home_stale`,
`environment_unknown`, `invented_code`, `""`; forms direct/wrapped/joined.

## Evidence (real runs, Darwin arm64 Go 1.25.5)

Published runner executed exactly as published from `.temp/`
against a fresh disposable dir; full log in evidence resource.

- `./TASK-260909-xtvqf3_rev2_run-gate.sh <fresh-dir>`: exit 0
  - baseline diagnostics gate: exit 0, 6/6 named TestGate tests PASS
  - baseline framing gate: exit 0, 2/2 named TestGateFraming tests PASS
  - baseline full-package suites: exit 0
  - resolve mutant (admits `mcp_layer_missing`): exit 1,
    `TestGateResolveRejectsForeignCodes` reports
    `resolve accepted foreign "mcp_layer_missing"` (direct/wrapped/joined)
  - layer mutant (admits `resolve_invocation_failed`): exit 1,
    `TestGateLayerRejectsForeignCodes` (prior R3 survivor killed)
  - refusal mutant (admits `resolve_invocation_failed`): exit 1,
    `TestGateRefusalRejectsForeignCodes` (prior R3 survivor killed)
  - joined-positive mutant (rejects joined `resolve_invocation_failed`):
    exit 1, `TestGateJoinedOwnedAccepted` reports
    `joined owned resolve "resolve_invocation_failed" lost`
    (rev1 reviewer survivor killed)
  - framing mutant (equality on the exact complete Detail): exit 1,
    `TestGateFramingSingleDetailAtRealResolver` at
    `run -> Resolver.Resolve (ExecRunner, fixed absent binary) -> Emit/Line`
    reports `does not carry code "resolve_invocation_failed"`;
    `TestGateFramingCompanionsStayFramed` exit 0 (3/3 byte-exact);
    diagnostics gate under same mutant exit 0
  - diagnostics.go restored byte-equal after every mutant
- `./TASK-260909-xtvqf3_rev2_run-gate.sh --self-check`: exit 0,
  5/5 negative probes trip (stale destination, missing overlays,
  zero selection, baseline failure, surviving mutant)

## R1 review fixes (rev1 -> rev2)

- R1 fail-closed runner: published names and copy/adoption commands agree
  (runner embeds `TASK-260909-xtvqf3_rev2_*` sources); fresh destination
  required, stale refused; exact source identity pinned; every
  setup/extraction/copy/mutation/restoration step checked; baselines require
  named PASS lines; mutants require exit 1 plus named behavioral assertion
  (compile errors / `no tests to run` are runner failures); 5 negative
  self-checks included. Two real latent defects found by executing as
  published and fixed: `PIPESTATUS` reset under `set -u`, `git archive`
  empty-when-run-from-subdirectory, `tar -C` silent on `/../` paths.
- R2 owned positives: every owned code asserted in direct/wrapped/joined;
  joined typed-nil behavior (both orders + all-typed-nil Join) pinned;
  `TestGateJoinedOwnedAccepted` kills the exact reviewer joined-positive
  survivor; `TestGateCoverageCounts` pins 18 / 44 / 168 and states the
  mutable-Code (resolve 6, layer 2, refusal 2) vs fixed-code
  (UsageError, axconfig.Error) vs call-site-selected (defaults_unresolvable,
  plan_*, env_unsupported, exec_provider_missing, ax_handoff_failed) scope.
- R3 framing companions: split into `TestGateFramingCompanionsStayFramed`
  with byte-exact `Line()` assertions, executed independently under the
  mutant (exit 0) while the exempted-detail test fails (exit 1).

## Preservation and bounds

- No edits to `internal/`, `cmd/`, `SPEC.md`, or candidate blobs; only
  `.temp/TASK-260909-xtvqf3-gate-rev2/` and `.temp/TASK-260909-xtvqf3-publish/`
  (gitignored) written. Disposable copies restored byte-equal.
- Broad prefix-contains / whole-framing deletions noted as useful but not
  narrowing; gate uses only the five single-member narrowings above.
- Production call sites pinned in vectors: resolve
  `fragment.Resolver.Resolve`, layer `composition.Value.CheckLaunchBoundary`,
  refusal `systemprompt.Select/ProbeFiles via PrepareLaunch`.
- No tags/installs/real ax/model calls/hosted CI/private-record changes.

## Adoption

See `TASK-260909-xtvqf3_rev2_adoption.md`. Next diagnostics revision should
copy both gate test files verbatim, replace the hand-selected `strangers`
list with derived foreign sets, and add joined owned/typed-nil positives.
