# TASK-260905-30zs8t drafting report — stage (a) core

Branch: `feat/agent-environments-stage-a` (curator), 4 signed commits,
linear, no push. Spec authority: curator-spec `f39f4a9`
(`protocol/environments.md` rev 1.1, schemas v1, conformance v1).

## Commits (all `Good "git" signature for oparin@me.com`)

- `aa1e9ea9` core libraries + unit tests (pkgversion, contextpkg,
  contextresolve, contextlock, contextstore, contextaudit,
  contextmaterialize, envmarker)
- `8a0f4ca0` conformance subset tests + `platform-cases.tsv` ledger rows
- `4f4773f4` envprofile manager + profile CLI rows + README tools section
- `7238412c` follow-up: audit warnings surfacing, scoped use/clear,
  listing discipline, remaining refusal tests

## Package map

- `pkgversion` — strict SemVer 2.0 `v`-tags without build metadata;
  closed npm-shaped ranges (caret incl. 0.x/0.0.x, tilde, comparators,
  x-ranges, partial coercion, `-0` upper bounds, `||`); `latest` = `*`;
  hyphen ranges and `v` inside ranges rejected; prerelease same-triple
  rule; total order. Entry: `ParseTag`, `ParseRange`,
  `Range.Satisfies`, `Compare`, `Sort`.
- `contextpkg` — `agent-context.json` / `agent-mcp.json` strict readers
  (unknown fields rejected; `weights` root-parsed; modules with
  `path`/`environments`/`class`; requires with exactly one of
  range/tag/revision; stdio bare `command`; `https` URL grammar;
  `env_names` grammar + manager-reserved exclusion). Diagnostics exactly
  `context_manifest_invalid`, `mcp_declaration_invalid`,
  `profile_module_missing`, `profile_module_bytes_invalid`,
  `profile_selector_unknown_environment`.
- `contextresolve` — joint resolution (root + overlays + skills + MCP,
  downward re-selection, termination); `context_range_conflict`,
  `context_version_mismatch`; weights manifest → agreeing edges → root
  map (`context_weight_conflict`, `context_weights_not_root`).
- `contextlock` — context-lock-v1: (kind,name) order, `commit` |
  `state_sha256` pins, effective weights, requirer chains, overlay flag;
  CCJ-1 bytes, `lock_sha256` as `sha256:<hex>`.
- `contextstore` — commit-/state-keyed immutable entries under
  `<home>/contexts`; git kind via `gitops.Extract`, state kind for
  path/local.
- `contextaudit` — `context-secret-material` (unpinnable, blocking,
  scoped waivers) over modules, manifests, `CONTEXT.md`;
  `context-system-module-present` always-warn. Entry: `Detect`,
  `DetectFiles`, `SystemModules`.
- `contextmaterialize` — `curator-root-context-v2` header, chapter
  parts in emitted weight order under both precedence primitives,
  no-chapter/zero-module cases, joining, system-prompt output, hashes.
- `envmarker` — `agent-environment-marker-v1` strict reader/writer.
- `envprofile` — profile store, git/path install (resolve + strict
  audit + store + lock), update with old-lock-stands, guarded remove
  with purge, linked switching (scope-wide attempt, versioned backups
  retention 5, markers, `profile_use_partial`), sync, builtin `local`
  `default` migration.

CLI (`cmd/curator/profile.go` in `run()` + usage): install, list, use
(+`--env`, `--clear`), update (`--all`), remove (`--purge`), sync;
`compose` refuses with the schema-2 bound.

## Vector families passed (candidate lane, spec `f39f4a9` conformance/v1)

- context-versions.json: 15 version + 1 ordering + 50 range + 63
  satisfies cases — PASS
- environments.json: 28 resolution + 3 lock cases (CCJ-1 bytes,
  lock_sha256) — PASS
- environments.json: 4 header cases — PASS
- environments.json: 19 materialization cases — 12 monolithic +
  system-prompt sets byte-for-byte PASS; 7 referenced/MCP subtests skip
  (opt-in stage (b))
- context-detectors.json: 3 pattern classes + 12 cases — PASS
- Final tally on committed tree: 5 top-level PASS, 7 sub-skips, 0 FAIL
- rc.9 root lacks these files: each family skips with registered
  root-content class there (candidate lane runs them).

## Gate outputs (real exit codes, standalone processes)

