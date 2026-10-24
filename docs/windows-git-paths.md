# Windows Git metadata paths

New Git clones on Windows require a projected metadata path of at most **259
UTF-16 code units**, excluding the terminating NUL. Curator computes this before
creating the destination or its cache parents. It expands existing short-name
ancestors and resolves the absolute path, so a relative operand, an 8.3 alias or
an extended path prefix does not increase the budget. For junctions, the longer
of the expanded visible path and its resolved target is charged.

Curator reserves the longest fixed metadata layout for both SHA-1 and SHA-256:
loose objects, pack keep/promisor/bitmap files, commit-graph layers and
incremental multi-pack-index bitmaps and chain locks. The longest reserved tail
is 126 units including the separator. A normal clone root therefore has a
**133-unit maximum**, including its drive or UNC prefix. The bound applies to
the expanded clone root, not just the manager home or its displayed spelling.

For example, a project MCP cache adds `/project-mcp-repos/` and a 64-character
source-identity digest to the manager home. Its expanded manager-home path must
fit within 50 units. Other cache namespaces use their actual destination path
and receive the same check. Use a short manager home rather than relying on
`core.longpaths` to make an over-budget metadata root usable.

The typed `git_metadata_path_too_long` refusal reports the projected metadata
length, the measured metadata bound, the expanded root length and its root
bound. It creates no destination, parent directory or half-cloned cache. The
manager preserves this refusal through `profile_source_invalid`; unrelated
existing home files remain unchanged. Existing caches are not relocated or
removed by this check. Other Git errors still use the existing clone cleanup.

The measurement uses real Git clone, fetch and exact object reads with a
disabled ambient long-path setting and enabled process/local settings. On Git
for Windows 2.56.0.windows.2, a SHA-1 loose-object root of 204 units succeeds and
205 fails; real packed file transfer succeeds at 190 and fails at 191. Both
transitions put the last usable metadata pathname at 259 units. The hosted
`TestWindowsGitMetadataLimit` repeats these adjacent controls instead of assuming
that the nominal Win32 MAX_PATH value is usable. The admission policy reserves
the larger SHA-256 and maintenance layouts before the remote format is known.

This is a fixed metadata-layout budget, not a guarantee for arbitrary remote
ref names, custom templates or third-party Git extensions. Those inputs can
still cause Git to refuse an operation. Unix path behavior is unchanged.

For Windows fixture runs, use a short `TMP` or the hosted runner's `RUNNER_TEMP`.
`testgitenv.ShortTempDir` creates private, task-lifetime fixture homes there
without embedding long test names. Explicit over-budget fixtures still exercise
the refusal and its no-write invariant. Run:

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -v ./internal/gitops ./internal/testgitenv
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -v ./internal/envprofile -run 'Test(WindowsMCPMetadataRootRefusal|MCPWindowsNarrowingControls)$'
```

Keep logs and bounded overlay controls in task-scoped `.temp/`; hosted Test
artifacts retain the complete Go JSON stream and platform gate output.

A short private Git metadata layout could remove this root-length limitation.
It remains a follow-up requiring a storage/identity/cleanup design; this change
does not redirect caches, add path aliases or relax privacy and no-follow gates.
