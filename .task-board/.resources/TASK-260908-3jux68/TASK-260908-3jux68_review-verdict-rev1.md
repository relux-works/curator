# TASK-260908-3jux68 independent review verdict — CR-TASK-260908-3jux68-1 revision 1

**Verdict: ACCEPT.**

Reviewer: claude-fable-5-1 (independent, read-only). Reviewed the exact candidate: base `9a6025d169a49b4cd692486bf8808f4dfc2d3044`, candidate tree `74601fb31e0b436aafa4fd89050bc9911bf6596b`. I recomputed the candidate tree OID from the worktree with a throwaway index (`git read-tree HEAD; git add -A; git write-tree`) and it equals the CR tree OID. Working tree is uncommitted at HEAD 9a6025d (no producer commit).

Shell: bash invoked from zsh, `set -o pipefail`; every gate run as a standalone process. Nothing was accepted from the producer's evidence without rerun.

## 1. Byte-exactness against relux-agents-infra 459742ea67e3c6b84169520b92d74fe7f73e3002

Inventory from `git -C /Users/administrator/Developer/ReluxWorks/relux-agents-infra-main ls-tree 459742e .instructions/`: 15 files = AGENTS.md + INSTRUCTIONS.md + 13 INSTRUCTIONS_*.md. INSTRUCTIONS_EXTERNAL_RESOURCES.md is absent upstream; candidate deletes it. AGENTS.md is a byte alias of INSTRUCTIONS.md and is correctly not duplicated (core README records this).

Per-module `git show <pin>:.instructions/<name> | shasum -a 256` versus local file: **14 of 14 OK**, digests identical to the candidate `SOURCES.sha256` and to the producer's byte table (1ce362ad… INSTRUCTIONS.md … ffae7f51… DIAGRAMS.md). `shasum -a 256 -c SOURCES.sha256`: 0 non-OK rows.

Split: core 9 (index, PLATFORM, STRUCTURE, BROWSER_AUTOMATION, TOOLS, SKILLS, SKILL_TRIGGERS, DIAGRAMS, DOCS), workflow 2 (TESTING, WORKFLOW), style 1, claude 1 (REMOTE_AGENTS, `environments: ["claude_code"]`), attachments 1. BROWSER_AUTOMATION → core per the tool-capability rule, recorded in core README. Every leaf README carries `source: relux-works/relux-agents-infra@459742ea… .instructions/<file>` lines (14 lines, one per module).

## 2. Manifests via the curator parser (oracle)

Throwaway `go run` main under curator control-root `.temp/` (deleted afterwards) calling `internal/contextpkg.LoadManifest` + `ValidateModules` with registered envs {claude_code, codex_cli, opencode, pi}: **exit 0**, no errors, no warnings for all six packages. Observed: core weight 100 / 9 modules; workflow 70 / 2; claude 50 / 1 with envs=[claude_code]; attachments 40 / 1, skills=1; style 10 / 1; ivan hasContext=false, modules=0, weights map {attachments 40, claude 50, core 100, style 10, workflow 70}, contexts=5, skills=3, mcp=0.

Weights match Decision 0012; umbrella weights map equals leaf weights; no edge weights; all five ranges `^1.0` admit `1.0.0`; versions all 1.0.0.

## 3. validate.sh and tests

- `bash scripts/validate.sh` → exit 0 (`manifests, modules, weights, ranges: OK`; `sources: 14 module digests OK`; PASS).
- `python3 -m unittest discover -s tests -v` → exit 0, 9 tests OK. Tests execute the production entry point `scripts/validate.sh` (which invokes `scripts/check_sources.py`) in disposable copies.
- `python3 tests/mutants.py` → exit 0; all 7 producer narrowing mutants killed (trailing-lf, unknown-env, weight, digest, sources-digest, inventory, duplicate), each with suite exit 1 and the named probe failing.

## 4. Reviewer's own mutants (all bit)

| Mutant | Result |
| --- | --- |
| A: umbrella range `^2.0` for core | validate.sh exit 1: `umbrella range '^2.0' does not admit relux-root-context-core (1, 0, 0)` |
| B: edge `weight` on umbrella requires.contexts.core | exit 1: `umbrella must not mix edge weights with the weights map` |
| C: extra undeclared module on disk WITH matching SOURCES row | exit 1: `sources: checksum, disk and manifest inventories differ` (manifest inventory gate bites independently of digests) |
| D: token-preserving weakening — keep `hashlib.sha256` token, compare only 8 hex chars and exempt the all-zero digest | behavioral suite FAILED: `test_sources_drift` (mutant killed; harness runs the behavioral suite, not a static checker) |

## 5. Bounded rows (accepted as declared bounds)

- Tags NOT created; six signed tags are the orchestrator's step after acceptance. Correct.
- Umbrella `requires.skills` = pdf, skill-creator, agents-attachments at `^0.1` with `git@github.com:relux-works/skill-{pdf,creator,agents-attachments}.git`; these match TASK-260908-2kihaw producer evidence §4 lines exactly. `requires.mcp` deferred to B3 and documented in umbrella README. No invented sources.
- `.gitignore` has `.temp/` and `__pycache__/`.
- SOURCES.sha256 is a reviewed baseline, not a signature (README says so); coordinated module+hash forgery is out of scope of the offline gate. Native include expansion, live tag/skill resolution, and non-macOS platforms are unverified.
- Coverage: 3 of 6 rework rows have committed tests driving the production entry point (row 1 selectors/weights, row 2 offline drift, row 4 partially via local gate); rows 3, 5, 6 are inspection/oracle bounds. Producer reported this honestly; acceptable for a content-package repo with no runtime.
- Signature check of the landed commit is the integration path's job; base 9a6025d was verified signed in the previous review cycle and this CR adds no commit.

## Required changes

None.
