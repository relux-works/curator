# TASK-260922-1wvwc3 results — F-M1b: versioned capability table, drift refusal, release-ready

Developer handoff evidence. Worktree: skill-agents-management
`.temp/STORY-260922-h3epwn/worktree`, branch
`task-board/story/STORY-260922-h3epwn`, work UNCOMMITTED (integration is
the orchestrator's step). No tag cut: the signed annotated tag is the
orchestrator's step per the brief.

## Expected release

Next patch of v0.5.x after the latest tag v0.5.16 is **v0.5.17**.
CHANGELOG.md is release-ready (F-M1b entry first under Unreleased, names
the token and the expected version). Tags verified:
`git tag --list 'v0.5*' | sort -V | tail` ends at v0.5.16.

## Grammar version token (launcher cites this)

`permission-grammar-v1` — const `agentic.PermissionGrammarV1`
(`pkg/agentic/system.go`), stated in README (capability-table bullet)
and CHANGELOG. Drift refusals name it in the diagnostic.

## Capability table (environment, tool release) -> grammar version

| environment | verified release | grammar | yolo |
|---|---|---|---|
| `claude-code` | 2.1.261 | `permission-grammar-v1` | `--dangerously-skip-permissions` |
| `codex` | 0.153.2 | `permission-grammar-v1` | `--dangerously-bypass-approvals-and-sandbox` |
| `pi` | 0.84.2 | `permission-grammar-v1` | unsupported (`ErrPermissionModeUnsupported`) |
| `pi-native` | 0.84.2 | `permission-grammar-v1` | unsupported (`ErrPermissionModeUnsupported`) |

Each plugin holds its own environment's rows (`verifiedReleases` in
`policy.go`); `agentic.LookupReleaseCapability` is the single reader.
There is deliberately NO central `map[SystemID]` table: the
single-source guard forbids a second binding table, and Decision 0013
D5 keeps the mapping with each plugin. All four packages' tests plus
the singlesource guard pass (see validation).

## New request members and sentinels

- `LaunchRequest.ToolRelease`: established release; read ONLY on the
  interactive yolo path. Ignored elsewhere (pinned by test, not implied).
- `LaunchRequest.NativeArgs`: caller native args, forwarded VERBATIM
  after the module-spelled argv (yolo bypass lands before them, 0018
  item 1). Refused outside interactive (`ErrNativeArgsNotInteractive`).
- `ErrPermissionModeUnverifiedRelease` (drift: unpinned/newer/empty
  release, or a row naming an unimplemented grammar), `ErrNativePolicyUnknown`
  (unknown policy form; caller maps to usage exit 2),
  `ErrToolReleaseUndetected` (probe failure; caller passes `""`).

## Behavior

- Yolo: prefix-duplicate check (F-M1a, first, untouched) -> release
  lookup -> support read -> closed-grammar scan of flag positions ->
  argv `[model, effort, bypass, ...verbatim]`. Any unknown `-c` key /
  `--permission-mode` value (separate, `=`, attached-short for `-c`;
  malformed included) refuses as usage, never resolved into a claim.
- Native: no lookup, no scan, verbatim forward at ANY release
  (0018 item 4: no argv inspection; raw bypass may pass untracked).
- Pi/pi-native yolo: lookup first (drift), then the unchanged
  unsupported refusal; the verified release stays unsupported.
- Known policy selectors under yolo are classified but forwarded
  verbatim — 0018 item 4's conflict table is a later leaf (documented
  in README + CHANGELOG). Headless detection over the same args is
  F-L1's (item 5). No launcher/curator/ax/task-board change.

## Parsing rule (owned here, `internal/nativeargs`)

`--` ends flag parsing; before it, flag-shaped (len>1, dash-leading)
elements are flags; positionals, a lone `-`, and everything from `--`
on are prompt text, never parsed. `=`-forms read as their flag
(`SplitFlagValue`, first-`=` cut). Scans use `FlagIndexes` only.

## Probes (`ProbeToolRelease`, per-plugin + dispatcher)

Resolve via the plugin's own `ResolveBinary`, run `<binary> --version`
(`internal/toolprobe`: exact argv, ctx + own timeout, 64KiB cap, handed
env or empty — never ambient), parse per-tool output. Real captures
used for the grammars (stdout, exit 0): claude `2.1.274 (Claude Code)`
(first field), codex `codex-cli 0.153.4` (exact tagged pair), pi
`0.84.2` (whole answer). Installed claude/codex are NEWER than the
pins — drift fires on this machine by design. `pi` (wrapper) has no
probe: the wrapper's version is not pi's release (dispatcher returns
undetected; covered by test). Verified by probing installed tools with
`--help` only (no session): `-c<attached>`, `--config=`, and
`--permission-mode=` all parse (exit 0 + help), bogus flags error;
codex's own tip confirms `--` separates values.

