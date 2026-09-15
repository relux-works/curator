# TASK-260908-25z3wj producer evidence

State: ready for review; task_delta for the first leaf. Candidate remains UNCOMMITTED on the managed Story branch. Parent owns checkpoint and signed delivery.

## Scope and decisions

Added internal/defaults/defaults.go, external-package behavioral tests, .scripts/defaults-mutants.py, and an honest internal-milestone README section. No main.go, SPEC, dependency, CI, install, ax, home, or LOGBOOK/control-root changes. The existing fragment.ParseJSON and fragment.HomeVariable are reused; the parser does not strip sig before closed-schema rejection. The Go-testing skill's TUI-specific helpers are inapplicable to this dependency-free filesystem package; real filesystem tests follow the explicit brief.

Schema and defaults are required. Locked is optional and defaults false, matching the SPEC's “when the machine file carries locked: true” condition. Operator locked has no authority. Both files are validated, including ignored operator entries. Values remain strings with separate presence; no model/effort admission occurs. Files stores validated entries privately; the zero value means no configuration. ConfigPaths receives process XDG/home inputs explicitly, with no fragment environment or ambient test input. A missing file or parent is optional; dangling final/parent links, loops, unreadable files and nonregular files refuse. Nonregular files are rejected before ReadFile to avoid blocking on special files.

## Behavioral AC coverage

**8 of 8 leaf behavioral AC rows driven** through exported production APIs by the named tests below. These tests are in the uncommitted review candidate, as required by the managed-worktree contract; no claim of committed tests before checkpoint.

| AC row | Production call site | Named test |
|---|---|---|
| Strict schema, known envs, member types, null/unknown/duplicate refusals | defaults.Load -> loadFile -> parse -> fragment.ParseJSON/HomeVariable | TestLoadRejectsInvalid (28 cases, each machine and locked-machine/operator paths); TestLoadKnownEnvironmentsAndPresence |
| Explicit presence, empty values and unadmitted strings survive | defaults.Load and Files.Resolve | TestLoadKnownEnvironmentsAndPresence; TestResolveExplicitEmptyOverrides |
| Missing files optional; read/parse failure cannot yield fallback | defaults.Load -> inspect/ReadFile/parse | TestLoadFilesystem; TestLoadRejectsInvalid (also verifies no partial files leak on errors) |
| Flags > operator > machine per member | Files.Resolve | TestResolvePrecedence (64 combinations) |
| Locked members refuse flags including equal/empty values; unset members accept flags; ignored operator supplies nothing | Files.Resolve | TestResolveLocks/model and /effort; TestResolveExplicitEmptyOverrides (operator locked ignored) |
| Per-member origins and unresolved members retained | Files.Resolve | TestResolvePrecedence; TestResolveLocks; TestLoadFilesystem/absent |
| Launcher machine and XDG/default operator locations from explicit process inputs | defaults.ConfigPaths | TestConfigPaths |
| Invalid resolution env refuses, while known opencode config remains valid | Files.Resolve and defaults.Load | TestLoadKnownEnvironmentsAndPresence |

Scope-bound facts (not launcher end-to-end claims): executable wiring is explicitly excluded, so **0 main-pipeline launch rows driven for defaults**. No tests exercise actual /etc or user-home configuration. Filesystem disappearance between inspection and read is fail-closed by code, but concurrent replacement races and special-device/FIFO behavior are not experimentally exercised. Permission-denied test ran on this non-root host; it explicitly skips root because root bypasses permission bits. Relative XDG normalization/admission is not specified here. Shared strict JSON parser's entire rejection vocabulary is not exhaustively mutated by this leaf. Root-kind and schema-type refusals also have independent required-member/schema-value refusals; no claim that every redundant clause has independently been defeated.

## Commands personally executed

| Command | Real exit | Evidence |
|---|---:|---|
| go test ./internal/defaults -count=1 -cover | 0 | defaults-test-01.log; 97.8% statements |
| python3 .scripts/defaults-mutants.py .temp/TASK-260908-25z3wj/mutants | 0 | mutants-01.log; each child gate is expected-red exit 1 |
| make check | 0 | make-check-01.log; build, fmt-check, vet, all tests, all race tests |
| git diff --check | 0 | scope/self-review; no whitespace errors |

