# TASK-260916-bn5kvb — review evidence

Changed only the umbrella manifest/README, scripts/validate.sh, tests/test_validate.py and tests/mutants.py. Both MCP requirements use git@github.com:relux-works/relux-mcp.git and ^1.0. The offline gate requires exactly figma and safari with the specified source/range.

## Validation run by this producer

Shell: zsh; shell validator executed explicitly with bash. Commands were standalone, without tee or pipelines.

- `bash scripts/validate.sh`: exit 0; 14 source digests checked.
- `python3 -m unittest discover -s tests -v`: exit 0; 14 tests. Includes missing family, missing/extra member, wrong Git source and wrong range for both MCP members through the shipped shell entry point.
- `bash -n scripts/validate.sh`: exit 0.
- `git diff --check`: exit 0.
- Python byte comparison against HEAD 66d86a5287cc82b8aa6a48b3d13fd71ddbf65bd7: exit 0; 24/24 tracked files outside the umbrella across the five other packages unchanged.
- Parser oracle: `go run . ../..` from `.temp/TASK-260916-bn5kvb-oracle`: exit 0; LoadManifest and ValidateModules passed for 6/6 packages, with all four registered environments and zero warnings. The Go command compiled and executed copied production contextpkg, identifiers, pkgversion and protocoljson source from Curator checkout f750344f5e6fe3ef539485981c12e24240cca1ed. No control-root writes. Oracle source attached separately for reproduction; copy these four internal directories and use module github.com/relux-works/curator, go 1.25.5.
- `git ls-remote --tags git@github.com:relux-works/relux-mcp.git`: exit 0. figma/v1.0.0 tag object 41f912c93f17d49b6dd5fe65af0a90bfd1da6692; safari/v1.0.0 tag object 453f5fbc396524523d810a05014381671b07dfd6. Both peel to 027f55b7582591604c8dc963684879671c95dbcd. A temporary clone (exit 0) and git show (exit 0) confirm the tagged packages/figma and packages/safari agent-mcp.json declare the matching names and version 1.0.0, satisfying ^1.0.

- `python3 tests/mutants.py`: exit 0; 9/9 narrowing mutants killed. Each mutated behavioral suite actually exited 1 (expected failing gate evidence), naming the intended failing probe. New mutations admit a figma-only inventory or restrict source/range checking to figma; both were detected. Existing seven mutations also detected. This measures these nine mutations only, not exhaustive schema coverage.

## Bounds

The local validator is an offline repository expectation gate, not the full Curator schema or network resolver. Parser coverage is 6/6 local packages; remote evidence verifies published tags and package versions, not a full profile materialization or live MCP launch. No cross-platform claim. No previously attached test evidence was substituted for these runs. The configured landing suite is left to the handoff runtime, per campaign rules.

No important anomaly or architecture decision arose. LOGBOOK.md was not edited, as prohibited by campaign rules; this outcome records the task findings.
