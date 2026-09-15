# TASK-260908-3jux68 results rev2

## Candidate and findings

Managed Story candidate, UNCOMMITTED against checkpoint `9a6025d169a49b4cd692486bf8808f4dfc2d3044`. No producer commit, push, tag, release, install or home mutation. Integration and independent reviewer acceptance remain orchestrator-owned.

Preserved and checked the partial candidate from the earlier runs. All 14 shipped modules equal `relux-works/relux-agents-infra@459742ea67e3c6b84169520b92d74fe7f73e3002` byte for byte. Inventory derived with `git ls-tree` from that pin; AGENTS.md equals INSTRUCTIONS.md and is not duplicated. EXTERNAL_RESOURCES is absent upstream and removed. BROWSER_AUTOMATION belongs in core as a tool-capability module. All leaf READMEs name source paths. Versions remain 1.0.0; weights remain 100/70/50/40/10. Umbrella skills exactly match sibling B2 evidence section 4; MCP stays deferred. `.gitignore` excludes `.temp/` and Python caches.

`scripts/check_sources.py`, invoked by `scripts/validate.sh`, rejects missing/malformed/duplicate inventory and SHA256 drift. Tests invoke the production shell entry point in disposable copies. Trailing-LF test updates the disposable digest baseline to isolate the format gate; its narrowing mutant permits exactly the extra-LF suffix class ending in period + two LF.

## Commands personally executed

Shell: zsh with `set -o pipefail`; each gate ran as a standalone process, without tee or pipelines.

