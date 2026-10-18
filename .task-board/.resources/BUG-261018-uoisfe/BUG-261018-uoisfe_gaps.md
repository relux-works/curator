# Native skills scope and gaps

## Current handoff — round 7

Current signed candidate: `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`, tree `769581956b4fd7bfd02c2567d7465b60335f5b4b`. The Rework round 7 section and BUG-261018-uoisfe_validation-r7.md supersede earlier implementation/validation statements. Earlier sections are retained as historical evidence; the old late-handler owner statement no longer describes the current shared admission guard.

## Rework round 3 — unrelated-takeover-argument-discard

This section supersedes the round-2 parser-scope exclusion. The deciding
review held rows 1–11 and rejected row 12; this revision changes only the CLI
consent boundary and its tests. The previous surface implementation, Muse,
opencode, managed-home layout, specification files and root README are unchanged.

`cmd/curator/main.go:283` calls `rejectsTakeover` before help, argument-discarding
wrappers, configuration loading and provider dispatch. The closed command
allowlist lives in `cmd/curator/command_help.go:128`. It admits consent only for
profile install with --use, profile use, profile sync, profile update and
env resolve with --repair. The env parser retains its existing --repair requirement. Operation parsers retain their remaining
restrictions, including repair/dry-run. Go boolean spellings are preserved;
arguments following -- remain literal operands.

`TestNativeSkillsTakeoverEveryRegisteredCommand` drives production run() for
all 67 unrelated registered commands, rejecting consent with exit 2 both with
and without help, with successful no-consent help controls for every command.
The help controls are intentionally stated as help, not proof of successful
live operations: group topics have no executable action and UI/login commands
require interactive input. `TestNativeSkillsTakeoverArgumentDiscardRegression`
also drives successful actual list, update, gc, version and --version calls,
then rejects consent for those same calls without state changes.
`TestNativeSkillsTakeoverActivationBoundary` covers install/resolve without
activation, boolean spellings and the argument separator. The existing
`TestNativeSkillsCLIOperations` exercises all five real activating operations.

The original names of the activation/reinstall tests are retained. Three
no-use consent assertions in import_cli_test.go/profile_git_reinstall_test.go
and the takeover-only cases of profile_install_matrix_test.go now require
exit 2 and preserved state, matching the tightened row-12 requirement.

Coverage: **12 of 12 attack-surface rows driven; 5 of 5 AC rows driven**.
Rows 1–11 keep the round-2 coverage map and narrowing controls; row 12 is
replaced by the closed registry sweep above. There is no excluded registered
command and no legacy-wrapper parser exemption in this revision.

| Mutant | Narrowed gate | Named failing tests | Survival bound |
|---|---|---|---|
| list-only-consent-bypass | Retains every rejection except list, which silently discards --takeover | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Killed: child exit 1 with both named behavioral failures, observed list exit 0 instead of 2; no survivor |

Out-of-contract rows: none. Explicit round-2 scope exclusions remain Muse,
opencode and managed-home layout. Open specification questions remain:
1. Muse needs a native multi-root ownership/marker/backup contract.
2. Adapter-ledger ownership/coexistence with an active profile needs specification.
3. Opencode keeps its machine-global native surface; profile ownership needs clarification.

Exact validation identities and results follow below.

The preliminary duplicate resolve guard was removed after the existing
claude-takeover-without-repair control survived (parent exit 1, child exit 0).
The existing env parser remains the single owner of that condition; its
unchanged control now produces its named behavioral failure again (parent
exit 0, mutant child exit 1). No native-surface row was reimplemented.

Historical control diagnostics (not counted as successful kills):

| Mutant | Narrowed gate | Named failing test | Survival bound |
|---|---|---|---|
| claude-takeover-without-repair on preliminary duplicate-guard snapshot | Allows one adapter through the existing resolve parser | None; child exited 0, parent exited 1 | Survived because the new outer duplicate still rejected consent; this establishes no kill for the inner condition. The duplicate was removed, and the unchanged control now fails TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code as intended. |
| list-only-consent-bypass, preliminary anchor | Intended list-only bypass; plant did not apply | None; parent failed the anchor check | Not-applied survivor; no behavioral property established. The unique switch anchor was corrected, and the current child exits 1 with both named list failures. |

Rework scope versus the accepted round-2 snapshot: six files, all under
`cmd/curator/`. No changes outside that module were introduced in this rework.
The existing native implementation and docs remain in the full candidate.
`TestNativeSkillsTakeoverEveryRegisteredCommand` is at
`cmd/curator/native_skills_flags_test.go:17`; the direct regression is at line 54,
the activation boundary at line 76, and its narrowing control at line 102.
All previously committed reviewer/native tests retain their original names.

The Story worktree remains uncommitted at checkpoint
`5d8e6b07f7f485557d9bbb065c9692348ca065c2`. Validation uses detached signed
snapshot `3e00fcc8abe75e411a88773c65c4042ae2448eb0` (signature G), tree
`2fe068da665b203ed48e52cfcb7fba74de96e37e`, equal to the current worktree.
The fresh protected-trunk comparison reports no path overlap.

Retained native controls, refreshed in the current full replay:

| Mutant | Narrowed behavior | Named failing test | Survival bound |
|---|---|---|---|
| pi-native-only | Only pi writes old staging instead of native skills | `TestNativeSkillsUseAndSync` | Killed; child exit 1, parent 0; no survivor |
| codex-link-drift | Ignores a mismatched recorded link only for Codex | `TestNativeSkillsRepair` | Killed; child exit 1, parent 0; no survivor |
| dangling-foreign-consent | Allows only the retired relative-link class without consent | `TestNativeSkillsTakeover/foreign` | Killed; child exit 1, parent 0; no survivor |
| plain-file-consent | Allows only regular files without consent | `TestNativeSkillsTakeover/file` | Killed; child exit 1, parent 0; no survivor |
| skills-parent-traversal | Allows only a symlinked skills component | `TestNativeSkillsLinkedParent` | Killed; child exit 1, parent 0; no survivor |
| unnamed-takeover-directory | Treats only unnamed directory as recorded during takeover | `TestNativeSkillsConvergenceAndSwitch` | Killed; child exit 1, parent 0; no survivor |
| one-ledger-name | Bypasses ledger ownership only for skills/myskill | `TestNativeSkillsLedger` | Killed; child exit 1, parent 0; no survivor |
| one-old-staging-name | Retains only old staging myskill | `TestNativeSkillsConvergenceAndSwitch` | Killed; child exit 1, parent 0; no survivor |
| one-retired-name | Retains only retired native myskill | `TestNativeSkillsConvergenceAndSwitch` | Killed; child exit 1, parent 0; no survivor |
| backup-directory-collision | Ignores only an existing directory generation | `TestNativeSkillsBackupCollision` | Killed; child exit 1, parent 0; no survivor |
| directory-marker-is-absence | Treats only a directory marker read failure as absence | `TestNativeSkillsUnreadableMarker` | Killed; child exit 1, parent 0; no survivor |
| claude-takeover-without-repair | Allows only Claude consent without repair | `TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code` | Killed; child exit 1, parent 0; no survivor |

