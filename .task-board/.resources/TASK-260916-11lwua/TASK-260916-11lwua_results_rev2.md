# TASK-260916-11lwua — CLI aliases claude/codex, curator side — rev2 recovery results

## Why rev2 exists

Rev1 implementation was complete, but Change Request CR-TASK-260916-11lwua-1
revision 1 failed remote-gate validation (`scripts/remote-gate.sh`, exit 1):
`go test` was green on every lane, but the Windows platform-case gate failed
on exactly one skip:

```
FAIL  skip with an unrecognised reason on windows: cmd/curator :: TestRunDispatchNormalizesAliasOperand
      reason: no sh on this platform
```

(`TASK-260916-11lwua_change-request_rev1-validation.log`, run 35093912447.)

## The fix (2 edits, both in cmd/curator/envalias_test.go)

1. `TestRunDispatchNormalizesAliasOperand`: skip reason
   `no sh on this platform` → `no bash on this platform`. This is the
   established reason for POSIX-shell fixtures in this package
   (`TestUmbrellaDispatchesToProvider` uses the identical scripted
   `curator-run` fixture; `TestImplementedCommandWinsOverProvider` skips
   with the same reason and no ledger row — the exact posture this test
   now has). It matches `.github/ci/skip-classes.tsv`
   `host-capability: no bash → allow`, so Tier 2 admits it. No CI or
   ledger file was touched.
2. New `TestNormalizeRunEnvOperand`: pure-Go table test of the
   run-dispatch rewrite decision (alias→canonical, flag-value/unknown/
   non-run passthrough). No fixtures, no skips — it runs on every
   platform including Windows, so the AC headline behavior
   (`run claude/codex` → `claude_code/codex_cli`) is asserted on Windows
   even though the exec-plumbing test must skip there like its siblings.

No production code changed in rev2; the rev1 implementation was re-read
in full and kept. Spot checks re-verified: `EnvLockKey` keys on the knob
head only, so the alias spelling cannot bypass a system lock; `ByID`
still refuses raw aliases (`environment_unknown` pin test); wire ids,
markers, fragments, defaults, spec untouched.

## Scope readings (inherited from rev1, re-verified against code + spec)

- `env status` takes no environment operand in code
  (`cmd/curator/env.go`: `env status [--check] [--json]`, positionals are
  usage) or in the spec rule (profiles/manager.md: the operand is
  `env resolve`, `profile use --env`, `env unmanage --env` only). No
  operand was added; `TestEnvStatusPrintsCanonicalAfterAliasUse` proves
  status prints canonical ids only after alias-driven operations.
- `env unmanage` does not exist in this tree (spec names it; no
  implementation). Nothing to wire; its future implementation must call
  `NormalizeEnvID` on `--env`.

## Evidence (story worktree, real exit codes, shell sh)

- `gofmt -l cmd internal` → no output, exit 0
- `go build ./...` → exit 0
- `go vet ./cmd/curator/ ./internal/envregistry/` → exit 0
- `go test ./internal/envregistry/ -count=1` → ok, exit 0
- `go test ./cmd/curator/ -count=1 -run '<11 alias tests>'` → 11/11
  PASS incl. new `TestNormalizeRunEnvOperand`, exit 0
- Neighbors `go test ./cmd/curator/ -count=1 -run
  'TestEnv|TestProfile|TestUmbrella|TestImplemented|TestRun'` → ok,
  exit 0
- `go test ./internal/config/ -count=1` → ok, exit 0
- `golangci-lint run ./cmd/curator/... ./internal/envregistry/...` →
  0 issues, exit 0
- Gate simulation with the REAL `.github/ci/platform-case-gate.sh`
  under `CI_GATE_GOOS=windows` (synthetic single-skip stream, empty
  ledger — isolates Tier 2 classification; script at /tmp/alias-ev/
  in the run environment, not part of the deliverable):
  - reason `no bash on this platform` → `allowed-host-capability`,
    gate exit 0
  - control reason `no sh on this platform` → reproduces the exact
    rev1 `FAIL ... FATAL-unclassified` text, gate exit 1
- Mutant probe: `normalizeRunEnvOperand` stubbed to identity →
  `TestNormalizeRunEnvOperand` + `TestRunDispatchNormalizesAliasOperand`
  both FAIL (exit 1, killed, no survivors); restored → green, exit 0
- NOT run: full `go test ./...` / landing suite (brief: narrow only;
  the remote gate runs at handoff).

## Coverage notes

- Driven rows (through `run()`): unchanged from rev1, plus the new
  pure-Go rewrite test that additionally runs on Windows.
- Bounds: helper-direct knob/rewrite tables, the wire-layer pin, and
  the Windows skip of the exec-plumbing dispatch test (same bound as
  the sibling POSIX-fixture tests).
- Windows/macOS/Linux full-suite proof comes from the remote gate at
  handoff; the local gate simulation above proves only the skip-reason
  classification.
