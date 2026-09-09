# TASK-260909-xtvqf3 results: diagnostics CR2 conformance gate

Status: ready for review (developer handoff).
Scope: task-scoped gate only; no production changes.

## Baseline identity (exact candidate)

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
- Verified via `git ls-tree` + `git archive` into disposable copies
  (`/tmp/gate-evidence`, `/tmp/gate-trial2`); original workspace untouched
  (`git status --short` empty, `.temp/` gitignored).

## Gate artifacts (attached, task-scoped)

- `TASK-260909-xtvqf3_gate-conformance_test.go` sha256 `b93a2f483546e21e5a23072126ec9feb38d3b2b533464eafcf678e3eb93b0a92`
- `TASK-260909-xtvqf3_gate-framing_test.go` sha256 `71bba3181d23d427773cee1a605e14c8e587bed2dd4acfe000abd1c6812729d4`
- `TASK-260909-xtvqf3_vectors.json` sha256 `8c47ee10dd810a33d455d4a04373039fdcd89a21a1f19bed226b4ded12d6805e`
- `TASK-260909-xtvqf3_run-gate.sh` sha256 `b3a503faadaef5244fbf1a4373dcf97b23ae994c8a9e4408928bc74134981683`
- `TASK-260909-xtvqf3_gate-evidence.log` (full runner + supplement)
- `TASK-260909-xtvqf3_adoption.md` (adoption instructions)

Vectors derived at runtime from SPEC section 6 Codes column plus owner
constants (`fragment.Code*`, `composition.Code*`, `systemprompt.Code*`);
foreign = normative minus own; extras `environment_home_stale`,
`environment_unknown`, `invented_code`, `""`; forms direct/wrapped/joined.

## Evidence (real runs, Darwin arm64 Go 1.25.5)

Runner: `.temp/TASK-260909-xtvqf3-gate/run_gate.sh /tmp/gate-evidence`
(full log in gate-evidence resource).

- baseline diagnostics gate `go test ./internal/diagnostics -run TestGate`: exit 0
- baseline framing gate `go test ./cmd/curator-run -run TestGateFraming`: exit 0
- baseline full with gate `go test ./internal/diagnostics ./cmd/curator-run`: exit 0
- resolve mutant (admits `mcp_layer_missing`): `TestGateResolveRejectsForeignCodes`
  `resolve accepted foreign "mcp_layer_missing" as "mcp_layer_missing" (direct/wrapped/joined)`, exit 1
- layer mutant (admits `resolve_invocation_failed`): `TestGateLayerRejectsForeignCodes`
  `layer accepted foreign "resolve_invocation_failed" as "resolve_invocation_failed"`, exit 1
- refusal mutant (admits `resolve_invocation_failed`): `TestGateRefusalRejectsForeignCodes`
  `refusal accepted foreign ...`, exit 1
- framing mutant (equality on complete Detail
  `could not run /nonexistent-missing-curator-xtvqf3 env resolve pi --profile normal\ncurator-run: usage: forged ...`):
  `TestGateFramingSingleDetailAtRealResolver` at
  `run -> Resolver.Resolve (ExecRunner, fixed absent binary) -> Emit/Line`
  `diagnostic line "curator-run: usage: forged ..." does not carry code "resolve_invocation_failed"`, exit 1;
  diagnostics gate under same mutant still exit 0 (other framing intact)

Hole proof (supplement): old `TestCodeOfRejectsForeignFamily` exits 0 for
both layer and refusal mutants (SURVIVED); new `TestGateLayer/Refusal` exit 1
(KILLED). Each mutant leaves other owners passing, proving single-member.

## Preservation and bounds

- No edits to `internal/`, `cmd/`, `SPEC.md`, or candidate blobs; only
  `.temp/TASK-260909-xtvqf3-gate/` (gitignored) written.
- Disposable copies via `git archive`; restored byte-equal after each mutant.
- Broad prefix-contains / whole-framing deletions noted as useful but not
  narrowing; gate uses only single-member narrowings above.
- Production call sites pinned in vectors: resolve `fragment.Resolver.Resolve`,
  layer `composition.Value.CheckLaunchBoundary`, refusal `systemprompt.Select/ProbeFiles via PrepareLaunch`.

## Adoption

See `TASK-260909-xtvqf3_adoption.md`. Next diagnostics revision should copy
both gate test files verbatim and replace the hand-selected `strangers` list.