Local validation is green: all 85 served packages covered; 28 CLI/envprofile
groups cover all 431 CLI and 381 envprofile top-level names. Twenty-one groups
were executed on the current identity; seven unchanged envprofile groups reuse
exact green evidence from the preceding signed snapshot, as recorded in
`r3-reuse-identity.json`. Eighty-three unrelated packages reuse the round-2
stream bound to unchanged source/tests/configuration/toolchain/environment.
The full merged stream has 3,494 unique top-level names and 36 classified skips.
`r3-merge-local.py` and the production platform-case gate both exit 0.
The initial missing-group control exited 1 as required, so missing group
evidence cannot masquerade as full coverage.

Lint (private task cache), go vet, go build, diff check and gofmt are green
(exit 0, no gofmt output). The earlier shared-cache lint exit 1 referenced
unrelated removed worktrees and is preserved as a failed diagnostic attempt.
The initial activation-matrix exit 1 exposed old no-use consent expectations;
the unchanged test names now assert usage refusal and state preservation,
and the full refreshed package is green.

Hosted validation is green on attempt 2, with 20 successful jobs and two
workflow-conditional skips. `gh run watch 39117060646 --exit-status` exits 0.
Attempt 1 remains recorded as exit 1: eighteen successful jobs and two failed
Complete runner cleanup steps after successful repository steps. The legitimate
`--failed` retry exits 0 and repeats only those two jobs on the same signed head.
Both pass including cleanup; no status was forged and no local mirror replaced
required checks. In-progress retry requests were refused (exit 1); the later
accepted retry waited for busy shared Linux runners.

The downloaded Windows test artifact is bound to this exact candidate and
contains all 67 registry subcommand passes, the four new tests, all twelve
retained narrowing controls and the five-operation behavioral suite. The
artifact digest and raw stream hash are recorded in r3-hosted-windows-summary.json.

Fresh integrity after hosted completion verifies worktree/tree equality, G
signature, exact server head, preserved Story checkpoint, unchanged root README
and no overlap with fresh protected trunk. Sanitized complete logs and their
raw/sanitized SHA-256 manifest are attached as round-3 validation outcomes.

Home safety: each new behavioral case calls `profileHome` (profile_test.go:18)
to redirect manager and native adapter roots to temporary directories and
`pinOperatorHome` (import_cli_test.go:21) to redirect HOME/USERPROFILE.
The mutation child executes those same production-entry fixtures. No native
operator-home path is used as a write target.

## Historical round-2 record (superseded where round 3 says so)


## Round 2 — native skills for the three single-home adapters

The round-2 brief selects option A from the preceding Stop-The-Line packet.
The earlier blocked assessment is historical. Muse is explicitly excluded from
this leaf; no native Muse multi-root contract is invented. Opencode and managed
home layouts retain their existing behavior. No specification text or root
README is changed.

Candidate: `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b`, signed and verified `G`.
Story checkpoint: `5d8e6b07f7f485557d9bbb065c9692348ca065c2`.
The Story branch remains at that checkpoint, with the candidate uncommitted;
the signed candidate is a separate gate snapshot. Its hosted gate is
https://github.com/relux-works/curator/actions/runs/39094506341.
Validation status is recorded in the accompanying round-2 evidence report;
no queued or running job is treated as green.

Measured coverage: **12 of 12 surface rows driven**, covering **5 of 5 AC rows**.
The following map names the production calls and the narrowing controls. No
surface-table row is omitted. Row 12 has the explicit parser-scope bound below. Each gate control retains
the gate for neighboring inputs and admits one named exception. The suite
requires exit 1 with the named behavioral failure; build/setup failures do not
count as kills. The map uses the shared coverage-map template's per-row
structure; the brief's concrete destination/error/ownership surfaces refine
its destination-conflicts, read-boundary, error-surface, and selector-path types.

| Surface row | Production entry and call site | Attacking tests | Killed narrowing mutant | Out-of-contract inputs |
|---|---|---|---|---|
| 1. Absent target | `UseWithPolicy` / `SyncWithPolicy` → `materializeOne`, switch.go:690; CLI `run` → install/use/sync/update/resolve | `TestNativeSkillsUseAndSync`, `TestNativeSkillsInstallAndUpdate`, `TestNativeSkillsCLIOperations`; existing `TestGlobalAddPublishesProfileLockBeforeNativeMaterialization` and `TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization` | `pi-native-only` → `TestNativeSkillsUseAndSync` | None claimed |
| 2. Recorded manager link | `Resolve` → `nativeSkillsState`, native_skills.go:65; `repairUnderLock`, managed.go:2648 | `TestNativeSkillsRepair` (status, restore, dry-run and non-current bounds), `TestNativeSkillsUseAndSync` | `codex-link-drift` → `TestNativeSkillsRepair` | None claimed |
| 3. Foreign symlink | `UseWithPolicy` → `materializeOne` inventory, switch.go:690; five CLI operations through `run` | `TestNativeSkillsTakeover/foreign`, `TestNativeSkillsCLIOperations` | `dangling-foreign-consent` → `TestNativeSkillsTakeover/foreign` | None claimed |
| 4. Other unrecorded entry | `UseWithPolicy` → `materializeOne` inventory, switch.go:690 | `TestNativeSkillsTakeover/directory`, `/file`, `/store`, `/relative-store` | `plain-file-consent` → `TestNativeSkillsTakeover/file` | None claimed |
| 5. Path components | `UseWithPolicy` → `managedPath` via `inPlaceSkillsLedger` and `materializeOne` | `TestNativeSkillsLinkedParent/ordinary`, `/takeover` | `skills-parent-traversal` → `TestNativeSkillsLinkedParent` | None claimed |
| 6. Unnamed entries | five CLI operations through `run` → native materializer; `SyncWithPolicy` and `UseWithPolicy` | `TestNativeSkillsCLIOperations`, `TestNativeSkillsConvergenceAndSwitch`, `TestNativeSkillsEmptyProfileLeavesUnnamedParent` | `unnamed-takeover-directory` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 7. Adapter ledger | `UseWithPolicy` → `inPlaceSkillsLedger`, native_skills.go:31; `StatusOf` → `scopeHomes`, status.go:1023 | `TestNativeSkillsLedger`, `TestNativeSkillsUnrecordedLedger` | `one-ledger-name` → `TestNativeSkillsLedger` | None claimed |
| 8. Recorded old staging | `SyncWithPolicy` → `materializeOne` recorded-surface removal | `TestNativeSkillsConvergenceAndSwitch` | `one-old-staging-name` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 9. Profile switch | `UseWithPolicy` → `materializeOne` removal/replacement; scoped selection | `TestNativeSkillsConvergenceAndSwitch`, `TestNativeSkillsInstallAndUpdate` | `one-retired-name` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 10. Backup generation | `UseWithPolicy` → `openBackupWithDirectoryRead`, switch.go:1013 | `TestNativeSkillsBackupCollision` (collision and next-generation positive control) | `backup-directory-collision` → `TestNativeSkillsBackupCollision` | None claimed |
| 11. Unreadable marker | `UseWithPolicy` → native `envmarker.Read` before inventory and writes | `TestNativeSkillsUnreadableMarker` | `directory-marker-is-absence` → `TestNativeSkillsUnreadableMarker` | None claimed |
| 12. Flag acceptance | CLI `run` → profile handlers and `cmdEnvResolve` | `TestNativeSkillsCLIOperations` (all five); `TestNativeSkillsTakeoverFlagBoundary` (negative built-in rows and resolve restrictions) | `claude-takeover-without-repair` → `TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code` | Legacy generic argument parsing: explicit bound below |

### Mutant evidence

