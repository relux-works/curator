# TASK-260916-3na4vf — revision 1 review
VERDICT: CHANGES_REQUESTED

Candidate: 3ab7b2174e80ec6f80e6b109f7aeee46ce3f7b27
Base/HEAD: abaadf43772341d0196e72a4ca9914017dc8f512
Tracked worktree matches candidate (git diff --exit-code candidate: 0). No reviewer code edits or commits.

## Required correction
README.md:22-28 omits the explicit requirement in 3na4vf-brief.md step 3: explain that a context or MCP requirement selects a package within the repository using directory. The repository-wide version explanation is correct, but the root README never mentions directory. Add a sentence such as: "A context or MCP requirement selects a package within the repository using its directory field." This is the sole requested change; publish a new revision for review. Decision 0012 section 1, lines 179-184 supports this distinction.

## Checks independently performed
- Exact diff inspected: six manifests, seven READMEs, validator, two test files; module bytes unchanged.
- All 6/6 manifest versions equal 1.0.1. All package README versions updated.
- Both 2/2 MCP directories correct; git and range unchanged.
- scripts/validate.sh change is necessary: former exact dictionary equality rejected any directory. Updated equality admits only the specified package directory and retains git/range constraints.
- Root README correctly describes repository-wide releases and preserves never-retag guidance.
- git diff --check: exit 0.
- bash scripts/validate.sh: exit 0.
  validate: manifests, modules, weights, ranges: OK
  sources: 14 module digests OK
  validate: module bytes: OK
  validate: PASS
- python3 -m unittest discover -s tests -v: exit 0; 15/15 tests OK (12.626s explicit-exit replay).
  Passing tests: duplicate_row, extra_mcp, mcp_directory, mcp_inventory, mcp_range, mcp_source, missing_mcp, missing_row, missing_sources, source_drift, sources_drift, trailing_lf, unknown_env, valid, weight_drift.
- python3 tests/mutants.py: exit 0; 10/10 narrowing mutants killed.
  mcp-inventory -> test_mcp_inventory
  mcp-source -> test_mcp_source
  mcp-directory -> test_mcp_directory
  trailing-lf -> test_trailing_lf
  unknown-env -> test_unknown_env
  weight -> test_weight_drift
  digest -> test_source_drift
  sources-digest -> test_sources_drift
  inventory -> test_missing_row
  duplicate -> test_duplicate_row
  Each behavioral suite exited 1 with the expected named FAIL; harness ended "all narrowing mutants killed".
Commands launched from zsh; validation entry point is bash scripts/validate.sh. Validator and unit tests were replayed once because the initial combined shell invocation did not expose individual process exit codes. Mutation harness ran once to completion. No producer test evidence substituted for these results.

## Resolver compatibility and bounds
Read local Curator internal/envprofile/gitsource.go Candidates/Manifest/packageOf and internal/contextresolve/contextresolve.go range selection and version comparison. Candidates parses version tags with pkgversion.ParseTag; ranges select highest satisfying tag; manifests read beneath directory and must match selected tag version.
Given publication of this candidate at v1.0.1, and no higher satisfying tag or conflicting constraints, umbrella plus five contexts resolve to v1.0.1 with matching manifests.
Read-only git ls-tree at relux-mcp v1.0.0 confirms packages/figma/agent-mcp.json and packages/safari/agent-mcp.json; git show confirms both versions 1.0.0. MCP requirements therefore resolve with correct directories at that tag under the stated tag set.
No remaining declaration defect identified. Actual profile install was not run. Publication of v1.0.1 is still required; remote tag freshness, skill dependency availability, transport/authentication, machine allowlists and runtime provisioning remain unverified. Local layout inspection does not prove live installation.
Coverage is bounded to the local validation entry point and ten specified narrowing mutants; it is not an end-to-end Curator install test or exhaustive gate mutation coverage.

## Lifecycle
CR revision 1 was published and HEAD remains the recorded base. spawn goal reports no active goal (run not goal-bound); no directives recorded. Route to to-dev for the documentation correction. No LOGBOOK.md edit: campaign rules prohibit those writes; finding recorded in this outcome and board notes.
