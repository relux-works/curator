# TASK-260910-gocke2 results — manager-mcp-install-surfacing (S4, `s4-warn`)

Story STORY-260910-1lf0m5, wave 1. Spec: curator-spec `23dafa7`
(`protocol/environments.md` §2.2/§2.3/§9.1/§9.2/§10.3/§12/§12.1/§12.2),
vectors `conformance/v1/vectors/environments-env-passthrough.json`.

Shipped profile: **`s4-warn`** (warning release). `s4-enforce` is
implemented behind the same option and follows in a later release.

## Per-file changes

Production:

- `internal/envfragment/envfragment.go` — S4 core. `S4Profile`
  (`s4-warn`/`s4-enforce`, :216/:220), shipped default
  `ActiveS4Profile = S4Warn` (:227, the one constant the flip moves),
  diagnostics `mcp_env_passthrough_unlisted` /
  `mcp_env_passthrough_dropped` (:201/:204), `MigrationHint` (:230),
  `EffectivePassable` (:247), `ResolvePassthrough` (:264, reserved
  exclusion first and silent, then the effective list; warn/dropped
  warnings), `BoundEnvNames` (:314) reimplemented as the active-profile
  resolution with nil read as an absent knob (pre-S4 behaviour
  preserved under `s4-warn`).
- `internal/envregistry/envregistry.go` — `MachineConfig` gains
  `PassableEnvNamesSet` (:548): absent knob vs explicit null.
- `internal/config/environments.go` — `Environments` gains
  `PassableEnvNamesSet` (:36), set on parse (:228); `render` (:926)
  renders the §12.1 schema default `[]` for absent, `null` for
  explicit null, the list otherwise. Merge is raw-map level, so
  presence flows through the system overlay unchanged.
- `internal/contextmaterialize/mcp.go` — `MCPDeclaration` (:134) and
  `FormatDeclarationRows` (:152): byte-exact §2.3 rows, ascending
  package-name byte order, compact JSON arrays (`SetEscapeHTML(false)`,
  nil → `[]`), `-`/`[]` for http, one LF per row, no rows for empty.
- `internal/envprofile/surfacing.go` (new) — `AllowlistEmptyWarning`
  (:15, `mcp_package_allowlist_empty`, states every declaration
  package in the closure is admitted) and `surfacingRows` (:29, lock
  MCP members + store manifests → rows; unreadable reported, never
  absent, never fatal).
- `internal/envprofile/envprofile.go` — `DiagAllowlistEmpty` (:70);
  `Info.Surfacing` (:172); install (:695), reinstall move (:826) and
  reinstall-identical (:803), update move (:1044) and
  update-identical (:1027) compute surfacing + the allowlist warning
  after the audit gate and before (or without, when nothing moves)
  lock publication; `activateReinstall` threads surfacing (:883).
- `internal/envprofile/managed.go` — `buildFragment` resolves the
  §10.3 bound once per resolution through `ResolvePassthrough` with
  the machine knob + presence + active profile (:1487), warning into
  the verdict warnings.
- `internal/envprofile/status.go` — posture: `S4Profile`,
  `PassableEnvNames` (nil = unbounded), machine `Warnings`,
  `MCPDeclarations` per reported scope (:128–:147); computed in
  `StatusOf` (:218+) via `declarationScopes` (:494, machine first
  then scoped currents; unreadable locks/declarations warned, never
  empty); currency untouched by warnings.
- `cmd/curator/profile.go` — install/import/update print `Info.Surfacing`
  verbatim to stdout after warnings, before activation/result lines.
- `cmd/curator/env.go` — `machineFromConfig` (:17) threads the
  effective `passable_env_names` + presence into resolve and status.
- `cmd/curator/envstatus.go` — `s4_profile:` + effective list row,
  machine warnings, per-scope declaration groups; `formatPassable`
  (:104, nil → `unbounded`).
- `CHANGELOG.md` — Unreleased S4 warning-release entry.
- `.github/ci/platform-cases.tsv` — ledger row for
  `TestEnvironmentsEnvPassthroughVectors` (all platforms,
  `root-content`).

Tests (all committed, all in-repo layout):