| Mutant | What it narrows the gate or behavior to | Named failing test | Survival bound |
|---|---|---|---|
| pi-native-only | pi alone falls back to staging, omitting its native write while the other adapters stay native | TestNativeSkillsUseAndSync | Killed; no survivor |
| codex-link-drift | wrong-target links are ignored only for Codex native skills | TestNativeSkillsRepair | Killed; no survivor |
| dangling-foreign-consent | dangling relative links below the retired-manager spelling bypass consent | TestNativeSkillsTakeover/foreign | Killed; no survivor |
| plain-file-consent | unrecorded regular files alone bypass consent; other kinds retain the gate | TestNativeSkillsTakeover/file | Killed; no survivor |
| skills-parent-traversal | managedPath traverses a linked component named skills; other components retain refusal | TestNativeSkillsLinkedParent | Killed; no survivor |
| unnamed-takeover-directory | takeover removes one unrequested directory named unnamed | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| one-ledger-name | only skills/myskill loses ledger exclusion | TestNativeSkillsLedger | Killed; no survivor |
| one-old-staging-name | cleanup retains one old recorded staging name | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| one-retired-name | cleanup retains one obsolete native skill name | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| backup-directory-collision | existing directories alone bypass the next-generation collision check | TestNativeSkillsBackupCollision | Killed; no survivor |
| directory-marker-is-absence | a directory at the marker path alone is treated as absent | TestNativeSkillsUnreadableMarker | Killed; no survivor |
| claude-takeover-without-repair | bare resolve accepts takeover only for claude_code | TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code | Killed; no survivor |

All mutation plants keep relevant source tokens and execute behavioral Go
suites. No new runtime gate attests behavior by scanning source text. Mutation
anchor checks only validate control construction; a token hit cannot produce a
passing kill verdict.

### Required items and implementation

- Registry-native discovery and marker ownership: `native_skills.go:19`,
  `switch.go:690`; tested by use/sync/install/update and five-operation CLI tests.
- Consent, exact symlink backup, no-follow and preflight: `switch.go:505`,
  `switch.go:690`, `switch.go:1013`, existing `nofollow.go`; takeover, linked-parent,
  unreadable-marker and backup-collision tests prove the native-home refusals.
- Ledger ownership and non-current reporting: `native_skills.go:31`,
  `native_skills.go:65`, `status.go:1023`; ledger and unreadable-ledger tests.
- Staging convergence and changing skill sets: `switch.go:690` recorded removal;
  convergence and scoped-switch tests. Removal failures now propagate.
- Current-only repair: `managed.go:2648`, `native_skills.go:187`; repair restores
  native links, dry-run stays read-only, and non-current explicit profiles do not
  alter native state. Managed-home planning and layout remain unchanged.
- Unchanged-lock update: `envprofile.go:1317`, `native_skills.go:209`; the CLI
  update subtest drives refusal and consented retry against an unchanged lock.
- Short documentation: `docs/native-profile-skills.md`.

All added tests pin HOME and USERPROFILE plus native adapter homes to temporary
directories. Mutants inherit those fixtures. No real operator home is an input.
All previously committed reviewer tests retain their original names and assertions.
Four global-lock-publication tests now target native `skills/` instead of retired
staging; their shared helper additionally requires the exact native marker path.
No reviewer test is removed.

### Explicit limits and out-of-scope rows

- Muse native multi-root behavior: excluded by round-2 Scope; unchanged.
- Opencode native and managed homes: excluded by round-2 Scope; unchanged layouts.
- Specification text and root README: excluded by round-2 Scope; untouched.
- Native-home refusal assertions cover the refusing entry/home. The existing
  M11 per-adapter partial-switch model remains; this leaf does not claim a new
  cross-adapter transaction.
- Backup collision is a deterministic stale-directory-read fault injection at
  the real UseWithPolicy entry point, not a timing-dependent live-race assay.
- The flag boundary drives supported profile/env built-in rows plus status;
  external provider argument parsers and legacy top-level/global handlers are outside
  this leaf's `internal/envprofile` switch/status/repair scope. In particular, the
  pre-existing top-level `list` and `update` wrappers discard trailing arguments
  (`cmd/curator/command_help.go:27,30`), and `global update` does likewise
  (`main.go:1614`). This leaf does not claim universal rejection of `--takeover`
  by every unrelated command. It proves consent routing on the five specified
  profile operations and rejection on the enumerated profile/env/status rows.
  No legacy-global operation gains takeover authority.

The rework diff stays in `internal/envprofile`, except
`cmd/curator/native_skills_test.go` required to exercise actual CLI flag routing,
`cmd/curator/global_lock_publication_test.go` whose four existing publication and
refusal tests must exercise the newly specified native destination, and the
explicitly requested short note under `docs/`. No registry facts needed
changing. No unrelated production source is rewritten.

### Open specification questions

1. **Muse multi-root ownership:** define independent native XDG data/config root
   ownership, marker and backup placement, scoped currency, drift and repair
   before enabling in-place Muse switching. This is a separate follow-up.
2. **Adapter ledger coexistence:** define resolution when a current profile names
   a legacy-global ledger-owned skill. This leaf leaves the entry and ledger in
   place, never adopts them into the environment marker, and reports non-current.
3. **Opencode:** clarify machine-global discovery versus profile scopes under the
   existing split-brain note; retain today's behavior until that contract exists.

The CLI-only flag mutant uses the existing CLI helper-process lane, preserving
actual run() routing and assertions while avoiding a whole-package GOROOT lock
that these flag cases do not inspect. The native suite and all twelve controls
were replayed successfully while the full CLI group was running.

### Round-2 local validation verdict

The full local suite covers all 85 served packages (3490 unique top-level test names), with no package failure or missing inventory row. Platform-case gate exit 0 classifies all 36 skips. Native suite exit 0 kills all twelve controls with named behavioral failures (child exit 1). Gate self-test exits 0 (267 pass, 2 explicit skips), and lint, vet, build, gofmt inspection and diff-check each exit 0. Exact commands, raw/selected streams, identity reuse and failed setup attempts are in the validation evidence. Full hosted run 39094506341 now reports success on the same signed candidate: 20 successful jobs, with only the branch/dispatch-conditional rose-air and candidate-suite jobs skipped. gh run watch --exit-status exits 0.

### Windows portability correction

The prior third snapshot `eb21808188d767af6e2f912b9dc1208d3c7418a6` had 19 successful hosted jobs and one failed Windows Test job (gate exit 1); its full hosted verdict was failure. Relative symlink creation on Windows returns native separators. The new tests now capture the original Readlink text before the operation and require the refusal, backup or untouched entry to preserve exactly that captured text. They never normalize the observed result. Only the three native test files changed; production and previously committed reviewer tests are unchanged.

| Historical mutant | What it narrowed on Windows | Intended named failure | Survival bound |
|---|---|---|---|
| dangling-foreign-consent on eb218081 | Nothing: slash-only matching missed Windows link text | None; a backup-literal assertion failed for an unrelated reason | Survivor/invalid control on Windows. Old kill evidence is bounded to POSIX spelling. The fourth snapshot normalizes only the matcher, preserving exact-link assertions. Its fresh local replay is killed by TestNativeSkillsTakeover/foreign; the fresh hosted Windows artifact now proves all 12 controls killed, with named child exit-1 failures. |

