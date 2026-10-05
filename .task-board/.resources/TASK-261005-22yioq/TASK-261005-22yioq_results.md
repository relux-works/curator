# TASK-261005-22yioq — Publish CIP drafts 0002-0006 (results)

Worktree: curator-spec branch task-board/story/STORY-261005-2z4y92 at b0caf8d.
Source: relux-works/curator main @ 3d9aa987 (verified == GitHub remote main via ls-remote; blob URLs for draft + evidence return 200).

## Files

- cips/CIP-0002-project-context-in-managed-launches.md (from .research/261004_CIP-0002-project-context-in-managed-launches.md)
- cips/CIP-0003-claude-managed-home-credential-modes.md (from .research/261004_CIP-0003-claude-managed-home-credential-modes.md)
- cips/CIP-0004-shell-hook-without-sourcing-and-path-append.md (from .research/261004_CIP-0004-shell-hook-no-source-path-append.md)
- cips/CIP-0005-audit-backends-and-cli-secret-transport.md (from .research/261004_CIP-0005-audit-backends-and-cli-secret-transport.md)
- cips/CIP-0006-legacy-provider-settings-and-mcp-opt-outs.md (from .research/261004_CIP-0006-legacy-provider-settings-and-mcp-optouts.md)
- cips/README.md index: 5 reservation rows converted to linked Draft rows.

## Transformations (docs only)

- Status normalized to exactly `Draft`; each file states the operator has made no acceptance decision.
- H1 titles aligned to the reserved index titles (0004/0006 reworded to match; substance unchanged).
- Each drafts single relative companion-evidence link rewritten to a backticked relux-works/curator .research/ path + main @ 3d9aa987 citation (repo convention per docs/security-audit-2026-09.md); evidence cited, not copied. 0004s extra Evidence metadata line folded into Current state so all files carry exactly the template metadata + section list.
- No normative protocol/schema, CHANGELOG, LOGBOOK, or decision-file edits. Personal-data grep over cips/ clean.

## Gates (real exit codes)

- python tools/validate.py (venv, jsonschema 4.25.1): exit 0 — 73 schemas, 1294 vectors, local links incl. new CIPs.
- python -B -m unittest discover -s tools -p test_*.py: exit 0 — 672/672 OK (clean undisturbed rerun).
- python tools/verify_skillfile_sources_independence.py: exit 0.
- git diff --check: exit 0; trailing-whitespace grep over 6 touched files: none; gofmt -l tools: clean.
- go test ./tools/...: exit 1 — ONE pre-existing failure (TestRC14ReleaseMetadataPinsCandidateAndPreservesSourceBaseline, main_test.go:920), reproduced identically with cips/ changes stashed. Cause: the test run regenerates conformance/v1/vectors/environments-read-failure.json (+ manifest/candidate pins); unrelated to this docs change. Side-effect edits reverted; tree contains only the intended cips/ change set.
- lychee: not installed, not run (CI-only). Substitute: validate.py local-link gate green; all 6 external hosts in new files curl-checked (200; rfc-editor 302 redirect); github issue #105 URL 200.

## Findings

- Pre-existing: go test regenerates conformance files and fails its own rc.14 pin check on a clean tree. No action taken (out of docs scope); recorded here instead of repo LOGBOOK per task instruction not to touch LOGBOOK.
- First full unittest run showed 2 transient failures while a parallel step of this same session mutated conformance/; clean rerun is 672/672 green. Counts as environmental confound, not a defect.