| Command | Exit / result |
| --- | --- |
| `bash scripts/validate.sh` | 0; manifests/modules/weights/ranges OK; 14 digests OK; PASS |
| `python3 -m unittest discover -s tests -v` | 0; 9 tests passed |
| `go run ./cmd/b1-oracle <absolute Story worktree>` in `.temp/oracle` | 0; six manifests accepted, module counts 1/1/9/0/1/2 |
| `bash -n scripts/validate.sh` | 0 |
| Python `ast.parse` on scripts/check_sources.py and tests/*.py | 0 |
| `git diff --check` | 0 |
| Python source comparison using `git show <pin>:.instructions/<basename>` for each shipped file, asserting inventory equality from `git ls-tree`, plus AGENTS alias comparison | 0; all 14 exact |
| Python comparison of every `.temp/oracle/internal/**/*.go` with current curator control-root source | 0; all copied production source exact |

Oracle is a throwaway copy under this worktree, not a modification to curator. Its main calls `internal/contextpkg.LoadManifest` and `ValidateModules` on every package with the four registered environments. Oracle output:

```
relux-root-context-attachments: LoadManifest + ValidateModules OK (1 modules)
relux-root-context-claude: LoadManifest + ValidateModules OK (1 modules)
relux-root-context-core: LoadManifest + ValidateModules OK (9 modules)
relux-root-context-ivan: LoadManifest + ValidateModules OK (0 modules)
relux-root-context-style: LoadManifest + ValidateModules OK (1 modules)
relux-root-context-workflow: LoadManifest + ValidateModules OK (2 modules)
```

No prior green validation accepted as current-run evidence. Full landing suite intentionally reserved for runtime handoff (exactly once). Shellcheck is unavailable on PATH; shell syntax check and Python AST checks ran, but no shellcheck/linter pass is claimed. No Windows/Linux or live profile-install validation performed.

## Coverage and bounds

**0 of 6 AC rows driven under the literal committed-test requirement**: new tests are deliberately uncommitted until runtime snapshot/integration. Six rows are the numbered rework requirements. Candidate tests drive the offline-drift portion of row 2 and environment/weight negatives of row 1; other portions remain inspection or oracle bounds, not whole-row behavioral coverage.

| Rework row | Production call site / evidence | Bound |
| --- | --- | --- |
| 1 refresh/split/selectors/weights | `scripts/validate.sh`; Validation.test_unknown_env, test_weight_drift; source git-show comparison | Exact selectors/weights and full upstream inventory inspected; native include expansion not tested |
| 2 pins/offline checksum | `scripts/validate.sh` → `scripts/check_sources.py`; test_source_drift, test_sources_drift, test_missing_sources, test_missing_row, test_duplicate_row | README pins inspected; coordinated module+checksum forgery is not detected by unsigned offline baseline |
| 3 .gitignore | file inspection | No committed behavioral test |
| 4 manager-valid manifests | Curator LoadManifest + ValidateModules via throwaway oracle | Oracle executed, not a committed test; local validator only checks a schema subset |
| 5 1.0.0/no tags | manifest/workflow inspection | No live tag resolution or publication test |
| 6 umbrella skills/deferred MCP | Curator LoadManifest + B2 section 4 comparison | No live skill dependency install/resolution |

Additional named tests: `Validation.test_valid` and `Validation.test_trailing_lf` execute `scripts/validate.sh`. New source-drift gate has negative and narrowing evidence; inherited generic schema clauses are not exhaustively mutation-covered. No source-token-search attestation gate exists: checks operate on parsed JSON or actual bytes. Mutation harness changes behavior and executes the entire behavioral suite.

Independent review and signed PR delivery are pending, not asserted. Findings are recorded in board notes and this resource as the permitted logbook surface; LOGBOOK.md edits are expressly forbidden.

## Byte table

```text
1ce362adb54d612cd6c0a465fe105f878b3b70ca6db0583fccbd242b1db283ec  packages/relux-root-context-core/context/INSTRUCTIONS.md
261475bf17cc210bac8b850efb458357a155d62d29291b00a94f7e3cbc014cd9  packages/relux-root-context-core/context/INSTRUCTIONS_PLATFORM.md
43f0f117bdb96dfb2d62c3ce036dfdece46ced7aceb7ced4c0910cffbf292eb2  packages/relux-root-context-core/context/INSTRUCTIONS_TOOLS.md
4dcfbed5b38cd15f241fd973915b7dc8aac8c61405a3cb67246392817a31f03f  packages/relux-root-context-claude/context/INSTRUCTIONS_REMOTE_AGENTS.md
5b7539e1af4d44a4391d9fe7e0f4d3da09ec74bedc34a77703587240b499eb5a  packages/relux-root-context-workflow/context/INSTRUCTIONS_WORKFLOW.md
655e2960311de5d31f96e1561f7d949ee1b4d0f8ec19fb99faf53872b523aa4c  packages/relux-root-context-core/context/INSTRUCTIONS_SKILLS.md
6bd6def53361fa5d1b26bb8baced60a1f5b56db565f216efcf838a59cdb02378  packages/relux-root-context-core/context/INSTRUCTIONS_STRUCTURE.md
73ddf3882a7260393998da77d4502623bd81473d483de71e6e5418facda69077  packages/relux-root-context-attachments/context/INSTRUCTIONS_ATTACHMENTS.md
7c17ffa77d1393969d8cece10c85ef4e95b4471f7af1227746b7e186fde2a26a  packages/relux-root-context-style/context/INSTRUCTIONS_STYLE.md
8cef94d9ffc2a5534a22d0b8b31a21eb87de35b4bc43ede1e64b89b1178ab082  packages/relux-root-context-workflow/context/INSTRUCTIONS_TESTING.md
b9cba5fc6818c4e157e507ba7b9f3c4d57c2927929a5170c9b1abb91d2761f02  packages/relux-root-context-core/context/INSTRUCTIONS_DOCS.md
be152470a9744d047d1a8b198900a67fc0c8a9e1b01fb6151b54bc7c48103748  packages/relux-root-context-core/context/INSTRUCTIONS_BROWSER_AUTOMATION.md
dbb373fc1fc0b86e0277712cbbcaefd6a57e48cc455206b188102be0c8d2b08b  packages/relux-root-context-core/context/INSTRUCTIONS_SKILL_TRIGGERS.md
ffae7f515787fb8e141f39328c044808eec082bb1b88d3a53ec672ae4b0d793b  packages/relux-root-context-core/context/INSTRUCTIONS_DIAGRAMS.md
```

## Narrowing mutants personally executed

`python3 tests/mutants.py`: exit **0**. Each disposable mutant executed `python3 -m unittest discover -s tests -v` with real exit **1** (expected red: named negative test observed the weakened validator admit the prohibited input). No survivor among these seven probes. Gate logic remains present in every mutation.

| Mutant | Narrowing exception | Named failing test | Suite exit |
| --- | --- | --- | --- |
| trailing-lf | permits period followed by two LF | test_trailing_lf | 1 |
| unknown-env | permits future_env only | test_unknown_env | 1 |
| weight | permits drift when leaf is 101 | test_weight_drift | 1 |
| digest | permits modified bytes prefixed changed + LF | test_source_drift | 1 |
| sources-digest | permits all-zero expected digest | test_sources_drift | 1 |
| inventory | permits an inventory missing one of 14 rows | test_missing_row | 1 |
| duplicate | permits duplicate first checksum path only | test_duplicate_row | 1 |

The mutant suite tests gates through `bash scripts/validate.sh`, never helper-only calls. The current candidate's ordinary nine-test suite exited 0; these seven separate mutated suites exited 1 by design.
