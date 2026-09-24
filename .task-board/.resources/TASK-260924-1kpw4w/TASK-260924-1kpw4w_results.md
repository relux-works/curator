# TASK-260924-1kpw4w results

## Implementation

- Added draft skill-manifest schema 9 support for `dependencies.skills.*.directory`, with the shared Skillfile directory grammar, root normalization to `.`, selected-folder containment and link checks, required `SKILL.md`, and dependency-key/frontmatter-name matching.
- Carried normalized directories through closure identity/unification, shared repository snapshots, package lock identity, frozen install input, source-audit records, and marker-v5 schema 9 validation.
- Added parser/conformance tests, resolution-vector tests, same-directory diamond and different-directory conflict tests, schema-8 lock-byte compatibility, a compiled-CLI install/refresh/reinstall audit test, and a CHANGELOG entry.

## Accepted amendment fixtures

Source: accepted amendment `TASK-260924-2am4qa_change-request_rev2.patch`, SHA-256 `d2b62645e78aade7809c3c0fe5a3a35b89da3a4530f966c7498e32d089f66f77`; candidate tree `b3663bb3`.

Copied from `conformance/draft-sources-v1/` into `internal/skillspec/testdata/draft-sources-v1/`. A recursive comparison against an archive of candidate tree `b3663bb3` exited 0. Vector hashes:

- `manifest-dependency-directories.json`: `fd72d52fd10963c23b1a93a79e847da2c106c2f58940773bfe4dcb8dd38d5a51`
- The following schema-case files have the same bytes and SHA-256 in both `schema-cases/agent-skill-v9/` and `schema-cases/csk-skill-v9/`:
  - `invalid-directory-absolute.json`: `35a7543cb9c6c3309daa3f61997500e6ff161e03b2041360e9e4cba454947b7e`
  - `invalid-directory-backslash.json`: `4bd973a4ff6a91fb08e5a9fb9d5e565bea0334c7c5c7b1276e9755221755658d`
  - `invalid-directory-empty.json`: `f4fa3de75ef27017713d63dc432da6ddf2976de279dcc2851056c18113efb5a4`
  - `invalid-directory-escape.json`: `aa6977b72c228e5e0f1175b8e7a268e2421a0cb0cbd1b67b460200e4332a0f5c`
  - `invalid-directory-glob.json`: `51e26dbf8003a5730819f34269a1278b6ef7b5cdb49aab57ae5695ab5fe7073c`
  - `invalid-directory-parent-component.json`: `35d4498e73d1d058263ecdc48ceae140bc42b860a1a898416244edb6eaa76361`
  - `valid-directory-absent.json`: `b046676baef4a667571ee9c60eb77781a76e7eedd3b5ccbe5d6a6e6dd56d87a5`
  - `valid-directory-root.json`: `f0b2a09a02f3e431e46737135b7d7282b56ce9dcd8f834c2ba963f4e7cf1567a`
  - `valid-directory-subfolder.json`: `96525b88d0f2d92b4ba0036ff9b60a94cfe14bc07af9caad096487d9466b4b16`
  - `invalid.json`: `0282ff1290b90081925f8a18be4a3c1dd86c51df883508e0f74e2285533f0ebe`
  - `valid.json`: `b08b365927d1e036a5e0254c8a8af5bea3758379f345a1b4e2963c12d9bf7784`

The runtime parser test consumes the nine directory-specific schema cases in each suite. The generic `valid.json`/`invalid.json` pair is copied byte-for-byte but is not used as a directory gate: that pair tests schema-version acceptance, and the invalid sample is schema 8, which remains supported.

## Verification

Green commands (real exit code 0):

- `go test ./internal/closure -count=1` — full closure package, 213.265s.
- `go test ./internal/skillspec -count=1` — 11.207s.
- `go test ./internal/skillspec ./internal/manifest ./internal/identifiers ./internal/marker -count=1`.
- `go test ./internal/crossconformance -run '^TestDraftSourcesCLIDirectoryManifestDependencyInstallRefreshAndAudit$|^TestDraftSourcesCLILocalSkillScriptDependencies$' -count=1` — 81.825s; real CLI resolution, lock, install, refresh, reinstall, marker and audit cover the directory diamond.
- `go build -o "$TMPDIR/TASK-260924-1kpw4w_curator" ./cmd/curator`.
- `go vet ./internal/skillspec ./internal/closure ./internal/manifest ./internal/marker ./internal/install ./internal/crossconformance`.
- `golangci-lint run` — 0 issues.
- `git diff --check`.

One broader optional run did not pass: `go test ./internal/install -run '^TestDraft' -count=1` exited 1 after the package timeout (601.163s), panicking in existing `TestDraftBuildsPublishReceipt3OnBothArms` while `transaction.Engine.saveJournal` was blocked in `os.CreateTemp`. The focused compiled-CLI install/refresh/audit path above passed. The configured landing suite was left for handoff and was not run manually.