- `internal/envfragment/envfragment_test.go` — S4 matrix (both
  profiles × absent/null/list/empty/reserved), effective defaults,
  shipped-default pin.
- `internal/contextmaterialize/mcp_test.go` — row bytes incl.
  ordering, http dash, space/quote args, empty/nil edges.
- `internal/config/environments_test.go` — knob presence × rendering
  (absent→`[]`, null→null, list, explicit `[]`, invalid rejected).
- `internal/envprofile/envpassthrough_conformance_test.go` (new,
  ledger-listed) — executes EVERY case of
  `environments-env-passthrough.json` (7 resolution incl. 2 negative,
  6 allowlist incl. refusal + negative, 4 schema, 6 surfacing incl. 2
  negative rows, 2 order) through the production entries, with the
  `root-content` skip when the root lacks the family.
- `internal/envprofile/surfacing_test.go` (new) — git install rows
  (http+stdio, ordering), git update candidate rows, warning gate
  across install/update/status (incl. non-empty silent), resolve
  knob matrix through `Resolve`, status posture, unreadable-manifest
  §8.4 edge.
- `cmd/curator/profile_surfacing_test.go` (new), `cmd/curator/env_test.go` —
  CLI: install lists the stdio row before the installed report +
  warns; update lists the new row; declaration-free install warns
  with no rows; status posture text + JSON; configured-list posture.

## How each AC line is met

- "Install output lists stdio commands" —
  `cmd/curator/profile.go` prints `Info.Surfacing` (exact
  `mcp-declaration … command=<stdio> …` rows); covered by
  `TestProfileInstallListsStdioCommands` (CLI) and
  `TestInstallSurfacesMCPDeclarations` (entry point).
- "warnings implemented" — `mcp_package_allowlist_empty` at install,
  update, status (`AllowlistEmptyWarning`, emitted at the five
  `envprofile.go` sites + `StatusOf`); `mcp_env_passthrough_unlisted`
  / `mcp_env_passthrough_dropped` at resolve (`managed.go:1487`);
  covered by the vector test, `TestAllowlistWarningOperations`,
  `TestResolvePassthroughKnobs`, and the CLI warning tests.
- Both S4 profiles behind one option, `s4-warn` shipped —
  `envfragment.ActiveS4Profile`; absent knob unbounded + unlisted
  warning naming variables + knob + hint; enforce absent = empty
  with drop warning; explicit null unbounded silent under both;
  reserved excluded (`contextpkg.ReservedEnvName`, silent).
- §2.3 rows at install/update after audit, before publish; repeated
  by status; closed columns/order; no rows when empty; posture shows
  profile + effective list; warnings never non-current — as above;
  order proven at the observable edges (blocked audit ⇒ no rows;
  rows describe the candidate lock).
- Vector test from `CURATOR_CONFORMANCE_ROOT` with root-content skip
  + ledger row — `TestEnvironmentsEnvPassthroughVectors` + the
  `platform-cases.tsv` row; skip reasons verified (`… is not set`,
  `… publishes no environments-env-passthrough vector`).
- CHANGELOG S4 warning-release entry — done, with migration hint and
  the `s4-enforce`-follows note.

## Validation transcripts (exit codes real)

Shell: `bash`, repo root, `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1`
unless noted. Shared host: sibling agents' suites ran concurrently;
see the contention note.

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0, no findings.
- `gofmt -l internal cmd` → clean (repo-wide `gofmt -l .` lists only
  pre-existing `.task-board/.resources` scratch files, untouched).
- `golangci-lint run` on the six touched packages → exit 0, 0 issues
  (after two `#nosec G101` diagnostic-code annotations and
  unused-parameter renames in the new test stub).
- `go test ./internal/envfragment/ ./internal/contextmaterialize/`
  → ok (0.5s/0.7s).
- `go test ./internal/envprofile/ -run TestEnvironmentsEnvPassthroughVectors`
  → ok, 25/25 subtests pass (15.3s); both skip paths verified.
