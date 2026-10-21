# Brief: the audit network-declaration detector scans Markdown only inside agent-facing context (Curator, regression fix, URGENT)

**Repository:** curator. **Branch:** `fix/audit-markdown-scope`, from current `origin/main`.

## Bug

`internal/audit/audit.go` `detectWithOpaquePaths` (about line 832) scans **every `.md` file anywhere in the snapshot** for URLs. That was added by CIP-0021 (PR #151). Before, only `scripts/` and the manifest were scanned. Result: installing `project-management` (task-board) from the repository root audit-blocks with `high audit.capability.network-undeclared`. The hits are historical URL citations in `.research/*.md`, `.task-board/.resources/**` and `tools/board-cli/vendor/**/README.md` (a build root). Curator is now the only install path for task-board, so this blocks Gate A.

## Correct scope (normative source: curator-spec core §3, the agent-facing context rules)

Markdown is scanned only where it can reach the agent:
- the root entries eligible for agent-facing context: `SKILL.md`, `agents/`, `references/`, `.skill_triggers/`, `assets/`, `templates/`, `examples/`, `data/`, plus `scripts/` only when the package exports no commands;
- **minus**, at every depth, the core §3 exclusion patterns (`.git .github .gitlab-ci.yml .venv __pycache__ *.pyc node_modules tests test __tests__ README* CHANGELOG* LICENSE* Makefile setup.py pyproject.toml requirements*.txt .DS_Store .gitignore`);
- **minus** whole `runtime_roots` and `build_roots` subtrees from the validated manifest.

Reuse Curator's existing implementation of that eligibility and exclusion set, which already exists for context assembly and the `skill.build_root_in_prompt_context` check. Do not write a second, divergent copy. Keep the existing non-Markdown scanning (`scripts/`, the manifest) exactly as it is today. Keep the CIP-0021 selected-directory rule: the audit subject is the selected package tree.

## Tests (each test states the claim it proves)

- Regression: a package with URLs in `.research/notes.md`, `.task-board/.resources/x.md` and `<build_root>/vendor/m/README.md` installs with no network finding.
- Positive controls: an undeclared URL in `SKILL.md`, in `references/guide.md` and in `assets/a.md` is still `audit.capability.network-undeclared`; a declared host passes.
- Boundary: a `README.md` inside `references/` is excluded (core §3 pattern); `scripts/x.md` behaves according to the exports rule.
- The CIP-0021 vectors and the existing audit, conformance and network-contact vector tests stay green.
- A narrowing mutant that restores "every .md" fails the regression; a mutant that drops `references/` from the scan fails the positive control.
- Bump the audit `RulesetVersion` if the detector change requires it under the cache-reuse rule from round 7 of CIP-0021 (`TestReviewPriorMarkdownAllowance`). Explain your reasoning either way.

## Run

`go build ./...`, `go vet ./...`, `env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 ./internal/audit/... ./internal/envprofile/... ./internal/skillcheck/...`, and the full default suite in a normal-.git validation clone with a short TMPDIR and a private GOCACHE. End to end, if feasible: install a local copy of a package shaped like project-management (repo-root `SKILL.md`, `.research/` with URLs, a build root with vendor READMEs) through `curator profile install <path>` in a temporary HOME, and show it admits.

## Hand-off

Results file, branch pushed (signed), handover. Changelog line in the results file (no CHANGELOG.md edit).
