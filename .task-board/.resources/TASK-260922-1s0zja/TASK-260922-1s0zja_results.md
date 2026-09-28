# TASK-260922-1s0zja results — F-M1a: permission-mode member + provider mapping (revision 2)

Developer handoff evidence. Worktree: skill-agents-management
`.temp/STORY-260922-h3epwn/worktree`, branch
`task-board/story/STORY-260922-h3epwn`, uncommitted (integration is the
orchestrator's step). No release tag cut (F-M1b cuts it).

Revision 2 answers the rev1 verdict (`TASK-260922-1s0zja_review-verdict-rev1.md`):
one P1 (F1, third-site ownership guard) and one P2 (F2, documented error
contract). No product behaviour change beyond F1's guard and F2's documents;
everything the reviewer judged correct in rev1 stands as committed.

## Pinned tool releases (unchanged from rev1)

Goldens are pinned to the releases the repository already verifies its
interactive grammar against (plugin `args.go` comments):

| env | pinned release | yolo mapping | evidence for the spelling |
|---|---|---|---|
| `claude_code` | claude 2.1.261 (repo pin) | `--dangerously-skip-permissions` | module's own exec spelling (golden-captured from real launches at source `ed48781`); corroborated by installed `claude --help` at 2.1.274 ("Bypass all permission checks", exit 0) and Decision 0018's verification at 2.1.273 |
| `codex_cli` | codex 0.153.2 (repo pin) | `--dangerously-bypass-approvals-and-sandbox` | module's own exec spelling (golden-captured); corroborated by installed `codex --help` at 0.153.4 (exit 0) and Decision 0018's verification at 0.153.4 (both helps list it) |
| `pi` | pi 0.84.2 (repo pin = installed) | none — refuse `ErrPermissionModeUnsupported` | installed `pi --help` at 0.84.2 (exit 0): no bypass flag; only `--approve, -a` ("Trust project-local files for this run") and `--no-approve, -na`, which 0018 records as not equivalent. `agents-infra pi --help` passes the same help through, so the wrapper adds no approval flag either. Applies to both `pi` (wrapper) and `pi-native` |

The (environment, tool release) capability table that re-verifies these per
release is F-M1b (TASK-260922-1wvwc3), not this leaf.

## What was built (rev1, retained)

- `pkg/agentic/system.go`: `PermissionMode` type (`native`/`yolo`, zero value
  means native), `LaunchRequest.PermissionMode` member, `Resolve()` as the
  single reader of the value rule (exact match after trim; near-misses
  refused), `IsZero()`; interactive grammar comment admits exactly this one
  optional member under 0018 (cites 0018 + 0013 D5).
- `pkg/agentic/plan.go`: sentinels `ErrPermissionModeUnknown`,
  `ErrPermissionModeNotInteractive`, `ErrPermissionModeUnsupported`,
  `ErrPermissionModeDuplicate`; `BuildPlan` validates unknown-value (every
  mode) and non-zero-outside-interactive before `PrepareLaunchRequest` runs,
  so a refused value triggers no `Argv` construction. (Accuracy correction
  vs rev1: the static `sys.Capabilities()` declaration is read before
  validation at `plan.go:249` — no operational dispatch follows from it, but
  the rev1 wording "no plugin surface whatsoever" was overbroad. The tests
  measure no-`Argv`, and that is the claim kept here.)
- `claude` / `codex` `args.go`: one const per plugin holding the bypass
  spelling, referenced by both the exec branch (unchanged behavior) and the
  interactive branch (yolo appends it after model+effort — the exec grammar's
  relative order — before any prompt text, which interactive never carries);
  plugin-level repeats of unknown/scope refusals; yolo + composition prefix
  already carrying the flag refused, not de-duplicated.
- `pi` / `pinative` `args.go`: yolo refused with
  `ErrPermissionModeUnsupported` (evidence above); plugin-level
  unknown/scope repeats.
- `internal/argvguard/literal.go`: `LiteralSites` occurrence counter shared
  by the plugin proofs (string literals only; comments are not spellings;
  malformed source is an error).
- `internal/regress/interactive_test.go`: `yoloFlag` per case; native sweep
  unchanged (bypass flags absent); new yolo sweep asserts exactly one
  occurrence of the system's own flag and no other marker, and
  `ErrPermissionModeUnsupported` for pi-native.
- README (interactive member, per-plugin mappings, guard note) + CHANGELOG
  (Unreleased). No launcher, vendorplugin, CLI, or tag changes.

## Rework in revision 2

- F1 (P1): `pkg/agentic/systems/claude/argvguard_test.go` — replaced the
  plugin-scoped count plus the "agy is non-empty" assertion with the
  module-wide closed allowlist
  `TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites`: `LiteralSites` over all
  non-test module sources must report exactly the claude const
  (`pkg/agentic/systems/claude/args.go` / `bypassPermissionsFlag`) and agy's
  construction site (`pkg/agentic/systems/agy/args.go` / `Args`). A third
  spelling anywhere — a new package or a second site inside either plugin —
  fails it, as does a known site going silent. Committed narrowing:
  `TestTheTwoSiteProofBitesOnAThirdSpelling` with the reviewer's
  `pkg/agentic/review_third_spelling.go` reproduction and a second-site-inside-agy
  row (both must count three). No agy argv change, no golden change.
  `--dangerously-bypass-approvals-and-sandbox` needed no equivalent change:
  its codex proof was already module-wide and the module holds no second
  site for it. Final shapes: codex flag = module-wide one-site, claude flag =
  module-wide two-site.
- F2 (P2): `CHANGELOG.md` and `README.md` now name the two entry points
  separately — through `BuildPlan` a yolo composition is refused earlier with
  `ErrCompositionNotInteractive`, and `ErrPermissionModeDuplicate` is the
  direct plugin `Argv` path's refusal — and the README interactive paragraph
  acknowledges the optional typed permission-mode member. The two README
  guard-shape sentences were updated to the F1 proof (module-wide two-site).

## Exact goldens (all `reflect.DeepEqual` on whole argv, via real `BuildPlan`)

| plugin | native (zero and explicit — byte-identical) | yolo |
|---|---|---|
| claude | `--model <m> --effort <e>` (unchanged test) | `--model <m> --effort <e> --dangerously-skip-permissions`, flag count == 1 |
| codex | `-m <m> -c model_reasoning_effort="<e>"` (unchanged test) | `... --dangerously-bypass-approvals-and-sandbox`, flag count == 1, no other marker |
| pi | `pi --model <m>` (unchanged test) | refused `ErrPermissionModeUnsupported` |
| pinative | `--model <v>/<m> --thinking <e>` (unchanged test) | refused `ErrPermissionModeUnsupported` |

Effortless variants (model flag[s] alone + yolo flag) and stdin-detached
assertions included per plugin. All pre-existing tests unmodified and green.

## Narrowing mutants: executed and killed

Rev1 rows 1–4b (rerun green in the suite since; not hand-re-executed — the
reviewer independently re-ran rows 1 and 3a and killed both):

| # | bound (production gate) | narrowing mutant | killing test | result |
|---|---|---|---|---|
| 1 | unknown value — `PermissionMode.Resolve` (`pkg/agentic/system.go`), called by core + all 4 plugins | admit `"auto"` as native (one extra value; gate still refuses everything else) | `pkg/agentic/TestBuildPlanRefusesAnUnknownPermissionModeInEveryMode/near-misses_are_refused,_not_guessed_at` | FAIL `value "auto": err = <nil>`, exit 1 |
| 2 | interactive-only scope — `buildPlan` (`pkg/agentic/plan.go`) | `mode != interactive` narrowed to `mode == exec` (dry-run/managed-session admit yolo) | `pkg/agentic/TestBuildPlanRefusesANonZeroPermissionModeOutsideInteractiveLaunches/{dry-run,managed-session}/*` | FAIL `err = <nil>`, exit 1 |
| 3a | duplicate spelling — claude `interactiveArgs` | prefix scan narrowed to `Prefix[0]` only (duplicate planted at index 1 admitted) | `claude/TestTheYoloRefusalsAreNamedAndTotal/yolo_with_the_bypass_flag_already_in_...` | FAIL `Argv err = <nil>`, exit 1 |
| 3b | duplicate spelling — codex interactive branch | same narrowing as 3a | `codex/TestTheYoloRefusalsAreNamedAndTotal/yolo_with_the_bypass_flag_already_in_...` | FAIL `Argv err = <nil>`, exit 1 |
| 4a | unsupported — pi `interactiveArgs` | yolo condition narrowed with `&& model == ""` (unreachable: empty model returns earlier; yolo admitted) | `pi/TestAnInteractiveYoloLaunchIsRefusedAsUnsupported` | FAIL `BuildPlan err = <nil>`, exit 1 |
| 4b | unsupported — pinative `Args` | yolo condition narrowed with unsatisfiable conjunct (yolo admitted) | `pinative/TestAnInteractiveYoloLaunchIsRefusedAsUnsupported` | FAIL `BuildPlan err = <nil>`, exit 1 |

New in revision 2 (executed by me against the working tree in this run):

| # | bound (production gate) | narrowing mutant | killing test | result |
|---|---|---|---|---|
| 5 | claude bypass ownership — `TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites` (`pkg/agentic/systems/claude/argvguard_test.go`) over `argvguard.LiteralSites` | reviewer's verbatim rev1 reproduction planted as a real file: `pkg/agentic/review_third_spelling.go` spelling `--dangerously-skip-permissions` in a third package | `claude/TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites` | FAIL `unexpected bypass-flag spelling at pkg/agentic/review_third_spelling.go:2 reviewThirdSpelling`, exit 1; file removed afterwards, gate re-run green (exit 0). `git status` confirms no leftover |

Committed in-suite narrowings (not hand-executed): each argvguard proof has
planted-site tests (claude: third-package + second-site-inside-agy rows);
duplicate tests plant the flag at a non-zero prefix index; flagless-prefix
narrowing proves the duplicate refusal is the flag's.

## argvguard ownership proof (revised)

- codex: `--dangerously-bypass-approvals-and-sandbox` occurs exactly once in
  all non-test module sources (the const); the AST signature carries the flag,
  so the gate + narrowing + mutant battery prove no second construction, and
  the occurrence count proves no second spelling. Shape: module-wide one-site.
- claude: `--dangerously-skip-permissions` occurs at exactly two known sites
  in all non-test module sources — the claude const and agy's `Args` — held
  by a closed module-wide allowlist. The AST signature still excludes the
  flag (agy's golden-pinned exec spelling is the reason, unchanged); the
  occurrence proof is what keeps that sharing to exactly these two sites.
  Shape: module-wide two-site.
- pi/pinative spell no bypass flag; nothing to guard.

## Decisions and boundaries (unchanged from rev1)

- Explicit `native` outside interactive is refused (`ErrPermissionModeNotInteractive`),
  per the brief's "any non-zero value": outside this mode no mapping exists
  for the member to select, so silent acceptance would be a lie.
- Duplicate check reads `Composition.Prefix`, the only caller-supplied argv
  fragment on `LaunchRequest`. Through `BuildPlan` the composition gate
  refuses first (fail closed either way; pinned by test); the duplicate
  sentinel fires for direct plugin holders. Exact-element match only:
  `=`-forms and the wider conflict table are F-M1b's parsing rule + F-L1.
- End-to-end "yolo + raw bypass after `--`" at the launcher is F-L1's item-4
  table; this leaf pins the agents-management half (the plugin never emits a
  duplicate).
- No `NativeArgs` member added: splicing caller argv without F-M1b's
  flag-vs-prompt parsing rule would be half a feature.

## Validation run by me (bash, `set -o pipefail`)

- `go build -mod=mod ./...` — exit 0.
- `go vet -mod=mod` over the 7 narrow packages — exit 0.
- `go test -mod=mod ./pkg/agentic/ ./pkg/agentic/systems/claude/
  ./pkg/agentic/systems/codex/ ./pkg/agentic/systems/pi/
  ./pkg/agentic/systems/pinative/ ./internal/argvguard/
  ./internal/regress/ -count=1` — all 7 packages `ok`, exit 0.
- `go test -mod=mod ./pkg/agentic/systems/agy/ -count=1` — `ok` (the gate
  now names agy's site; agy itself untouched).
- New gate + committed mutants: `go test ./pkg/agentic/systems/claude/ -run
  'TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites|TestTheTwoSiteProofBitesOnAThirdSpelling'
  -count=1` — pass, exit 0.
- Real-file third-site mutant (row 5 above): planted → FAIL exit 1 with the
  naming diagnostic; removed → gate green exit 0.
- `gofmt -l pkg/ internal/ tools/` — empty (clean).
- Full `go test ./...` accepted from the handoff runtime (board validation
  command runs once at handoff; not run here per campaign rules).

## Files changed

`pkg/agentic/{system,plan,interactive_test}.go`,
`pkg/agentic/systems/{claude,codex}/{args,argvguard_test,interactive_test}.go`,
`pkg/agentic/systems/pi/{args,interactive_test}.go`,
`pkg/agentic/systems/pinative/{args,pinative_test}.go`,
`internal/argvguard/{literal.go,argvguard_test.go}`,
`internal/regress/interactive_test.go`, `README.md`, `CHANGELOG.md`.
Revision 2 touches only `pkg/agentic/systems/claude/argvguard_test.go`,
`README.md`, and `CHANGELOG.md` relative to revision 1.

Ready for review.