Fresh local native-candidate-05 and native-cli-05 each exit 0. All twelve current controls are killed by their named behavioral tests; lint-07, vet-03, build-05, diff-check-06, gofmt-05, merge-local-03 and platform-gate-03 exit 0. All 28 refreshed envprofile/CLI groups exit 0 and replace their package events in the merged stream; the other 83 package identities are unchanged and explicitly reused. The sanitized evidence stream also passes the platform-case gate (exit 0). The fourth signed candidate and current full hosted run are named above. The current Windows Test job is green: downloaded evidence records 48 passing native test/subtest events and all 12 controls with their named child exit-1 failures. The complete workflow now reports success: 20 successful jobs and 2 explicitly conditional skips.

### Round-2 hosted verdict

Full hosted gate **success** on exact signed candidate b32823ac920d8b0485c4a7a7111b48fff8e3ca2b: [run 39094506341](https://github.com/relux-works/curator/actions/runs/39094506341). All 20 applicable jobs passed, including Test on Linux/macOS/Windows, both Race lanes and all three gate self-tests. The rose-air main-only lane and optional candidate-spec workflow-dispatch lane were not executed and are explicitly conditional skips. No fallback or synthetic status was used. The standalone gh run watch --exit-status command exited 0. The source-tree/signature/remote-head check exits 0 and preserves the Story checkpoint; all task changes remain uncommitted.

Checklist interpretation: item 1 follows the winning Round-2 scope (claude_code, codex_cli, pi); Muse is explicitly excluded and unchanged. Item 14 has no new runtime source-scanning gate to qualify; all mutation plants preserve relevant tokens and execute the behavioral suites.

### Acceptance-criteria mapping

Measured AC coverage: **5 of 5 AC rows driven** (row 5 has the explicit row-12 legacy-parser bound in the surface map).

| AC | Production call sites | Named tests / evidence |
|---|---|---|
| 1. Five operations, three adapters | CLI run → profile install/use/sync/update and cmdEnvResolve; materializeOne (switch.go:690), syncCurrentNativeSkills (native_skills.go:209), repairUnderLock (managed.go:2648) | TestNativeSkillsCLIOperations, TestNativeSkillsUseAndSync, TestNativeSkillsInstallAndUpdate, TestNativeSkillsRepair |
| 2. Closed takeover and no-follow | UseWithPolicy → materializeOne inventory, managedPath, openBackupWithDirectoryRead | TestNativeSkillsTakeover, TestNativeSkillsLinkedParent, TestNativeSkillsBackupCollision, TestNativeSkillsUnreadableMarker, TestNativeSkillsCLIOperations |
| 3. Ledger entries retained/reported | UseWithPolicy → inPlaceSkillsLedger; StatusOf → scopeHomes/nativeSkillsState | TestNativeSkillsLedger, TestNativeSkillsUnrecordedLedger |
| 4. Recorded staging converges | SyncWithPolicy → materializeOne recorded removal | TestNativeSkillsConvergenceAndSwitch |
| 5. Twelve attack rows, temp homes | Real UseWithPolicy/SyncWithPolicy/Install/Update/Resolve/StatusOf and CLI run entry points in the twelve-row map | TestNativeSkillsNarrowingMutants and all named native/CLI tests; temporary HOME/USERPROFILE/native-home fixtures; 12/12 controls killed locally and on hosted Windows |


---

# BUG-261018-uoisfe \u2014 native skills surface: blocked before implementation

## Constraint and exact decision

The brief requires Muse in-place skills at `<native home>/<SkillsDir>/<name>` using the registry's `data/muse/skills`, recorded as environment marker surface paths, while managed homes and specification text remain unchanged. The current native Muse home is the native config directory, not the synthetic four-parent root used by managed homes. Independent native XDG config/data roots therefore cannot be represented by the requested one-home path model.

An ownership/layout decision is required before implementing this leaf: either (A) explicitly defer Muse from this leaf, keeping its existing behavior and deliver Claude Code/Codex/Pi now; or (B) authorize a separate native Muse multi-root ownership contract identifying the native data destination, marker placement, backup placement, independent XDG overrides, and which current scope owns each root. Recommendation: A for this narrowly scoped bug, with B as an explicit follow-up. If all four adapters must ship together, B must be settled first. This is a requested product/architecture decision, not a request for implementation permission.

## Evidence and failed assumption

- `internal/envprofile/switch.go:65`, `:584`, `:710`: existing skills are staged under `.agent-context/skills`; the bug exists.
- `internal/envprofile/switch.go:82`: the in-place switch registry contains Claude Code, Codex, Opencode and Pi; Muse does not participate. Merely changing a skill directory does not add Muse to profile use/sync/install/update.
- `internal/envregistry/envregistry.go:233` and `:237`: Muse has four independently redirected parents, with managed-relative skill path `data/muse/skills`.
- `internal/envprofile/muse.go:13` and `internal/envprofile/managed.go:164`: native Muse resolves to `XDG_CONFIG_HOME/muse` (or the default native config directory).
- `internal/envprofile/managed.go:70`: ManagedHomeDir for Muse is the synthetic manager-owned root. It is a different kind of root from nativeHome.
- `internal/envprofile/muse_test.go:256`: existing TestMuseDirectoriesAndIsolation checks the synthetic managed layout.
- `internal/envmarker/envmarker.go:328`: surface paths must pass PortablePath; paths escaping to an independent data directory cannot be recorded.

Counterexample using temporary homes only: config root `config-root`, independent data root `independent-data-root`. The literal required concatenation produces `config-root/muse/data/muse/skills/sample`. A native XDG data surface would instead be `independent-data-root/muse/skills/sample`. Their relative route escapes the config home and is rejected by the marker path validator. Treating the synthetic managed root as the native home, assuming a common XDG parent, or allowing `..`/absolute paths would disguise the missing ownership model. None was implemented. No experiment changed permissions, symlinks or files in the real operator home.

## Checks actually run

Source checkpoint: ce5cd3ecc45e3798e967cfaa02fc71e9d599c47e. Fresh fetched origin/main: 5d8e6b07f7f485557d9bbb065c9692348ca065c2. The Story checkout is 15 commits behind local main; it was not switched/rebased/committed. `git diff --quiet HEAD origin/main -- internal/envprofile internal/envregistry internal/envmarker` exited 0, establishing unchanged diagnosed source despite the older checkpoint.

`go test -count=1 -timeout=2m -v ./internal/envprofile -run '^Test(NATSkillMuseNativeLayoutConstraint|MuseDirectoriesAndIsolation)$'` exited **0** on Go 1.25.5, darwin/arm64. Both named tests passed. The first was a temporary diagnostic placed in the package for this command, then removed; its source and raw output are attached separately. It proves the conflicting current path model, not the requested feature. The second is an unchanged existing managed-home test. No previously committed test was changed or removed.

Full local suite, lint, build, property suites, mutation controls and hosted gate: **not run**. There is no implementation candidate or signed candidate to qualify. No suite is reported as green by proxy. No background process remains. No handoff was attempted because its preconditions are unsatisfied.

## Required-item coverage and gaps

Measured desired AC coverage: **0 of 5 AC rows driven**. Diagnostic checks are not acceptance evidence.

| AC | Relevant production site | Implementation/test status |
| --- | --- | --- |
| 1: native links on current scopes and operations | switch.go:584, switch.go:710; managed.go:2451 Resolve path | Blocked on Muse ownership/layout; no behavior changed. |
| 2: closed takeover, identical-link-text backup, unrequested entries and no-follow | switch.go:735 inventory; switch.go:972 openBackup | Existing root/staging machinery inspected; native-skill behavioral tests and controls unrun. |
| 3: adapter ledger untouched and non-current reporting | import.go:391 readSkillsLedger; status.go:335 StatusOf | No changes; required reporting not implemented or tested. |
| 4: recorded staging converges away | switch.go:826 recorded-path removal | No changes; convergence not implemented or tested. |
| 5: negative tests and mutation, real HOME untouched | temporary TestNATSkillMuseNativeLayoutConstraint only | No AC gate evidence; diagnostic uses temporary HOME/USERPROFILE/XDG only. |

The brief contains no attack-surface table: this is a brief gap, not a waived coverage-map requirement. No rows were silently classified as out of contract. Explicit exclusions from the brief: specification text, managed-home behavior and Opencode behavior. There is no rework diff, and the repository working tree is unchanged after removing the probe.

| Mutant | Narrowed gate | Named failing test | Survival bound |
| --- | --- | --- | --- |
| None executed | Not established | None | All mutation claims remain unverified; no handoff admissible. |

## Open specification questions

1. Muse: define a native multi-root ownership model before applying synthetic managed-relative registry paths to native XDG homes; define marker/backup locations, drift and repair for independent roots, scoped currents, and symlink component boundaries.
2. Adapter ledger: where a current profile names a legacy-global ledger-owned skill, clarify coexistence and resolution while preserving the ledger, link, and non-current reporting required by this brief.
3. Opencode: retain existing machine-global skills behavior; clarify how in-place profile skills and global skills coexist under the existing split-brain note.

## Logbook \u2014 2026-10-18

The native skills defect is confirmed in source, but the four-adapter brief assumes a native root model Muse does not have. A bounded production-nativeHome/registry/marker-validation diagnostic demonstrated that the literal mapping cannot reach an independently configured native XDG data root. Stop-The-Line applied before product edits. Viable alternatives, recommendation and the exact scope/ownership decision are above. Outcome is blocked, not review-ready.


## Rework round 3: successor handoff recovery


The predecessor implemented round 3 and attached the full validation evidence, but Change Request revision 2 validation exited 1 before hosted execution: the required signing environment was absent and SSH signing requested a passphrase. That failed validation is not a green gate.

Recovery introduced no repository changes. Sourcing the host's prescribed agent-git-env restored the configured agent socket and signing key. A detached exact-tree signing probe verified G; the Story branch was neither committed nor moved.

| Check rerun by successor | Exit | Evidence |
|---|---:|---|
| Four takeover regression tests via production run(), including all 67 unrelated registered command cases | 0 | resume/regression.log |
| list-only-consent-bypass narrowing mutant | 0 parent, 1 child | TestNativeSkillsTakeoverEveryRegisteredCommand/list and TestNativeSkillsTakeoverArgumentDiscardRegression/list fail as intended |
| go build -o task-local artifact ./cmd/curator | 0 | resume/build.log |
| git diff --check | 0 | No whitespace diagnostics |
| Fresh origin/main fetch and changed-path intersection | 0 | resume/freshness.json; no overlapping paths, root README unchanged |
| Detached signing probe | 0 | Signature G; exact candidate tree |
| Provider read of prior hosted gate | 0 | Run 39117060646, attempt 2, completed/success on 3e00fcc8abe75e411a88773c65c4042ae2448eb0 |

Full local coverage is reused from the already attached round-3 evidence on the identical source, test, configuration and environment identity: 85 packages, 3494 top-level names, explicit unchanged-evidence bounds in validation-r3 and its full logs. It was not replayed wholesale by this successor. Accepted rows 1-11 remain unchanged. Coverage remains 12 of 12 surface rows and 5 of 5 AC rows, with help-control limitations explicitly preserved in the coverage map.

| Mutant | Narrowed gate | Named failing tests | Survivor bound |
|---|---|---|---|
| list-only-consent-bypass | Only list silently accepts takeover; all other command rejections remain | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Killed by both named failures; no survivor |

The successor invoked handoff from a shell with the prescribed signing environment. Handoff returned exit 0, status to-review, checklist 23/23. No fresh hosted workflow was observed during that command. The runtime publishes the Change Request after the producer turn ends, so its own validation result is not yet established by this note. The runtime must also inherit the prescribed agent signing/transport environment; shell-local sourcing proves signing readiness only in this successor shell. The earlier signing failure is retained. All changes remain uncommitted at the recorded Story checkpoint. No real HOME was used as a test write target. No control-root files were edited directly.


## Handoff audit warning

The successful role handoff reported run_wrote_outside_worktree under warn policy. Its unattributed rows name unrelated tasks BUG-261018-b2jnyg, BUG-261018-dbeghc and STORY-261018-m1ak35, with samples 2026-10-18T11:23:42.155598Z through 2026-10-18T11:26:57.971668Z; two other rows are explicitly attributed to other runs. This successor made no direct control-root writes: all task evidence and state writes used task-board CLI. The six rows attributed to this run's own chain are task/parent activity, progress and outcome resources. The warning is preserved, not cleared or treated as proof that the unrelated writes belong to this producer.


## Rework round 3: second recovery — runtime signing blocker

# Rework round 3 — second recovery: CR validator signing blocker

Task: BUG-261018-uoisfe — in-place-profile-skills-miss-native-surface.
This run made no repository source changes. The row-12 fix remains at
cmd/curator/main.go:283 and cmd/curator/command_help.go:128; behavioral sweep,
regression witness and list-only narrowing mutant remain in
cmd/curator/native_skills_flags_test.go:17, :54 and :102.
Rows 1–11 are unchanged; coverage remains the previously documented
12 of 12 attack-surface rows and 5 of 5 AC rows, with the production call
sites, positive-control bounds and retained reviewer names in the existing
coverage map. Muse, opencode and managed homes remain out of scope.

Candidate: 3e00fcc8abe75e411a88773c65c4042ae2448eb0, signature G.
Observed worktree tree: 2fe068da665b203ed48e52cfcb7fba74de96e37e,
identical to the signed candidate. Story HEAD stays at checkpoint
5d8e6b07f7f485557d9bbb065c9692348ca065c2; all changes remain uncommitted.
Fresh origin/main is 9c1605a74436d2448f759bd4b66ae6685237d83b;
the changed-path intersection is empty and root README.md is unchanged.

| Check | Real exit | Evidence / bound |
| --- | ---: | --- |
| `go test -count=1 -timeout=4m -v ./cmd/curator -run '^TestNativeSkillsTakeover(EveryRegisteredCommand|ArgumentDiscardRegression|ActivationBoundary|RegistryNarrowingMutant)$'` | 0 | Four tests; all 67 unrelated command cases; run directly in this run |
| `go build -o .temp/BUG-261018-uoisfe/resume2-curator ./cmd/curator` | 0 | Direct build in this run |
| `git diff --check` | 0 | Empty diagnostic log |
| Temporary-index identity / signature / fresh-trunk overlap verification | 0 | No real index or branch changes |
| Host-prescribed environment + task-local `ssh-keygen -Y sign -n git` probe using the configured public signing key | 0 | Agent signing works in the correctly bootstrapped shell; signature output stayed private and ignored |
| Fresh provider read of full hosted gate | 0 | Run 39117060646, attempt 2, completed/success on the exact signed candidate |
| CR revision 2 runtime `sh scripts/remote-gate.sh` | 1 | Previously recorded; signing fails before hosted execution; not green |
| CR revision 3 runtime `sh scripts/remote-gate.sh` | 1 | Previously recorded; same signing failure before hosted execution; not green |

Full local suite, lint, vet and platform-case evidence are reused from the
unchanged candidate and environment, not claimed as freshly rerun: 85 packages,
3494 top-level names, 28 CLI/envprofile groups, with exact prior/current identity
bounds already attached in validation-r3 and its 121-payload evidence bundle.
Fresh provider URL: https://github.com/relux-works/curator/actions/runs/39117060646
No new full hosted run was dispatched; the exact candidate already has a green
full hosted gate. That result does not overwrite the two failing CR validations.

| Mutant | Gate narrowed to | Named tests that fail | Survivor bound |
| --- | --- | --- | --- |
| list-only-consent-bypass | Only `list` may silently accept takeover; all other consent rejections stay present | TestNativeSkillsTakeoverEveryRegisteredCommand/list; TestNativeSkillsTakeoverArgumentDiscardRegression/list | Killed, child exit 1; no survivor |

## Constraint, recovery attempts and required owner action

The successor prompt described a missing handoff, but persisted notes and CR
validation logs show both preceding runs called handoff and the post-turn CR
construction subsequently failed. The failure is not a missing product change.
Both runtime logs stop at `git commit-tree -S` in scripts/remote-gate.sh:47:
SSH signing tries to load the private signing-key file and requests a passphrase.
There is no test, build or hosted result from those failed commands.

Sourcing the host-prescribed `agent-git-env` works in this run's shell, and a
fresh task-local signing probe exits 0. The preceding recovery already sourced
that environment before handoff; CR revision 3 still failed afterward. Therefore
shell-local bootstrap is not evidence that the post-turn validator has the same
environment. The runtime's actual environment was not exported or inspected;
its precise missing/overriding variable remains unverified.

The supported handoff, worktree and spawn CLI help exposes no producer command
to synchronously republish this candidate with an explicit validator environment.
Changing the gate, repository config, shared git config or runtime installation
to hide the failure would exceed the row-12 rework scope. Restarting another
session's shared runtime is prohibited by the host state-preservation rules.
No private key, alternate socket, unsigned commit, forged hosted status or
validation bypass was used. No control-root file was directly edited.

Recommended recovery: the runtime owner must bootstrap the process performing
post-turn CR validation with the prescribed agent transport/signing environment
(configured public signing key and swarm socket), then retry the producer's
normal CR publication. If that process needs a restart, its owner must arrange
it safely between affected sessions. Unlocking keys is needed only if that
correctly bootstrapped process actually reports unavailable identities; the
current shell signing probe shows available signing authority.

Alternatives: an owner-supported per-validator environment configuration would
avoid a shared-runtime restart if available. A product-code or gate-script
workaround would widen this leaf and is not recommended. More no-change
handoffs from this shell would repeat the already disproven recovery.

Status: blocked on runtime-owner signing-environment repair. No handoff is
claimed for this run, and no further unchanged CR attempt is knowingly queued.
The existing gaps remain: Muse multi-root ownership, adapter-ledger interplay,
and opencode's machine-global native surface.


## Publication retry — round 4

No source, test, documentation, configuration or candidate file changed in this
run. The entire materialized worktree still equals signed candidate
`3e00fcc8abe75e411a88773c65c4042ae2448eb0`, tree
`2fe068da665b203ed48e52cfcb7fba74de96e37e`; signature verification reports `G`.
Story HEAD remains at checkpoint
`5d8e6b07f7f485557d9bbb065c9692348ca065c2`; the candidate stays uncommitted.
The isolated-index whole-tree comparison and signature check exited 0. The
real index and Story branch were not modified.

Full local, hosted and mutant evidence from round 3 is reused on the unchanged
candidate; those suites were not rerun by this publication-only producer. The
configured post-turn validator reruns automatically after handoff. Its result
is pending and is not claimed green here. The approved host environment was
sourced for this run's signing/identity check and handoff; the round-4 brief
states that the validator launch environment is repaired. Both earlier runtime
snapshot-signing failures (CR revisions 2 and 3, exit 1) remain recorded above.

No refresh-candidate was requested by the board at preflight. This retry
preserves the exact candidate identity. Coverage remains the recorded 12 of 12
surface rows and 5 of 5 AC rows; no accepted row was reopened. Muse multi-root,
adapter-ledger coexistence and opencode remain the recorded open spec questions.
No product operation or test used the real HOME as a write target. Only board
CLI mutations update the authoritative board; no control-root file was edited.


### Publication retry refusal — checklist sequencing blocker

The prescribed-environment `task-board handoff BUG-261018-uoisfe --role developer`
ran synchronously and exited **1**. It refused because checklist item 24 is
unchecked: "Post-turn Change Request validation succeeds under the prescribed
runtime signing environment; two prior signing failures remain recorded".
A fresh compact board read exited 0 and confirms items 1-23 checked, item 24
unchecked, and status development. The handoff also warned that it could not
compare board content for the current run and recorded no progress for that
command; that warning is retained, not interpreted as a clean audit.

The candidate is unchanged and no source repair is needed. The earlier two
signing failures remain above. This retry has no successful post-turn validator
result: a future validation is neither a passing command nor evidence that can
truthfully check item 24. No checklist item was checked by inference. No
refresh-candidate refusal occurred; refresh is unrelated to this blocker.

Constraint: the handoff prerequisite demands an outcome described by the brief
as occurring automatically after the producer turn. The approved signing
environment fixes the former signing issue but cannot supply that future
outcome before this handoff. Altering the requirement, falsely checking it or
substituting prior manual hosted evidence would weaken the acceptance contract.

Required external input: the runtime/board owner must either run and record the
actual configured CR validation under the approved environment and check item
24 from that evidence, or correct the checklist/lifecycle sequencing so the
post-turn gate is enforced after producer handoff. Recommendation: keep the
validation mandatory and repair its sequencing. Then retry handoff of the same
unchanged candidate. No request to weaken signing, checks or review is made.
Status is routed to blocked on this external lifecycle prerequisite; no review
handoff or successful new CR validation is claimed.


## Publication retry — round 5

The unchanged worktree was verified against signed candidate
`3e00fcc8abe75e411a88773c65c4042ae2448eb0`: both trees are
`2fe068da665b203ed48e52cfcb7fba74de96e37e`, and Git signature verification
reports `G` (identity command exit 0). The Story checkpoint remains
`5d8e6b07f7f485557d9bbb065c9692348ca065c2`; the candidate delta remains
uncommitted. No source, test, documentation or configuration file changed.

This retry only updates board evidence and requests the normal developer
handoff. The orchestrator reports the validator signing environment repaired
and removed the impossible pre-handoff post-turn-validation checklist item.
No such checklist item was added here. Existing exact-candidate local/hosted
evidence is retained; no suite was rerun in this publication-only retry.
Configured post-turn validation is pending and is not claimed green here.
No candidate refresh was required by the observed unchanged checkpoint/tree.
The open Muse multi-root, adapter-ledger and opencode questions remain.


## Publication retry — round 5 successor

No source, test, documentation or configuration file changed. The fresh
standalone identity check exited 0: worktree tree
`2fe068da665b203ed48e52cfcb7fba74de96e37e` equals signed candidate
`3e00fcc8abe75e411a88773c65c4042ae2448eb0` (signature `G`). Story HEAD
remains checkpoint `5d8e6b07f7f485557d9bbb065c9692348ca065c2` with the
candidate delta uncommitted; root README is unchanged. No refresh-candidate
was required by the observed checkpoint/tree identity.

The predecessor's configured CR revision 4 validation ran and **failed,
exit 1**: [hosted run 39137598202](https://github.com/relux-works/curator/actions/runs/39137598202).
Snapshot signing and publication succeeded. The validation log records five
jobs failing at `Complete runner`, with cleanup permission-denied output.
This infrastructure failure remains failed evidence; it is not a product
pass or a successful configured validation. The original immutable resource
`BUG-261018-uoisfe_change-request_rev4-validation.log` is retained.

This publication-only successor did not rerun tests, build, lint or hosted CI.
Previously attached exact-candidate local and hosted evidence is retained
with its existing bounds. The next configured validator runs automatically
after the required developer handoff; its result is unknown at this point.
No checklist prerequisite about that future result was added or checked.
Coverage remains the recorded 12 of 12 surface rows and 5 of 5 AC rows;
Muse multi-root, adapter-ledger coexistence and opencode remain open spec
questions. No product operation or test touched the real HOME.


## Publication retry — round 5 recovery attempt 2

Fresh standalone identity verification exited 0. The unchanged worktree tree
`2fe068da665b203ed48e52cfcb7fba74de96e37e` equals signed candidate
`3e00fcc8abe75e411a88773c65c4042ae2448eb0`, signature `G`. Story HEAD
remains checkpoint `5d8e6b07f7f485557d9bbb065c9692348ca065c2`; the candidate
delta remains uncommitted. No source, test, documentation or configuration file
changed, and root README is unchanged. No observed refresh prerequisite applies.

CR revision 5 configured validation failed with exit 1: hosted run
[39146093379](https://github.com/relux-works/curator/actions/runs/39146093379).
Its persisted validation log records Lint and Naming gate failures at
`Complete runner`, with cleanup permission-denied output. This remains failed
evidence, never a successful configured validation. The immutable
`BUG-261018-uoisfe_change-request_rev5-validation.log` is retained.

This publication-only retry reran identity/signature verification, not tests,
build, lint or hosted CI. Previously attached exact-candidate evidence is reused
with its recorded bounds. The configured validator reruns automatically after
handoff; its next result is unknown. No future-validation checklist item was
added or checked. Coverage remains the recorded 12 of 12 attack-surface rows
and 5 of 5 AC rows, with the existing coverage map, tests and mutant table
unchanged. Muse multi-root, adapter-ledger interplay and opencode remain open
spec questions. No product operation or test touched the real HOME.

### Logbook

Recovery attempt 2 resumed from the persisted revision 5 validation failure.
Verified candidate identity and signature under the approved agent Git
environment. No repository edits, commits, branch moves or direct control-root
writes. Updated this gaps resource and attached task-scoped identity and retry
evidence before requesting developer handoff. The previous failed validator
and the next pending validator are kept distinct.


## Rework round 7

Finding `unrelated-takeover-argument-discard`, row 12: resolve takeover reached
successful help before the handler could require a true activating repair.
The permanent reviewer witness first failed with real exit 1: missing repair,
`--repair=false`, and `--repair=0` each returned 0 instead of usage exit 2.
The neighboring `--repair --dry-run --takeover --help` bypass was also driven:
`dry-run-witness-before.log` exited 1 for the same successful-help mechanism.
It is covered in this leaf, rather than left behind another help short-circuit.

`rejectsTakeover`, `cmd/curator/command_help.go:128`, now reads repair at :144,
dry-run at :146, normalizes `help <command>` at :154, and requires true repair
and false dry-run for resolve at :162. It runs from `cli.run`,
`cmd/curator/main.go:283`, before every help return (:288/:300/:315). The existing
admission diagnostic is retained. The late repair/takeover and dry-run/takeover
guards were removed from `cmd/curator/env.go` before :139/:147, leaving one
production admission check for help and dispatch. Standalone dry-run and
preserve-lock requirements still belong to the handler.

Rework is bounded to `cmd/curator/command_help.go`, `cmd/curator/env.go`,
`cmd/curator/native_skills_flags_test.go`, and the existing row-12 plant in
`internal/envprofile/native_skills_mutants_test.go:32`. The last path moves its
claude-only plant to the shared guard while retaining the original test/subtest
name. Accepted native materialization rows are unchanged. Root README,
CHANGELOG, Muse, opencode, managed-home layout and spec text are unchanged.

Named behavioral tests, all through `runProfile -> run -> cli.run`:

- `TestReviewNativeTakeoverHelpBoundary`, native_skills_flags_test.go:102:
  permanent reviewer witness, missing/false/zero repair and true dry-run refusal;
  no-consent help and true non-dry repair help controls; fixture state preserved.
- `TestNativeSkillsTakeoverHelpCrossProduct`, :127: every registry topic,
  every requested help syntax (both prefix and suffix flag placement), five
  takeover forms, seven install/resolve activation variants, plus five dry-run
  forms for each resolve variant. All no-consent help controls return 0.
- `TestNativeSkillsTakeoverHelpNarrowingMutant`, :241: two behavioral plants
  below. The revision-1 list-wrapper regression and list-only mutant retain
  their original names; every previously committed reviewer test is retained.

The complete help table below records **595 matrix rows**, **595 no-consent
positive controls**, and **2975 consent invocations** over **72 registry topics**.
Group/interactive commands use help controls; this does not claim successful
live UI/login operations. Actual no-flag list/update/gc/version controls remain
in `TestNativeSkillsTakeoverArgumentDiscardRegression`. Operands after `--`
are literals, as exercised by `TestNativeSkillsTakeoverActivationBoundary`.
All product calls pin HOME/USERPROFILE and native-home overrides to temporary
fixtures; development-tool cache/config reads are not claimed absent.

Coverage: **12 of 12 surface rows driven; 5 of 5 AC rows driven**. Rows 1–11
retain their accepted coverage map, tests and narrowing controls. Their native
tests and all twelve current plants were replayed in envprofile group 8; group
11 refreshes the CLI source scanner. No surface-table row is out of contract.
The round-2 scope exclusions remain Muse multi-root support, opencode native
surface changes, managed-home layout changes and spec text. Open spec questions:

1. Muse multi-root native ownership: how should independent XDG roots coexist
   with the one-home in-place ownership model? This leaf preserves Muse behavior.
2. Adapter-ledger coexistence: which scope arbitrates a ledgered legacy global
   entry when a current profile names it? This leaf leaves it untouched/non-current.
3. Opencode's machine-global surface: how should profile-local ownership interact
   with it? This leaf preserves today's staging behavior.

### Round-7 mutant table

| Mutant | What it narrows the gate to | Named failing test | Result / bound |
|---|---|---|---|
| `resolve-help-missing-repair` | Keeps repair refusal for dispatch but allows missing/false/zero repair only with a help flag | `TestReviewNativeTakeoverHelpBoundary` | Child exit 1, actual successful forbidden help exit 0; parent `TestNativeSkillsTakeoverHelpNarrowingMutant/missing-repair` exit 0 |
| `resolve-help-dry-run` | Keeps dry-run takeover refusal for dispatch but admits true dry-run only with a help flag | `TestReviewNativeTakeoverHelpBoundary` | Child exit 1 for `--repair --dry-run --takeover --help` returning 0; parent `TestNativeSkillsTakeoverHelpNarrowingMutant/dry-run` exit 0 |
| `list-only-consent-bypass` | Only list bypasses the otherwise present admission gate | `TestNativeSkillsTakeoverEveryRegisteredCommand/list`; `TestNativeSkillsTakeoverArgumentDiscardRegression/list` | Child exit 1 with observed successful admission; parent control exit 0 |
| `claude-takeover-without-repair` (relocated) | Only exact `claude_code` resolve operand bypasses true-repair admission | `TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code` | Child exit 1: proceeds to `profile_unknown` exit 1 instead of usage 2. Bound: demonstrates lost admission refusal, not successful repair; the new help controls directly witness forbidden success |
| Historical handler-only `claude-takeover-without-repair` plant | Weakened the retired late handler check while the new early guard remained | None; survivor | Exploratory envprofile suite exit 1, mutant child exit 0. Bound: the downstream check was redundant and ceased to be the production admission boundary. Replaced at the shared guard; no claim this survivor was killed |

The original rows 1–11 retain their named narrowing controls in the coverage
map. Current envprofile group 8 reruns all twelve plants with intended named
behavioral failures. Both new help plants preserve the repair token and execute
the behavioral witness, rather than a static checker. Setup/compiler/timeouts
are not counted as kills. Historical reds remain reds in the validation ledger.

### Help sweep table

72 registry topics; 595 matrix rows, 595 successful no-consent controls and 2975 consent invocations. All through runProfile -> run -> cli.run. No live UI/login operation is claimed.

| Command | Activation | `cmd -h` | `cmd --help` | `-h cmd` | `--help cmd` | `help cmd` |
|---|---|---:|---:|---:|---:|---:|
| `bootstrap` | `none` | 2 | 2 | 2 | 2 | 2 |
| `init` | `none` | 2 | 2 | 2 | 2 | 2 |
| `add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `install` | `none` | 2 | 2 | 2 | 2 | 2 |
| `update` | `none` | 2 | 2 | 2 | 2 | 2 |
| `upgrade` | `none` | 2 | 2 | 2 | 2 | 2 |
| `status` | `none` | 2 | 2 | 2 | 2 | 2 |
| `list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `project` | `none` | 2 | 2 | 2 | 2 | 2 |
| `project add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `project resolve` | `none` | 2 | 2 | 2 | 2 | 2 |
| `project refresh` | `none` | 2 | 2 | 2 | 2 | 2 |
| `skill` | `none` | 2 | 2 | 2 | 2 | 2 |
| `skill check` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global init` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global status` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global install` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global adopt` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global update` | `none` | 2 | 2 | 2 | 2 | 2 |
| `global upgrade` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile install` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile install` | `--use` | 0 | 0 | 0 | 0 | 0 |
| `profile install` | `--use=true` | 0 | 0 | 0 | 0 | 0 |
| `profile install` | `--use=1` | 0 | 0 | 0 | 0 | 0 |
| `profile install` | `--use=TRUE` | 0 | 0 | 0 | 0 | 0 |
| `profile install` | `--use=false` | 2 | 2 | 2 | 2 | 2 |
| `profile install` | `--use=0` | 2 | 2 | 2 | 2 | 2 |
| `profile import` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile use` | `none` | 0 | 0 | 0 | 0 | 0 |
| `profile update` | `none` | 0 | 0 | 0 | 0 | 0 |
| `profile remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile sync` | `none` | 0 | 0 | 0 | 0 | 0 |
| `profile compose` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile compose <profile> add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile compose <profile> remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `profile compose <profile> list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=true` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=1` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=TRUE` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=false` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--dry-run=false` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--dry-run=0` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair --dry-run=false` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair --dry-run=0` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=true --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=true --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=true --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=true --dry-run=false` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=true --dry-run=0` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=1 --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=1 --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=1 --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=1 --dry-run=false` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=1 --dry-run=0` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=TRUE --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=TRUE --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=TRUE --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=TRUE --dry-run=false` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=TRUE --dry-run=0` | 0 | 0 | 0 | 0 | 0 |
| `env resolve` | `--repair=false --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=false --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=false --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=false --dry-run=false` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=false --dry-run=0` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0 --dry-run` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0 --dry-run=true` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0 --dry-run=1` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0 --dry-run=false` | 2 | 2 | 2 | 2 | 2 |
| `env resolve` | `--repair=0 --dry-run=0` | 2 | 2 | 2 | 2 | 2 |
| `env status` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env config` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env config show` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env config set` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env config unset` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env migrate` | `none` | 2 | 2 | 2 | 2 | 2 |
| `env unmanage` | `none` | 2 | 2 | 2 | 2 | 2 |
| `run` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hybrid` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hybrid add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hybrid remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hybrid list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hybrid status` | `none` | 2 | 2 | 2 | 2 | 2 |
| `audit` | `none` | 2 | 2 | 2 | 2 | 2 |
| `gc` | `none` | 2 | 2 | 2 | 2 | 2 |
| `shell-init` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hook` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hook approve` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hook approvals` | `none` | 2 | 2 | 2 | 2 | 2 |
| `hook revoke` | `none` | 2 | 2 | 2 | 2 | 2 |
| `ui` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config show` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-ssh` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-ssh add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-ssh list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-ssh remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-https` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-https add` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-https login` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-https list` | `none` | 2 | 2 | 2 | 2 | 2 |
| `config build-https remove` | `none` | 2 | 2 | 2 | 2 | 2 |
| `version` | `none` | 2 | 2 | 2 | 2 | 2 |

Each cell applies to `--takeover`, `--takeover=true`, `--takeover=false`, `--takeover=0`, and `-takeover`. Without takeover each corresponding help control exits 0. Inputs after `--` remain operands, per the retained activation-boundary regression.

### Round-7 validation and candidate identity

Signed candidate `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`, tree `769581956b4fd7bfd02c2567d7465b60335f5b4b`, signature G. Worktree remains uncommitted on checkpoint `5d8e6b07f7f485557d9bbb065c9692348ca065c2` and matches that tree; actual index is empty. Fresh trunk `b9caa20901ee859eec20e649940693033e6297d4` has no candidate-path overlap.

Full composed local suite: 85/85 served packages, 3497 top-level names. Fresh: every CLI group (15), envprofile groups 8/11, full scriptworker/scriptpolicy/crossconformance, and the three install binary consumers. Reused unchanged identity: 79 packages and unchanged envprofile/install tests, explicitly bounded in validation-r7 and reuse-identity.json. All current direct test processes exited 0; missing-group evidence control intentionally exited 1. Raw and sanitized platform-case gates, build, vet, lint, format and signed-tree integrity checks exited 0.

[Hosted full gate](https://github.com/relux-works/curator/actions/runs/39158440563) succeeded on the exact candidate: 20 successful jobs and two conditional skips, no failed job or infrastructure failure. Historical failures and superseded cancelled runs are separated from current passing evidence. The configured post-turn validation remains a subsequent gate, with no advance success claim.

See BUG-261018-uoisfe_validation-r7.md, BUG-261018-uoisfe_validation-logs-r7.txt and BUG-261018-uoisfe_evidence-manifest-r7.json for exact commands, real exits, stream coverage and hashes.
