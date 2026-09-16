# Brief — TASK-260916-3na4vf (umbrella requires.mcp directory + coordinated 1.0.1)

Why: `curator profile install git@github.com:relux-works/relux-root-context.git --directory packages/relux-root-context-ivan --range '^1.0' --use --takeover` on host e11-1 refused with
`profile_source_invalid: mcp_declaration_invalid: agent-mcp.json is absent at <snapshot>` because relux-mcp hosts its packages at `packages/figma/agent-mcp.json` and `packages/safari/agent-mcp.json`, while the umbrella's `requires.mcp` entries carry no `directory`. Curator's manifest parser admits `directory` for `requires.mcp` entries (internal/contextpkg: allowed keys git, range, tag, revision, directory).

Facts about Curator's resolver you must respect (installed curator main-04550e2, internal/envprofile/gitsource.go + internal/contextresolve):
- Version candidates of a git source are ONLY tags of the form `vX.Y.Z` (pkgversion.ParseTag). Tags like `core/v1.0.0` or `figma/v1.0.0` are invisible to it. The working tags today: relux-root-context `v1.0.0` (abaadf43) and relux-mcp `v1.0.0` (027f55b).
- For a range requirement the resolver picks the highest `vX.Y.Z` tag satisfying the range and REQUIRES the manifest `version` at that tag (inside `directory`) to equal the tag version. The repository tag namespace is shared by all six packages, so a new tag `v1.0.1` requires every package manifest to say `1.0.1`.

Do exactly:
1. `packages/relux-root-context-ivan/agent-context.json`: add `"directory": "packages/figma"` to `requires.mcp.figma` and `"directory": "packages/safari"` to `requires.mcp.safari`; keep `git` and `range` unchanged; keep key order tidy.
2. Bump `"version"` to `1.0.1` in all six `packages/*/agent-context.json`. Update the six package READMEs (`Version \`1.0.0\`` → `1.0.1`) and the umbrella README line about requires.mcp (mention the `directory` fields; the satisfying relux-mcp tag is `v1.0.0`, a repository-wide tag).
3. Root README tag convention (lines ~22-26): the current text (`<short>/vX.Y.Z`, e.g. `core/v1.0.0`) contradicts curator-spec Decision 0012 §2 ("Version tags are strict SemVer 2.0 with a mandatory `v` prefix: `v<major>.<minor>.<patch>`", unique per repository) and the implementation. Rewrite it: version tags are repository-wide `vX.Y.Z`; all six packages are released together at one version; a context or MCP requirement addresses a package inside the repository with `directory` (Decision 0012 §1 requirement keys). Keep "never retag or move a published tag". Do not edit curator-spec.
4. If `tests/test_validate.py` or `tests/mutants.py` pin `1.0.0`, update them; run `bash scripts/validate.sh` and `python3 -m pytest tests -q` (or `python3 tests/test_validate.py` if that is how they run) — attach outputs with exit codes as `TASK-260916-3na4vf_evidence.md`.
5. Tick the checklist and `task-board handoff TASK-260916-3na4vf --role developer`. No manual commits on the Story branch; no tags; no pushes.
