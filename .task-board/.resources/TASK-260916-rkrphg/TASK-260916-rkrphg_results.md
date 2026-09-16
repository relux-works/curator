# TASK-260916-rkrphg: launcher-accept-claude-codex-aliases — results

Spec: curator-spec main 2d6d497 (profiles/manager.md CLI aliases).
Control root: curator-agent-launcher. Worktree: `.temp/STORY-260916-prdjid/worktree`.

## What changed

`curator-run <env-id>` accepts `claude` and `codex` as aliases of the
canonical `claude_code` and `codex_cli` ids. The operand normalizes to the
canonical id before any validation or lookup; the `curator env resolve`
argv, the default tracked ax session name, and all provenance lines carry
the canonical id only. Aliases are never persisted. Any other spelling
keeps the existing refusal. Wire ids, fragment schemas, defaults.json keys,
markers, and locks are unchanged.

Files:
- `internal/cli/cli.go`: added `NormalizeEnvID` (exact, case-sensitive,
  `claude`→`claude_code`, `codex`→`codex_cli`, else identity); `Parse`
  stores the normalized operand; `Usage` lists the aliases.
- `cmd/curator-run/main.go`: defensive `inv.EnvID =
  cli.NormalizeEnvID(inv.EnvID)` after `Parse`, before `fragment.Resolve`
  and `execution.Prepare`, so resolve argv and default ax name are canonical
  even if the parser is bypassed.
- `SPEC.md` (§3 table + parsing rules, §4.1 resolve, §4.6 default name),
  `README.md` (pipeline step 2, supported environments, tracked default
  name), `internal/cli` usage, `cmd/curator-run/testdata/help.golden`:
  document aliases, normalization-before-lookup, canonical-only outputs.
- Goldens: `internal/cli/testdata/cases.golden` (+5 alias shapes);
  `cmd/curator-run/testdata/forbidden-*.golden` (help-text refresh only);
  `pipeline-*.golden` unchanged (canonical behavior identical).

## Tests (production entry points)

- `internal/cli`: `TestNormalizeEnvID` (closed table + idempotence +
  near-miss/case/whitespace pass-through), `TestParseNormalizesAliases`
  (both spellings through `Parse`, native tail never normalized,
  near-misses verbatim), 5 new golden shapes.
- `cmd/curator-run` (`run` entry): `TestRunAliasesBehaveAsCanonical`
  (alias vs canonical byte-identical stdout/stderr, canonical resolve argv,
  flagged form), `TestRunAliasFragmentNeverAccepted` (alias-named fragment
  refused as `resolve_fragment_invalid`), `TestRunUnknownSpellingsStillRefused`
  (`Claude`/`CODEX`/`claudes`/`codexx`/`not_registered` verbatim argv +
  `resolve_environment_unknown`).
- `cmd/curator-run` (full pipeline): `TestProductionAliasEquivalence`
  (both aliases × tracked/untracked through `run` with real parser, tagged
  admission, fake provider/ax; canonical resolve argv, canonical default ax
  name, payload matches the canonical `pipeline-<id>-<tracked>.golden`
  byte-for-byte).
- `internal/defaults`: `TestLoadRejectsAliasKeys` (`claude`/`codex` keys
  rejected as `defaults_config_invalid` in both files; `Resolve(alias)`
  rejected).

## Evidence (shell: bash, `set -o pipefail`, real exit codes)

- `go build ./...` → exit 0
- `gofmt -l cmd internal` → exit 0, empty (FMT_CLEAN)
- `go vet ./...` → exit 0
- `go test ./internal/cli -count=1` → exit 0
- `go test ./internal/cli -run 'TestNormalizeEnvID|TestParseNormalizesAliases|TestParseAccepted|TestParseRejected' -count=1` → exit 0
- `go test ./internal/cli -run TestGolden -count=1` (before regen) → exit 1
  expected (new shapes); after `-update` + review → `go test ./internal/cli -count=1` exit 0
- `go test ./internal/defaults -count=1` → exit 0
- `go test ./internal/defaults -run TestLoadRejectsAliasKeys -count=1 -v` → exit 0
- `go test ./cmd/curator-run -run 'TestRunAliasesBehaveAsCanonical|TestRunAliasFragmentNeverAccepted|TestRunUnknownSpellingsStillRefused|TestRunHelpGolden' -count=1` → exit 0
- `go test ./cmd/curator-run -run 'TestRunAliasesBehaveAsCanonical|TestRunAliasFragmentNeverAccepted|TestRunUnknownSpellingsStillRefused|TestRunHelpGolden|TestRunUsageErrorsExit2|TestRunMapping' -count=1 -v` → exit 0
- `go test ./cmd/curator-run -run TestProductionForbiddenFlags -count=1` (before help refresh) → exit 1
  expected (help text); after `UPDATE_PIPELINE_GOLDENS=1` regen + review →
  rerun without UPDATE → exit 0
- `go test ./cmd/curator-run -run 'TestProductionPipelineGoldens|TestProductionAliasEquivalence' -count=1` → exit 0

Not run: full `go test ./...` / `make check` / `make race` (brief: host
stalls on big suites; remote gate runs at handoff). No cross-platform claim;
this host only.

## Narrowing mutants (all killed, sources restored)

- M1 (drop `claude` mapping): `go test ./internal/cli -run
  'TestNormalizeEnvID|TestParseNormalizesAliases'` → exit 1 (both tests fail
  on `claude`).
- M2 (`codex` maps to itself): same gate → exit 1 (both tests fail on `codex`).
- M3 (over-broad: also map `Claude`): same gate → exit 1 (verbatim tests fail).
- M4 (`NormalizeEnvID` identity): `go test ./cmd/curator-run -run
  TestRunAliasesBehaveAsCanonical` → exit 1 (alias resolve argv verbatim,
  not canonical). Proves `run()` production-entry coverage.
- After each mutant: bytes restored; `go test ./internal/cli -count=1` and
  `go test ./cmd/curator-run -run TestRunAliasesBehaveAsCanonical` return
  exit 0.

Coverage: rows "driven" only via `Parse` and `run` entry points above.
Helper-direct `NormalizeEnvID` table is a bound. No survivors.

## Notes

- `mapping.Resolve`, `fragment` parser, `defaults` keys unchanged: they
  accept only canonical ids, so aliases cannot persist and fragments stay
  `launch-env-fragment-v1` with canonical `environment`.
- `main.go` normalization is idempotent defense-in-depth; `cli.Parse` is the
  primary gate.
