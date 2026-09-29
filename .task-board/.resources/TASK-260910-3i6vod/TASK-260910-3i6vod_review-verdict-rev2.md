# TASK-260910-3i6vod review verdict — rev2 ACCEPTED (reviewer claude-opus-5-5)

Candidate: base 3f60f7f0, tree 4a1579d6, 2 paths (README.md +15, SECURITY.md new 9 lines). Worktree matches diff.

## Content vs AC / spec rc.13 (23435129)
- README.md:96-109 "Installed command security" and SECURITY.md:3-9 "Installed command execution" both state: shim-launched commands run as the invoking user with their OS privileges; portable assurance is not a user-identity change or full OS sandbox; script-worker-v1 (opt-in `execution_policy`) is the enforced script path, others declared-only; `execution.mode: verified` is the provider-backed, non-fallback enforcement path. Matches core.md §4.1.1 (line 215), §4.2.1 (756), assurance.md §1/§2/§5, profiles/manager.md §3.1 (declared-only vs enforced).
- No overclaiming: SECURITY.md:9 says this release ships no verified provider; README keeps the existing "ships no platform provider" line (README.md:113). SECURITY.md:7 says script-worker controls are not a kernel sandbox. Code: script-worker-v1 implemented (internal/scriptworker, internal/runtimestore/enforced.go).
- Links: all pinned to spec commit 23435129 (exists in curator-spec); heading anchors verified against the headings at that commit (GitHub slugs 411-…, 421-…, 1-closed-selection, 2-platform-neutral-provider-contract, 5-failure-rules). Relative link SECURITY.md#installed-command-execution resolves.

## Tests (zsh, set -o pipefail, real exit)
`go test ./cmd/curator -run 'TestDraftDocs|Doc|Readme|README' -count=1 -v` → ok, exit=0 (includes README-reading TestEveryCurrentnessCodeIsDocumented, TestDraftDocsPinExamples).

## Gate flake
Run 36467710280 macOS failure TestPathInstallCapturesDirtyUntrackedInsideGit (internal/envprofile) builds its own temp git repo; it does not read README.md/SECURITY.md. A docs-only diff cannot affect it (BUG-260928-uyak0e).

## Bounds
- Docs-only: no committed test pins the new sections, so no mutant is killable; stated bound (repo has no general docs link checker; anchors verified manually above).
- No CHANGELOG/LOGBOOK edits in the diff.