- `go build ./...` — 0
- `go vet ./...` — 0
- `gofmt -l cmd internal` — clean
- `golangci-lint run` (9 new packages + cmd/curator) — 0, 0 issues
- `go test -race` (9 new packages) — all ok
- `ledger-consistency.sh` — 0, 103 rows
- `gate-selftest.sh` — 0 (81 passed); `no-broad-suppression.sh` — 0
- `test-gate.sh` with candidate root (full suite + platform-case
  enforcement) — 0 (`go test exit=0, platform-case gate exit=0`;
  5 new ledger rows ok; 26 skips all tolerated/allowed)
- `go test ./cmd/curator/ -count=1` on final tree — 0 (278s)
- `go test` on 9 new packages, final tree — 0
- Follow-up delta (7238412c, post-gate): covered by the re-runs above
  (envprofile full, cmd/curator full, interop subset, lint/vet).

## Unit/CLI tests: 81 functions across the 9 packages + cmd/curator

Every gate has negative tests; every gate ships a narrowing mutant
below. Production call sites are named in each test's doc comment.

## Mutants (all killed; `cmp`-verified clean reverts, packages re-green)

| Mutant | Narrowing (gate stays, admits exactly one) | Named failing test | Exit |
|---|---|---|---|
| M1 pkgversion: `-` primitive → `*` | hyphen ranges parse | TestHyphenRangesAreRejected | 1 |
| M2 contextpkg: `surprise` key allowed | one unknown field admitted | TestUnknownFieldIsRejected | 1 |
| M3 contextaudit: bearer `{20,}`→`{21,}` | 20-char tokens admitted | TestMCPArgsAndURLAreInScope | 1 |
| M4 contextlock: order check from index 2 | one inversion admitted | TestCanonicalOrderIsEnforced | 1 |
| M5 envprofile: ledger check only with prior marker | unrecorded overwrite admitted | TestUsePartialRecordsNothing | 1 |
| M6 contextaudit: aws severity→warning, token match kept | blocking lost, token kept; unit + behavioral suites run | TestSecretIsBlocking + TestConformanceContextDetectors | 1 / 1 |

## AC coverage: 8 of 8 brief items delivered; 7 of 12 CLI rows driven

- Items 1–7 fully driven through production entry points by named
  committed tests (libs via unit + vector tests; install/list/use/
  update/remove/sync via `envprofile.*` + `run()` CLI tests).
- Item 8 (migration): `EnsureDefault` builtin `local` default umbrella
  driven (TestEnsureDefaultCreatesLocalProfile); three sub-behaviors are
  stated bounds: overlays input (no schema-2 knob), waivers knob (no
  machine-config surface), global↔lock declaration wiring + fetch-only
  global update/upgrade (cross-epic skill-pipeline ownership).
- CLI rows: install/list/use/update/remove/sync + compose-refusal
  driven (11 CLI tests); 5 `env` rows are stage-(b) bounds (no stub
  ships); `profile use --target` names `environment_target_unknown`
  (driven).
- Other bounds: precedence always defaults; retention default 5 (no
  knob); native-home `$HOME` defaults are implementation choices (tests
  pin all four variables); `environment_backup_exists` untriggered by
  test (race-only path); LOGBOOK.md unwritten per the brief's ban
  (incident recorded here + board notes).

## Skipped items with reasons

- Referenced-form + MCP expected sets: opt-in skips (`CURATOR_STAGE_B=1`),
  stage (b) per brief.
- `env` family, managed homes/seeds/passthrough, MCP channel files,
  composition CLI, manager-config schema 2, `curator run`/ax: out of
  scope per brief (stage b/c/d).

## Anomalies

- Unpinned switching tests wrote into the operator's real homes
  (`~/.claude/CLAUDE.md` overwritten, `~/.codex/AGENTS.md` replaced by
  a dangling symlink, markers + backups in four homes). Cause: two
  tests did not pin home variables and the early ledger check allowed
  overwriting unrecorded files when no prior marker existed. Fixed:
  all tests pin homes; the ledger fails any write to an unrecorded
  file. Operator files restored byte-identical from the operation's own
  backups (`cmp`-verified; copies kept in /tmp/stagea-restore-evidence/
  per the scratch-file rule). Net regression cover:
  TestUsePartialRecordsNothing + mutant M5.
- Full-suite `go test ./cmd/curator/` takes ~4–8 min: split from the
  unit runs per the headless single-call bound; tails are in this
  report, streams in /tmp/stagea-*.txt.