make check ran once directly at publication after all edits and restoration. No gate used tee or a pipe chain. No earlier attached validation was accepted in place of a rerun. Mutants run the full behavioral defaults suite with -count=1, verify the named failing test (not merely any error), and restore exact candidate bytes from memory rather than Git. Source-search admission gates are not introduced; the sig mutant nevertheless preserves the sig token while changing behavior.

## Narrowing mutant evidence

20 of 20 mutants killed; every child gate below genuinely failed with exit 1 (expected failure because it admits a forbidden case). No surviving mutant.

| Mutant | Gate narrowed to / admitted case | Named failing test | Exit | Survivor bound |
|---|---|---|---:|---|
| schema-version | admit schema v2 only | TestLoadRejectsInvalid/wrong-schema | 1 | none |
| locked-null | admit null lock only | TestLoadRejectsInvalid/locked-null | 1 | none |
| defaults-null | admit null defaults only | TestLoadRejectsInvalid/defaults-null | 1 | none |
| unknown-env | admit future env only | TestLoadRejectsInvalid/unknown-env | 1 | none |
| empty-entry | admit empty pi entry only | TestLoadRejectsInvalid/empty-entry | 1 | none |
| entry-null | admit null env entry only | TestLoadRejectsInvalid/null-entry | 1 | none |
| member-null | admit null string members only | TestLoadRejectsInvalid/model-null | 1 | none |
| extra-member | admit extra entry member only | TestLoadRejectsInvalid/unknown-member | 1 | none |
| sig | admit top-level sig only; token retained | TestLoadRejectsInvalid/unknown-sig | 1 | none |
| required-schema | allow absent schema only | TestLoadRejectsInvalid/missing-schema | 1 | none |
| required-defaults | allow absent defaults only | TestLoadRejectsInvalid/missing-defaults | 1 | none |
| duplicate | admit duplicate model only in strict JSON reader | TestLoadRejectsInvalid/duplicate-model | 1 | none |
| broken-link | misclassify ENOENT symlink targets only as absence | TestLoadFilesystem/broken-link | 1 | none |
| permission | treat permission-denied reads only as absence | TestLoadFilesystem/read-permission | 1 | none |
| directory | treat directories only as absence | TestLoadFilesystem/directory | 1 | none |
| resolve-env | admit future resolve environment only | TestLoadKnownEnvironmentsAndPresence | 1 | none |
| model-lock | enforce locks only for effort | TestResolveLocks/model | 1 | none |
| effort-lock | enforce locks only for model | TestResolveLocks/effort | 1 | none |
| operator-ignore | ignore operator only when machine sets effort | TestResolveLocks/model | 1 | none |
| presence | ignore explicitly empty flags only | TestResolveExplicitEmptyOverrides | 1 | none |

## Remaining lineup/pipeline scope

Next leaf must consume the real tagged agents-management module; map real compatibility and Lineup ordering; resolve still-unset model and effort without invented recommendations; wire the exported API into the production pipeline with process configuration inputs; emit per-member origins each launch; preserve module admission/refusal behavior and --effort diagnostics; and test actual executable paths. This first leaf claims none of those behaviors or full launcher defaults delivery. ax.json and other pipeline stages remain outside this task.

## Lifecycle and operational notes

The mandated initial development transition returned exit 1 because the task lacked an estimate. Recorded estimate 5 and repeated transition successfully (exit 0). No operator directives were pending. LOGBOOK writes are explicitly prohibited by this leaf's brief; findings are persisted in this outcome and board notes instead. Source-text-gate checklist item is not applicable. Candidate snapshot/checkpoint, repository freshness and integration belong to the parent; this worker did not switch, rebase or commit the managed Story branch.

Baseline HEAD: 13b28c9a8916464e7253551808ae9969d6aa0186

Platform: macOS-26.6.2-arm64-arm-64bit-Mach-O

go version go1.25.5 darwin/arm64