## Goldens (exact argv via real BuildPlan; negatives via errors.Is)

- Positive per release: yolo + pinned release -> F-M1a argv
  byte-identical (claude `--model m --effort e --dangerously-...`,
  codex `-m m -c model_reasoning_effort="e" --dangerously-...`), plus
  yolo + pinned + NativeArgs -> bypass BEFORE the verbatim suffix.
- Negative pair per tool: yolo + newer (`2.1.274`/`0.153.4`/`0.85.0`,
  `9.9.9`) or `""` -> `ErrPermissionModeUnverifiedRelease` (names token
  + verified releases); native + same -> builds verbatim.
- Unknown forms: codex `-c`/`--config` separate, `=`, attached-short,
  dangling, no-`=`, empty/whitespace key, case-exact; claude mode
  separate, `=`, empty, dangling, case-exact — all usage refusals.
- Narrowings: all six claude modes, all five codex keys +
  `mcp_servers.*`, `--configx`/`-C` (exact longs, case-sensitive
  shorts) build and forward; native + unknown forms forwards.
- Parsing both directions: same text refused pre-`--`, forwarded
  post-`--` (unknown key, unknown mode, bypass-as-prompt vs duplicate);
  positional and lone `-` forward.
- Probes: fake-binary release rows, argv capture (`--version` exactly),
  unresolvable/no-PATH/non-zero/garbage/oversize/ctx-fired refusals.
- F-M1a fallout: six one-line `ToolRelease` pins (2 claude, 2 codex, pi,
  pinative) + regress `toolRelease` per case + new regress drift test.
  No F-M1a production line reshaped (additive gates only).

## Narrowing mutants: executed and killed (exit 1 each)

Shell: bash, `set -o pipefail`. Each mutant applied, narrow `-run`
killing test observed FAIL, then exact-reverse restored (grep confirms
zero `MUTANT` markers remain; full narrow suite re-run green after).

| # | bound (production gate) | narrowing mutant | killing test | result |
|---|---|---|---|---|
| B1 | drift-unverified — `LookupReleaseCapability` (`pkg/agentic/system.go`) | admit any unknown non-empty release as verified | claude `TestYoloRefusesDriftWithANamedDiagnostic` (2.1.274, 9.9.9 rows) | FAIL `err = <nil>`, exit 1 (`""` row still refused: true narrowing) |
| B2 | detection-failure — same gate | admit `""` as the first verified row | same test (`unestablished` row only) | FAIL `err = <nil>`, exit 1 |
| B3 | unknown codex `-c` key — `checkConfigOverride` (`codex/policy.go`) | refuse unknown keys only past 10 chars | codex `TestYoloRefusesUnknownConfigKeys` (all `future_key` len-10 rows) | FAIL `err = <nil>`, exit 1 |
| B4 | unknown claude mode — `scanNativePolicy` (`claude/policy.go`) | refuse unknown modes only past 11 chars | claude `TestYoloRefusesUnknownPermissionModes` (11-char and shorter rows) | FAIL `err = <nil>`, exit 1 |
| B5 | parsing rule — `FlagIndexes` (`internal/nativeargs`) | drop the `--` break | nativeargs `TestFlagIndexesStopsAtTheSeparator` (`[1 2 3]` vs `[1]`) + claude `TestPromptTextIsNeverParsedAsAFlag` (prompt refused as policy/duplicate) | FAIL, exit 1, both levels |
| B6a | native-args duplicate, claude — `scanNativePolicy` | check index 0 only | claude `TestYoloRefusesTheBypassFlagInNativeArgs/at_a_non-zero_index` | FAIL `err = <nil>`, exit 1 |
| B6b | native-args duplicate, codex — same shape | check index 0 only | codex twin test, same row | FAIL `err = <nil>`, exit 1 |
| B7 | NativeArgs scope — `buildPlan` (`pkg/agentic/plan.go`) | `mode != interactive` narrowed to `mode == exec` | core `TestBuildPlanRefusesNativeArgsOutsideInteractiveLaunches` (dry-run, managed-session) | FAIL `err = <nil>`, exit 1 |
| B8 | probe parse — `parseToolRelease` (`claude/probe.go`) | accept any first field | claude parse table + `TestProbeToolReleaseFailsClosed/an_unparsable_answer` | FAIL, exit 1, both levels |

Helper-direct bounds disclosed: `parseToolRelease` tables and the
synthetic grammar-v2 lookup row (production rows are all v1). All other
rows drive `BuildPlan` or `agentic.ProbeToolRelease`.

