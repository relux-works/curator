# TASK-260916-11lwua — rev3 rework results (canonical knob paths in env config output)

## The blocking finding (rev2 verdict, one item)

`cmd/curator/envconfig.go`: `env config set/unset` normalized the knob for
lookup but printed the ORIGINAL spelling in messages — `unset forms.claude`
printed `unset forms.claude`, absent-knob and lock refusals quoted
`forms.codex` / `isolation.acme.claude`. Fixed; existing unset tests only
checked exit status.

## The fix (production: `cmd/curator/envconfig.go` only)

- `cmdEnvConfigSet` and `cmdEnvConfigUnset`: after
  `segments = normalizeEnvKnob(segments)`, reassign
  `knob = strings.Join(segments, ".")`, so the lock-key lookup, the lock
  refusal, both not-set refusals, and the `unset <knob>` success line all
  name the canonical knob. For non-alias knobs `join(split(k)) == k`, so no
  behavior change outside aliases. `set` success still prints the value JSON
  only (no knob to leak); `show` likewise.
- No other production file touched in rev3; no wire-id, schema,
  defaults.json, or spec change. The withdrawn system_prompt_files finding
  was NOT implemented per the rework brief.

## New tests (`cmd/curator/envalias_test.go`, both through `run()`)

- `TestEnvConfigAliasOutputPrintsCanonicalKnob`: for `forms.claude` and
  `forms.codex` — not-set refusal names the quoted canonical knob (exact
  quoted-alias negative assertion); set prints the value JSON only;
  unset prints exactly `unset <canonical>`; canonical spellings produce
  byte-identical output (controls).
- `TestEnvConfigAliasLockRefusalPrintsCanonicalKnob`: with
  `environments.isolation` locked (the one alias-bearing lockable table),
  set/unset via `isolation.acme.claude` / `isolation.acme.codex` refuse
  naming the quoted canonical knob; canonical controls refuse with
  identical bytes; refusals write nothing to the machine file.

## Evidence (story worktree, shell bash, real exit codes)

- `gofmt -l cmd internal` → no output, exit 0
- `go build ./...` → exit 0
- `go vet ./cmd/curator/ ./internal/envregistry/` → exit 0
- `go test ./internal/envregistry/ -count=1` → ok, exit 0
- `go test ./cmd/curator/ -count=1 -run 'EnvConfig|Normalize|RunDispatch|Alias'`
  → ok, 18/18 PASS incl. the 2 new tests, exit 0 (rerun on the final tree)
- `golangci-lint run ./cmd/curator/... ./internal/envregistry/...` →
  0 issues, exit 0
- Kill proof (behavioral mutant restoring rev2 output: canonical join
  replaced by an identity rejoin, import kept live): both new tests FAIL
  with the reviewer's exact reproductions
  (`environments knob "forms.claude" is not set`,
  `environments knob "isolation.acme.claude" is locked by ...`), exit 1;
  restored → green, exit 0. No survivors.
- End-to-end reviewer probes with a built binary (`CURATOR_CONFIG` on a
  disposable file): `set forms.claude referenced` → `"referenced"`, exit 0;
  `unset forms.claude` → `unset forms.claude_code`, exit 0;
  `unset forms.codex` (absent) →
  `curator: environments knob "forms.codex_cli" is not set`, exit 1;
  machine file holds no alias bytes.
- NOT run: full `go test ./...` / landing suite (brief: narrow only; the
  remote gate runs at handoff).

## Bounds (orchestrator ruling, recorded not claimed)

- `curator env status` takes no environment operand in this tree (matrix
  output only) and `env unmanage` does not exist, so the literal
  "env status accepts both spellings" AC item is satisfied by the canonical
  matrix output alone — no operand was added, no unmanage invented.
- `run` dispatch normalization is proven through a scripted provider plus
  the cross-platform pure-Go rewrite test; real launcher launch/refusal is
  the sibling task's scope. No LOGBOOK.md edit (campaign rules forbid it).