- New envprofile tests (7) → all pass (74s).
- New CLI tests (5), individually and combined → all pass (59s combined).
- Full `internal/envprofile` in 5 bounded `-run` chunks (partition
  verified complete/disjoint): A–G ok 103s, I ok 171s, L–O ok 61s,
  P–R ok 98s, S–W ok 111s. (One unchunked run hit the 600s go-test
  timeout at 99% completion under parallel load — cumulative time,
  not a hang; the chunks are the evidence.)
- Full `internal/config`: only pre-existing failures remain —
  `TestManagerConfigV2Vectors` (22 subtests) and the schema-case
  suites (19 subtests) fail IDENTICALLY at baseline and with this
  change (verified by HEAD-vs-S4 `comm` diff: zero fixed, zero
  regressed). Every remaining diff is a sibling-owned knob this
  task must not touch: `provider_directories`, `system_module_waivers`,
  `transitive_system_modules` (plus sibling overlay-discriminator
  schema cases). This change's scope renders correctly:
  `passable_env_names:[]` now matches where the vector expects `[]`,
  `null` where it expects `null`, and
  `schema2-passable-env-invalid-name` passes.
- Full `cmd/curator` in 13 bounded `-run` chunks (partition verified
  173/173, disjoint, counts sum exactly): A–B ok 38s; C1
  (CLI/Check/Classify/Currentness) ok 209s; C3 (Config) ok 2s;
  CompiledInventory ok 48s; the 3 slow compiled tests singly ok
  248s/240s/179s; D–E ok 130s; F–G ok 76s; H–O ok 6s; P-a ok 22s;
  P-b (install set, incl. the 3 new surfacing tests) ok 62s; P-c ok
  42s; R–S ok 67s; T–W ok 1s. (Two unchunked C attempts died first:
  one on the 8m go-test timeout with 5 parallel compiled builds
  stuck in `os/exec` child waits, one on the 9m session watchdog —
  both contention artifacts; every member passes in the splits.)
- `go test ./internal/envregistry/` → ok (0.4s).

## Out of scope (deliberate)

- `s4-enforce` as default (later release flips the one constant).
- Sibling config knobs (`provider_directories`, `system_module_waivers`,
  `transitive_system_modules`) and sibling overlay-discriminator schema
  cases — their vector failures are pre-existing and byte-identical.
- Closed MCP launch interpreter contract, E3 codex seed, S1/S3
  defaults, SPEC_PIN (left at the released pin), spec edits.
- Pinned-root updates emit no surfacing (no audit gate runs, no
  candidate resolves); identical-lock install/update still surface
  the candidate set (documented in code).
- Full `MachineConfig` wiring from file knobs (forms/isolation/XDG
  still take `DefaultMachineConfig` in production): only
  `passable_env_names` is threaded, which is all this task needs
  (finding below).

## Findings / spec gaps (reported, spec not patched)

1. Production resolve/status built `Machine` from
   `DefaultMachineConfig()`, so NO file knob (forms, isolation, XDG,
   passable) reached resolution. This task threads `passable_env_names`;
   the rest is a follow-up for the owning story.
2. §10.3's "passed variable outside a configured list" under `s4-warn`
   is unreachable by construction (the list bounds, silently) — the
   vectors confirm with the silent explicit-list case. Implemented
   per vectors.
3. `env status` reads MCP manifests from the store; an unreadable one
   warns (codeless, §8.4-style) rather than failing or hiding.
4. One combined CLI run flaked at 300s with zero per-test output on
   the shared host; the identical set passes in 59s and every member
   passes individually. Treated as contention flake, recorded here.
5. Host outage at handoff: macOS `syspolicyd` down, so every
   ad-hoc-signed binary launch (incl. the board CLI) wedged in
   `dyld_start` (unsigned binaries launched fine). Final board writes
   (`set_notes`, item 10, handoff) went through a `/tmp` copy of the
   current CLI with only the signature stripped
   (`codesign --remove-signature`; code bytes identical, original
   untouched). A sibling agent was observed waiting on syspolicyd
   recovery; host admin follow-up recommended.

Checklist: all definition-of-done items satisfied. The only red in
the narrow scope is the pre-existing, sibling-owned
`manager-config-v2` vector/schema surface (byte-identical at baseline,
zero regressions), which goes green when sibling tasks land their
knobs.
