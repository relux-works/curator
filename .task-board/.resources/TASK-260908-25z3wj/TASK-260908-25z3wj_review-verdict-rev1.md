# TASK-260908-25z3wj CR1 review verdict

Verdict: accepted. No blocking findings. repeat-of: none.

Reviewed revision 1, candidate tree 5b154a523b71ec5daabb97a55e81d9af1d3a5c86 against base 13b28c9a8916464e7253551808ae9969d6aa0186. Repository delta is present and limited to internal/defaults/defaults.go, its external-package tests, .scripts/defaults-mutants.py and README.md. Candidate bytes match the workspace; all archived files were verified against the candidate tree after selected mutations. No repository code was modified by this review. Attacks ran only in a disposable archive under .temp/review-25z3wj/candidate.

## AC coverage assessed before code review

8 of 8 scoped behavioral AC rows driven. Tests belong to the uncommitted CR snapshot, as required by the managed Story contract; checkpoint will commit them. No main launch integration is claimed (0 main-pipeline rows driven, intentionally outside this leaf).

| Row | Production entry and named driving test |
| --- | --- |
| Strict schema, known env/member validation | Load -> loadFile -> parse -> fragment.ParseJSON/HomeVariable; TestLoadRejectsInvalid (28 cases in both machine and locked-machine/operator contexts), TestLoadKnownEnvironmentsAndPresence |
| Empty presence and unadmitted strings | Load / Files.Resolve; TestLoadKnownEnvironmentsAndPresence, TestResolveExplicitEmptyOverrides |
| Optional absence versus errors / no partial result | Load; TestLoadFilesystem, TestLoadRejectsInvalid |
| Per-member precedence | Files.Resolve; TestResolvePrecedence, all 64 presence combinations |
| Machine locks / ignored operator | Files.Resolve; TestResolveLocks/model and /effort, TestResolveExplicitEmptyOverrides |
| Origin metadata and unresolved members | Files.Resolve; TestResolvePrecedence, TestResolveLocks, TestLoadFilesystem/absent |
| Explicit process path derivation | ConfigPaths; TestConfigPaths |
| Unknown resolve env rejected / known opencode accepted | Load / Files.Resolve; TestLoadKnownEnvironmentsAndPresence |

## Independent verification

- go test ./internal/defaults -count=1 -v -cover: exit 0, 97.8% statements. All filesystem cases ran, including real permission denial; no root skip.
- Selected original narrowing harness, executed in the exact-tree archive with the full defaults behavioral suite and -count=1: 8 of 8 killed, each child exit 1 with its named failing test. Harness exit 0. See attached summary and logs.
- Attacks: extra-member -> TestLoadRejectsInvalid/unknown-member; sig -> TestLoadRejectsInvalid/unknown-sig; broken-link -> TestLoadFilesystem/broken-link; permission -> TestLoadFilesystem/read-permission; model-lock -> TestResolveLocks/model; effort-lock -> TestResolveLocks/effort; operator-ignore -> TestResolveLocks/model; presence -> TestResolveExplicitEmptyOverrides.
- These exercise unknown-member admission, read-error-as-absence, bypass paths around the lock, and absent-versus-explicit-empty confusion. The sig mutant preserves the token and runs behavioral tests, not a static search checker.
- git diff --check: exit 0. Candidate restoration verified against Git object bytes.
- Accepted existing exact-CR runtime evidence TASK-260908-25z3wj_change-request_rev1-validation.log: make check exit 0, build, fmt-check, vet, all tests and race tests. Did not repeat this full gate. Accepted producer evidence and its 20-mutant report for the remaining 12 attacks; did not claim independent reruns of them.

## Architecture and policy assessment

Load calls the strict JSON reader directly, without fragment canonicalization that could strip sig. Known environment IDs agree with SPEC 4.2, including opencode; strings are neither admitted nor normalized. Files entries are private, and both Load errors and lock refusals return zero results. Resolution applies flags/operator/machine precedence independently and ignores all operator values for a machine-named locked environment, while accepting flags for unset machine members. Operator locked is non-authoritative.

Validating ignored operator content is consistent with SPEC 4.3's unconditional existing-file read/parse-error contract: ignored means no value authority, not exemption from file validity. README documents this explicitly. ConfigPaths takes explicit process inputs and does no ambient environment lookup or writes. No main/SPEC/dependency/home/install/CI changes or hidden installation side effects were found. The reusable harness restores bytes in finally and checks named behavioral failures.

Bounds retained: no actual /etc or home reads in tests; no executable wiring, module Lineup, admission or launch diagnostics. Concurrent file replacement between stat and read and FIFO/device experiments remain unproven as stated by the producer; static nonregular paths are rejected before ReadFile, and directories/links/loops/read errors have real filesystem tests. No source-text admission gate is introduced. Full shared JSON parser mutation coverage is outside this leaf.

## Checklist and lifecycle

Implementation matches the scoped AC, fits the existing dependency-free internal package architecture, tests pass, and refusal behavior was attacked. Conditional changes-request routing is not applicable because the verdict is accepted. LOGBOOK/control-root writes are prohibited by this task; this task-scoped artifact is the persisted review record. Run goal query reports not goal-bound; no operator directives were present.

Accept CR1 via accept_cr. This is a non-final task_delta: accepted means integrating, not done or landed. A new bound developer/implementer run owns checkpoint; the following leaf still owns real tagged-module lineup, main pipeline wiring, origins/effort diagnostics and executable tests. Parent retains full Story delivery lifecycle.