## Validation run by me (bash, `set -o pipefail`, exit codes real)

- `go build -mod=mod ./...` — exit 0 (post-restore).
- `go vet -mod=mod ./pkg/agentic/... ./internal/nativeargs/
  ./internal/toolprobe/ ./internal/regress/` — exit 0.
- `gofmt -l pkg/ internal/ tools/` — empty (clean).
- Narrow suites, final tree, all `ok`, exit 0 (sequential `-p 1`,
  `-count=1`): `pkg/agentic`, `systems/{claude,codex,pi,pinative}`,
  `internal/{nativeargs,toolprobe,regress}` (8/8), plus blast radius
  `systems/{agy,qwen,gemini,muse}`, `internal/{argvguard,gosources}`,
  `pkg/{vendorplugin/...,localruntime,providerlimits,plugin,
  inferenceengine/...}`, `internal/{launchenv,runtimeenv,mcpjson,
  ident}` — all `ok`, exit 0.
- `./tools/...` NOT run (untouched; build covers compile), full
  `./...` accepted from the handoff runtime per campaign rules.
- Baseline before changes: the F-M1a 7-package narrow set green.
- Environment anomaly (verified, see below): the machine SIGKILLs any
  binary executed from a temp dir (`/tmp` and `$TMPDIR` both; a copied
  `/bin/echo` dies with -9 while the same bytes run fine from `$HOME`).
  All test runs above therefore export
  `TMPDIR=$HOME/.cache/tmptest GOTMPDIR=$HOME/.cache/gotmp`. `go build`
  and narrow `-run` mutant kills ran before the restriction appeared;
  post-restore green runs all use the redirection. No repo file depends
  on it.

## Files

Modified: `CHANGELOG.md`, `README.md`,
`internal/regress/interactive_test.go`, `pkg/agentic/{plan,system}.go`,
`{claude,codex}/args.go`, `codex/argvguard_test.go` (+3 allowlist
entries with reasons; guard proven to report the new sites first),
`{claude,codex}/interactive_test.go`, `pi/{args,interactive_test}.go`,
`pinative/{args,pinative_test}.go`.
Added: `internal/nativeargs/*`, `internal/toolprobe/*`,
`pkg/agentic/capability_test.go`, `{claude,codex}/{policy,probe}.go`,
`{claude,codex}/{policy,probe}_test.go`, `pi/policy{,_test}.go`,
`pinative/{policy,probe}.go`, `pinative/{policy,probe}_test.go`.

## Findings and decisions (logbook substitute; no LOGBOOK edits per campaign rules)

- Table is per-plugin rows, not a central map: the single-source guard
  forbids `map[SystemID]` outside the registry, and D5 keeps the
  mapping with each plugin. README carries the unified table.
- Native never inspects (0018 item 4), so unknown-form refusal is
  yolo-only; native + unknown forms forwarding is a committed row, not
  an oversight. Known-conflict refusal (item 4 table) is an explicit
  later leaf, not silently included.
- Yolo argv order with NativeArgs is bypass-before-suffix (0018 item 1);
  empty-suffix yolo argv is byte-identical to F-M1a.
- F-M1a prefix-duplicate check stays FIRST in the yolo branch
  (release-independent exact match), preserving its no-ToolRelease
  test unmodified; lookup follows, then the grammar scan.
- `-c` closed set: the two module transports + three 0018 policy keys
  + `mcp_servers.*` (composition-evidenced). Help-example keys
  (`model`, `shell_environment_policy.inherit`) are intentionally
  OUTSIDE it (fail closed; re-verification extends the set).
- Pi rows duplicate the `0.84.2` string across two plugins on purpose:
  the table keys (environment, release) — two environments sharing one
  tool release.
- No `NativeArgs`/`ToolRelease` plumbing in vendorplugin: tracked
  launches never carry them (fidelity check is field-wise; new zero
  fields unaffected).

## Acceptance mapping

1) Table versioned; unpinned/newer refuses yolo with a named
   diagnostic, native forwards verbatim (golden pair per tool) — done.
2) Unknown `-c` key + unknown claude mode refused as usage exit 2 —
   done (`ErrNativePolicyUnknown`, caller maps to exit 2).
3) Parsing-rule rows, prompt never parsed as flag — done, both
   directions per tool.
4) One narrowing mutant per refusal bound, executed and killed — done
   (B1-B7 + B6b/B8 bonus, table above).
5) Release tagged + CHANGELOG names the token — CHANGELOG names
   `permission-grammar-v1` and expects v0.5.17; the tag itself is the
   orchestrator's step per the brief (not cut here).

Ready for review.

revision 2 = revision 1 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).