Mutant attacks were run in a disposable copy; each mutated test exited 1 for its expected assertion:

- Allowing glob metacharacters in `ValidDirectory`: `go test ./internal/skillspec -run '^TestDraftManifestDependencyDirectoryGrammar$' -count=1` rejected the mutant at the `glob` case (`valid=false, error=<nil>`).
- Bypassing directory mismatch unification: `go test ./internal/closure -run '^TestDraftManifestDependencyDirectorySameNameDifferentFolderConflicts$' -count=1` observed no `source_name_conflict`.
- Narrowing the selected `SKILL.md` name check: `go test ./internal/closure -run '^TestDraftManifestDependencyDirectoryRequiresMatchingSkillName$' -count=1` observed no refusal.
- Weakening the symlink refusal class: `go test ./internal/closure -run 'TestDraftManifestDependencyDirectoryResolutionVectors/symlinked-directory-escape' -count=1` observed `source_member_invalid` instead of `source_selection_invalid`.
- Admitting a selected folder without `SKILL.md`: `go test ./internal/closure -run 'TestDraftManifestDependencyDirectoryResolutionVectors/folder-without-skill-md' -count=1` failed because the expected direct `no SKILL.md` diagnostic was lost; the later content-hash gate still refused the package.

For absent-directory compatibility, deterministic schema-8 dependency resolution produced lock SHA-256 `sha256:685572494b1612df0e2f8be58102a9bb01037d404f513c0e756da5103fd5d66c` both from the unmodified Story base and this candidate. The golden assertion is in `TestLegacyDependencyWithoutDirectoryLockBytes`.

## Board state

The initial requested transition to `development` was refused because prerequisite `TASK-260924-2am4qa` is still `integrating`; the task remains `backlog`. Its accepted revision-2 patch and candidate tree were used as instructed. No repository commit was made.

## Handoff gate

`task-board handoff TASK-260924-1kpw4w --role developer` exited 1. The board refused the `to-review` transition because prerequisite `TASK-260924-2am4qa` remains `integrating` and dependencies must be done first. The implementation and evidence are prepared; handoff is waiting for that prerequisite's integration state to reach `done`.

## CHANGELOG entry (for release prep)

Per CHANGELOG POLICY (2026-09-24) the hunk below was reverted to the base bytes in the worktree (`git diff CHANGELOG.md` is empty after the revert) and is recorded here verbatim for release prep:

```
- Draft schema-9 skill-manifest dependencies may select a skill package below a
  pinned repository with `directory`. Directory grammar, containment,
  `SKILL.md` name checks, closure unification, lock identity, and source-audit
  records all bind the selected folder; omitted `directory` remains the
  repository root.
```

Note: the "and a CHANGELOG entry" phrase in the Implementation section above refers to this same entry, now held here instead of in the tree per policy.

## Handoff

Bound developer run finishing the refused handoff after the orchestrator removed the `blocked_by TASK-260924-2am4qa` link:

- `set_estimate(TASK-260924-1kpw4w, estimate=estimated, scale=fibonacci, number=8)` exited 0 (the board requires an estimate before `development`; sibling implementation tasks in the epic carry fibonacci 5/8).
- `set_status(TASK-260924-1kpw4w, status=development)` exited 0; task, story `STORY-260924-1ckno7`, and epic are now `development`.
- CHANGELOG.md hunk reverted to base bytes; no other file changed. `git status --short` shows only the 18 modified product/test files plus two deliverable test assets (`internal/closure/directory_dependency_test.go`, `internal/skillspec/testdata/`); no stray files in the worktree.
- No repository commit was made; no LOGBOOK.md edit (per 2026-09-24 rule this resource is the logbook citation).
- The "one broader optional run did not pass" note in Verification stands as recorded: `go test ./internal/install -run '^TestDraft'` exited 1 on a pre-existing package-timeout panic, unrelated to this change; no test or build step was re-run in this bound run beyond `git diff`/`git status` inspection, and all green commands cited under Verification were run by the prior implementation run with real exit code 0.

DoD (all 7 board items already checked, cited here):

1. Directory-dependency behavior row — Verification, mutant attacks, absent-directory lock bytes above.
2. Code per description and AC — Implementation, Accepted amendment fixtures.
3. Tests written and passing — Verification.
4. Lint clean — `golangci-lint run` 0 issues, `go vet`, `git diff --check` under Verification.
5. Build/validation run, build not broken — `go build -o "$TMPDIR/..." ./cmd/curator` under Verification.
6. Task-scoped outcome artifact attached — this updated `TASK-260924-1kpw4w_results.md` resource.
7. Logbook when relevant — this resource is the record; LOGBOOK.md untouched per policy.
